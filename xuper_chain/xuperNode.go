package xuper_chain

import (
	"context"

	"gitee.com/chunanyong/zorm"
	"github.com/ccfos/nightingale/v6/models"
)

func InsertXuperNode(ctx context.Context, xuperNode *models.XuperNode) error {
	_, err := zorm.Transaction(ctx, func(ctx context.Context) (interface{}, error) {
		zorm.Delete(ctx, xuperNode)
		_, err := zorm.Insert(ctx, xuperNode)
		return nil, err
	})
	return err
}

func GetXuperNode(ctx context.Context) ([]models.XuperNode, error) {
	f := zorm.NewSelectFinder(models.XuperNodeStructTableName)
	var xuperNodes []models.XuperNode
	err := zorm.Query(ctx, f, &xuperNodes, nil)
	return xuperNodes, err
}

func GetXuperNodeByRootName(ctx context.Context, rootnet string) ([]models.XuperNode, error) {
	f := zorm.NewSelectFinder(models.XuperNodeStructTableName)
	f.Append("where rootnet = ?", rootnet)
	var xuperNodes []models.XuperNode
	err := zorm.Query(ctx, f, &xuperNodes, nil)
	return xuperNodes, err
}
