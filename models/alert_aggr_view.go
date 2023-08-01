package models

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"gitee.com/chunanyong/zorm"
	"github.com/ccfos/nightingale/v6/pkg/ctx"
	"github.com/toolkits/pkg/slice"
)

const AlertAggrViewTableName = "alert_aggr_view"

// AlertAggrView 在告警聚合视图查看的时候，要存储一些聚合规则
type AlertAggrView struct {
	zorm.EntityStruct
	Id       int64  `json:"id" column:"id"`
	Name     string `json:"name" column:"name"`
	Rule     string `json:"rule" column:"rule"`
	Cate     int    `json:"cate" column:"cate"`
	CreateAt int64  `json:"create_at" column:"create_at"`
	CreateBy int64  `json:"create_by" column:"create_by"`
	UpdateAt int64  `json:"update_at" column:"update_at"`
}

func (v *AlertAggrView) GetTableName() string {
	return AlertAggrViewTableName
}

func (v *AlertAggrView) DB2FE() error {
	return nil
}

func (v *AlertAggrView) Verify() error {
	v.Name = strings.TrimSpace(v.Name)
	if v.Name == "" {
		return errors.New("name is blank")
	}

	v.Rule = strings.TrimSpace(v.Rule)
	if v.Rule == "" {
		return errors.New("rule is blank")
	}

	var validFields = []string{
		"cluster",
		"group_id",
		"group_name",
		"rule_id",
		"rule_name",
		"severity",
		"runbook_url",
		"target_ident",
		"target_note",
	}

	arr := strings.Split(v.Rule, "::")
	for i := 0; i < len(arr); i++ {
		pair := strings.Split(arr[i], ":")
		if len(pair) != 2 {
			return errors.New("rule invalid")
		}

		if !(pair[0] == "field" || pair[0] == "tagkey") {
			return errors.New("rule invalid")
		}

		if pair[0] == "field" {
			// 只支持有限的field
			if !slice.ContainsString(validFields, pair[1]) {
				return fmt.Errorf("unsupported field: %s", pair[1])
			}
		}
	}

	return nil
}

func (v *AlertAggrView) Add(ctx *ctx.Context) error {
	if err := v.Verify(); err != nil {
		return err
	}

	now := time.Now().Unix()
	v.CreateAt = now
	v.UpdateAt = now
	v.Cate = 1
	return Insert(ctx, v)
}

func (v *AlertAggrView) Update(ctx *ctx.Context) error {
	if err := v.Verify(); err != nil {
		return err
	}
	v.UpdateAt = time.Now().Unix()
	finder := zorm.NewUpdateFinder(AlertAggrViewTableName)
	finder.Append("name=?,rule=?,cate=?,update_at=?,create_by=? WHERE id=?", v.Name, v.Rule, v.Cate, v.UpdateAt, v.CreateBy, v.Id)
	return UpdateFinder(ctx, finder)
	//return DB(ctx).Model(v).Select("name", "rule", "cate", "update_at", "create_by").Updates(v).Error
}

// AlertAggrViewDel: userid for safe delete
func AlertAggrViewDel(ctx *ctx.Context, ids []int64, createBy ...interface{}) error {
	if len(ids) == 0 {
		return nil
	}
	finder := zorm.NewDeleteFinder(AlertAggrViewTableName).Append("WHERE id in (?) ", ids)
	if len(createBy) > 0 {
		finder.Append(" and create_by = ?", createBy)
	}
	return UpdateFinder(ctx, finder)
	/*
		if len(createBy) > 0 {
			return DB(ctx).Where("id in ? and create_by = ?", ids, createBy).Delete(new(AlertAggrView)).Error
		}

		return DB(ctx).Where("id in ?", ids).Delete(new(AlertAggrView)).Error
	*/
}

func AlertAggrViewGets(ctx *ctx.Context, createBy interface{}) ([]AlertAggrView, error) {
	lst := make([]AlertAggrView, 0)
	finder := zorm.NewSelectFinder(AlertAggrViewTableName).Append("WHERE create_by = ? or cate = 0", createBy)
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

func AlertAggrViewGet(ctx *ctx.Context, where string, args ...interface{}) (*AlertAggrView, error) {
	lst := make([]AlertAggrView, 0)
	finder := zorm.NewSelectFinder(AlertAggrViewTableName)
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
