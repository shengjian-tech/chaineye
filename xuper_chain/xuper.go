package xuper_chain

import (
	"context"
	"encoding/hex"
	"strconv"
	"sync"
	"time"

	"gitee.com/chunanyong/zorm"

	"github.com/ccfos/nightingale/v6/models"
	"github.com/toolkits/pkg/logger"
	"github.com/xuperchain/xuper-sdk-go/v2/xuper"
	"github.com/xuperchain/xuperchain/service/pb"
)

func SyncXuperBlockTimer(xuperSdkYmlPath string) {
	duration := time.Duration(3) * time.Second
	for {
		time.Sleep(duration)
		SyncXuperBlock(xuperSdkYmlPath)
	}
}

var syncMap sync.Map

// 三秒同步一次 最新区块，以及交易数量
func SyncXuperBlock(xuperSdkYmlPath string) {
	defer func() {
		if err := recover(); err != nil {
			logger.Error(err)
		}
	}()
	// 查询 xuper_node 表获取所要拉取的链上交易信息 循环拉取
	ctx := context.TODO()
	xuperNodes, err := GetXuperNode(ctx)
	if err != nil || len(xuperNodes) <= 0 {
		logger.Error("get xuper node info failed")
		return
	}
	for _, node := range xuperNodes {
		// 生成对应链节点的client
		xuperClient, err := xuper.New(node.Node, xuper.WithConfigFile(xuperSdkYmlPath))
		if err != nil {
			logger.Errorf("sync xuper chain block failed, rootnet: %s, node: %s, err: %s", node.RootNet, node.Node, err.Error())
			return
		}
		defer xuperClient.Close()
		// 查询网络状态 看有多少条链
		b, err := xuperClient.QuerySystemStatus()
		if err != nil {
			logger.Errorf("get chain system status failed, rootnet: %s, node: %s, err: %s", node.RootNet, node.Node, err.Error())
			return
		}
		if len(b.SystemsStatus.BcsStatus) <= 0 {
			return
		}
		for _, bcsStatus := range b.SystemsStatus.BcsStatus {
			value, ok := syncMap.Load(node.RootNet + bcsStatus.Bcname)
			if !ok {
				logger.Errorf("once sync rootnet: %s, bcName: %s", node.RootNet, bcsStatus.Bcname)
				syncMap.Store(node.RootNet+bcsStatus.Bcname, true)
				return
			}
			if value == true {
				go SyncXuperData(xuperSdkYmlPath, bcsStatus, ctx, node)
			}
		}
	}
}

func SyncXuperData(xuperSdkYmlPath string, chain *pb.BCStatus, ctx context.Context, node models.XuperNode) {
	defer func() {
		if err := recover(); err != nil {
			logger.Error(err)
		}
	}()
	// 加锁
	syncMap.Store(node.RootNet+chain.Bcname, false)
	// 生成对应链节点的client
	xuperClient, err := xuper.New(node.Node, xuper.WithConfigFile(xuperSdkYmlPath))
	if err != nil {
		logger.Errorf("sync xuper chain block failed, rootnet: %s, node: %s, err: %s", node.RootNet, node.Node, err.Error())
		return
	}
	defer xuperClient.Close()
	// 查询数据库 该网络该链下的最新区块高度
	heightInDB, err := GetHeightInDBByChain(ctx, node.RootNet, chain.Bcname)
	if err != nil {
		logger.Errorf("get height max in db failed, rootnet: %s, chainName: %s, err:%s",
			node.RootNet,
			chain.Bcname,
			err.Error())
		return
	}
	var dataToDB models.XuperStruct
	// 第一次同步 数据库没有该条链的区块数据 所以先插入一条
	if heightInDB <= 0 {
		// 解析区块
		b2, err := xuperClient.QueryBlockByHeight(1, xuper.WithQueryBcname(chain.Bcname))
		if err != nil {
			logger.Errorf("query block failed, rootnet: %s, chainName: %s, height: %d ,err:%s",
				node.RootNet,
				chain.Bcname,
				1,
				err.Error())
			return
		}
		dataToDB.BlockHeight = 1
		dataToDB.BlockHash = hex.EncodeToString(b2.Blockid)
		// 截取十位时间戳  到秒级
		s := strconv.FormatInt(b2.Block.Timestamp, 10)[0:10]
		timeStamp, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			logger.Error("timestamp string to int64 failed", err.Error())
			return
		}
		dataToDB.Timestamp = timeStamp
		dataToDB.BlockTxCount = int64(b2.Block.TxCount)
		dataToDB.TotalTxCount = int64(b2.Block.TxCount)
		dataToDB.RootNet = node.RootNet
		dataToDB.ChainName = chain.Bcname
		dataToDB.Id = strconv.FormatInt(dataToDB.BlockHeight, 10) + node.RootNet + chain.Bcname

		// 同时插入 t_xuper_day_tx 记录
		var dayTx models.XuperDayTx
		dayTx.Day = time.Unix(timeStamp, 0).Format("2006-01-02")
		dayTx.Id = dayTx.Day + node.RootNet + chain.Bcname
		dayTx.DayTxCount = int64(b2.Block.TxCount)
		dayTx.TotalTxCount = int64(b2.Block.TxCount)
		dayTx.ChainName = chain.Bcname
		dayTx.RootNet = node.RootNet
		// 写入数据库
		_, err = zorm.Transaction(ctx, func(ctx context.Context) (interface{}, error) {
			_, err := zorm.Insert(ctx, &dataToDB)
			if err != nil {
				return nil, err
			}
			_, err = zorm.Insert(ctx, &dayTx)
			return nil, err
		})
		if err != nil {
			return
		}

	} else {
		txCounts, err := GetTxTotal(ctx, node.RootNet, chain.Bcname)
		if err != nil {
			logger.Errorf("get height max tsx in db failed, rootnet: %s, chainName: %s, err:%s",
				node.RootNet,
				chain.Bcname,
				err.Error())
			return
		}
		// 如果查询到  则不是第一次同步 查询最新区块和数据库记录区块的区间。循环同步
		var newHeight = chain.Block.Height
		for i := heightInDB + 1; i <= newHeight; i++ {
			// 解析区块
			b2, err := xuperClient.QueryBlockByHeight(i, xuper.WithQueryBcname(chain.Bcname))
			if err != nil {
				// TODO 可注释
				// logger.Errorf("query block failed, rootnet: %s, chainName: %s, height: %d ,err:%s",
				// 	node.RootNet,
				// 	chain.Bcname,
				// 	i,
				// 	err.Error())
				continue
			}
			var dataToDB models.XuperStruct
			dataToDB.BlockHeight = i
			dataToDB.BlockHash = hex.EncodeToString(b2.Blockid)
			// 截取十位时间戳  到秒级
			s := strconv.FormatInt(b2.Block.Timestamp, 10)[0:10]
			timeStamp, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				logger.Error("timestamp string to int64 failed", err.Error())
				continue
			}
			dataToDB.Timestamp = timeStamp
			dataToDB.BlockTxCount = int64(b2.Block.TxCount)
			txCounts = txCounts + int64(b2.Block.TxCount)
			dataToDB.TotalTxCount = txCounts
			dataToDB.RootNet = node.RootNet
			dataToDB.ChainName = chain.Bcname
			dataToDB.Id = strconv.FormatInt(dataToDB.BlockHeight, 10) + node.RootNet + chain.Bcname

			// 查询数据库 t_xuper_day_tx 当天是否有记录。有直接在记录基础上累加。没有则证明是新的一天，插入一条新数据
			dayNow := time.Unix(timeStamp, 0).Format("2006-01-02")
			var dayTx models.XuperDayTx

			dayTx.RootNet = node.RootNet
			dayTx.ChainName = chain.Bcname
			dayTx.Id = dayNow + node.RootNet + chain.Bcname
			dayTx.Day = dayNow

			xuperDayTxs, err := GetXuperDayTx(ctx, dayTx.Id)
			if err != nil {
				logger.Errorf("query t_xuper_day_tx failed, rootnet: %s, chainName: %s, day: %d ,err:%s",
					node.RootNet,
					chain.Bcname,
					dayNow,
					err.Error())
			}
			// 最近一次的记录
			xuperDayTxsYestoday, err := GetLastDayTx(ctx, node.RootNet, chain.Bcname)
			if err != nil {
				logger.Errorf("query t_xuper_day_tx failed, rootnet: %s, chainName: %s, day: %d ,err:%s",
					node.RootNet,
					chain.Bcname,
					time.Unix(timeStamp, 0).Format("2006-01-02"),
					err.Error())
			}
			// 证明已经有当天数据 在当天原有的基础上累加
			if len(xuperDayTxs) > 0 {
				if len(xuperDayTxsYestoday) == 1 && xuperDayTxs[0].Day == xuperDayTxsYestoday[0].Day {
					dayTx.TotalTxCount = xuperDayTxs[0].TotalTxCount + int64(b2.Block.TxCount)
					dayTx.DayTxCount = dayTx.TotalTxCount
				} else if len(xuperDayTxsYestoday) == 2 {
					dayTx.TotalTxCount = xuperDayTxs[0].TotalTxCount + int64(b2.Block.TxCount)
					dayTx.DayTxCount = dayTx.TotalTxCount - xuperDayTxsYestoday[1].TotalTxCount
				}
			} else {
				// 证明是新的一天
				if len(xuperDayTxsYestoday) > 0 {
					dayTx.TotalTxCount = xuperDayTxsYestoday[0].TotalTxCount + int64(b2.Block.TxCount)
					dayTx.DayTxCount = int64(b2.Block.TxCount)
				}
			}

			// 写入数据库
			_, err = zorm.Transaction(ctx, func(ctx context.Context) (interface{}, error) {
				_, err := zorm.Insert(ctx, &dataToDB)
				if err != nil {
					return nil, err
				}
				if len(xuperDayTxs) <= 0 {
					_, err = zorm.Insert(ctx, &dayTx)
				} else {
					_, err = zorm.Update(ctx, &dayTx)
				}
				return nil, err
			})
			if err != nil {
				continue
			}
		}
	}
	// 解锁
	syncMap.Store(node.RootNet+chain.Bcname, true)
}

// 定期清除区块数据，交易数据，每天执行一次。删除10天之前的数据
func DeleteXuperDataTimer() {
	duration := time.Duration(24) * time.Hour
	for {
		time.Sleep(duration)
		DeleteXuperData()
	}
}

// 定时删除十天前的同步数据
func DeleteXuperData() {
	defer func() {
		if err := recover(); err != nil {
			logger.Error(err)
		}
	}()
	// 获取数据库最新一条区块高度
	ctx := context.TODO()
	oldTime := time.Now().Add(-time.Hour * 24 * 10)
	oldDay := oldTime.Format("2006-01-02")
	// 删除 oldTime之前的数据
	_, err := zorm.Transaction(ctx, func(ctx context.Context) (interface{}, error) {
		finder := zorm.NewDeleteFinder(models.XuperStructTableName)
		finder.Append("where timestamp <= ?", oldTime.Unix())
		_, err := zorm.UpdateFinder(ctx, finder)
		if err != nil {
			return nil, err
		}
		finderTx := zorm.NewDeleteFinder(models.XuperDayTxStructTableName)
		finderTx.Append("where block_day <= ?", oldDay)
		_, err = zorm.UpdateFinder(ctx, finderTx)
		return nil, err
	})
	if err != nil {
		logger.Error("delete old block info in db XuperStruct failed", err.Error())
		return
	}
}

func GetHeightInDBByChain(ctx context.Context, rootNet string, chainName string) (int64, error) {
	var height int64
	// 查询数据库最新区块高度
	f := zorm.NewSelectFinder(models.XuperStructTableName, " Max(block_height) ")
	f.Append("WHERE rootnet = ? and chain_name = ? ", rootNet, chainName)
	_, err := zorm.QueryRow(ctx, f, &height)
	if err != nil {
		logger.Error("get txs in db by height failed", err.Error())
		return height, err
	}
	return height, nil
}
