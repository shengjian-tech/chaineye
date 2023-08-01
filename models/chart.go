package models

import (
	"gitee.com/chunanyong/zorm"
	"github.com/ccfos/nightingale/v6/pkg/ctx"
)

const ChartTableName = "chart"

type Chart struct {
	// 引入默认的struct,隔离IEntityStruct的方法改动
	zorm.EntityStruct
	Id      int64  `json:"id" column:"id"`
	GroupId int64  `json:"group_id" column:"group_id"`
	Configs string `json:"configs" column:"configs"`
	Weight  int    `json:"weight" column:"weight"`
}

func (c *Chart) GetTableName() string {
	return ChartTableName
}

func ChartsOf(ctx *ctx.Context, chartGroupId int64) ([]Chart, error) {
	objs := make([]Chart, 0)
	finder := zorm.NewSelectFinder(ChartTableName).Append("WHERE group_id = ? order by weight asc", chartGroupId)
	err := zorm.Query(ctx.Ctx, finder, &objs, nil)
	//err := DB(ctx).Where("group_id = ?", chartGroupId).Order("weight").Find(&objs).Error
	return objs, err
}

func (c *Chart) Add(ctx *ctx.Context) error {
	return Insert(ctx, c)
}

func (c *Chart) Update(ctx *ctx.Context, selectFields ...string) error {
	return Update(ctx, c, selectFields)
	//return DB(ctx).Model(c).Select(selectField, selectFields...).Updates(c).Error
}

func (c *Chart) Del(ctx *ctx.Context) error {
	finder := zorm.NewDeleteFinder(ChartTableName).Append("WHERE id=?", c.Id)
	return UpdateFinder(ctx, finder)
	//return DB(ctx).Where("id=?", c.Id).Delete(&Chart{}).Error
}
