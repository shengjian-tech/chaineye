package models

import (
	"gitee.com/chunanyong/zorm"
	"github.com/ccfos/nightingale/v6/pkg/ctx"
	"github.com/ccfos/nightingale/v6/pkg/poster"
)

const UserGroupMemberTableName = "user_group_member"

type UserGroupMember struct {
	// 引入默认的struct,隔离IEntityStruct的方法改动
	zorm.EntityStruct
	Id      int64 `column:"id"`
	GroupId int64 `column:"group_id"`
	UserId  int64 `column:"user_id"`
}

func (ugm *UserGroupMember) GetTableName() string {
	return UserGroupMemberTableName
}

func (UserGroupMember) DB2FE() error {
	return nil
}

func MyGroupIds(ctx *ctx.Context, userId int64) ([]int64, error) {
	ids := make([]int64, 0)
	finder := zorm.NewSelectFinder(UserGroupMemberTableName, "group_id").Append("WHERE user_id=?", userId)
	err := zorm.Query(ctx.Ctx, finder, &ids, nil)
	//err := DB(ctx).Model(&UserGroupMember{}).Where("user_id=?", userId).Pluck("group_id", &ids).Error
	return ids, err
}

func MemberIds(ctx *ctx.Context, groupId int64) ([]int64, error) {
	ids := make([]int64, 0)
	finder := zorm.NewSelectFinder(UserGroupMemberTableName, "user_id").Append("WHERE group_id=?", groupId)
	err := zorm.Query(ctx.Ctx, finder, &ids, nil)
	//err := DB(ctx).Model(&UserGroupMember{}).Where("group_id=?", groupId).Pluck("user_id", &ids).Error
	return ids, err
}

func UserGroupMemberCount(ctx *ctx.Context, where string, args ...interface{}) (int64, error) {
	finder := zorm.NewSelectFinder(UserGroupMemberTableName, "count(*)")
	AppendWhere(finder, where, args...)
	return Count(ctx, finder)
	//return Count(DB(ctx).Model(&UserGroupMember{}).Where(where, args...))
}

func UserGroupMemberAdd(ctx *ctx.Context, groupId, userId int64) error {
	num, err := UserGroupMemberCount(ctx, "user_id=? and group_id=?", userId, groupId)
	if err != nil {
		return err
	}

	if num > 0 {
		// already exists
		return nil
	}

	obj := &UserGroupMember{
		GroupId: groupId,
		UserId:  userId,
	}

	return Insert(ctx, obj)
}

func UserGroupMemberDel(ctx *ctx.Context, groupId int64, userIds []int64) error {
	if len(userIds) == 0 {
		return nil
	}
	finder := zorm.NewDeleteFinder(UserGroupMemberTableName).Append("WHERE group_id = ? and user_id in (?)", groupId, userIds)
	return UpdateFinder(ctx, finder)
	//return DB(ctx).Where("group_id = ? and user_id in ?", groupId, userIds).Delete(&UserGroupMember{}).Error
}

func UserGroupMemberGetAll(ctx *ctx.Context) ([]*UserGroupMember, error) {
	if !ctx.IsCenter {
		lst, err := poster.GetByUrls[[]*UserGroupMember](ctx, "/v1/n9e/user-group-members")
		return lst, err
	}

	lst := make([]*UserGroupMember, 0)
	finder := zorm.NewSelectFinder(UserGroupMemberTableName)
	err := zorm.Query(ctx.Ctx, finder, &lst, nil)
	//err := DB(ctx).Find(&lst).Error
	return lst, err
}
