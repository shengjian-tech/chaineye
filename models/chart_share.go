package models

import (
	"gitee.com/chunanyong/zorm"
	"github.com/ccfos/nightingale/v6/pkg/ctx"
)

const ChartShareTableName = "chart_share"

type ChartShare struct {
	// 引入默认的struct,隔离IEntityStruct的方法改动
	zorm.EntityStruct
	Id           int64  `json:"id" column:"id"`
	Cluster      string `json:"cluster" column:"cluster"`
	DatasourceId int64  `json:"datasource_id" column:"datasource_id"`
	Configs      string `json:"configs" column:"configs"`
	CreateBy     string `json:"create_by" column:"create_by"`
	CreateAt     int64  `json:"create_at" column:"create_at"`
}

func (cs *ChartShare) GetTableName() string {
	return ChartShareTableName
}

func (cs *ChartShare) Add(ctx *ctx.Context) error {
	return Insert(ctx, cs)
}

func ChartShareGetsByIds(ctx *ctx.Context, ids []int64) ([]ChartShare, error) {
	lst := make([]ChartShare, 0)
	if len(ids) == 0 {
		return lst, nil
	}
	finder := zorm.NewSelectFinder(ChartShareTableName).Append("WHERE id in (?) order by id asc", ids)
	err := zorm.Query(ctx.Ctx, finder, &lst, nil)
	//err := DB(ctx).Where("id in ?", ids).Order("id").Find(&lst).Error
	return lst, err
}
