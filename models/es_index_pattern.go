package models

import (
	"fmt"
	"time"

	"gitee.com/chunanyong/zorm"
	"github.com/ccfos/nightingale/v6/pkg/ctx"

	"errors"
)

const EsIndexPatternTableName = "es_index_pattern"

type EsIndexPattern struct {
	// 引入默认的struct,隔离IEntityStruct的方法改动
	zorm.EntityStruct
	Id                         int64  `json:"id" column:"id"`
	DatasourceId               int64  `json:"datasource_id" column:"datasource_id"`
	Name                       string `json:"name" column:"name"`
	TimeField                  string `json:"time_field" column:"time_field"`
	AllowHideSystemIndices     int    `json:"-" column:"allow_hide_system_indices"`
	AllowHideSystemIndicesBool bool   `json:"allow_hide_system_indices"`
	FieldsFormat               string `json:"fields_format" column:"fields_format"`
	CreateAt                   int64  `json:"create_at" column:"create_at"`
	CreateBy                   string `json:"create_by" column:"create_by"`
	UpdateAt                   int64  `json:"update_at" column:"update_at"`
	UpdateBy                   string `json:"update_by" column:"update_by"`
}

func (t *EsIndexPattern) GetTableName() string {
	return EsIndexPatternTableName
}

func (r *EsIndexPattern) Add(ctx *ctx.Context) error {
	esIndexPattern, err := EsIndexPatternGet(ctx, "datasource_id = ? and name = ?", r.DatasourceId, r.Name)
	if err != nil {
		return fmt.Errorf("failed to query es index pattern:%w", err)
	}

	if esIndexPattern != nil {
		return errors.New("es index pattern datasource and name already exists")
	}
	r.FE2DB()
	return Insert(ctx, r)
	//return DB(ctx).Create(r).Error
}

func EsIndexPatternDel(ctx *ctx.Context, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	return DeleteByIds(ctx, EsIndexPatternTableName, ids)
	//return DB(ctx).Where("id in ?", ids).Delete(new(EsIndexPattern)).Error
}

func (ei *EsIndexPattern) Update(ctx *ctx.Context, eip EsIndexPattern) error {
	if ei.Name != eip.Name || ei.DatasourceId != eip.DatasourceId {
		exists, err := EsIndexPatternExists(ctx, ei.Id, eip.DatasourceId, eip.Name)
		if err != nil {
			return err
		}

		if exists {
			return errors.New("EsIndexPattern already exists")
		}
	}

	eip.Id = ei.Id
	eip.CreateAt = ei.CreateAt
	eip.CreateBy = ei.CreateBy
	eip.UpdateAt = time.Now().Unix()
	eip.FE2DB()
	return Update(ctx, &eip, nil)
	//return DB(ctx).Model(ei).Select("*").Updates(eip).Error
}

func (dbIndexPatten *EsIndexPattern) DB2FE() {
	if dbIndexPatten.AllowHideSystemIndices == 1 {
		dbIndexPatten.AllowHideSystemIndicesBool = true
	}
}

func (feIndexPatten *EsIndexPattern) FE2DB() {
	if feIndexPatten.AllowHideSystemIndicesBool {
		feIndexPatten.AllowHideSystemIndices = 1
	}
}

func EsIndexPatternGets(ctx *ctx.Context, where string, args ...interface{}) ([]*EsIndexPattern, error) {
	objs := make([]*EsIndexPattern, 0)
	finder := zorm.NewSelectFinder(EsIndexPatternTableName)
	AppendWhere(finder, where, args...)
	err := zorm.Query(ctx.Ctx, finder, &objs, nil)
	//err := DB(ctx).Where(where, args...).Find(&objs).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query es index pattern:%w", err)
	}

	for _, i := range objs {
		i.DB2FE()
	}
	return objs, nil
}

func EsIndexPatternGet(ctx *ctx.Context, where string, args ...interface{}) (*EsIndexPattern, error) {
	lst := make([]*EsIndexPattern, 0)
	finder := zorm.NewSelectFinder(EsIndexPatternTableName)
	AppendWhere(finder, where, args...)
	err := zorm.Query(ctx.Ctx, finder, &lst, nil)
	//err := DB(ctx).Where(where, args...).Find(&lst).Error
	if err != nil {
		return nil, err
	}

	if len(lst) == 0 {
		return nil, nil
	}

	lst[0].DB2FE()

	return lst[0], nil
}

func EsIndexPatternGetById(ctx *ctx.Context, id int64) (*EsIndexPattern, error) {
	return EsIndexPatternGet(ctx, "id=?", id)
}

func EsIndexPatternExists(ctx *ctx.Context, id, datasourceId int64, name string) (bool, error) {
	finder := zorm.NewSelectFinder(EsIndexPatternTableName, "count(*)").Append("WHERE id <> ? and datasource_id = ? and name = ?", id, datasourceId, name)

	count := 0
	_, err := zorm.QueryRow(ctx.Ctx, finder, &count)
	if err != nil {
		return false, err
	}

	/*
		session := DB(ctx).Where("id <> ? and datasource_id = ? and name = ?", id, datasourceId, name)

		var lst []EsIndexPattern
		err := session.Find(&lst).Error
		if err != nil {
			return false, err
		}
	*/
	if count == 0 {
		return false, nil
	}
	return true, nil
}
