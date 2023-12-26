package models

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"gitee.com/chunanyong/zorm"
	"github.com/ccfos/nightingale/v6/pkg/ctx"
	"github.com/ccfos/nightingale/v6/pkg/poster"
)

const TargetTableName = "target"

type Target struct {
	// 引入默认的struct,隔离IEntityStruct的方法改动
	zorm.EntityStruct
	Id           int64             `json:"id" column:"id"`
	GroupId      int64             `json:"group_id" column:"group_id"`
	GroupObj     *BusiGroup        `json:"group_obj"`
	Ident        string            `json:"ident" column:"ident"`
	Note         string            `json:"note" column:"note"`
	Tags         string            `json:"-" column:"tags"`
	TagsJSON     []string          `json:"tags"`
	TagsMap      map[string]string `json:"tags_maps"` // internal use, append tags to series
	UpdateAt     int64             `json:"update_at" column:"update_at"`
	HostIp       string            `json:"host_ip" column:"host_ip"` //ipv4，do not needs range select
	AgentVersion string            `json:"agent_version" column:"agent_version"`
	UnixTime     int64             `json:"unixtime"`
	Offset       int64             `json:"offset"`
	TargetUp     float64           `json:"target_up"`
	MemUtil      float64           `json:"mem_util"`
	CpuNum       int               `json:"cpu_num"`
	CpuUtil      float64           `json:"cpu_util"`
	OS           string            `json:"os"`
	Arch         string            `json:"arch"`
	RemoteAddr   string            `json:"remote_addr"`
}

func (t *Target) GetTableName() string {
	return TargetTableName
}

func (t *Target) DB2FE() error {
	return nil
}

func (t *Target) FillGroup(ctx *ctx.Context, cache map[int64]*BusiGroup) error {
	if t.GroupId <= 0 {
		return nil
	}

	bg, has := cache[t.GroupId]
	if has {
		t.GroupObj = bg
		return nil
	}

	bg, err := BusiGroupGetById(ctx, t.GroupId)
	if err != nil {
		return fmt.Errorf("failed to get busi group:%w", err)
	}

	t.GroupObj = bg
	cache[t.GroupId] = bg
	return nil
}

func TargetStatistics(ctx *ctx.Context) (*Statistics, error) {
	if !ctx.IsCenter {
		s, err := poster.GetByUrls[*Statistics](ctx, "/v1/n9e/statistic?name=target")
		return s, err
	}

	return StatisticsGet(ctx, TargetTableName)

	/*
		var stats []*Statistics
		err := DB(ctx).Model(&Target{}).Select("count(*) as total", "max(update_at) as last_updated").Find(&stats).Error
		if err != nil {
			return nil, err
		}

		return stats[0], nil
	*/
}

func TargetDel(ctx *ctx.Context, idents []string) error {
	if len(idents) == 0 {
		panic("idents empty")
	}
	finder := zorm.NewDeleteFinder(TargetTableName).Append("WHERE ident in (?)", idents)
	return UpdateFinder(ctx, finder)
	//return DB(ctx).Where("ident in ?", idents).Delete(new(Target)).Error
}

func buildTargetWhere(ctx *ctx.Context, selectField string, bgids []int64, dsIds []int64, query string, downtime int64) *zorm.Finder {
	finder := zorm.NewSelectFinder(TargetTableName, selectField).Append("WHERE 1=1")
	//session := DB(ctx).Model(&Target{})
	finder.SelectTotalCount = false
	if len(bgids) > 0 {
		//session = session.Where("group_id in (?)", bgids)
		finder.Append("and group_id in (?)", bgids)
	}

	if len(dsIds) > 0 {
		//session = session.Where("datasource_id in (?)", dsIds)
		finder.Append("and datasource_id in (?)", dsIds)
	}

	if downtime > 0 {
		//session = session.Where("update_at < ?", time.Now().Unix()-downtime)
		finder.Append("and update_at < ?", time.Now().Unix()-downtime)
	}

	if query != "" {
		arr := strings.Fields(query)
		for i := 0; i < len(arr); i++ {
			q := "%" + arr[i] + "%"
			//session = session.Where("ident like ? or note like ? or tags like ?", q, q, q)
			finder.Append("and (ident like ? or note like ? or tags like ?)", q, q, q)
		}
	}

	return finder
}

func TargetTotalCount(ctx *ctx.Context) (int64, error) {
	finder := zorm.NewSelectFinder(TargetTableName, "count(*)")
	return Count(ctx, finder)
	//return Count(DB(ctx).Model(new(Target)))
}

func TargetTotal(ctx *ctx.Context, bgids []int64, dsIds []int64, query string, downtime int64) (int64, error) {
	finder := buildTargetWhere(ctx, "count(*)", bgids, dsIds, query, downtime)
	return Count(ctx, finder)
	//return Count(buildTargetWhere(ctx, bgids, dsIds, query, downtime))
}

func TargetGets(ctx *ctx.Context, bgids []int64, dsIds []int64, query string, downtime int64, limit, offset int) ([]*Target, error) {
	lst := make([]*Target, 0)
	finder := buildTargetWhere(ctx, "*", bgids, dsIds, query, downtime)
	finder.Append("order by ident asc ")
	page := zorm.NewPage()
	page.PageSize = limit
	page.PageNo = offset / limit
	err := zorm.Query(ctx.Ctx, finder, &lst, page)
	//err := buildTargetWhere(ctx, bgids, dsIds, query, downtime).Order("ident").Limit(limit).Offset(offset).Find(&lst).Error
	if err == nil {
		for i := 0; i < len(lst); i++ {
			lst[i].TagsJSON = strings.Fields(lst[i].Tags)
		}
	}
	return lst, err
}

// 根据 groupids, tags, hosts 查询 targets
func TargetGetsByFilter(ctx *ctx.Context, query []map[string]interface{}, limit, offset int) ([]*Target, error) {
	lst := make([]*Target, 0)
	finder, page := TargetFilterQueryBuild(ctx, "*", query, limit, offset)
	finder.Append("order by ident asc ")
	err := zorm.Query(ctx.Ctx, finder, &lst, page)
	//err := session.Order("ident").Find(&lst).Error
	cache := make(map[int64]*BusiGroup)
	for i := 0; i < len(lst); i++ {
		lst[i].TagsJSON = strings.Fields(lst[i].Tags)
		lst[i].FillGroup(ctx, cache)
	}

	return lst, err
}

func TargetCountByFilter(ctx *ctx.Context, query []map[string]interface{}) (int64, error) {
	finder, _ := TargetFilterQueryBuild(ctx, "count(*)", query, 0, 0)
	return Count(ctx, finder)
	//return Count(session)
}

func MissTargetGetsByFilter(ctx *ctx.Context, query []map[string]interface{}, ts int64) ([]*Target, error) {
	lst := make([]*Target, 0)
	finder, page := TargetFilterQueryBuild(ctx, "*", query, 0, 0)
	finder.Append("and update_at < ?", ts)
	//session = session.Where("update_at < ?", ts)
	finder.Append("order by ident asc")
	err := zorm.Query(ctx.Ctx, finder, &lst, page)
	//err := session.Order("ident").Find(&lst).Error
	return lst, err
}

func MissTargetCountByFilter(ctx *ctx.Context, query []map[string]interface{}, ts int64) (int64, error) {
	finder, _ := TargetFilterQueryBuild(ctx, "count(*)", query, 0, 0)
	finder.Append("and update_at < ?", ts)
	return Count(ctx, finder)
	//session = session.Where("update_at < ?", ts)
	//return Count(session)
}

func TargetFilterQueryBuild(ctx *ctx.Context, selectField string, query []map[string]interface{}, limit, offset int) (*zorm.Finder, *zorm.Page) {
	finder := zorm.NewSelectFinder(TargetTableName, selectField).Append("WHERE 1=1")
	finder.SelectTotalCount = false
	if len(query) > 0 {
		finder.Append("and (1=1 ")
	}
	//session := DB(ctx).Model(&Target{})
	for _, q := range query {
		//tx := DB(ctx).Model(&Target{})
		for k, v := range q {
			//tx = tx.Or(k, v)
			finder.Append("or "+k+"=?", v)
		}
		//session = session.Where(tx)
	}
	if len(query) > 0 {
		finder.Append(")")
	}
	var page *zorm.Page

	if limit > 0 {
		page = zorm.NewPage()
		page.PageSize = limit
		page.PageNo = offset / limit
		//session = session.Limit(limit).Offset(offset)
	}

	return finder, page
}

func TargetGetsAll(ctx *ctx.Context) ([]*Target, error) {
	if !ctx.IsCenter {
		lst, err := poster.GetByUrls[[]*Target](ctx, "/v1/n9e/targets")
		return lst, err
	}

	lst := make([]*Target, 0)
	finder := zorm.NewSelectFinder(TargetTableName)
	err := zorm.Query(ctx.Ctx, finder, &lst, nil)
	//err := DB(ctx).Model(&Target{}).Find(&lst).Error
	for i := 0; i < len(lst); i++ {
		lst[i].FillTagsMap()
	}
	return lst, err
}

func TargetUpdateNote(ctx *ctx.Context, idents []string, note string) error {
	finder := zorm.NewUpdateFinder(TargetTableName).Append("note=?,update_at=? WHERE ident in (?)", note, time.Now().Unix(), idents)
	return UpdateFinder(ctx, finder)
	/*
		return DB(ctx).Model(&Target{}).Where("ident in ?", idents).Updates(map[string]interface{}{
			"note":      note,
			"update_at": time.Now().Unix(),
		}).Error
	*/
}

func TargetUpdateBgid(ctx *ctx.Context, idents []string, bgid int64, clearTags bool) error {

	finder := zorm.NewUpdateFinder(TargetTableName).Append("group_id=?,update_at=?", bgid, time.Now().Unix())

	/*
		fields := map[string]interface{}{
			"group_id":  bgid,
			"update_at": time.Now().Unix(),
		}
	*/
	if clearTags {
		//fields["tags"] = ""
		finder.Append(",tags=?", "")
	}
	finder.Append("WHERE ident in (?)", idents)
	return UpdateFinder(ctx, finder)
	//return DB(ctx).Model(&Target{}).Where("ident in ?", idents).Updates(fields).Error
}

func TargetGet(ctx *ctx.Context, where string, args ...interface{}) (*Target, error) {
	lst := make([]Target, 0)
	finder := zorm.NewSelectFinder(TargetTableName)
	AppendWhere(finder, where, args...)
	err := zorm.Query(ctx.Ctx, finder, &lst, nil)
	//err := DB(ctx).Where(where, args...).Find(&lst).Error
	if err != nil {
		return nil, err
	}

	if len(lst) == 0 {
		return nil, nil
	}

	lst[0].TagsJSON = strings.Fields(lst[0].Tags)

	return &lst[0], nil
}

func TargetGetById(ctx *ctx.Context, id int64) (*Target, error) {
	return TargetGet(ctx, "id = ?", id)
}

func TargetGetByIdent(ctx *ctx.Context, ident string) (*Target, error) {
	return TargetGet(ctx, "ident = ?", ident)
}

func TargetGetTags(ctx *ctx.Context, idents []string) ([]string, error) {
	finder := zorm.NewSelectFinder(TargetTableName, "DISTINCT tags")
	//session := DB(ctx).Model(new(Target))

	arr := make([]string, 0)
	if len(idents) > 0 {
		//session = session.Where("ident in ?", idents)
		finder.Append("WHERE ident in (?)", idents)
	}
	err := zorm.Query(ctx.Ctx, finder, &arr, nil)
	//err := session.Select("distinct(tags) as tags").Pluck("tags", &arr).Error
	if err != nil {
		return nil, err
	}

	cnt := len(arr)
	if cnt == 0 {
		return []string{}, nil
	}

	set := make(map[string]struct{})
	for i := 0; i < cnt; i++ {
		tags := strings.Fields(arr[i])
		for j := 0; j < len(tags); j++ {
			set[tags[j]] = struct{}{}
		}
	}

	cnt = len(set)
	ret := make([]string, 0, cnt)
	for key := range set {
		ret = append(ret, key)
	}

	sort.Strings(ret)

	return ret, err
}

func (t *Target) AddTags(ctx *ctx.Context, tags []string) error {
	for i := 0; i < len(tags); i++ {
		if !strings.Contains(t.Tags, tags[i]+" ") {
			t.Tags += tags[i] + " "
		}
	}

	arr := strings.Fields(t.Tags)
	sort.Strings(arr)

	finder := zorm.NewUpdateFinder(TargetTableName).Append("tags=?,update_at=? WHERE id=?", strings.Join(arr, " ")+" ", time.Now().Unix(), t.Id)
	return UpdateFinder(ctx, finder)
	/*
		return DB(ctx).Model(t).Updates(map[string]interface{}{
			"tags":      strings.Join(arr, " ") + " ",
			"update_at": time.Now().Unix(),
		}).Error
	*/

}

func (t *Target) DelTags(ctx *ctx.Context, tags []string) error {
	for i := 0; i < len(tags); i++ {
		t.Tags = strings.ReplaceAll(t.Tags, tags[i]+" ", "")
	}
	finder := zorm.NewUpdateFinder(TargetTableName).Append("tags=?,update_at=? WHERE id=?", t.Tags, time.Now().Unix(), t.Id)
	return UpdateFinder(ctx, finder)
	/*
		return DB(ctx).Model(t).Updates(map[string]interface{}{
			"tags":      t.Tags,
			"update_at": time.Now().Unix(),
		}).Error
	*/
}

func (t *Target) FillTagsMap() {
	t.TagsJSON = strings.Fields(t.Tags)
	t.TagsMap = make(map[string]string)
	m := make(map[string]string)
	for _, item := range t.TagsJSON {
		arr := strings.Split(item, "=")
		if len(arr) != 2 {
			continue
		}
		m[arr[0]] = arr[1]
	}

	t.TagsMap = m
}

func (t *Target) FillMeta(meta *HostMeta) {
	t.MemUtil = meta.MemUtil
	t.CpuUtil = meta.CpuUtil
	t.CpuNum = meta.CpuNum
	t.UnixTime = meta.UnixTime
	t.Offset = meta.Offset
	t.OS = meta.OS
	t.Arch = meta.Arch
	t.RemoteAddr = meta.RemoteAddr
}

func TargetIdents(ctx *ctx.Context, ids []int64) ([]string, error) {
	ret := make([]string, 0)

	if len(ids) == 0 {
		return ret, nil
	}
	finder := zorm.NewSelectFinder(TargetTableName, "ident").Append("WHERE id in (?)", ids)
	err := zorm.Query(ctx.Ctx, finder, &ret, nil)

	//err := DB(ctx).Model(&Target{}).Where("id in ?", ids).Pluck("ident", &ret).Error
	return ret, err
}

func TargetIds(ctx *ctx.Context, idents []string) ([]int64, error) {
	ret := make([]int64, 0)

	if len(idents) == 0 {
		return ret, nil
	}
	finder := zorm.NewSelectFinder(TargetTableName, "id").Append("WHERE ident in (?)", idents)
	err := zorm.Query(ctx.Ctx, finder, &ret, nil)
	//err := DB(ctx).Model(&Target{}).Where("ident in ?", idents).Pluck("id", &ret).Error
	return ret, err
}

func IdentsFilter(ctx *ctx.Context, idents []string, where string, args ...interface{}) ([]string, error) {
	arr := make([]string, 0)
	if len(idents) == 0 {
		return arr, nil
	}
	finder := zorm.NewSelectFinder(TargetTableName, "ident").Append("WHERE ident in (?)", idents)
	if where != "" {
		finder.Append("and "+where, args...)
	}

	err := zorm.Query(ctx.Ctx, finder, &arr, nil)
	//err := DB(ctx).Model(&Target{}).Where("ident in ?", idents).Where(where, args...).Pluck("ident", &arr).Error
	return arr, err
}

func (m *Target) UpdateFieldsMap(ctx *ctx.Context, fields map[string]interface{}) error {
	entityMap := zorm.NewEntityMap(TargetTableName)
	entityMap.PkColumnName = m.GetPKColumnName()
	for k, v := range fields {
		entityMap.Set(k, v)
	}
	entityMap.Set(m.GetPKColumnName(), m.Id)
	_, err := zorm.Transaction(ctx.Ctx, func(ctx context.Context) (interface{}, error) {
		return zorm.UpdateEntityMap(ctx, entityMap)
	})
	return err
	//return DB(ctx).Model(m).Updates(fields).Error
}
