package models

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"errors"

	"gitee.com/chunanyong/zorm"
	"github.com/ccfos/nightingale/v6/pkg/ctx"
	"github.com/ccfos/nightingale/v6/pkg/poster"
	"github.com/toolkits/pkg/logger"
	"github.com/toolkits/pkg/str"
)

const DatasourceTableName = "datasource"

type Datasource struct {
	// 引入默认的struct,隔离IEntityStruct的方法改动
	zorm.EntityStruct
	Id             int64                  `json:"id" column:"id"`
	Name           string                 `json:"name" column:"name"`
	Description    string                 `json:"description" column:"description"`
	PluginId       int64                  `json:"plugin_id" column:"plugin_id"`
	PluginType     string                 `json:"plugin_type" column:"plugin_type"`           // prometheus
	PluginTypeName string                 `json:"plugin_type_name" column:"plugin_type_name"` // Prometheus Like
	Category       string                 `json:"category" column:"category"`                 // timeseries
	ClusterName    string                 `json:"cluster_name" column:"cluster_name"`
	Settings       string                 `json:"-" column:"settings"`
	SettingsJson   map[string]interface{} `json:"settings"`
	Status         string                 `json:"status" column:"status"`
	HTTP           string                 `json:"-" column:"http"`
	HTTPJson       HTTP                   `json:"http"`
	Auth           string                 `json:"-" column:"auth"`
	AuthJson       Auth                   `json:"auth"`
	CreatedAt      int64                  `json:"created_at" column:"created_at"`
	UpdatedAt      int64                  `json:"updated_at" column:"updated_at"`
	CreatedBy      string                 `json:"created_by" column:"created_by"`
	UpdatedBy      string                 `json:"updated_by" column:"updated_by"`
	IsDefault      bool                   `json:"is_default" column:"is_default"`
	Transport      *http.Transport        `json:"-"`
}

type Auth struct {
	BasicAuth         bool   `json:"basic_auth"`
	BasicAuthUser     string `json:"basic_auth_user"`
	BasicAuthPassword string `json:"basic_auth_password"`
}

type HTTP struct {
	Timeout             int64             `json:"timeout"`
	DialTimeout         int64             `json:"dial_timeout"`
	TLS                 TLS               `json:"tls"`
	MaxIdleConnsPerHost int               `json:"max_idle_conns_per_host"`
	Url                 string            `json:"url"`
	Headers             map[string]string `json:"headers"`
}

func (h HTTP) IsLoki() bool {
	if strings.Contains(h.Url, "loki") {
		return true
	}

	for k := range h.Headers {
		tmp := strings.ToLower(k)
		if strings.Contains(tmp, "loki") {
			return true
		}
	}

	return false
}

type TLS struct {
	SkipTlsVerify bool `json:"skip_tls_verify"`
}

func (ds *Datasource) GetTableName() string {
	return DatasourceTableName
}

func (ds *Datasource) Verify() error {
	if str.Dangerous(ds.Name) {
		return errors.New("Name has invalid characters")
	}

	err := ds.FE2DB()
	return err
}

func (ds *Datasource) Update(ctx *ctx.Context, selectFields ...string) error {
	if err := ds.Verify(); err != nil {
		return err
	}
	ds.UpdatedAt = time.Now().Unix()
	return Update(ctx, ds, selectFields)
	//return DB(ctx).Model(ds).Select(selectField, selectFields...).Updates(ds).Error
}

func (ds *Datasource) Add(ctx *ctx.Context) error {
	if err := ds.Verify(); err != nil {
		return err
	}

	now := time.Now().Unix()
	ds.CreatedAt = now
	ds.UpdatedAt = now
	return Insert(ctx, ds)
}

func DatasourceDel(ctx *ctx.Context, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	return DeleteByIds(ctx, DatasourceTableName, ids)
	//return DB(ctx).Where("id in ?", ids).Delete(new(Datasource)).Error
}

func DatasourceGet(ctx *ctx.Context, id int64) (*Datasource, error) {
	var ds Datasource
	finder := zorm.NewSelectFinder(DatasourceTableName).Append("WHERE id=?", id)
	_, err := zorm.QueryRow(ctx.Ctx, finder, &ds)
	//err := DB(ctx).Where("id = ?", id).First(&ds).Error
	if err != nil {
		return nil, err
	}
	return &ds, ds.DB2FE()
}

func (ds *Datasource) Get(ctx *ctx.Context) error {
	finder := zorm.NewSelectFinder(DatasourceTableName).Append("WHERE id=?", ds.Id)
	_, err := zorm.QueryRow(ctx.Ctx, finder, ds)
	//err := DB(ctx).Where("id = ?", ds.Id).First(ds).Error
	if err != nil {
		return err
	}
	return ds.DB2FE()
}

func GetDatasources(ctx *ctx.Context) ([]Datasource, error) {
	if !ctx.IsCenter {
		lst, err := poster.GetByUrls[[]Datasource](ctx, "/v1/n9e/datasources")
		if err != nil {
			return nil, err
		}
		for i := 0; i < len(lst); i++ {
			lst[i].FE2DB()
		}
		return lst, nil
	}

	dss := make([]Datasource, 0)
	finder := zorm.NewSelectFinder(DatasourceTableName)
	err := zorm.Query(ctx.Ctx, finder, &dss, nil)
	//err := DB(ctx).Find(&dss).Error

	for i := 0; i < len(dss); i++ {
		dss[i].DB2FE()
	}

	return dss, err
}

func GetDatasourceIdsByEngineName(ctx *ctx.Context, engineName string) ([]int64, error) {
	if !ctx.IsCenter {
		lst, err := poster.GetByUrls[[]int64](ctx, "/v1/n9e/datasource-ids?name="+engineName)
		return lst, err
	}

	//dss := make([]Datasource, 0)
	ids := make([]int64, 0)
	finder := zorm.NewSelectFinder(DatasourceTableName, "id").Append("WHERE cluster_name = ?", engineName)
	err := zorm.Query(ctx.Ctx, finder, &ids, nil)
	/*
		err := DB(ctx).Where("cluster_name = ?", engineName).Find(&dss).Error
		if err != nil {
			return ids, err
		}

		for i := 0; i < len(dss); i++ {
			ids = append(ids, dss[i].Id)
		}
	*/
	return ids, err
}

func GetDatasourcesCountByName(ctx *ctx.Context, name string) (int64, error) {
	finder := zorm.NewSelectFinder(DatasourceTableName, "count(*)")
	//session := DB(ctx).Model(&Datasource{})
	if name != "" {
		//session = session.Where("name = ?", name)
		finder.Append("WHERE name = ?", name)
	}

	return Count(ctx, finder)
}

func GetDatasourcesCountBy(ctx *ctx.Context, typ, cate, name string) (int64, error) {
	finder := zorm.NewSelectFinder(DatasourceTableName, "count(*)").Append("WHERE 1=1")
	//session := DB(ctx).Model(&Datasource{})

	if name != "" {
		arr := strings.Fields(name)
		for i := 0; i < len(arr); i++ {
			qarg := "%" + arr[i] + "%"
			//session = session.Where("name =  ?", qarg)
			finder.Append("and name =  ?", qarg)
		}
	}

	if typ != "" {
		//session = session.Where("plugin_type = ?", typ)
		finder.Append("and plugin_type = ?", typ)
	}

	if cate != "" {
		//session = session.Where("category = ?", cate)
		finder.Append("and category = ?", cate)
	}

	return Count(ctx, finder)
}

func GetDatasourcesGetsBy(ctx *ctx.Context, typ, cate, name, status string) ([]*Datasource, error) {
	finder := zorm.NewSelectFinder(DatasourceTableName).Append("WHERE 1=1")
	//session := DB(ctx)

	if name != "" {
		arr := strings.Fields(name)
		for i := 0; i < len(arr); i++ {
			qarg := "%" + arr[i] + "%"
			//session = session.Where("name =  ?", qarg)
			finder.Append("and name =  ?", qarg)
		}
	}

	if typ != "" {
		//session = session.Where("plugin_type = ?", typ)
		finder.Append("and plugin_type = ?", typ)
	}

	if cate != "" {
		//session = session.Where("category = ?", cate)
		finder.Append("and category = ?", cate)
	}

	if status != "" {
		//session = session.Where("status = ?", status)
		finder.Append("and status = ?", status)
	}
	finder.Append("order by id desc")
	lst := make([]*Datasource, 0)
	err := zorm.Query(ctx.Ctx, finder, &lst, nil)
	//err := session.Order("id desc").Find(&lst).Error
	if err == nil {
		for i := 0; i < len(lst); i++ {
			lst[i].DB2FE()
		}
	}
	return lst, err
}

func GetDatasourcesGetsByTypes(ctx *ctx.Context, typs []string) (map[string]*Datasource, error) {
	lst := make([]*Datasource, 0)
	m := make(map[string]*Datasource)
	finder := zorm.NewSelectFinder(DatasourceTableName).Append("WHERE plugin_type in (?)", typs)
	err := zorm.Query(ctx.Ctx, finder, &lst, nil)
	//err := DB(ctx).Where("plugin_type in ?", typs).Find(&lst).Error
	if err == nil {
		for i := 0; i < len(lst); i++ {
			lst[i].DB2FE()
			m[lst[i].Name] = lst[i]
		}
	}
	return m, err
}

func (ds *Datasource) FE2DB() error {
	if ds.SettingsJson != nil {
		b, err := json.Marshal(ds.SettingsJson)
		if err != nil {
			return err
		}
		ds.Settings = string(b)
	}

	b, err := json.Marshal(ds.HTTPJson)
	if err != nil {
		return err
	}
	ds.HTTP = string(b)

	b, err = json.Marshal(ds.AuthJson)
	if err != nil {
		return err
	}
	ds.Auth = string(b)

	return nil
}

func (ds *Datasource) DB2FE() error {
	if ds.Settings != "" {
		err := json.Unmarshal([]byte(ds.Settings), &ds.SettingsJson)
		if err != nil {
			return err
		}
	}

	if ds.HTTP != "" {
		err := json.Unmarshal([]byte(ds.HTTP), &ds.HTTPJson)
		if err != nil {
			return err
		}
	}

	if ds.HTTPJson.Timeout == 0 {
		ds.HTTPJson.Timeout = 10000
	}

	if ds.HTTPJson.DialTimeout == 0 {
		ds.HTTPJson.DialTimeout = 10000
	}

	if ds.HTTPJson.MaxIdleConnsPerHost == 0 {
		ds.HTTPJson.MaxIdleConnsPerHost = 100
	}

	if ds.Auth != "" {
		err := json.Unmarshal([]byte(ds.Auth), &ds.AuthJson)
		if err != nil {
			return err
		}
	}

	return nil
}

func DatasourceGetMap(ctx *ctx.Context) (map[int64]*Datasource, error) {
	lst := make([]*Datasource, 0)
	var err error
	if !ctx.IsCenter {
		lst, err = poster.GetByUrls[[]*Datasource](ctx, "/v1/n9e/datasources")
		if err != nil {
			return nil, err
		}
		for i := 0; i < len(lst); i++ {
			lst[i].FE2DB()
		}
	} else {
		finder := zorm.NewSelectFinder(DatasourceTableName)
		err := zorm.Query(ctx.Ctx, finder, &lst, nil)
		//err := DB(ctx).Find(&lst).Error
		if err != nil {
			return nil, err
		}

		for i := 0; i < len(lst); i++ {
			err := lst[i].DB2FE()
			if err != nil {
				logger.Warningf("get ds:%+v err:%v", lst[i], err)
				continue
			}
		}
	}

	ret := make(map[int64]*Datasource)
	for i := 0; i < len(lst); i++ {
		ret[lst[i].Id] = lst[i]
	}

	return ret, nil
}

func DatasourceStatistics(ctx *ctx.Context) (*Statistics, error) {
	if !ctx.IsCenter {
		s, err := poster.GetByUrls[*Statistics](ctx, "/v1/n9e/statistic?name=datasource")
		return s, err
	}
	var stats Statistics
	finder := zorm.NewSelectFinder(DatasourceTableName, "count(*) as Total , max(updated_at) as LastUpdated")
	_, err := zorm.QueryRow(ctx.Ctx, finder, &stats)
	return &stats, err
	//return StatisticsGet(ctx, DatasourceTableName)
	/*
		session := DB(ctx).Model(&Datasource{}).Select("count(*) as total", "max(updated_at) as last_updated")

		var stats []*Statistics
		err := session.Find(&stats).Error
		if err != nil {
			return nil, err
		}

		return stats[0], nil
	*/
}
