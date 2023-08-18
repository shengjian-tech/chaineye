package models

import (
	"fmt"
	"log"
	"os"
	"time"

	"gitee.com/chunanyong/zorm"
	"github.com/ccfos/nightingale/v6/pkg/ctx"
	"github.com/ccfos/nightingale/v6/pkg/poster"

	"github.com/toolkits/pkg/runner"
	"github.com/toolkits/pkg/str"
)

const ConfigsTableName = "configs"

type Configs struct {
	// 引入默认的struct,隔离IEntityStruct的方法改动
	zorm.EntityStruct
	Id   int64  `column:"id"`
	Ckey string `column:"ckey"`
	Cval string `column:"cval"`
}

func (Configs) GetTableName() string {
	return ConfigsTableName
}

func (c *Configs) DB2FE() error {
	return nil
}

// InitSalt generate random salt
func InitSalt(ctx *ctx.Context) {
	val, err := ConfigsGet(ctx, "salt")
	if err != nil {
		log.Fatalln("cannot query salt", err)
	}

	if val != "" {
		return
	}

	content := fmt.Sprintf("%s%d%d%s", runner.Hostname, os.Getpid(), time.Now().UnixNano(), str.RandLetters(6))
	salt := str.MD5(content)
	err = ConfigsSet(ctx, "salt", salt)
	if err != nil {
		log.Fatalln("init salt in mysql", err)
	}
}

func ConfigsGet(ctx *ctx.Context, ckey string) (string, error) {
	if !ctx.IsCenter {
		if !ctx.IsCenter {
			s, err := poster.GetByUrls[string](ctx, "/v1/n9e/config?key="+ckey)
			return s, err
		}
	}

	lst := make([]string, 0)
	finder := zorm.NewSelectFinder(ConfigsTableName, "cval").Append("WHERE ckey=?", ckey)
	err := zorm.Query(ctx.Ctx, finder, &lst, nil)
	//err := DB(ctx).Model(&Configs{}).Where("ckey=?", ckey).Pluck("cval", &lst).Error
	if err != nil {
		return "", fmt.Errorf("failed to query configs:%w", err)
	}

	if len(lst) > 0 {
		return lst[0], nil
	}

	return "", nil
}

func ConfigsSet(ctx *ctx.Context, ckey, cval string) error {
	finder := zorm.NewSelectFinder(ConfigsTableName, "count(*)").Append("WHERE ckey=?", ckey)
	num, err := Count(ctx, finder)
	//num, err := Count(DB(ctx).Model(&Configs{}).Where("ckey=?", ckey))
	if err != nil {
		return fmt.Errorf("failed to count configs:%w", err)
	}

	if num == 0 {
		// insert
		/*
			err = DB(ctx).Create(&Configs{
				Ckey: ckey,
				Cval: cval,
			}).Error
		*/
		err = Insert(ctx, &Configs{
			Ckey: ckey,
			Cval: cval,
		})
	} else {
		// update
		finder := zorm.NewUpdateFinder(ConfigsTableName).Append("cval=? WHERE ckey=?", cval, ckey)
		err = UpdateFinder(ctx, finder)
		//err = DB(ctx).Model(&Configs{}).Where("ckey=?", ckey).Update("cval", cval).Error
	}

	return err
}

func ConfigGet(ctx *ctx.Context, id int64) (*Configs, error) {
	objs := make([]Configs, 0)
	finder := zorm.NewSelectFinder(ConfigsTableName).Append("WHERE id=?", id)
	err := zorm.Query(ctx.Ctx, finder, &objs, nil)
	//err := DB(ctx).Where("id=?", id).Find(&objs).Error

	if len(objs) == 0 {
		return nil, nil
	}
	return &objs[0], err
}

func ConfigsGets(ctx *ctx.Context, prefix string, limit, offset int) ([]*Configs, error) {
	objs := make([]*Configs, 0)
	finder := zorm.NewSelectFinder(ConfigsTableName)
	//session := DB(ctx)
	if prefix != "" {
		//session = session.Where("ckey like ?", prefix+"%")
		finder.Append("WHERE ckey like ?", prefix+"%")
	}
	finder.Append(" order by id desc")
	page := zorm.NewPage()
	page.PageSize = limit
	page.PageNo = offset / limit
	finder.SelectTotalCount = false
	err := zorm.Query(ctx.Ctx, finder, &objs, page)
	//err := session.Order("id desc").Limit(limit).Offset(offset).Find(&objs).Error
	return objs, err
}

func (c *Configs) Add(ctx *ctx.Context) error {
	finder := zorm.NewSelectFinder(ConfigsTableName, "count(*)").Append("WHERE ckey=?", c.Ckey)
	num, err := Count(ctx, finder)
	//num, err := Count(DB(ctx).Model(&Configs{}).Where("ckey=?", c.Ckey))
	if err != nil {
		return fmt.Errorf("failed to count configs:%w", err)
	}
	if num > 0 {
		return fmt.Errorf("key is exists:%w", err)
	}

	// insert
	/*
		err = DB(ctx).Create(&Configs{
			Ckey: c.Ckey,
			Cval: c.Cval,
		}).Error
	*/
	err = Insert(ctx, &Configs{
		Ckey: c.Ckey,
		Cval: c.Cval,
	})

	return err
}

func (c *Configs) Update(ctx *ctx.Context) error {
	finder := zorm.NewSelectFinder(ConfigsTableName, "count(*)").Append("WHERE id<>? and ckey=?", c.Id, c.Ckey)
	num, err := Count(ctx, finder)
	//num, err := Count(DB(ctx).Model(&Configs{}).Where("id<>? and ckey=?", c.Id, c.Ckey))
	if err != nil {
		return fmt.Errorf("failed to count configs:%w", err)
	}
	if num > 0 {
		return fmt.Errorf("key is exists:%w", err)
	}
	return Update(ctx, c, nil)
	//err = DB(ctx).Model(&Configs{}).Where("id=?", c.Id).Updates(c).Error
}

func ConfigsDel(ctx *ctx.Context, ids []int64) error {
	return DeleteByIds(ctx, ConfigsTableName, ids)
	//return DB(ctx).Where("id in ?", ids).Delete(&Configs{}).Error
}

func ConfigsGetsByKey(ctx *ctx.Context, ckeys []string) (map[string]string, error) {
	objs := make([]Configs, 0)
	finder := zorm.NewSelectFinder(ConfigsTableName).Append("WHERE ckey in (?)", ckeys)
	err := zorm.Query(ctx.Ctx, finder, &objs, nil)
	//err := DB(ctx).Where("ckey in ?", ckeys).Find(&objs).Error
	if err != nil {
		return nil, fmt.Errorf("failed to gets configs:%w", err)
	}

	count := len(ckeys)
	kvmap := make(map[string]string, count)
	for i := 0; i < count; i++ {
		kvmap[ckeys[i]] = ""
	}

	for i := 0; i < len(objs); i++ {
		kvmap[objs[i].Ckey] = objs[i].Cval
	}

	return kvmap, nil
}
