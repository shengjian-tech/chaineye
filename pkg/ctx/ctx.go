package ctx

import (
	"context"

	"gitee.com/chunanyong/zorm"
	"github.com/ccfos/nightingale/v6/conf"
)

type Context struct {
	DB        *zorm.DBDao
	CenterApi conf.CenterApi
	Ctx       context.Context
	IsCenter  bool
}

func NewContext(ctx context.Context, db *zorm.DBDao, isCenter bool, centerApis ...conf.CenterApi) *Context {
	var api conf.CenterApi
	if len(centerApis) > 0 {
		api = centerApis[0]
	}

	return &Context{
		Ctx:       ctx,
		DB:        db,
		CenterApi: api,
		IsCenter:  isCenter,
	}
}

// set db to Context
func (c *Context) SetDB(db *zorm.DBDao) {
	c.DB = db
}

// get context from Context
func (c *Context) GetContext() context.Context {
	return c.Ctx
}

// get db from Context
func (c *Context) GetDB() *zorm.DBDao {
	return c.DB
}
