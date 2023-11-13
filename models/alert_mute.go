package models

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"gitee.com/chunanyong/zorm"
	"github.com/ccfos/nightingale/v6/pkg/ctx"
	"github.com/ccfos/nightingale/v6/pkg/ormx"
	"github.com/ccfos/nightingale/v6/pkg/poster"
	"github.com/toolkits/pkg/logger"

	"errors"
)

type TagFilter struct {
	Key    string              `json:"key"`   // tag key
	Func   string              `json:"func"`  // `==` | `=~` | `in` | `!=` | `!~` | `not in`
	Value  string              `json:"value"` // tag value
	Regexp *regexp.Regexp      // parse value to regexp if func = '=~' or '!~'
	Vset   map[string]struct{} // parse value to regexp if func = 'in' or 'not in'
}

func GetTagFilters(jsonArr ormx.JSONArr) ([]TagFilter, error) {
	if jsonArr == nil || len([]byte(jsonArr)) == 0 {
		return []TagFilter{}, nil
	}

	bFilters := make([]TagFilter, 0)
	err := json.Unmarshal(jsonArr, &bFilters)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(bFilters); i++ {
		if bFilters[i].Func == "=~" || bFilters[i].Func == "!~" {
			bFilters[i].Regexp, err = regexp.Compile(bFilters[i].Value)
			if err != nil {
				return nil, err
			}
		} else if bFilters[i].Func == "in" || bFilters[i].Func == "not in" {
			arr := strings.Fields(bFilters[i].Value)
			bFilters[i].Vset = make(map[string]struct{})
			for j := 0; j < len(arr); j++ {
				bFilters[i].Vset[arr[j]] = struct{}{}
			}
		}
	}

	return bFilters, nil
}

const TimeRange int = 0
const Periodic int = 1
const AlertMuteTableName = "alert_mute"

type AlertMute struct {
	zorm.EntityStruct
	Id                int64          `json:"id" column:"id"`
	GroupId           int64          `json:"group_id" column:"group_id"`
	Note              string         `json:"note" column:"note"`
	Cate              string         `json:"cate" column:"cate"`
	Prod              string         `json:"prod" column:"prod"`
	DatasourceIds     string         `json:"-" column:"datasource_ids"`     // datasource ids
	DatasourceIdsJson []int64        `json:"datasource_ids"`                // for fe
	Cluster           string         `json:"cluster" column:"cluster_name"` // take effect by clusters, seperated by space
	Tags              ormx.JSONArr   `json:"tags" column:"tags"`
	Cause             string         `json:"cause" column:"cause"`
	Btime             int64          `json:"btime" column:"btime"`
	Etime             int64          `json:"etime" column:"etime"`
	Disabled          int            `json:"disabled" column:"disabled"` // 0: enabled, 1: disabled
	CreateBy          string         `json:"create_by" column:"create_by"`
	UpdateBy          string         `json:"update_by" column:"update_by"`
	CreateAt          int64          `json:"create_at" column:"create_at"`
	UpdateAt          int64          `json:"update_at" column:"update_at"`
	ITags             []TagFilter    `json:"-"`                                      // inner tags
	MuteTimeType      int            `json:"mute_time_type" column:"mute_time_type"` //  0: mute by time range, 1: mute by periodic time
	PeriodicMutes     string         `json:"-" column:"periodic_mutes"`
	PeriodicMutesJson []PeriodicMute `json:"periodic_mutes"`
	Severities        string         `json:"-" column:"severities"`
	SeveritiesJson    []int          `json:"severities"`
}

type PeriodicMute struct {
	EnableStime      string `json:"enable_stime"`        // split by space: "00:00 10:00 12:00"
	EnableEtime      string `json:"enable_etime"`        // split by space: "00:00 10:00 12:00"
	EnableDaysOfWeek string `json:"enable_days_of_week"` // eg: "0 1 2 3 4 5 6"
}

func (m *AlertMute) GetTableName() string {
	return AlertMuteTableName
}

func AlertMuteGetById(ctx *ctx.Context, id int64) (*AlertMute, error) {
	return AlertMuteGet(ctx, "id=?", id)
}

func AlertMuteGet(ctx *ctx.Context, where string, args ...interface{}) (*AlertMute, error) {
	lst := make([]AlertMute, 0)
	finder := zorm.NewSelectFinder(AlertMuteTableName)
	AppendWhere(finder, where, args...)
	err := zorm.Query(ctx.Ctx, finder, &lst, nil)
	//err := DB(ctx).Where(where, args...).Find(&lst).Error
	if err != nil {
		return nil, err
	}

	if len(lst) == 0 {
		return nil, nil
	}
	err = lst[0].DB2FE()
	return &lst[0], err
}

func AlertMuteGets(ctx *ctx.Context, prods []string, bgid int64, query string) ([]AlertMute, error) {
	lst := make([]AlertMute, 0)

	//session := DB(ctx)
	finder := zorm.NewSelectFinder(AlertMuteTableName).Append("WHERE 1=1")
	if bgid != -1 {
		//session = session.Where("group_id = ?", bgid)
		finder.Append("and group_id = ?", bgid)
	}

	if len(prods) > 0 {
		//session = session.Where("prod in (?)", prods)
		finder.Append("and prod in (?)", prods)
	}

	if query != "" {
		arr := strings.Fields(query)
		for i := 0; i < len(arr); i++ {
			qarg := "%" + arr[i] + "%"
			//session = session.Where("cause like ?", qarg)
			finder.Append("and cause like ?", qarg)
		}
	}
	finder.Append("order by id desc")
	err := zorm.Query(ctx.Ctx, finder, &lst, nil)
	//err = session.Order("id desc").Find(&lst).Error
	for i := 0; i < len(lst); i++ {
		lst[i].DB2FE()
	}
	return lst, err
}

func AlertMuteGetsByBG(ctx *ctx.Context, groupId int64) ([]AlertMute, error) {
	lst := make([]AlertMute, 0)
	finder := zorm.NewSelectFinder(AlertMuteTableName).Append("WHERE group_id=? order by id desc", groupId)
	err := zorm.Query(ctx.Ctx, finder, &lst, nil)
	//err = DB(ctx).Where("group_id=?", groupId).Order("id desc").Find(&lst).Error
	for i := 0; i < len(lst); i++ {
		lst[i].DB2FE()
	}
	return lst, err
}

func (m *AlertMute) Verify() error {
	if m.GroupId < 0 {
		return errors.New("group_id invalid")
	}

	if IsAllDatasource(m.DatasourceIdsJson) {
		m.DatasourceIdsJson = []int64{0}
	}

	if m.Etime <= m.Btime {
		return fmt.Errorf("oops... etime(%d) <= btime(%d)", m.Etime, m.Btime)
	}

	if err := m.Parse(); err != nil {
		return err
	}

	if len(m.ITags) == 0 {
		return errors.New("tags is blank")
	}

	return nil
}

func (m *AlertMute) Parse() error {
	var err error
	m.ITags, err = GetTagFilters(m.Tags)
	if err != nil {
		return err
	}

	return nil
}

func (m *AlertMute) Add(ctx *ctx.Context) error {
	if err := m.Verify(); err != nil {
		return err
	}

	if err := m.FE2DB(); err != nil {
		return err
	}

	now := time.Now().Unix()
	m.CreateAt = now
	m.UpdateAt = now
	return Insert(ctx, m)
}

func (m *AlertMute) Update(ctx *ctx.Context, arm AlertMute) error {

	arm.Id = m.Id
	arm.GroupId = m.GroupId
	arm.CreateAt = m.CreateAt
	arm.CreateBy = m.CreateBy
	arm.UpdateAt = time.Now().Unix()

	err := arm.Verify()
	if err != nil {
		return err
	}

	if err := arm.FE2DB(); err != nil {
		return err
	}
	_, err = zorm.Transaction(ctx.Ctx, func(ctx context.Context) (interface{}, error) {
		return zorm.UpdateNotZeroValue(ctx, &arm)
	})
	return err
	//return DB(ctx).Model(m).Select("*").Updates(arm).Error
}

func (m *AlertMute) FE2DB() error {
	idsBytes, err := json.Marshal(m.DatasourceIdsJson)
	if err != nil {
		return err
	}
	m.DatasourceIds = string(idsBytes)

	periodicMutesBytes, err := json.Marshal(m.PeriodicMutesJson)
	if err != nil {
		return err
	}
	m.PeriodicMutes = string(periodicMutesBytes)

	if len(m.SeveritiesJson) > 0 {
		severtiesBytes, err := json.Marshal(m.SeveritiesJson)
		if err != nil {
			return err
		}
		m.Severities = string(severtiesBytes)
	}

	return nil
}

func (m *AlertMute) DB2FE() error {
	err := json.Unmarshal([]byte(m.DatasourceIds), &m.DatasourceIdsJson)
	if err != nil {
		return err
	}
	err = json.Unmarshal([]byte(m.PeriodicMutes), &m.PeriodicMutesJson)
	if err != nil {
		return err
	}

	if m.Severities != "" {
		err = json.Unmarshal([]byte(m.Severities), &m.SeveritiesJson)
		if err != nil {
			return err
		}
	}

	return err
}

func (m *AlertMute) UpdateFieldsMap(ctx *ctx.Context, fields map[string]interface{}) error {
	return UpdateFieldsMap(ctx, m, m.Id, fields)
	//return DB(ctx).Model(m).Updates(fields).Error
}

func AlertMuteDel(ctx *ctx.Context, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	finder := zorm.NewDeleteFinder(AlertMuteTableName).Append("WHERE id in (?)", ids)
	return UpdateFinder(ctx, finder)
	//return DB(ctx).Where("id in ?", ids).Delete(new(AlertMute)).Error
}

func AlertMuteStatistics(ctx *ctx.Context) (*Statistics, error) {
	//stats := make([]Statistics, 0)
	if !ctx.IsCenter {
		s, err := poster.GetByUrls[*Statistics](ctx, "/v1/n9e/statistic?name=alert_mute")
		return s, err
	}

	// clean expired first
	buf := int64(30)
	finder := zorm.NewDeleteFinder(AlertMuteTableName).Append("WHERE etime < ? and mute_time_type = 0", time.Now().Unix()-buf)
	err := UpdateFinder(ctx, finder)
	//err := DB(ctx).Where("etime < ? and mute_time_type = 0", time.Now().Unix()-buf).Delete(new(AlertMute)).Error
	if err != nil {
		return nil, err
	}
	return StatisticsGet(ctx, AlertMuteTableName)
	/*
			session := DB(ctx).Model(&AlertMute{}).Select("count(*) as Total", "max(update_at) as last_updated")
			err = session.Find(&stats).Error
			if err != nil {
				return nil, err
			}

		return stats[0], nil
	*/
}

func AlertMuteGetsAll(ctx *ctx.Context) ([]*AlertMute, error) {
	// get my cluster's mutes
	lst := make([]*AlertMute, 0)
	if !ctx.IsCenter {
		lst, err := poster.GetByUrls[[]*AlertMute](ctx, "/v1/n9e/alert-mutes")
		if err != nil {
			return nil, err
		}
		for i := 0; i < len(lst); i++ {
			lst[i].FE2DB()
		}
		return lst, err
	}

	finder := zorm.NewSelectFinder(AlertMuteTableName)
	//session := DB(ctx).Model(&AlertMute{})
	err := zorm.Query(ctx.Ctx, finder, &lst, nil)
	//err := session.Find(&lst).Error
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(lst); i++ {
		lst[i].DB2FE()
	}

	return lst, err
}

func AlertMuteUpgradeToV6(ctx *ctx.Context, dsm map[string]Datasource) error {
	lst := make([]AlertMute, 0)
	finder := zorm.NewSelectFinder(AlertMuteTableName)
	err := zorm.Query(ctx.Ctx, finder, &lst, nil)
	//err := DB(ctx).Find(&lst).Error
	if err != nil {
		return err
	}

	for i := 0; i < len(lst); i++ {
		var ids []int64
		if lst[i].Cluster == "$all" {
			ids = append(ids, 0)
		} else {
			clusters := strings.Fields(lst[i].Cluster)
			for j := 0; j < len(clusters); j++ {
				if ds, exists := dsm[clusters[j]]; exists {
					ids = append(ids, ds.Id)
				}
			}
		}

		b, err := json.Marshal(ids)
		if err != nil {
			continue
		}
		lst[i].DatasourceIds = string(b)

		if lst[i].Prod == "" {
			lst[i].Prod = METRIC
		}

		if lst[i].Cate == "" {
			lst[i].Cate = PROMETHEUS
		}

		err = lst[i].UpdateFieldsMap(ctx, map[string]interface{}{
			"datasource_ids": lst[i].DatasourceIds,
			"prod":           lst[i].Prod,
			"cate":           lst[i].Cate,
		})
		if err != nil {
			logger.Errorf("update alert rule:%d datasource ids failed, %v", lst[i].Id, err)
		}
	}
	return nil
}
