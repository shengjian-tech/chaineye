package models

import (
	"context"

	"gitee.com/chunanyong/zorm"
	"github.com/ccfos/nightingale/v6/pkg/ctx"
	"github.com/pkg/errors"
	"github.com/toolkits/pkg/str"
)

const ChartGroupTableName = "chart_group"

type ChartGroup struct {
	// 引入默认的struct,隔离IEntityStruct的方法改动
	zorm.EntityStruct
	Id          int64  `json:"id" column:"id"`
	DashboardId int64  `json:"dashboard_id" column:"dashboard_id"`
	Name        string `json:"name" column:"name"`
	Weight      int    `json:"weight" column:"weight"`
}

func (cg *ChartGroup) GetTableName() string {
	return ChartGroupTableName
}

func (cg *ChartGroup) Verify() error {
	if cg.DashboardId <= 0 {
		return errors.New("Arg(dashboard_id) invalid")
	}

	if str.Dangerous(cg.Name) {
		return errors.New("Name has invalid characters")
	}

	return nil
}

func (cg *ChartGroup) Add(ctx *ctx.Context) error {
	if err := cg.Verify(); err != nil {
		return err
	}

	return Insert(ctx, cg)
}

func (cg *ChartGroup) Update(ctx *ctx.Context, selectFields ...string) error {
	if err := cg.Verify(); err != nil {
		return err
	}
	return Update(ctx, cg, selectFields)
	//return DB(ctx).Model(cg).Select(selectField, selectFields...).Updates(cg).Error
}

func (cg *ChartGroup) Del(ctx *ctx.Context) error {
	/*
		return DB(ctx).Transaction(func(tx *zorm.DBDao) error {
			if err := tx.Where("group_id=?", cg.Id).Delete(&Chart{}).Error; err != nil {
				return err
			}

			if err := tx.Where("id=?", cg.Id).Delete(&ChartGroup{}).Error; err != nil {
				return err
			}

			return nil
		})
	*/
	_, err := zorm.Transaction(ctx.Ctx, func(ctx context.Context) (interface{}, error) {

		f1 := zorm.NewDeleteFinder(ChartTableName).Append("WHERE group_id=?", cg.Id)
		_, err := zorm.UpdateFinder(ctx, f1)
		if err != nil {
			return nil, err
		}
		f2 := zorm.NewDeleteFinder(ChartGroupTableName).Append("WHERE id=?", cg.Id)
		return zorm.UpdateFinder(ctx, f2)
	})
	return err

}

func NewDefaultChartGroup(ctx *ctx.Context, dashId int64) error {
	return Insert(ctx, &ChartGroup{
		DashboardId: dashId,
		Name:        "Default chart group",
		Weight:      0,
	})
}

func ChartGroupIdsOf(ctx *ctx.Context, dashId int64) ([]int64, error) {
	ids := make([]int64, 0)

	finder := zorm.NewSelectFinder(ChartGroupTableName, "id").Append("WHERE dashboard_id = ?", dashId)
	err := zorm.Query(ctx.Ctx, finder, &ids, nil)
	//err := DB(ctx).Model(&ChartGroup{}).Where("dashboard_id = ?", dashId).Pluck("id", &ids).Error
	return ids, err
}

func ChartGroupsOf(ctx *ctx.Context, dashId int64) ([]ChartGroup, error) {
	objs := make([]ChartGroup, 0)
	finder := zorm.NewSelectFinder(ChartGroupTableName).Append("WHERE dashboard_id = ? order by weight asc", dashId)
	err := zorm.Query(ctx.Ctx, finder, &objs, nil)
	//err := DB(ctx).Where("dashboard_id = ?", dashId).Order("weight").Find(&objs).Error
	return objs, err
}
