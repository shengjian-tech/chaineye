package models

import (
	"fmt"
	"time"

	"gitee.com/chunanyong/zorm"
	"github.com/ccfos/nightingale/v6/pkg/ctx"
	"github.com/ccfos/nightingale/v6/pkg/poster"
)

const AlertingEnginesTableName = "alerting_engines"

type AlertingEngines struct {
	// 引入默认的struct,隔离IEntityStruct的方法改动
	zorm.EntityStruct
	Id            int64  `json:"id" column:"id"`
	Instance      string `json:"instance" column:"instance"`
	EngineCluster string `json:"cluster" column:"engine_cluster"`
	DatasourceId  int64  `json:"datasource_id" column:"datasource_id"`
	Clock         int64  `json:"clock" column:"clock"`
}

func (e *AlertingEngines) GetTableName() string {
	return AlertingEnginesTableName
}

// UpdateCluster 页面上用户会给各个n9e-server分配要关联的目标集群是什么
func (e *AlertingEngines) UpdateDatasourceId(ctx *ctx.Context, id int64) error {
	finder := zorm.NewSelectFinder(AlertingEnginesTableName, "count(*)").Append("WHERE id<>? and instance=? and datasource_id=?", e.Id, e.Instance, id)
	count, err := Count(ctx, finder)
	//count, err := Count(DB(ctx).Model(&AlertingEngines{}).Where("id<>? and instance=? and datasource_id=?", e.Id, e.Instance, id))
	if err != nil {
		return err
	}

	if count > 0 {
		return fmt.Errorf("instance %s and datasource_id %d already exists", e.Instance, id)
	}

	e.DatasourceId = id
	return UpdateColumn(ctx, AlertingEnginesTableName, e.Id, "datasource_id", e.DatasourceId)
	//return DB(ctx).Model(e).Select("datasource_id").Updates(e).Error
}

func AlertingEngineAdd(ctx *ctx.Context, instance string, datasourceId int64) error {
	finder := zorm.NewSelectFinder(AlertingEnginesTableName, "count(*)").Append("WHERE instance=? and datasource_id=?", instance, datasourceId)
	count, err := Count(ctx, finder)
	//count, err := Count(DB(ctx).Model(&AlertingEngines{}).Where("instance=? and datasource_id=?", instance, datasourceId))
	if err != nil {
		return err
	}

	if count > 0 {
		return fmt.Errorf("instance %s and datasource_id %d already exists", instance, datasourceId)
	}

	err = Insert(ctx, &AlertingEngines{
		Instance:     instance,
		DatasourceId: datasourceId,
		Clock:        time.Now().Unix(),
	})

	/*
		err = DB(ctx).Create(&AlertingEngines{
			Instance:     instance,
			DatasourceId: datasourceId,
			Clock:        time.Now().Unix(),
		}).Error
	*/
	return err
}

func AlertingEngineDel(ctx *ctx.Context, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	finder := zorm.NewDeleteFinder(AlertingEnginesTableName).Append("WHERE id in (?)", ids)
	return UpdateFinder(ctx, finder)

	//return DB(ctx).Where("id in ?", ids).Delete(new(AlertingEngines)).Error
}

func AlertingEngineGetDatasourceIds(ctx *ctx.Context, instance string) ([]int64, error) {
	ids := make([]int64, 0)
	finder := zorm.NewSelectFinder(AlertingEnginesTableName, "datasource_id").Append("WHERE instance=?", instance)
	err := zorm.Query(ctx.Ctx, finder, &ids, nil)
	return ids, err
	/*
		err := DB(ctx).Where("instance=?", instance).Find(&objs).Error
		if err != nil {
			return []int64{}, err
		}

		if len(objs) == 0 {
			return []int64{}, nil
		}
		var ids []int64
		for i := 0; i < len(objs); i++ {
			ids = append(ids, objs[i].DatasourceId)
		}

		return ids, nil
	*/
}

// AlertingEngineGets 拉取列表数据，用户要在页面上看到所有 n9e-server 实例列表，然后为其分配 cluster
func AlertingEngineGets(ctx *ctx.Context, where string, args ...interface{}) ([]*AlertingEngines, error) {
	objs := make([]*AlertingEngines, 0)
	finder := zorm.NewSelectFinder(AlertingEnginesTableName)
	AppendWhere(finder, where, args...)
	finder.Append("order by instance asc")
	err := zorm.Query(ctx.Ctx, finder, &objs, nil)
	/*
		session := DB(ctx).Order("instance")
		if where == "" {
			err = session.Find(&objs).Error
		} else {
			err = session.Where(where, args...).Find(&objs).Error
		}
	*/
	return objs, err
}

func AlertingEngineGet(ctx *ctx.Context, where string, args ...interface{}) (*AlertingEngines, error) {
	lst, err := AlertingEngineGets(ctx, where, args...)
	if err != nil {
		return nil, err
	}

	if len(lst) == 0 {
		return nil, nil
	}

	return lst[0], nil
}

func AlertingEngineGetsClusters(ctx *ctx.Context, where string, args ...interface{}) ([]string, error) {
	arr := make([]string, 0)
	finder := zorm.NewSelectFinder(AlertingEnginesTableName, "DISTINCT engine_cluster").Append("WHERE engine_cluster != ?", "")
	if where != "" {
		finder.Append("and "+where, args...)
	}
	finder.Append("order by engine_cluster asc")
	err := zorm.Query(ctx.Ctx, finder, &arr, nil)
	/*
		session := DB(ctx).Model(new(AlertingEngines)).Where("engine_cluster != ''").Order("engine_cluster").Distinct("engine_cluster")
		if where == "" {
			err = session.Pluck("engine_cluster", &arr).Error
		} else {
			err = session.Where(where, args...).Pluck("engine_cluster", &arr).Error
		}
	*/
	return arr, err
}

func AlertingEngineGetsInstances(ctx *ctx.Context, where string, args ...interface{}) ([]string, error) {
	arr := make([]string, 0)
	finder := zorm.NewSelectFinder(AlertingEnginesTableName, "instance")
	AppendWhere(finder, where, args...)
	finder.Append("order by instance asc")
	err := zorm.Query(ctx.Ctx, finder, &arr, nil)
	/*
		session := DB(ctx).Model(new(AlertingEngines)).Order("instance")
		if where == "" {
			err = session.Pluck("instance", &arr).Error
		} else {
			err = session.Where(where, args...).Pluck("instance", &arr).Error
		}
	*/
	return arr, err
}

type HeartbeatInfo struct {
	Instance      string `json:"instance"`
	EngineCluster string `json:"engine_cluster"`
	DatasourceId  int64  `json:"datasource_id"`
}

func AlertingEngineHeartbeatWithCluster(ctx *ctx.Context, instance, cluster string, datasourceId int64) error {
	if !ctx.IsCenter {
		info := HeartbeatInfo{
			Instance:      instance,
			EngineCluster: cluster,
			DatasourceId:  datasourceId,
		}
		err := poster.PostByUrls(ctx, "/v1/n9e/server-heartbeat", info)
		return err
	}

	finder := zorm.NewSelectFinder(AlertingEnginesTableName, "count(*)").Append("WHERE instance=? and engine_cluster = ? and datasource_id=?", instance, cluster, datasourceId)
	total, err := Count(ctx, finder)
	//var total int64
	//err := DB(ctx).Model(new(AlertingEngines)).Where("instance=? and engine_cluster = ? and datasource_id=?", instance, cluster, datasourceId).Count(&total).Error
	if err != nil {
		return err
	}

	if total == 0 {
		err = Insert(ctx, &AlertingEngines{
			Instance:      instance,
			DatasourceId:  datasourceId,
			EngineCluster: cluster,
			Clock:         time.Now().Unix(),
		})
		// insert
		/*
			err = DB(ctx).Create(&AlertingEngines{
				Instance:      instance,
				DatasourceId:  datasourceId,
				EngineCluster: cluster,
				Clock:         time.Now().Unix(),
			}).Error
		*/
	} else {
		// updates
		finder := zorm.NewUpdateFinder(AlertingEnginesTableName).Append("clock=? WHERE instance=? and engine_cluster = ? and datasource_id=?", time.Now().Unix(), instance, cluster, datasourceId)
		err = UpdateFinder(ctx, finder)
		//fields := map[string]interface{}{"clock": time.Now().Unix()}
		//err = DB(ctx).Model(new(AlertingEngines)).Where("instance=? and engine_cluster = ? and datasource_id=?", instance, cluster, datasourceId).Updates(fields).Error
	}

	return err
}
