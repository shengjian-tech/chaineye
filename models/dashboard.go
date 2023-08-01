package models

import (
	"context"
	"strings"
	"time"

	"gitee.com/chunanyong/zorm"
	"github.com/ccfos/nightingale/v6/pkg/ctx"
	"github.com/pkg/errors"
	"github.com/toolkits/pkg/str"
)

const DashboardTableName = "dashboard"

type Dashboard struct {
	// 引入默认的struct,隔离IEntityStruct的方法改动
	zorm.EntityStruct
	Id       int64    `json:"id" column:"id"`
	GroupId  int64    `json:"group_id" column:"group_id"`
	Name     string   `json:"name" column:"name"`
	Tags     string   `json:"-" column:"tags"`
	TagsLst  []string `json:"tags"`
	Configs  string   `json:"configs" column:"configs"`
	CreateAt int64    `json:"create_at" column:"create_at"`
	CreateBy string   `json:"create_by" column:"create_by"`
	UpdateAt int64    `json:"update_at" column:"update_at"`
	UpdateBy string   `json:"update_by" column:"update_by"`
}

func (d *Dashboard) GetTableName() string {
	return DashboardTableName
}

func (d *Dashboard) Verify() error {
	if d.Name == "" {
		return errors.New("Name is blank")
	}

	if str.Dangerous(d.Name) {
		return errors.New("Name has invalid characters")
	}

	return nil
}

func (d *Dashboard) Add(ctx *ctx.Context) error {
	if err := d.Verify(); err != nil {
		return err
	}

	exists, err := DashboardExists(ctx, "group_id=? and name=?", d.GroupId, d.Name)
	if err != nil {
		return errors.WithMessage(err, "failed to count dashboard")
	}

	if exists {
		return errors.New("Dashboard already exists")
	}

	now := time.Now().Unix()
	d.CreateAt = now
	d.UpdateAt = now

	return Insert(ctx, d)
}

func (d *Dashboard) Update(ctx *ctx.Context, selectFields ...string) error {
	if err := d.Verify(); err != nil {
		return err
	}
	return Update(ctx, d, selectFields)
	//return DB(ctx).Model(d).Select(selectField, selectFields...).Updates(d).Error
}

func (d *Dashboard) Del(ctx *ctx.Context) error {
	cgids, err := ChartGroupIdsOf(ctx, d.Id)
	if err != nil {
		return err
	}

	_, err = zorm.Transaction(ctx.Ctx, func(ctx context.Context) (interface{}, error) {
		if len(cgids) == 0 {
			f1 := zorm.NewDeleteFinder(DashboardTableName).Append("WHERE id=?", d.Id)
			return zorm.UpdateFinder(ctx, f1)
		}
		f2 := zorm.NewDeleteFinder(ChartTableName).Append("WHERE group_id in (?)", cgids)
		_, err = zorm.UpdateFinder(ctx, f2)
		if err != nil {
			return nil, err
		}
		f3 := zorm.NewDeleteFinder(ChartGroupTableName).Append("WHERE dashboard_id=?", d.Id)
		_, err = zorm.UpdateFinder(ctx, f3)
		if err != nil {
			return nil, err
		}

		f4 := zorm.NewDeleteFinder(DashboardTableName).Append("WHERE id=?", d.Id)
		return zorm.UpdateFinder(ctx, f4)

	})
	return err

	/*
		if len(cgids) == 0 {
			return DB(ctx).Transaction(func(tx *zorm.DBDao) error {
				if err := tx.Where("id=?", d.Id).Delete(&Dashboard{}).Error; err != nil {
					return err
				}
				return nil
			})
		}

		return DB(ctx).Transaction(func(tx *zorm.DBDao) error {
			if err := tx.Where("group_id in ?", cgids).Delete(&Chart{}).Error; err != nil {
				return err
			}

			if err := tx.Where("dashboard_id=?", d.Id).Delete(&ChartGroup{}).Error; err != nil {
				return err
			}

			if err := tx.Where("id=?", d.Id).Delete(&Dashboard{}).Error; err != nil {
				return err
			}

			return nil
		})
	*/
}

func DashboardGet(ctx *ctx.Context, where string, args ...interface{}) (*Dashboard, error) {
	lst := make([]Dashboard, 0)
	finder := zorm.NewSelectFinder(DashboardTableName)
	AppendWhere(finder, where, args...)
	err := zorm.Query(ctx.Ctx, finder, &lst, nil)
	//err := DB(ctx).Where(where, args...).Find(&lst).Error
	if err != nil {
		return nil, err
	}

	if len(lst) == 0 {
		return nil, nil
	}

	lst[0].TagsLst = strings.Fields(lst[0].Tags)

	return &lst[0], nil
}

func DashboardCount(ctx *ctx.Context, where string, args ...interface{}) (num int64, err error) {
	finder := zorm.NewSelectFinder(DashboardTableName, "count(*)")
	AppendWhere(finder, where, args...)
	return Count(ctx, finder)
	//return Count(DB(ctx).Model(&Dashboard{}).Where(where, args...))
}

func DashboardExists(ctx *ctx.Context, where string, args ...interface{}) (bool, error) {
	num, err := DashboardCount(ctx, where, args...)
	return num > 0, err
}

func DashboardGets(ctx *ctx.Context, groupId int64, query string) ([]Dashboard, error) {
	finder := zorm.NewSelectFinder(DashboardTableName, "id, group_id, name, tags, create_at, create_by, update_at, update_by").Append("WHERE group_id=?", groupId)
	//session := DB(ctx).Where("group_id=?", groupId).Order("name")

	arr := strings.Fields(query)
	if len(arr) > 0 {
		for i := 0; i < len(arr); i++ {
			if strings.HasPrefix(arr[i], "-") {
				q := "%" + arr[i][1:] + "%"
				//session = session.Where("name not like ? and tags not like ?", q, q)
				finder.Append("and name not like ? and tags not like ?", q, q)
			} else {
				q := "%" + arr[i] + "%"
				//session = session.Where("(name like ? or tags like ?)", q, q)
				finder.Append("and (name like ? or tags like ?)", q, q)
			}
		}
	}

	finder.Append("order by name asc")

	objs := make([]Dashboard, 0)
	err := zorm.Query(ctx.Ctx, finder, &objs, nil)
	//err := session.Select("id", "group_id", "name", "tags", "create_at", "create_by", "update_at", "update_by").Find(&objs).Error
	if err == nil {
		for i := 0; i < len(objs); i++ {
			objs[i].TagsLst = strings.Fields(objs[i].Tags)
		}
	}

	return objs, err
}

func DashboardGetsByIds(ctx *ctx.Context, ids []int64) ([]Dashboard, error) {
	if len(ids) == 0 {
		return []Dashboard{}, nil
	}

	lst := make([]Dashboard, 0)
	finder := zorm.NewSelectFinder(DashboardTableName).Append("WHERE id in (?) order by name asc", ids)
	err := zorm.Query(ctx.Ctx, finder, &lst, nil)
	//err := DB(ctx).Where("id in ?", ids).Order("name").Find(&lst).Error
	return lst, err
}

func DashboardGetAll(ctx *ctx.Context) ([]Dashboard, error) {
	lst := make([]Dashboard, 0)
	finder := zorm.NewSelectFinder(DashboardTableName)
	err := zorm.Query(ctx.Ctx, finder, &lst, nil)
	//err := DB(ctx).Find(&lst).Error
	return lst, err
}
