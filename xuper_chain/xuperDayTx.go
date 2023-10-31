package xuper_chain

import (
	"context"

	"gitee.com/chunanyong/zorm"
	"github.com/ccfos/nightingale/v6/models"
	"github.com/toolkits/pkg/logger"
)

// 根据日期查询每日交易数
func GetXuperDayTx(ctx context.Context, id string) ([]models.XuperDayTx, error) {
	f := zorm.NewSelectFinder(models.XuperDayTxStructTableName)
	f.Append("where id = ? ", id)
	var xuperDayTx []models.XuperDayTx
	err := zorm.Query(ctx, f, &xuperDayTx, nil)
	return xuperDayTx, err
}

func GetXuperTenTxs(ctx context.Context, rootnet string, chainName string) ([]models.XuperDayTx, error) {
	f := zorm.NewSelectFinder(models.XuperDayTxStructTableName)
	f.Append("where rootnet = ? and chain_name = ? order by block_day asc", rootnet, chainName)
	p := zorm.NewPage()
	p.PageSize = 10
	p.PageNo = 1
	var xuperDayTx []models.XuperDayTx
	err := zorm.Query(ctx, f, &xuperDayTx, p)
	return xuperDayTx, err
}

// 查询数据库这个网络这条链最近2天的tx数据
func GetLastDayTx(ctx context.Context, rootnet string, chainName string) ([]models.XuperDayTx, error) {
	f := zorm.NewSelectFinder(models.XuperDayTxStructTableName)
	f.Append("where rootnet = ? AND chain_name = ? ORDER BY  block_day DESC", rootnet, chainName)
	p := zorm.NewPage()
	p.PageSize = 2
	p.PageNo = 1
	var xuperDayTx []models.XuperDayTx
	err := zorm.Query(ctx, f, &xuperDayTx, p)
	return xuperDayTx, err
}

// 查询指定网络指定链的最新交易总数
func GetTxTotal(ctx context.Context, rootnet string, chainName string) (int64, error) {
	var countInSection int64
	f := zorm.NewSelectFinder(models.XuperDayTxStructTableName, " MAX(total_tx_count) ")
	f.Append("where rootnet = ? AND chain_name = ? ", rootnet, chainName)
	_, err := zorm.QueryRow(ctx, f, &countInSection)
	if err != nil {
		logger.Error("get txs in db by height failed", err.Error())
		return countInSection, err
	}
	return countInSection, nil
}
