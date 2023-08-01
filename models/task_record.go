package models

import (
	"gitee.com/chunanyong/zorm"
	"github.com/ccfos/nightingale/v6/pkg/ctx"
	"github.com/ccfos/nightingale/v6/pkg/poster"
)

const TaskRecordTableName = "task_record"

type TaskRecord struct {
	// 引入默认的struct,隔离IEntityStruct的方法改动
	zorm.EntityStruct
	Id           int64  `json:"id" column:"id"`
	EventId      int64  `json:"event_id" column:"event_id"`
	GroupId      int64  `json:"group_id" column:"group_id"`
	IbexAddress  string `json:"ibex_address" column:"ibex_address"`
	IbexAuthUser string `json:"ibex_auth_user" column:"ibex_auth_user"`
	IbexAuthPass string `json:"ibex_auth_pass" column:"ibex_auth_pass"`
	Title        string `json:"title" column:"title"`
	Account      string `json:"account" column:"account"`
	Batch        int    `json:"batch" column:"batch"`
	Tolerance    int    `json:"tolerance" column:"tolerance"`
	Timeout      int    `json:"timeout" column:"timeout"`
	Pause        string `json:"pause" column:"pause"`
	Script       string `json:"script" column:"script"`
	Args         string `json:"args" column:"args"`
	CreateAt     int64  `json:"create_at" column:"create_at"`
	CreateBy     string `json:"create_by" column:"create_by"`
}

func (r *TaskRecord) GetTableName() string {
	return TaskRecordTableName
}

// create task
func (r *TaskRecord) Add(ctx *ctx.Context) error {
	if !ctx.IsCenter {
		err := poster.PostByUrls(ctx, "/v1/n9e/task-record-add", r)
		return err
	}

	return Insert(ctx, r)
}

// list task, filter by group_id, create_by
func TaskRecordTotal(ctx *ctx.Context, bgid, beginTime int64, createBy, query string) (int64, error) {
	finder := zorm.NewSelectFinder(TaskRecordTableName, "count(*)").Append("WHERE create_at > ? and group_id = ?", beginTime, bgid)
	//session := DB(ctx).Model(new(TaskRecord)).Where("create_at > ? and group_id = ?", beginTime, bgid)

	if createBy != "" {
		//session = session.Where("create_by = ?", createBy)
		finder.Append("and create_by = ?", createBy)
	}

	if query != "" {
		//session = session.Where("title like ?", "%"+query+"%")
		finder.Append("and title like ?", "%"+query+"%")
	}

	return Count(ctx, finder)
}

func TaskRecordGets(ctx *ctx.Context, bgid, beginTime int64, createBy, query string, limit, offset int) ([]*TaskRecord, error) {
	finder := zorm.NewSelectFinder(TaskRecordTableName).Append("WHERE create_at > ? and group_id = ?", beginTime, bgid)
	//session := DB(ctx).Where("create_at > ? and group_id = ?", beginTime, bgid).Order("create_at desc").Limit(limit).Offset(offset)

	if createBy != "" {
		//session = session.Where("create_by = ?", createBy)
		finder.Append("and create_by = ?", createBy)
	}

	if query != "" {
		//session = session.Where("title like ?", "%"+query+"%")
		finder.Append("and title like ?", "%"+query+"%")
	}
	finder.Append("order by create_at desc")
	finder.SelectTotalCount = false
	page := zorm.NewPage()
	page.PageSize = limit
	page.PageNo = offset / limit
	finder.SelectTotalCount = false
	lst := make([]*TaskRecord, 0)
	err := zorm.Query(ctx.Ctx, finder, &lst, page)
	//err := session.Find(&lst).Error
	return lst, err
}

// update is_done field
func (r *TaskRecord) UpdateIsDone(ctx *ctx.Context, isDone int) error {
	return UpdateColumn(ctx, TaskRecordTableName, r.Id, "is_done", isDone)
	//return DB(ctx).Model(r).Update("is_done", isDone).Error
}
