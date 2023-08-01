package models

import (
	"gitee.com/chunanyong/zorm"
	"github.com/ccfos/nightingale/v6/pkg/ctx"
)

const BusiGroupMemberTableName = "busi_group_member"

type BusiGroupMember struct {
	// 引入默认的struct,隔离IEntityStruct的方法改动
	zorm.EntityStruct
	Id          int64  `json:"id" column:"id"`
	BusiGroupId int64  `json:"busi_group_id" column:"busi_group_id"`
	UserGroupId int64  `json:"user_group_id" column:"user_group_id"`
	PermFlag    string `json:"perm_flag" column:"perm_flag"`
}

func (bg *BusiGroupMember) GetTableName() string {
	return BusiGroupMemberTableName
}

func (bgm *BusiGroupMember) DB2FE() error {
	return nil
}

func BusiGroupIds(ctx *ctx.Context, userGroupIds []int64, permFlag ...string) ([]int64, error) {
	if len(userGroupIds) == 0 {
		return []int64{}, nil
	}

	finder := zorm.NewSelectFinder(BusiGroupMemberTableName, "busi_group_id").Append("WHERE user_group_id in (?)", userGroupIds)

	//session := DB(ctx).Model(&BusiGroupMember{}).Where("user_group_id in ?", userGroupIds)
	if len(permFlag) > 0 {
		//session = session.Where("perm_flag=?", permFlag[0])
		finder.Append("and perm_flag=?", permFlag[0])
	}

	ids := make([]int64, 0)
	//err := session.Pluck("busi_group_id", &ids).Error
	err := zorm.Query(ctx.Ctx, finder, &ids, nil)
	return ids, err
}

func UserGroupIdsOfBusiGroup(ctx *ctx.Context, busiGroupId int64, permFlag ...string) ([]int64, error) {
	finder := zorm.NewSelectFinder(BusiGroupMemberTableName, "user_group_id").Append("WHERE busi_group_id = ?", busiGroupId)

	//session := DB(ctx).Model(&BusiGroupMember{}).Where("busi_group_id = ?", busiGroupId)
	if len(permFlag) > 0 {
		//session = session.Where("perm_flag=?", permFlag[0])
		finder.Append("and perm_flag=?", permFlag[0])
	}

	ids := make([]int64, 0)
	//err := session.Pluck("user_group_id", &ids).Error
	err := zorm.Query(ctx.Ctx, finder, &ids, nil)
	return ids, err
}

func BusiGroupMemberCount(ctx *ctx.Context, where string, args ...interface{}) (int64, error) {
	finder := zorm.NewSelectFinder(BusiGroupMemberTableName, "count(*)")
	AppendWhere(finder, where, args...)
	return Count(ctx, finder)
	//return Count(DB(ctx).Model(&BusiGroupMember{}).Where(where, args...))
}

func BusiGroupMemberAdd(ctx *ctx.Context, member BusiGroupMember) error {
	obj, err := BusiGroupMemberGet(ctx, "busi_group_id = ? and user_group_id = ?", member.BusiGroupId, member.UserGroupId)
	if err != nil {
		return err
	}

	if obj == nil {
		// insert
		return Insert(ctx, &BusiGroupMember{
			BusiGroupId: member.BusiGroupId,
			UserGroupId: member.UserGroupId,
			PermFlag:    member.PermFlag,
		})
	} else {
		// update
		if obj.PermFlag == member.PermFlag {
			return nil
		}

		finder := zorm.NewUpdateFinder(BusiGroupMemberTableName).Append("perm_flag=? WHERE busi_group_id = ? and user_group_id = ?", member.PermFlag, member.BusiGroupId, member.UserGroupId)
		return UpdateFinder(ctx, finder)
		//return DB(ctx).Model(&BusiGroupMember{}).Where("busi_group_id = ? and user_group_id = ?", member.BusiGroupId, member.UserGroupId).Update("perm_flag", member.PermFlag).Error
	}
}

func BusiGroupMemberGet(ctx *ctx.Context, where string, args ...interface{}) (*BusiGroupMember, error) {
	lst := make([]BusiGroupMember, 0)
	finder := zorm.NewSelectFinder(BusiGroupMemberTableName)
	AppendWhere(finder, where, args...)
	err := zorm.Query(ctx.Ctx, finder, &lst, nil)
	//err := DB(ctx).Where(where, args...).Find(&lst).Error
	if err != nil {
		return nil, err
	}

	if len(lst) == 0 {
		return nil, nil
	}

	return &lst[0], nil
}

func BusiGroupMemberDel(ctx *ctx.Context, where string, args ...interface{}) error {
	finder := zorm.NewDeleteFinder(BusiGroupMemberTableName)
	AppendWhere(finder, where, args...)
	return UpdateFinder(ctx, finder)
	//return DB(ctx).Where(where, args...).Delete(&BusiGroupMember{}).Error
}

func BusiGroupMemberGets(ctx *ctx.Context, where string, args ...interface{}) ([]BusiGroupMember, error) {
	lst := make([]BusiGroupMember, 0)
	finder := zorm.NewSelectFinder(BusiGroupMemberTableName).Append("WHERE "+where+" order by perm_flag asc", args...)
	err := zorm.Query(ctx.Ctx, finder, &lst, nil)
	//err := DB(ctx).Where(where, args...).Order("perm_flag").Find(&lst).Error
	return lst, err
}

func BusiGroupMemberGetsByBusiGroupId(ctx *ctx.Context, busiGroupId int64) ([]BusiGroupMember, error) {
	return BusiGroupMemberGets(ctx, "busi_group_id=?", busiGroupId)
}
