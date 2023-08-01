package models

import (
	"gitee.com/chunanyong/zorm"
	"github.com/ccfos/nightingale/v6/pkg/ctx"
)

const BuiltinCateTableName = "builtin_cate"

type BuiltinCate struct {
	// 引入默认的struct,隔离IEntityStruct的方法改动
	zorm.EntityStruct
	Id     int64  `json:"id" column:"id"`
	Name   string `json:"name" column:"name"`
	UserId int64  `json:"user_id" column:"user_id"`
}

func (b *BuiltinCate) GetTableName() string {
	return BuiltinCateTableName
}

// 创建 builtin_cate
func (b *BuiltinCate) Create(c *ctx.Context) error {
	return Insert(c, b)
}

// 删除 builtin_cate
func BuiltinCateDelete(ctx *ctx.Context, name string, userId int64) error {
	finder := zorm.NewDeleteFinder(BuiltinCateTableName).Append("WHERE name=? and user_id=?", name, userId)
	return UpdateFinder(ctx, finder)
	//return DB(c).Where("name=? and user_id=?", name, userId).Delete(&BuiltinCate{}).Error
}

// 根据 userId 获取 builtin_cate
func BuiltinCateGetByUserId(ctx *ctx.Context, userId int64) (map[string]BuiltinCate, error) {
	builtinCates := make([]BuiltinCate, 0)
	finder := zorm.NewSelectFinder(BuiltinCateTableName).Append("WHERE user_id=?", userId)
	err := zorm.Query(ctx.Ctx, finder, &builtinCates, nil)
	//err := DB(c).Where("user_id=?", userId).Find(&builtinCates).Error
	var builtinCatesMap = make(map[string]BuiltinCate)
	for _, builtinCate := range builtinCates {
		builtinCatesMap[builtinCate.Name] = builtinCate
	}

	return builtinCatesMap, err
}
