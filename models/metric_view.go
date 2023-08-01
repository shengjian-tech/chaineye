package models

import (
	"errors"
	"sort"
	"strings"
	"time"

	"gitee.com/chunanyong/zorm"
	"github.com/ccfos/nightingale/v6/pkg/ctx"
)

const MetricViewTableName = "metric_view"

// MetricView 在告警聚合视图查看的时候，要存储一些聚合规则
type MetricView struct {
	// 引入默认的struct,隔离IEntityStruct的方法改动
	zorm.EntityStruct
	Id       int64  `json:"id" column:"id"`
	Name     string `json:"name" column:"name"`
	Cate     int    `json:"cate" column:"cate"`
	Configs  string `json:"configs" column:"configs"`
	CreateAt int64  `json:"create_at" column:"create_at"`
	CreateBy int64  `json:"create_by" column:"create_by"`
	UpdateAt int64  `json:"update_at" column:"update_at"`
}

func (v *MetricView) GetTableName() string {
	return MetricViewTableName
}

func (v *MetricView) DB2FE() error {
	return nil
}

func (v *MetricView) Verify() error {
	v.Name = strings.TrimSpace(v.Name)
	if v.Name == "" {
		return errors.New("name is blank")
	}

	v.Configs = strings.TrimSpace(v.Configs)
	if v.Configs == "" {
		return errors.New("configs is blank")
	}

	return nil
}

func (v *MetricView) Add(ctx *ctx.Context) error {
	if err := v.Verify(); err != nil {
		return err
	}

	now := time.Now().Unix()
	v.CreateAt = now
	v.UpdateAt = now
	return Insert(ctx, v)
}

func (v *MetricView) Update(ctx *ctx.Context, name, configs string, cate int, createBy int64) error {
	if err := v.Verify(); err != nil {
		return err
	}

	v.UpdateAt = time.Now().Unix()
	v.Name = name
	v.Configs = configs
	v.Cate = cate

	if v.CreateBy == 0 {
		v.CreateBy = createBy
	}

	return Update(ctx, v, []string{"name", "configs", "cate", "update_at", "create_by"})
	//return DB(ctx).Model(v).Select("name", "configs", "cate", "update_at", "create_by").Updates(v).Error
}

// MetricViewDel: userid for safe delete
func MetricViewDel(ctx *ctx.Context, ids []int64, createBy ...interface{}) error {
	if len(ids) == 0 {
		return nil
	}

	finder := zorm.NewDeleteFinder(MetricViewTableName).Append("WHERE id in (?)", ids)

	if len(createBy) > 0 {
		//return DB(ctx).Where("id in ? and create_by = ?", ids, createBy[0]).Delete(new(MetricView)).Error
		finder.Append("and create_by = ?", createBy[0])
	}

	return UpdateFinder(ctx, finder)

	//return DB(ctx).Where("id in ?", ids).Delete(new(MetricView)).Error
}

func MetricViewGets(ctx *ctx.Context, createBy interface{}) ([]MetricView, error) {
	lst := make([]MetricView, 0)
	finder := zorm.NewSelectFinder(MetricViewTableName).Append("WHERE create_by = ? or cate = 0", createBy)
	err := zorm.Query(ctx.Ctx, finder, &lst, nil)
	//err := DB(ctx).Where("create_by = ? or cate = 0", createBy).Find(&lst).Error
	if err == nil && len(lst) > 1 {
		sort.Slice(lst, func(i, j int) bool {
			if lst[i].Cate < lst[j].Cate {
				return true
			}

			if lst[i].Cate > lst[j].Cate {
				return false
			}

			return lst[i].Name < lst[j].Name
		})
	}
	return lst, err
}

func MetricViewGet(ctx *ctx.Context, where string, args ...interface{}) (*MetricView, error) {
	lst := make([]MetricView, 0)
	finder := zorm.NewSelectFinder(MetricViewTableName)
	AppendWhere(finder, where, args...)
	err := zorm.Query(ctx.Ctx, finder, &lst, nil)
	//err := DB(ctx).Where(where, args...).Find(&lst).Error
	if err != nil {
		return nil, err
	}

	if len(lst) == 0 {
		return nil, nil
	}

	return &lst[0], nil
}
