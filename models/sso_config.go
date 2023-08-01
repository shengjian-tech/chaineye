package models

import (
	"gitee.com/chunanyong/zorm"
	"github.com/ccfos/nightingale/v6/pkg/ctx"
)

const SsoConfigTableName = "sso_config"

type SsoConfig struct {
	// 引入默认的struct,隔离IEntityStruct的方法改动
	zorm.EntityStruct
	Id      int64  `json:"id" column:"id"`
	Name    string `json:"name" column:"name"`
	Content string `json:"content" column:"content"`
}

func (b *SsoConfig) GetTableName() string {
	return SsoConfigTableName
}

func (b *SsoConfig) DB2FE() error {
	return nil
}

// get all sso_config
func SsoConfigGets(ctx *ctx.Context) ([]SsoConfig, error) {
	lst := make([]SsoConfig, 0)
	finder := zorm.NewSelectFinder(SsoConfigTableName)
	err := zorm.Query(ctx.Ctx, finder, &lst, nil)
	//err := DB(ctx).Find(&lst).Error
	return lst, err
}

// 创建 builtin_cate
func (b *SsoConfig) Create(c *ctx.Context) error {
	return Insert(c, b)
}

func (b *SsoConfig) Update(ctx *ctx.Context) error {
	return UpdateColumn(ctx, SsoConfigTableName, b.Id, "content", b.Content)
	//return DB(c).Model(b).Select("content").Updates(b).Error
}

// get sso_config coutn by name
func SsoConfigCountByName(ctx *ctx.Context, name string) (int64, error) {
	finder := zorm.NewSelectFinder(SsoConfigTableName, "count(*)").Append("WHERE name = ?", name)
	return Count(ctx, finder)
	/*
		var count int64
		err := DB(c).Model(&SsoConfig{}).Where("name = ?", name).Count(&count).Error
		return count, err
	*/
}
