package models

import (
	"context"

	"gitee.com/chunanyong/zorm"
	"github.com/ccfos/nightingale/v6/pkg/ctx"

	"github.com/toolkits/pkg/str"
)

const AdminRole = "Admin"

// if rule's cluster field contains `ClusterAll`, means it take effect in all clusters
const DatasourceIdAll = 0

func DB(ctx *ctx.Context) *zorm.DBDao {
	return ctx.DB
}

func Count(ctx *ctx.Context, finder *zorm.Finder) (int64, error) {
	var cnt int64
	//err := finder.Count(&cnt).Error
	_, err := zorm.QueryRow(ctx.Ctx, finder, &cnt)
	return cnt, err
}

func Exists(ctx *ctx.Context, finder *zorm.Finder) (bool, error) {
	//var cnt int64
	//return zorm.QueryRow(ctx.Ctx, finder, &cnt)
	//num, err := Count(tx)
	num, err := Count(ctx, finder)
	return num > 0, err
}

func Insert(ctx *ctx.Context, obj zorm.IEntityStruct) error {
	//return DB(ctx).Create(obj).Error
	_, err := zorm.Transaction(ctx.Ctx, func(ctx context.Context) (interface{}, error) {
		return zorm.Insert(ctx, obj)
	})
	return err
}

// CryptoPass crypto password use salt
func CryptoPass(ctx *ctx.Context, raw string) (string, error) {
	salt, err := ConfigsGet(ctx, "salt")
	if err != nil {
		return "", err
	}

	return str.MD5(salt + "<-*Uk30^96eY*->" + raw), nil
}

type Statistics struct {
	Total       int64
	LastUpdated int64
}

func StatisticsGet(ctx *ctx.Context, tableName string) (*Statistics, error) {
	var stats Statistics
	finder := zorm.NewSelectFinder(tableName, "count(*) as Total , max(update_at) as LastUpdated")
	//session := DB(ctx).Model(model).Select("count(*) as total", "max(update_at) as last_updated")

	_, err := zorm.QueryRow(ctx.Ctx, finder, &stats)

	//err := session.Find(&stats).Error
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

func MatchDatasource(ids []int64, id int64) bool {
	if id == DatasourceIdAll {
		return true
	}

	for _, i := range ids {
		if i == id {
			return true
		}
	}
	return false
}

func IsAllDatasource(datasourceIds []int64) bool {
	for _, id := range datasourceIds {
		if id == 0 {
			return true
		}
	}
	return false
}

type LabelAndKey struct {
	Label string `json:"label"`
	Key   string `json:"key"`
}

func LabelAndKeyHasKey(keys []LabelAndKey, key string) bool {
	for i := 0; i < len(keys); i++ {
		if keys[i].Key == key {
			return true
		}
	}
	return false
}

func UpdateFieldsMap(ctx *ctx.Context, entity zorm.IEntityStruct, idValue interface{}, fields map[string]interface{}) error {
	entityMap := zorm.NewEntityMap(AlertCurEventTableName)
	entityMap.PkColumnName = entity.GetPKColumnName()
	entityMap.Set(entity.GetPKColumnName(), idValue)
	for k, v := range fields {
		entityMap.Set(k, v)
	}
	_, err := zorm.UpdateEntityMap(ctx.Ctx, entityMap)
	return err
}

func UpdateColumn(ctx *ctx.Context, tableName string, idValue interface{}, column string, value interface{}) error {
	finder := zorm.NewUpdateFinder(tableName).Append(column+"=?", value).Append("WHERE id=?", idValue)
	return UpdateFinder(ctx, finder)
}

func DeleteByIds(ctx *ctx.Context, tableName string, ids []int64) error {
	finder := zorm.NewDeleteFinder(tableName).Append("WHERE id in (?)", ids)
	return UpdateFinder(ctx, finder)
}

func Update(ctx *ctx.Context, entity zorm.IEntityStruct, cols []string) error {
	_, err := zorm.Transaction(ctx.Ctx, func(ctx context.Context) (interface{}, error) {
		//指定仅更新的列
		if len(cols) > 0 {
			ctx, _ = zorm.BindContextOnlyUpdateCols(ctx, cols)
		}
		return zorm.Update(ctx, entity)
	})
	return err
}

func UpdateFinder(ctx *ctx.Context, finder *zorm.Finder) error {
	_, err := zorm.Transaction(ctx.Ctx, func(ctx context.Context) (interface{}, error) {
		return zorm.UpdateFinder(ctx, finder)
	})
	return err
}

func AppendWhere(finder *zorm.Finder, where string, args ...interface{}) {
	if where != "" {
		finder.Append("WHERE "+where, args...)
	}
}
