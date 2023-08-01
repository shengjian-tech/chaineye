package models

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gitee.com/chunanyong/zorm"
	"github.com/ccfos/nightingale/v6/pkg/ctx"
	"github.com/toolkits/pkg/logger"
)

const AlertHisEventTableName = "alert_his_event"

type AlertHisEvent struct {
	zorm.EntityStruct
	Id                 int64             `json:"id" column:"id"`
	Cate               string            `json:"cate" column:"cate"`
	IsRecovered        int               `json:"is_recovered" column:"is_recovered"`
	DatasourceId       int64             `json:"datasource_id" column:"datasource_id"`
	Cluster            string            `json:"cluster" column:"cluster"`
	GroupId            int64             `json:"group_id" column:"group_id"`
	GroupName          string            `json:"group_name" column:"group_name"` // busi group name
	Hash               string            `json:"hash" column:"hash"`
	RuleId             int64             `json:"rule_id" column:"rule_id"`
	RuleName           string            `json:"rule_name" column:"rule_name"`
	RuleNote           string            `json:"rule_note" column:"rule_note"`
	RuleProd           string            `json:"rule_prod" column:"rule_prod"`
	RuleAlgo           string            `json:"rule_algo" column:"rule_algo"`
	Severity           int               `json:"severity" column:"severity"`
	PromForDuration    int               `json:"prom_for_duration" column:"prom_for_duration"`
	PromQl             string            `json:"prom_ql" column:"prom_ql"`
	RuleConfig         string            `json:"-" column:"rule_config"` // rule config
	RuleConfigJson     interface{}       `json:"rule_config"`            // rule config for fe
	PromEvalInterval   int               `json:"prom_eval_interval" column:"prom_eval_interval"`
	Callbacks          string            `json:"-" column:"callbacks"`
	CallbacksJSON      []string          `json:"callbacks"`
	RunbookUrl         string            `json:"runbook_url" column:"runbook_url"`
	NotifyRecovered    int               `json:"notify_recovered" column:"notify_recovered"`
	NotifyChannels     string            `json:"-" column:"notify_channels"`
	NotifyChannelsJSON []string          `json:"notify_channels"`
	NotifyGroups       string            `json:"-" column:"notify_groups"`
	NotifyGroupsJSON   []string          `json:"notify_groups"`
	NotifyGroupsObj    []UserGroup       `json:"notify_groups_obj"`
	TargetIdent        string            `json:"target_ident" column:"target_ident"`
	TargetNote         string            `json:"target_note" column:"target_note"`
	TriggerTime        int64             `json:"trigger_time" column:"trigger_time"`
	TriggerValue       string            `json:"trigger_value" column:"trigger_value"`
	RecoverTime        int64             `json:"recover_time" column:"recover_time"`
	LastEvalTime       int64             `json:"last_eval_time" column:"last_eval_time"`
	Tags               string            `json:"-" column:"tags"`
	TagsJSON           []string          `json:"tags"`
	Annotations        string            `json:"-" column:"annotations"`
	AnnotationsJSON    map[string]string `json:"annotations"`                                    // for fe
	NotifyCurNumber    int               `json:"notify_cur_number" column:"notify_cur_number"`   // notify: current number
	FirstTriggerTime   int64             `json:"first_trigger_time" column:"first_trigger_time"` // 连续告警的首次告警时间
	ExtraConfig        interface{}       `json:"extra_config"`
}

func (e *AlertHisEvent) GetTableName() string {
	return AlertHisEventTableName
}

func (e *AlertHisEvent) Add(ctx *ctx.Context) error {
	return Insert(ctx, e)
}

func (e *AlertHisEvent) DB2FE() {
	e.NotifyChannelsJSON = strings.Fields(e.NotifyChannels)
	e.NotifyGroupsJSON = strings.Fields(e.NotifyGroups)
	e.CallbacksJSON = strings.Fields(e.Callbacks)
	e.TagsJSON = strings.Split(e.Tags, ",,")
	json.Unmarshal([]byte(e.Annotations), &e.AnnotationsJSON)
	json.Unmarshal([]byte(e.RuleConfig), &e.RuleConfigJson)
}

func (e *AlertHisEvent) FillNotifyGroups(ctx *ctx.Context, cache map[int64]*UserGroup) error {
	// some user-group already deleted ?
	count := len(e.NotifyGroupsJSON)
	if count == 0 {
		e.NotifyGroupsObj = []UserGroup{}
		return nil
	}

	for i := range e.NotifyGroupsJSON {
		id, err := strconv.ParseInt(e.NotifyGroupsJSON[i], 10, 64)
		if err != nil {
			continue
		}

		ug, has := cache[id]
		if has {
			e.NotifyGroupsObj = append(e.NotifyGroupsObj, *ug)
			continue
		}

		ug, err = UserGroupGetById(ctx, id)
		if err != nil {
			return err
		}

		if ug != nil {
			e.NotifyGroupsObj = append(e.NotifyGroupsObj, *ug)
			cache[id] = ug
		}
	}

	return nil
}

func AlertHisEventTotal(ctx *ctx.Context, prods []string, bgid, stime, etime int64, severity int, recovered int, dsIds []int64, cates []string, query string) (int64, error) {
	finder := zorm.NewSelectFinder(AlertHisEventTableName, "count(*)")
	finder.Append("WHERE last_eval_time between ? and ?", stime, etime)
	//session := DB(ctx).Model(&AlertHisEvent{}).Where("last_eval_time between ? and ?", stime, etime)

	if len(prods) > 0 {
		//session = session.Where("rule_prod in ?", prods)
		finder.Append("and rule_prod in (?)", prods)
	}

	if bgid > 0 {
		//session = session.Where("group_id = ?", bgid)
		finder.Append("and group_id = ?", bgid)
	}

	if severity >= 0 {
		//session = session.Where("severity = ?", severity)
		finder.Append("and severity = ?", severity)
	}

	if recovered >= 0 {
		//session = session.Where("is_recovered = ?", recovered)
		finder.Append("and is_recovered = ?", recovered)
	}

	if len(dsIds) > 0 {
		//session = session.Where("datasource_id in ?", dsIds)
		finder.Append("and datasource_id in (?)", dsIds)
	}

	if len(cates) > 0 {
		//session = session.Where("cate in ?", cates)
		finder.Append("and cate in (?)", cates)
	}

	if query != "" {
		arr := strings.Fields(query)
		for i := 0; i < len(arr); i++ {
			qarg := "%" + arr[i] + "%"
			//session = session.Where("rule_name like ? or tags like ?", qarg, qarg)
			finder.Append("and (rule_name like ? or tags like ?)", qarg, qarg)
		}
	}
	return Count(ctx, finder)
	//return Count(session)
}

func AlertHisEventGets(ctx *ctx.Context, prods []string, bgid, stime, etime int64, severity int, recovered int, dsIds []int64, cates []string, query string, limit, offset int) ([]AlertHisEvent, error) {
	finder := zorm.NewSelectFinder(AlertHisEventTableName)
	finder.Append("WHERE last_eval_time between ? and ?", stime, etime)
	//session := DB(ctx).Where("last_eval_time between ? and ?", stime, etime)

	if len(prods) != 0 {
		//session = session.Where("rule_prod in ?", prods)
		finder.Append("and rule_prod in (?)", prods)
	}

	if bgid > 0 {
		//session = session.Where("group_id = ?", bgid)
		finder.Append("and group_id = ?", bgid)
	}

	if severity >= 0 {
		//session = session.Where("severity = ?", severity)
		finder.Append("and severity = ?", severity)
	}

	if recovered >= 0 {
		//session = session.Where("is_recovered = ?", recovered)
		finder.Append("and is_recovered = ?", recovered)
	}

	if len(dsIds) > 0 {
		//session = session.Where("datasource_id in ?", dsIds)
		finder.Append("and datasource_id in (?)", dsIds)
	}

	if len(cates) > 0 {
		//session = session.Where("cate in ?", cates)
		finder.Append("and cate in (?)", cates)
	}

	if query != "" {
		arr := strings.Fields(query)
		for i := 0; i < len(arr); i++ {
			qarg := "%" + arr[i] + "%"
			//session = session.Where("rule_name like ? or tags like ?", qarg, qarg)
			finder.Append("and (rule_name like ? or tags like ?)", qarg, qarg)
		}
	}
	finder.Append("order by id desc")

	lst := make([]AlertHisEvent, 0)
	page := zorm.NewPage()
	page.PageSize = limit
	page.PageNo = offset / limit
	finder.SelectTotalCount = false
	err := zorm.Query(ctx.Ctx, finder, &lst, page)
	//err := session.Order("id desc").Limit(limit).Offset(offset).Find(&lst).Error

	if err == nil {
		for i := 0; i < len(lst); i++ {
			lst[i].DB2FE()
		}
	}

	return lst, err
}

func AlertHisEventGet(ctx *ctx.Context, where string, args ...interface{}) (*AlertHisEvent, error) {
	lst := make([]AlertHisEvent, 0)
	finder := zorm.NewSelectFinder(AlertHisEventTableName)
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
	lst[0].FillNotifyGroups(ctx, make(map[int64]*UserGroup))

	return &lst[0], nil
}

func AlertHisEventGetById(ctx *ctx.Context, id int64) (*AlertHisEvent, error) {
	return AlertHisEventGet(ctx, "id=?", id)
}

func (m *AlertHisEvent) UpdateFieldsMap(ctx *ctx.Context, fields map[string]interface{}) error {
	return UpdateFieldsMap(ctx, m, m.Id, fields)
	//return DB(ctx).Model(m).Updates(fields).Error
}

func AlertHisEventUpgradeToV6(ctx *ctx.Context, dsm map[string]Datasource) error {
	lst := make([]AlertHisEvent, 0)
	finder := zorm.NewSelectFinder(AlertHisEventTableName).Append("WHERE trigger_time > ? order by id desc", time.Now().Unix()-3600*24*30)
	page := zorm.NewPage()
	page.PageSize = 10000
	finder.SelectTotalCount = false
	err := zorm.Query(ctx.Ctx, finder, &lst, page)
	//err := DB(ctx).Where("trigger_time > ?", time.Now().Unix()-3600*24*30).Limit(10000).Order("id desc").Find(&lst).Error
	if err != nil {
		return err
	}

	for i := 0; i < len(lst); i++ {
		ds, exists := dsm[lst[i].Cluster]
		if !exists {
			continue
		}
		lst[i].DatasourceId = ds.Id

		ruleConfig := PromRuleConfig{
			Queries: []PromQuery{
				{
					PromQl:   lst[i].PromQl,
					Severity: lst[i].Severity,
				},
			},
		}
		b, _ := json.Marshal(ruleConfig)
		lst[i].RuleConfig = string(b)

		if lst[i].RuleProd == "" {
			lst[i].RuleProd = METRIC
		}

		if lst[i].Cate == "" {
			lst[i].Cate = PROMETHEUS
		}

		err = lst[i].UpdateFieldsMap(ctx, map[string]interface{}{
			"datasource_id": lst[i].DatasourceId,
			"rule_config":   lst[i].RuleConfig,
			"rule_prod":     lst[i].RuleProd,
			"cate":          lst[i].Cate,
		})
		if err != nil {
			logger.Errorf("update alert rule:%d datasource ids failed, %v", lst[i].Id, err)
		}
	}
	return nil
}

func EventPersist(ctx *ctx.Context, event *AlertCurEvent) error {
	has, err := AlertCurEventExists(ctx, "hash=?", event.Hash)
	if err != nil {
		return fmt.Errorf("event_persist_check_exists_fail: %v rule_id=%d hash=%s", err, event.RuleId, event.Hash)
	}

	his := event.ToHis(ctx)

	// 不管是告警还是恢复，全量告警里都要记录
	if err := his.Add(ctx); err != nil {
		return fmt.Errorf("add his event error:%v", err)
	}

	if has {
		// 活跃告警表中有记录，删之
		err = AlertCurEventDelByHash(ctx, event.Hash)
		if err != nil {
			return fmt.Errorf("event_del_cur_fail: %v hash=%s", err, event.Hash)
		}

		if !event.IsRecovered {
			// 恢复事件，从活跃告警列表彻底删掉，告警事件，要重新加进来新的event
			// use his id as cur id
			event.Id = his.Id
			if event.Id > 0 {
				if err := event.Add(ctx); err != nil {
					return fmt.Errorf("add cur event err:%v", err)
				}
			}
		}

		return nil
	}

	if event.IsRecovered {
		// alert_cur_event表里没有数据，表示之前没告警，结果现在报了恢复，神奇....理论上不应该出现的
		return nil
	}

	// use his id as cur id
	event.Id = his.Id
	if event.Id > 0 {
		if err := event.Add(ctx); err != nil {
			return fmt.Errorf("add cur event error:%v", err)
		}
	}

	return nil
}
