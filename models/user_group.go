package models

import (
	"context"
	"time"

	"gitee.com/chunanyong/zorm"
	"github.com/ccfos/nightingale/v6/pkg/ctx"
	"github.com/ccfos/nightingale/v6/pkg/poster"

	"github.com/pkg/errors"
	"github.com/toolkits/pkg/str"
)

const UserGroupTableName = "user_group"

type UserGroup struct {
	// 引入默认的struct,隔离IEntityStruct的方法改动
	zorm.EntityStruct
	Id       int64   `json:"id" column:"id"`
	Name     string  `json:"name" column:"name"`
	Note     string  `json:"note" column:"note"`
	CreateAt int64   `json:"create_at" column:"create_at"`
	CreateBy string  `json:"create_by" column:"create_by"`
	UpdateAt int64   `json:"update_at" column:"update_at"`
	UpdateBy string  `json:"update_by" column:"update_by"`
	UserIds  []int64 `json:"-"`
}

func (ug *UserGroup) GetTableName() string {
	return UserGroupTableName
}

func (ug *UserGroup) DB2FE() error {
	return nil
}
func (ug *UserGroup) Verify() error {
	if str.Dangerous(ug.Name) {
		return errors.New("Name has invalid characters")
	}

	if str.Dangerous(ug.Note) {
		return errors.New("Note has invalid characters")
	}

	return nil
}

func (ug *UserGroup) Update(ctx *ctx.Context, selectFields ...string) error {
	if err := ug.Verify(); err != nil {
		return err
	}
	return Update(ctx, ug, selectFields)
	//return DB(ctx).Model(ug).Select(selectField, selectFields...).Updates(ug).Error
}

func UserGroupCount(ctx *ctx.Context, where string, args ...interface{}) (num int64, err error) {
	finder := zorm.NewSelectFinder(UserGroupTableName, "count(*)")
	AppendWhere(finder, where, args...)
	return Count(ctx, finder)
	//return Count(DB(ctx).Model(&UserGroup{}).Where(where, args...))
}

func (ug *UserGroup) Add(ctx *ctx.Context) error {
	if err := ug.Verify(); err != nil {
		return err
	}

	num, err := UserGroupCount(ctx, "name=?", ug.Name)
	if err != nil {
		return errors.WithMessage(err, "failed to count user-groups")
	}

	if num > 0 {
		return errors.New("UserGroup already exists")
	}

	now := time.Now().Unix()
	ug.CreateAt = now
	ug.UpdateAt = now
	return Insert(ctx, ug)
}

func (ug *UserGroup) Del(ctx *ctx.Context) error {

	_, err := zorm.Transaction(ctx.Ctx, func(ctx context.Context) (interface{}, error) {
		f1 := zorm.NewDeleteFinder(UserGroupMemberTableName).Append("WHERE group_id=?", ug.Id)
		_, err := zorm.UpdateFinder(ctx, f1)
		if err != nil {
			return nil, err
		}
		return zorm.Delete(ctx, ug)
	})
	return err

	/*
		return DB(ctx).Transaction(func(tx *zorm.DBDao) error {
			if err := tx.Where("group_id=?", ug.Id).Delete(&UserGroupMember{}).Error; err != nil {
				return err
			}

			if err := tx.Where("id=?", ug.Id).Delete(&UserGroup{}).Error; err != nil {
				return err
			}

			return nil
		})
	*/
}

func UserGroupGet(ctx *ctx.Context, where string, args ...interface{}) (*UserGroup, error) {
	lst := make([]*UserGroup, 0)
	finder := zorm.NewSelectFinder(UserGroupTableName)
	AppendWhere(finder, where, args...)
	err := zorm.Query(ctx.Ctx, finder, &lst, nil)
	//err := DB(ctx).Where(where, args...).Find(&lst).Error
	if err != nil {
		return nil, err
	}

	if len(lst) == 0 {
		return nil, nil
	}

	return lst[0], nil
}

func UserGroupGetById(ctx *ctx.Context, id int64) (*UserGroup, error) {
	return UserGroupGet(ctx, "id = ?", id)
}

func UserGroupGetByIds(ctx *ctx.Context, ids []int64) ([]UserGroup, error) {
	lst := make([]UserGroup, 0)
	if len(ids) == 0 {
		return lst, nil
	}
	finder := zorm.NewSelectFinder(UserGroupTableName).Append("WHERE id in (?) order by name asc", ids)
	err := zorm.Query(ctx.Ctx, finder, &lst, nil)
	//err := DB(ctx).Where("id in ?", ids).Order("name").Find(&lst).Error
	return lst, err
}

func UserGroupGetAll(ctx *ctx.Context) ([]*UserGroup, error) {
	if !ctx.IsCenter {
		lst, err := poster.GetByUrls[[]*UserGroup](ctx, "/v1/n9e/users")
		return lst, err
	}

	lst := make([]*UserGroup, 0)
	finder := zorm.NewSelectFinder(UserGroupTableName)
	err := zorm.Query(ctx.Ctx, finder, &lst, nil)
	//err := DB(ctx).Find(&lst).Error
	return lst, err
}

func (ug *UserGroup) AddMembers(ctx *ctx.Context, userIds []int64) error {
	count := len(userIds)
	for i := 0; i < count; i++ {
		user, err := UserGetById(ctx, userIds[i])
		if err != nil {
			return err
		}
		if user == nil {
			continue
		}
		err = UserGroupMemberAdd(ctx, ug.Id, user.Id)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ug *UserGroup) DelMembers(ctx *ctx.Context, userIds []int64) error {
	return UserGroupMemberDel(ctx, ug.Id, userIds)
}

func UserGroupStatistics(ctx *ctx.Context) (*Statistics, error) {
	if !ctx.IsCenter {
		s, err := poster.GetByUrls[*Statistics](ctx, "/v1/n9e/statistic?name=user_group")
		return s, err
	}
	return StatisticsGet(ctx, UserGroupTableName)
	/*
		session := DB(ctx).Model(&UserGroup{}).Select("count(*) as total", "max(update_at) as last_updated")

		var stats []*Statistics
		err := session.Find(&stats).Error
		if err != nil {
			return nil, err
		}

		return stats[0], nil
	*/
}
