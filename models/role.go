package models

import (
	"errors"
	"fmt"

	"gitee.com/chunanyong/zorm"
	"github.com/ccfos/nightingale/v6/pkg/ctx"
)

const RoleTableName = "role"

type Role struct {
	// 引入默认的struct,隔离IEntityStruct的方法改动
	zorm.EntityStruct
	Id   int64  `json:"id" column:"id"`
	Name string `json:"name" column:"name"`
	Note string `json:"note" column:"note"`
}

func (Role) GetTableName() string {
	return RoleTableName
}

func (r *Role) DB2FE() error {
	return nil
}

func RoleGets(ctx *ctx.Context, where string, args ...interface{}) ([]Role, error) {
	objs := make([]Role, 0)
	finder := zorm.NewSelectFinder(RoleTableName)
	AppendWhere(finder, where, args...)
	err := zorm.Query(ctx.Ctx, finder, &objs, nil)
	//err := DB(ctx).Where(where, args...).Find(&objs).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query roles:%w", err)
	}
	return objs, nil
}

func RoleGetsAll(ctx *ctx.Context) ([]Role, error) {
	return RoleGets(ctx, "")
}

// 增加角色
func (r *Role) Add(ctx *ctx.Context) error {
	role, err := RoleGet(ctx, "name = ?", r.Name)
	if err != nil {
		return fmt.Errorf("failed to query user:%w", err)
	}

	if role != nil {
		return errors.New("role name already exists")
	}
	return Insert(ctx, r)
	//return DB(ctx).Create(r).Error
}

// 删除角色
func (r *Role) Del(ctx *ctx.Context) error {
	_, err := zorm.Delete(ctx.Ctx, r)
	return err
	//return DB(ctx).Delete(r).Error
}

// 更新角色
func (ug *Role) Update(ctx *ctx.Context, selectFields ...string) error {
	return Update(ctx, ug, selectFields)
	//return DB(ctx).Model(ug).Select(selectField, selectFields...).Updates(ug).Error
}

func RoleGet(ctx *ctx.Context, where string, args ...interface{}) (*Role, error) {
	lst := make([]Role, 0)
	finder := zorm.NewSelectFinder(RoleTableName)
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

func RoleCount(ctx *ctx.Context, where string, args ...interface{}) (num int64, err error) {
	finder := zorm.NewSelectFinder(RoleTableName, "count(*)")
	AppendWhere(finder, where, args...)
	return Count(ctx, finder)
	//return Count(DB(ctx).Model(&Role{}).Where(where, args...))
}
