package models

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"gitee.com/chunanyong/zorm"
	"github.com/ccfos/nightingale/v6/pkg/ctx"
	"github.com/toolkits/pkg/str"
)

const TaskTplTableName = "task_tpl"

type TaskTpl struct {
	// 引入默认的struct,隔离IEntityStruct的方法改动
	zorm.EntityStruct
	Id        int64    `json:"id" column:"id"`
	GroupId   int64    `json:"group_id" column:"group_id"`
	Title     string   `json:"title" column:"title"`
	Batch     int      `json:"batch" column:"batch"`
	Tolerance int      `json:"tolerance" column:"tolerance"`
	Timeout   int      `json:"timeout" column:"timeout"`
	Pause     string   `json:"pause" column:"pause"`
	Script    string   `json:"script" column:"script"`
	Args      string   `json:"args" column:"args"`
	Tags      string   `json:"-" column:"tags"`
	TagsJSON  []string `json:"tags"`
	Account   string   `json:"account" column:"account"`
	CreateAt  int64    `json:"create_at" column:"create_at"`
	CreateBy  string   `json:"create_by" column:"create_by"`
	UpdateAt  int64    `json:"update_at" column:"update_at"`
	UpdateBy  string   `json:"update_by" column:"update_by"`
}

func (t *TaskTpl) GetTableName() string {
	return TaskTplTableName
}

func (t *TaskTpl) DB2FE() error {
	return nil
}

func TaskTplTotal(ctx *ctx.Context, groupId int64, query string) (int64, error) {
	finder := zorm.NewSelectFinder(TaskTplTableName, "count(*)").Append("WHERE group_id = ?", groupId)
	//session := DB(ctx).Model(&TaskTpl{}).Where("group_id = ?", groupId)
	if query == "" {
		return Count(ctx, finder)
	}

	arr := strings.Fields(query)
	for i := 0; i < len(arr); i++ {
		arg := "%" + arr[i] + "%"
		//session = session.Where("title like ? or tags like ?", arg, arg)
		finder.Append("and (title like ? or tags like ?)", arg, arg)
	}

	return Count(ctx, finder)
}

func TaskTplGets(ctx *ctx.Context, groupId int64, query string, limit, offset int) ([]TaskTpl, error) {
	//session := DB(ctx).Where("group_id = ?", groupId).Order("title").Limit(limit).Offset(offset)
	finder := zorm.NewSelectFinder(TaskTplTableName).Append("WHERE group_id = ?", groupId)

	tpls := make([]TaskTpl, 0)
	if query != "" {
		arr := strings.Fields(query)
		for i := 0; i < len(arr); i++ {
			arg := "%" + arr[i] + "%"
			//session = session.Where("title like ? or tags like ?", arg, arg)
			finder.Append("and (title like ? or tags like ?)", arg, arg)
		}
	}

	finder.Append("order by title asc ")
	page := zorm.NewPage()
	page.PageSize = limit
	page.PageNo = offset / limit
	finder.SelectTotalCount = false
	err := zorm.Query(ctx.Ctx, finder, &tpls, page)
	//err := session.Find(&tpls).Error
	if err == nil {
		for i := 0; i < len(tpls); i++ {
			tpls[i].TagsJSON = strings.Fields(tpls[i].Tags)
		}
	}

	return tpls, err
}

func TaskTplGet(ctx *ctx.Context, where string, args ...interface{}) (*TaskTpl, error) {
	finder := zorm.NewSelectFinder(TaskTplTableName)
	AppendWhere(finder, where, args...)
	arr := make([]*TaskTpl, 0)
	err := zorm.Query(ctx.Ctx, finder, &arr, nil)
	//err := DB(ctx).Where(where, args...).Find(&arr).Error
	if err != nil {
		return nil, err
	}

	if len(arr) == 0 {
		return nil, nil
	}

	arr[0].TagsJSON = strings.Fields(arr[0].Tags)

	return arr[0], nil
}

func (t *TaskTpl) CleanFields() error {
	if t.Batch < 0 {
		return errors.New("arg(batch) should be nonnegative")
	}

	if t.Tolerance < 0 {
		return errors.New("arg(tolerance) should be nonnegative")
	}

	if t.Timeout < 0 {
		return errors.New("arg(timeout) should be nonnegative")
	}

	if t.Timeout == 0 {
		t.Timeout = 30
	}

	if t.Timeout > 3600*24 {
		return errors.New("arg(timeout) longer than one day")
	}

	t.Pause = strings.Replace(t.Pause, "，", ",", -1)
	t.Pause = strings.Replace(t.Pause, " ", "", -1)
	t.Args = strings.Replace(t.Args, "，", ",", -1)
	t.Tags = strings.Replace(t.Tags, "，", ",", -1)

	if t.Title == "" {
		return errors.New("arg(title) is required")
	}

	if str.Dangerous(t.Title) {
		return errors.New("arg(title) is dangerous")
	}

	if t.Script == "" {
		return errors.New("arg(script) is required")
	}

	if str.Dangerous(t.Args) {
		return errors.New("arg(args) is dangerous")
	}

	if str.Dangerous(t.Pause) {
		return errors.New("arg(pause) is dangerous")
	}

	if str.Dangerous(t.Tags) {
		return errors.New("arg(tags) is dangerous")
	}

	return nil
}

func (t *TaskTpl) Save(ctx *ctx.Context, hosts []string) error {
	if err := t.CleanFields(); err != nil {
		return err
	}
	finder := zorm.NewSelectFinder(TaskTplTableName, "count(*)").Append("WHERE group_id=? and title=?", t.GroupId, t.Title)
	cnt, err := Count(ctx, finder)
	//cnt, err := Count(DB(ctx).Model(&TaskTpl{}).Where("group_id=? and title=?", t.GroupId, t.Title))
	if err != nil {
		return err
	}

	if cnt > 0 {
		return fmt.Errorf("task template already exists")
	}

	_, err = zorm.Transaction(ctx.Ctx, func(ctx context.Context) (interface{}, error) {
		_, err := zorm.Insert(ctx, t)
		if err != nil {
			return nil, err
		}
		entityMaps := make([]zorm.IEntityMap, 0)
		for i := 0; i < len(hosts); i++ {
			host := strings.TrimSpace(hosts[i])
			if host == "" {
				continue
			}
			entityMap := zorm.NewEntityMap("task_tpl_host")
			entityMap.Set("id", t.Id)
			entityMap.Set("host", host)
			entityMaps = append(entityMaps, entityMap)
		}
		if len(entityMaps) > 0 {
			_, err = zorm.InsertEntityMapSlice(ctx, entityMaps)
		}
		return nil, err
	})
	return err
	/*
		return DB(ctx).Transaction(func(tx *zorm.DBDao) error {
			if err := tx.Create(t).Error; err != nil {
				return err
			}

			for i := 0; i < len(hosts); i++ {
				host := strings.TrimSpace(hosts[i])
				if host == "" {
					continue
				}

				err := tx.Table("task_tpl_host").Create(map[string]interface{}{
					"id":   t.Id,
					"host": host,
				}).Error

				if err != nil {
					return err
				}
			}

			return nil
		})
	*/
}

func (t *TaskTpl) Hosts(ctx *ctx.Context) ([]string, error) {
	arr := make([]string, 0)
	finder := zorm.NewSelectFinder("task_tpl_host", "host").Append("WHERE id=? order by ii asc ", t.Id)
	err := zorm.Query(ctx.Ctx, finder, &arr, nil)
	//err := DB(ctx).Table("task_tpl_host").Where("id=?", t.Id).Order("ii").Pluck("host", &arr).Error
	return arr, err
}

func (t *TaskTpl) Update(ctx *ctx.Context, hosts []string) error {
	if err := t.CleanFields(); err != nil {
		return err
	}

	finder := zorm.NewSelectFinder(TaskTplTableName, "count(*)").Append("WHERE group_id=? and title=? and id <> ?", t.GroupId, t.Title, t.Id)
	cnt, err := Count(ctx, finder)
	//cnt, err := Count(DB(ctx).Model(&TaskTpl{}).Where("group_id=? and title=? and id <> ?", t.GroupId, t.Title, t.Id))
	if err != nil {
		return err
	}

	if cnt > 0 {
		return fmt.Errorf("task template already exists")
	}

	_, err = zorm.Transaction(ctx.Ctx, func(ctx context.Context) (interface{}, error) {
		entityMap := zorm.NewEntityMap(TaskTplTableName)
		entityMap.PkColumnName = t.GetPKColumnName()
		entityMap.Set(t.GetPKColumnName(), t.Id)
		entityMap.Set("title", t.Title)
		entityMap.Set("batch", t.Batch)
		entityMap.Set("tolerance", t.Tolerance)
		entityMap.Set("timeout", t.Timeout)
		entityMap.Set("pause", t.Pause)
		entityMap.Set("script", t.Script)
		entityMap.Set("args", t.Args)
		entityMap.Set("tags", t.Tags)
		entityMap.Set("account", t.Account)
		entityMap.Set("update_by", t.UpdateBy)
		entityMap.Set("update_at", t.UpdateAt)
		_, err := zorm.UpdateEntityMap(ctx, entityMap)
		if err != nil {
			return nil, err
		}
		f1 := zorm.NewFinder().Append("DELETE FROM task_tpl_host WHERE id = ?", t.Id)
		_, err = zorm.UpdateFinder(ctx, f1)
		if err != nil {
			return nil, err
		}
		entityMaps := make([]zorm.IEntityMap, 0)
		for i := 0; i < len(hosts); i++ {
			host := strings.TrimSpace(hosts[i])
			if host == "" {
				continue
			}
			entityMap := zorm.NewEntityMap("task_tpl_host")
			entityMap.Set("id", t.Id)
			entityMap.Set("host", host)
			entityMaps = append(entityMaps, entityMap)
		}
		if len(entityMaps) > 0 {
			_, err = zorm.InsertEntityMapSlice(ctx, entityMaps)
		}
		return nil, err
	})

	return err

	/*
		return DB(ctx).Transaction(func(tx *zorm.DBDao) error {
			err := tx.Model(t).Updates(map[string]interface{}{
				"title":     t.Title,
				"batch":     t.Batch,
				"tolerance": t.Tolerance,
				"timeout":   t.Timeout,
				"pause":     t.Pause,
				"script":    t.Script,
				"args":      t.Args,
				"tags":      t.Tags,
				"account":   t.Account,
				"update_by": t.UpdateBy,
				"update_at": t.UpdateAt,
			}).Error

			if err != nil {
				return err
			}

			if err = tx.Exec("DELETE FROM task_tpl_host WHERE id = ?", t.Id).Error; err != nil {
				return err
			}

			for i := 0; i < len(hosts); i++ {
				host := strings.TrimSpace(hosts[i])
				if host == "" {
					continue
				}

				err := tx.Table("task_tpl_host").Create(map[string]interface{}{
					"id":   t.Id,
					"host": host,
				}).Error

				if err != nil {
					return err
				}
			}

			return nil
		})
	*/
}

func (t *TaskTpl) Del(ctx *ctx.Context) error {

	_, err := zorm.Transaction(ctx.Ctx, func(ctx context.Context) (interface{}, error) {
		f1 := zorm.NewFinder().Append("DELETE FROM task_tpl_host WHERE id = ?", t.Id)
		_, err := zorm.UpdateFinder(ctx, f1)
		if err != nil {
			return nil, err
		}
		return zorm.Delete(ctx, t)
	})
	return err
	/*
		return DB(ctx).Transaction(func(tx *zorm.DBDao) error {
			if err := tx.Exec("DELETE FROM task_tpl_host WHERE id=?", t.Id).Error; err != nil {
				return err
			}

			if err := tx.Delete(t).Error; err != nil {
				return err
			}

			return nil
		})
	*/
}

func (t *TaskTpl) AddTags(ctx *ctx.Context, tags []string, updateBy string) error {
	for i := 0; i < len(tags); i++ {
		if -1 == strings.Index(t.Tags, tags[i]+" ") {
			t.Tags += tags[i] + " "
		}
	}

	arr := strings.Fields(t.Tags)
	sort.Strings(arr)

	finder := zorm.NewUpdateFinder(TaskTplTableName).Append("tags=?,update_by=?,update_at=? WHERE id=?", strings.Join(arr, " ")+" ", updateBy, time.Now().Unix(), t.Id)
	return UpdateFinder(ctx, finder)
	/*
		return DB(ctx).Model(t).Updates(map[string]interface{}{
			"tags":      strings.Join(arr, " ") + " ",
			"update_by": updateBy,
			"update_at": time.Now().Unix(),
		}).Error
	*/
}

func (t *TaskTpl) DelTags(ctx *ctx.Context, tags []string, updateBy string) error {
	for i := 0; i < len(tags); i++ {
		t.Tags = strings.ReplaceAll(t.Tags, tags[i]+" ", "")
	}
	finder := zorm.NewUpdateFinder(TaskTplTableName).Append("tags=?,update_by=?,update_at=? WHERE id=?", t.Tags, updateBy, time.Now().Unix(), t.Id)
	return UpdateFinder(ctx, finder)
	/*
		return DB(ctx).Model(t).Updates(map[string]interface{}{
			"tags":      t.Tags,
			"update_by": updateBy,
			"update_at": time.Now().Unix(),
		}).Error
	*/
}

func (t *TaskTpl) UpdateGroup(ctx *ctx.Context, groupId int64, updateBy string) error {
	finder := zorm.NewUpdateFinder(TaskTplTableName).Append("group_id=?,update_by=?,update_at=? WHERE id=?", groupId, updateBy, time.Now().Unix(), t.Id)
	return UpdateFinder(ctx, finder)
	/*
		return DB(ctx).Model(t).Updates(map[string]interface{}{
			"group_id":  groupId,
			"update_by": updateBy,
			"update_at": time.Now().Unix(),
		}).Error
	*/
}
