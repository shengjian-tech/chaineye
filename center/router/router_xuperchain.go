package router

import (
	"encoding/hex"
	"math/big"
	"net/http"
	"strconv"
	"time"

	"github.com/ccfos/nightingale/v6/models"
	"github.com/ccfos/nightingale/v6/xuper_chain"
	"github.com/gin-gonic/gin"
	"github.com/toolkits/pkg/ginx"

	"github.com/xuperchain/xuper-sdk-go/v2/xuper"
	"github.com/xuperchain/xuperchain/service/pb"
)

// query contract count
func (rt *Router) getXuperChainContractCount(c *gin.Context) {
	chainName := ginx.QueryStr(c, "chain_name", "xuper")
	rootNet := ginx.QueryStr(c, "rootnet", "opennet")

	xuperNode, err := xuper_chain.GetXuperNodeByRootName(c, rootNet)
	if err != nil || len(xuperNode) <= 0 {
		ginx.Bomb(http.StatusInternalServerError, "query failed")
	}
	client, err := xuper.New(xuperNode[0].Node, xuper.WithConfigFile(rt.HTTP.XuperSdkYmlPath))
	if err != nil {
		ginx.Bomb(http.StatusInternalServerError, "query failed")
	}
	defer client.Close()
	// query contract counts
	csdr, err := client.QueryContractCount(xuper.WithQueryBcname(chainName))
	if err != nil {
		ginx.Bomb(http.StatusInternalServerError, "query failed")
	}
	// query block height 过滤条件不生效
	b, err := client.QuerySystemStatus(xuper.WithQueryBcname(chainName))
	if err != nil {
		ginx.Bomb(http.StatusInternalServerError, "query failed")
	}
	// query mint present address
	status, err := client.QueryBlockChainStatus(xuper.WithQueryBcname(chainName))
	if err != nil {
		ginx.Bomb(http.StatusInternalServerError, "query failed")
	}
	// query peer counts
	var peerCounts int

	// query tx counts
	txTotal, err := xuper_chain.GetTxTotal(c, rootNet, chainName)
	if err != nil {
		ginx.Bomb(http.StatusInternalServerError, "query failed")
	}
	if len(b.SystemsStatus.PeerUrls) == 0 {
		peerCounts = 1
	} else {
		peerCounts = len(b.SystemsStatus.PeerUrls)
	}

	ginx.NewRender(c).Data(
		gin.H{
			"count":     csdr.Data.ContractCount, // 返回合约数量
			"height":    status.Block.Height,     // 返回当前区块高度
			"proposer":  status.Block.Proposer,   // 返回当前打包区块的矿工地址
			"node_sum":  peerCounts,              // 返回当前链的节点数量 如果是
			"tx_counts": txTotal,                 // 返回当前链的交易总数
		}, nil)
}

// query block or tx
func (rt *Router) getXuperChainTx(c *gin.Context) {
	chainName := ginx.QueryStr(c, "chain_name", "xuper")
	rootNet := ginx.QueryStr(c, "rootnet", "opennet")

	input := ginx.QueryStr(c, "input", "")

	xuperNode, err := xuper_chain.GetXuperNodeByRootName(c, rootNet)
	if err != nil || len(xuperNode) <= 0 {
		ginx.Bomb(http.StatusInternalServerError, "query failed")
	}
	client, err := xuper.New(xuperNode[0].Node, xuper.WithConfigFile(rt.HTTP.XuperSdkYmlPath))
	if err != nil {
		ginx.Bomb(http.StatusInternalServerError, "query failed")
	}
	defer client.Close()

	// 全数字按照区块高度查询 转换异常 按照交易hash 或者 区块hash查询
	i, err := strconv.ParseUint(input, 10, 64)
	// err is nil 证明传入的是数字，按照区块高度查询
	if err == nil {
		b, err := client.QueryBlockByHeight(int64(i), xuper.WithQueryBcname(chainName))
		if err != nil {
			ginx.Bomb(http.StatusInternalServerError, "query failed")
		}
		br, err := blockToBlockRsp(b)
		if err != nil {
			ginx.Bomb(http.StatusInternalServerError, "query failed")
		}
		ginx.NewRender(c).Data(gin.H{
			"block": br,
		}, nil)
	} else {
		// err is not nil。 先按照交易hash 查询
		tx, err := client.QueryTxByID(input, xuper.WithQueryBcname(chainName))
		if err == nil {
			// query Block
			block, err := client.QueryBlockByID(hex.EncodeToString(tx.Blockid), xuper.WithQueryBcname(chainName))
			if err != nil {
				ginx.Bomb(http.StatusInternalServerError, "query failed")
			}
			tdr, err := TxToTxRsp(tx, block)
			if err != nil {
				ginx.Bomb(http.StatusInternalServerError, "query failed")
			}
			ginx.NewRender(c).Data(gin.H{
				"transaction": tdr,
			}, nil)
		} else {
			// 按照区块 hash 查询
			b, err := client.QueryBlockByID(input, xuper.WithQueryBcname(chainName))
			if err != nil {
				ginx.Bomb(http.StatusInternalServerError, "query failed")
			}
			br, err := blockToBlockRsp(b)
			if err != nil {
				ginx.Bomb(http.StatusInternalServerError, "query failed")
			}
			ginx.NewRender(c).Data(gin.H{
				"block": br,
			}, nil)
		}
	}
}

// 查询十天 每天的交易总数 返回前端生成折线图
func (rt *Router) getXuperChainTxLineChart(c *gin.Context) {
	chainName := ginx.QueryStr(c, "chain_name", "xuper")
	rootNet := ginx.QueryStr(c, "rootnet", "opennet")

	var dataRsp []string
	var txCountsRsp []int64

	// 查询时间戳 并转成日期 取当前日期后一天
	txs, err := xuper_chain.GetXuperTenTxs(c, rootNet, chainName)
	if err != nil {
		ginx.Bomb(http.StatusInternalServerError, "query failed")
	}
	for _, v := range txs {
		dataRsp = append(dataRsp, v.Day)
		txCountsRsp = append(txCountsRsp, v.DayTxCount)
	}
	ginx.NewRender(c).Data(gin.H{
		"data":   dataRsp,
		"counts": txCountsRsp,
	}, nil)
}

// 查询最新的十个区块  和 十次交易
func (rt *Router) getTxHistory(c *gin.Context) {
	chainName := ginx.QueryStr(c, "chain_name", "xuper")
	rootNet := ginx.QueryStr(c, "rootnet", "opennet")

	xuperNode, err := xuper_chain.GetXuperNodeByRootName(c, rootNet)
	if err != nil || len(xuperNode) <= 0 {
		ginx.Bomb(http.StatusInternalServerError, "query failed")
	}
	client, err := xuper.New(xuperNode[0].Node, xuper.WithConfigFile(rt.HTTP.XuperSdkYmlPath))
	if err != nil {
		ginx.Bomb(http.StatusInternalServerError, "query failed")
	}
	defer client.Close()
	// 获取最新区块高度，依次减1  减十次
	// query block height
	b, err := client.QueryBlockChainStatus(xuper.WithQueryBcname(chainName))
	if err != nil {
		ginx.Bomb(http.StatusInternalServerError, "query failed")
	}
	var lastHeight = b.Block.Height

	var txIds []string
	var blockRsps []BlockRsp

	for i := 0; i < 10; i++ {
		block, err := client.QueryBlockByHeight(lastHeight-int64(i), xuper.WithQueryBcname(chainName))
		if err != nil {
			ginx.Bomb(http.StatusInternalServerError, "query failed")
		}
		br, err := blockToBlockRsp(block)
		if err != nil {
			ginx.Bomb(http.StatusInternalServerError, "query failed")
		}
		blockRsps = append(blockRsps, br)
		txIds = append(txIds, br.Txs...)
		if len(txIds) == 10 {
			break
		}
	}
	var txs []TxDetailRsp
	// 填充最新十条交易信息返回
	for i := 0; i < len(txIds); i++ {
		tx, err := client.QueryTxByID(txIds[i], xuper.WithQueryBcname(chainName))
		if err != nil {
			ginx.Bomb(http.StatusInternalServerError, "query failed")
		}
		// query Block
		block, err := client.QueryBlockByID(hex.EncodeToString(tx.Blockid), xuper.WithQueryBcname(chainName))
		if err != nil {
			ginx.Bomb(http.StatusInternalServerError, "query failed")
		}
		tdr, err := TxToTxRsp(tx, block)
		if err != nil {
			ginx.Bomb(http.StatusInternalServerError, "query failed")
		}
		txs = append(txs, tdr)
	}

	ginx.NewRender(c).Data(gin.H{
		"blocks": blockRsps,
		"txs":    txs,
	}, nil)
}

/*
 * 获取超级链节点
 */
func (rt *Router) getXuperNodes(c *gin.Context) {
	xuperNodes, err := xuper_chain.GetXuperNode(c)
	if err != nil {
		ginx.Bomb(http.StatusInternalServerError, "获取超级链节点失败")
	}
	ginx.NewRender(c).Data(gin.H{
		"xuperNodes": xuperNodes,
	}, nil)
}

/*
同步 government 项目创建BaaS的网络名和节点数，从而定时同步数据,使用json接收
*/
func (rt *Router) syncXuperNode(c *gin.Context) {

	var xuperNode models.XuperNode
	c.BindJSON(&xuperNode)

	if xuperNode.RootNet == "" || xuperNode.Node == "" {
		ginx.Bomb(http.StatusInternalServerError, "BaaS网络名,节点IP:Port不能为空")
	}
	if xuperNode.RootNetName == "" {
		xuperNode.RootNetName = xuperNode.RootNet
	}

	err := xuper_chain.InsertXuperNode(c, &xuperNode)

	if err != nil {
		ginx.Bomb(http.StatusInternalServerError, "监控同步BaaS网络节点信息失败")
	}
	ginx.NewRender(c).Message("Sync xuper node success")
}

func TxToTxRsp(tx *pb.Transaction, block *pb.Block) (TxDetailRsp, error) {
	var rsp TxDetailRsp
	rsp.TxId = hex.EncodeToString(tx.Txid)
	rsp.BlockId = hex.EncodeToString(tx.Blockid)
	rsp.BlockHeight = block.Block.Height
	rsp.Coinbase = tx.Coinbase
	rsp.Miner = string(block.Block.Proposer)
	timestamp := block.Block.Timestamp / 1e9
	tm1 := time.Unix(timestamp, 0)
	rsp.BlockTimestamp = tm1.Local().Format("2006-01-02 15:04:05")
	rsp.Initiator = tx.Initiator
	for _, v := range tx.TxInputs {
		rsp.FromAddress = append(rsp.FromAddress, string(v.FromAddr))
	}
	for _, v := range tx.TxOutputs {
		rsp.ToAddresses = append(rsp.ToAddresses, string(v.ToAddr))
	}
	var amountInput = &big.Int{}
	var amountOuput = &big.Int{}
	for _, v := range tx.TxInputs {
		amountInput = amountInput.Add(FromAmountBytes(v.Amount), amountInput)
	}
	for _, v := range tx.TxOutputs {
		amountOuput = amountOuput.Add(FromAmountBytes(v.Amount), amountOuput)
	}
	rsp.FromTotal = amountInput
	rsp.ToTotal = amountOuput
	// 计算fee
	var gasUsed int64
	for _, v := range tx.ContractRequests {
		for _, v1 := range v.ResourceLimits {
			gasUsed += v1.Limit
		}
	}
	rsp.Fee = gasUsed
	tm := time.Unix(tx.Timestamp/1e9, 0)
	rsp.Date = tm.Local().Format("2006-01-02 15:04:05")
	for _, v := range tx.ContractRequests {
		rsp.Contracts = append(rsp.Contracts, v.ContractName)
	}
	return rsp, nil
}

func FromAmountBytes(buf []byte) *big.Int {
	n := big.Int{}
	n.SetBytes(buf)
	return &n
}

func blockToBlockRsp(block *pb.Block) (BlockRsp, error) {
	var rsp BlockRsp
	rsp.BlockHeight = block.Block.Height
	rsp.BlockId = hex.EncodeToString(block.Block.Blockid)
	rsp.Miner = string(block.Block.Proposer)
	rsp.NextHash = hex.EncodeToString(block.Block.NextHash)
	rsp.PreHash = hex.EncodeToString(block.Block.PreHash)
	timestamp := block.Block.Timestamp / 1e9
	tm := time.Unix(timestamp, 0)
	rsp.Timestamp = tm.Local().Format("2006-01-02 15:04:05")
	rsp.TxNumber = block.Block.TxCount
	for _, v := range block.Block.Transactions {
		rsp.Txs = append(rsp.Txs, hex.EncodeToString(v.Txid))
	}
	return rsp, nil
}

// type ContractCountReq struct {
// 	ChainName string `json:"chain_name"`
// }

// type TxQueryReq struct {
// 	ChainName string `json:"chain_name"`
// 	Input     string `json:"input"`
// }

type BlockRsp struct {
	BlockHeight int64    `json:"block_height"`
	BlockId     string   `json:"block_id"`
	Miner       string   `json:"miner"`
	NextHash    string   `json:"next_hash"`
	PreHash     string   `json:"pre_hash"`
	Timestamp   string   `json:"timestamp"`
	TxNumber    int32    `json:"tx_number"`
	Txs         []string `json:"txs"`
}

type TxDetailRsp struct {
	TxId           string   `json:"tx_id"`
	BlockId        string   `json:"block_id"`
	BlockHeight    int64    `json:"block_height"`
	Coinbase       bool     `json:"coinbase"`
	Miner          string   `json:"miner"`
	BlockTimestamp string   `json:"block_timestamp"`
	Initiator      string   `json:"initiator"`
	FromAddress    []string `json:"from_address"`
	ToAddresses    []string `json:"to_addresses"`
	FromTotal      *big.Int `json:"from_total"`
	ToTotal        *big.Int `json:"to_total"`
	Fee            int64    `json:"fee"`
	Date           string   `json:"date"`
	Contracts      []string `json:"contracts"`
}

const (
	ByTxId = iota + 1
	ByBlockHeight
	ByBlockId
)
