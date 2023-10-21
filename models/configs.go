package models

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"time"

	"gitee.com/chunanyong/zorm"
	"github.com/ccfos/nightingale/v6/pkg/ctx"
	"github.com/ccfos/nightingale/v6/pkg/poster"
	"github.com/ccfos/nightingale/v6/pkg/secu"
	"github.com/pkg/errors"

	"github.com/toolkits/pkg/logger"
	"github.com/toolkits/pkg/runner"
	"github.com/toolkits/pkg/str"
)

const ConfigsTableName = "configs"

type Configs struct {
	// 引入默认的struct,隔离IEntityStruct的方法改动
	zorm.EntityStruct
	Id        int64  `column:"id" json:"id"`
	Ckey      string `column:"ckey" json:"ckey"` //Unique field. Before inserting external configs, check if they are already defined as built-in configs.
	Cval      string `column:"cval" json:"cval"`
	Note      string `column:"note" json:"note"`
	External  int    `column:"external" json:"external"`   //Controls frontend list display: 0 hides built-in (default), 1 shows external
	Encrypted int    `column:"encrypted" json:"encrypted"` //Indicates whether the value(cval) is encrypted (1 for ciphertext, 0 for plaintext(default))
	CreateAt  int64  `column:"create_at" json:"create_at"`
	CreateBy  string `column:"create_by" json:"create_by"`
	UpdateAt  int64  `column:"update_at" json:"update_at"`
	UpdateBy  string `column:"update_by" json:"update_by"`
}

func (Configs) GetTableName() string {
	return ConfigsTableName
}

var (
	ConfigExternal  = 1 //external type
	ConfigEncrypted = 1 //ciphertext
)

func (c *Configs) DB2FE() error {
	return nil
}

const (
	SALT            = "salt"
	RSA_PRIVATE_KEY = "rsa_private_key"
	RSA_PUBLIC_KEY  = "rsa_public_key"
	RSA_PASSWORD    = "rsa_password"
)

// InitSalt generate random salt
func InitSalt(ctx *ctx.Context) {
	val, err := ConfigsGet(ctx, SALT)
	if err != nil {
		log.Fatalln("init salt in mysql", err)
	}

	if val != "" {
		return
	}

	content := fmt.Sprintf("%s%d%d%s", runner.Hostname, os.Getpid(), time.Now().UnixNano(), str.RandLetters(6))
	salt := str.MD5(content)
	err = ConfigsSet(ctx, SALT, salt)
	if err != nil {
		log.Fatalln("init salt in mysql", err)
	}

}
func InitRSAPassWord(ctx *ctx.Context) (string, error) {

	val, err := ConfigsGet(ctx, RSA_PASSWORD)
	if err != nil {
		return "", errors.WithMessage(err, "failed to get rsa password")
	}
	if val != "" {
		return val, nil
	}
	content := fmt.Sprintf("%s%d%d%s", runner.Hostname, os.Getpid(), time.Now().UnixNano(), str.RandLetters(6))
	pwd := str.MD5(content)
	err = ConfigsSet(ctx, RSA_PASSWORD, pwd)
	if err != nil {
		return "", errors.WithMessage(err, "failed to set rsa password")
	}
	return pwd, nil
}

func ConfigsGet(ctx *ctx.Context, ckey string) (string, error) { //select built-in type configs
	if !ctx.IsCenter {
		if !ctx.IsCenter {
			s, err := poster.GetByUrls[string](ctx, "/v1/n9e/config?key="+ckey)
			return s, err
		}
	}

	lst := make([]string, 0)
	finder := zorm.NewSelectFinder(ConfigsTableName, "cval").Append("WHERE ckey=? and external=? ", ckey, 0)
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
	return ConfigsSetWithUname(ctx, ckey, cval, "default")
}
func ConfigsSetWithUname(ctx *ctx.Context, ckey, cval, uName string) error { //built-in
	finder := zorm.NewSelectFinder(ConfigsTableName, "count(*)").Append("WHERE ckey=? and external=?", ckey, 0)
	num, err := Count(ctx, finder)
	//num, err := Count(DB(ctx).Model(&Configs{}).Where("ckey=?", ckey))
	if err != nil {
		return fmt.Errorf("failed to count configs:%w", err)
	}
	now := time.Now().Unix()
	if num == 0 {
		// insert
		/*
			err = DB(ctx).Create(&Configs{
				Ckey: ckey,
				Cval: cval,
			}).Error
		*/
		err = Insert(ctx, &Configs{
			Ckey:     ckey,
			Cval:     cval,
			CreateBy: uName,
			UpdateBy: uName,
			CreateAt: now,
			UpdateAt: now,
		})
	} else {
		// update
		finder := zorm.NewUpdateFinder(ConfigsTableName).Append("cval=?,update_by=?,update_at=? WHERE ckey=?", cval, uName, now, ckey)
		err = UpdateFinder(ctx, finder)
		//err = DB(ctx).Model(&Configs{}).Where("ckey=?", ckey).Update("cval", cval).Error
	}

	return err
}

func ConfigsSelectByCkey(ctx *ctx.Context, ckey string) ([]Configs, error) {
	objs := make([]Configs, 0)

	finder := zorm.NewSelectFinder(ConfigsTableName).Append("WHERE ckey=?", ckey)
	err := zorm.Query(ctx.Ctx, finder, &objs, nil)
	//err := DB(ctx).Where("ckey=?", ckey).Find(&objs).Error
	if err != nil {
		return nil, fmt.Errorf("failed to select conf:%w", err)
		//return nil, errors.WithMessage(err, "failed to select conf")
	}
	return objs, nil
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
	finder.Append("order by id desc")
	page := zorm.NewPage()
	page.PageSize = limit
	page.PageNo = offset / limit
	finder.SelectTotalCount = false
	err := zorm.Query(ctx.Ctx, finder, &objs, page)
	//err := session.Order("id desc").Limit(limit).Offset(offset).Find(&objs).Error
	return objs, err
}

func (c *Configs) Add(ctx *ctx.Context) error {
	finder := zorm.NewSelectFinder(ConfigsTableName, "count(*)").Append("WHERE ckey=? and external=? ", c.Ckey, c.External)
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
		Ckey:     c.Ckey,
		Cval:     c.Cval,
		External: c.External,
		CreateBy: c.CreateBy,
		UpdateBy: c.CreateBy,
		CreateAt: c.CreateAt,
		UpdateAt: c.CreateAt,
	})

	return err
}

func (c *Configs) Update(ctx *ctx.Context) error {
	finder := zorm.NewSelectFinder(ConfigsTableName, "count(*)").Append("WHERE id<>? and ckey=? and external=? ", c.Id, c.Ckey, c.External)
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

func ConfigsGetUserVariable(context *ctx.Context) ([]Configs, error) {
	objs := make([]Configs, 0)
	finder := zorm.NewSelectFinder(ConfigsTableName).Append("WHERE external = ? order by id desc ", ConfigExternal)
	err := zorm.Query(context.Ctx, finder, &objs, nil)
	//tx := DB(context).Where("external = ?", ConfigExternal).Order("id desc")
	//err := tx.Find(&objs).Error
	if err != nil {
		return nil, fmt.Errorf("failed to gets user variable:%w", err)
		//return nil, errors.WithMessage(err, "failed to gets user variable")
	}

	return objs, nil
}

func ConfigsUserVariableInsert(context *ctx.Context, conf Configs) error {
	conf.External = ConfigExternal
	conf.Id = 0
	err := userVariableCheck(context, conf.Ckey, conf.Id)
	if err != nil {
		return err
	}
	return Insert(context, &conf)
	//return DB(context).Create(&conf).Error
}

func ConfigsUserVariableUpdate(context *ctx.Context, conf Configs) error {
	err := userVariableCheck(context, conf.Ckey, conf.Id)
	if err != nil {
		return err
	}
	configOld, _ := ConfigGet(context, conf.Id)
	if configOld == nil || configOld.External != ConfigExternal { //not valid id
		return fmt.Errorf("not valid configs(id)")
	}
	return Update(context, &conf, []string{"ckey", "cval", "note", "encrypted", "update_by", "update_at"})
	//return DB(context).Model(&Configs{Id: conf.Id}).Select("ckey", "cval", "note", "encrypted").Updates(conf).Error
}
func isCStyleIdentifier(str string) bool {
	regex := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
	return regex.MatchString(str)
}
func userVariableCheck(context *ctx.Context, ckey string, id int64) error {
	var num int64
	var err error
	if !isCStyleIdentifier(ckey) {
		return fmt.Errorf("invalid key(%q), please use C-style naming convention ", ckey)
	}
	if id != 0 { //update
		finder := zorm.NewSelectFinder(ConfigsTableName, "count(*)").Append("WHERE id <> ? and ckey = ? and external=?", id, ckey, ConfigExternal)
		num, err = Count(context, finder)
		//err = DB(context).Where("id <> ? and ckey = ? and external=?", &id, ckey, ConfigExternal).Find(&objs).Error
	} else {
		finder := zorm.NewSelectFinder(ConfigsTableName, "count(*)").Append("WHERE ckey = ? and external=?", ckey, ConfigExternal)
		num, err = Count(context, finder)
		//err = DB(context).Where("ckey = ? and external=?", ckey, ConfigExternal).Find(&objs).Error
	}
	if err != nil {
		return err
	}
	if num == 0 {
		return nil
	}
	return fmt.Errorf("duplicate ckey value found: %s", ckey)
}

func ConfigsUserVariableStatistics(context *ctx.Context) (*Statistics, error) {
	if !context.IsCenter {
		return poster.GetByUrls[*Statistics](context, "/v1/n9e/statistic?name=user_variable")
	}
	statistics, err := StatisticsGet(context, ConfigsTableName)
	if err != nil {
		return nil, err
	}
	return statistics, nil
}

func ConfigUserVariableGetDecryptMap(context *ctx.Context, privateKey []byte, passWord string) (map[string]string, error) {

	if !context.IsCenter {
		ret, err := poster.GetByUrls[map[string]string](context, "/v1/n9e/user-variable/decrypt")
		if err != nil {
			return nil, err
		}
		return ret, nil
	}
	lst, err := ConfigsGetUserVariable(context)
	if err != nil {
		return nil, err
	}
	ret := make(map[string]string, len(lst))
	for i := 0; i < len(lst); i++ {
		if lst[i].Encrypted != ConfigEncrypted {
			ret[lst[i].Ckey] = lst[i].Cval
		} else {
			decCval, decErr := secu.Decrypt(lst[i].Cval, privateKey, passWord)
			if decErr != nil {
				logger.Errorf("RSA Decrypt failed: %v. Ckey: %s", decErr, lst[i].Ckey)
				decCval = ""
			}
			ret[lst[i].Ckey] = decCval
		}
	}

	return ret, nil
}
