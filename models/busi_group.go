package models

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gitee.com/chunanyong/zorm"
	"github.com/ccfos/nightingale/v6/pkg/ctx"
	"github.com/ccfos/nightingale/v6/pkg/poster"
)

const BusiGroupTableName = "busi_group"

type BusiGroup struct {
	// 引入默认的struct,隔离IEntityStruct的方法改动
	zorm.EntityStruct
	Id          int64                   `json:"id" column:"id"`
	Name        string                  `json:"name" column:"name"`
	LabelEnable int                     `json:"label_enable" column:"label_enable"`
	LabelValue  string                  `json:"label_value" column:"label_value"`
	CreateAt    int64                   `json:"create_at" column:"create_at"`
	CreateBy    string                  `json:"create_by" column:"create_by"`
	UpdateAt    int64                   `json:"update_at" column:"update_at"`
	UpdateBy    string                  `json:"update_by" column:"update_by"`
	UserGroups  []UserGroupWithPermFlag `json:"user_groups"`
	DB          *zorm.DBDao             `json:"-"`
}

func New(db *zorm.DBDao) *BusiGroup {
	return &BusiGroup{
		DB: db,
	}
}

type UserGroupWithPermFlag struct {
	UserGroup *UserGroup `json:"user_group"`
	PermFlag  string     `json:"perm_flag"`
}

func (bg *BusiGroup) GetTableName() string {
	return BusiGroupTableName
}

func (bg *BusiGroup) DB2FE() error {
	return nil
}

func (bg *BusiGroup) FillUserGroups(ctx *ctx.Context) error {
	members, err := BusiGroupMemberGetsByBusiGroupId(ctx, bg.Id)
	if err != nil {
		return err
	}

	if len(members) == 0 {
		return nil
	}

	for i := 0; i < len(members); i++ {
		ug, err := UserGroupGetById(ctx, members[i].UserGroupId)
		if err != nil {
			return err
		}
		bg.UserGroups = append(bg.UserGroups, UserGroupWithPermFlag{
			UserGroup: ug,
			PermFlag:  members[i].PermFlag,
		})
	}

	return nil
}

func BusiGroupGetMap(ctx *ctx.Context) (map[int64]*BusiGroup, error) {
	lst := make([]*BusiGroup, 0)
	var err error
	if !ctx.IsCenter {
		lst, err = poster.GetByUrls[[]*BusiGroup](ctx, "/v1/n9e/busi-groups")
		if err != nil {
			return nil, err
		}
	} else {
		finder := zorm.NewSelectFinder(BusiGroupTableName)
		err = zorm.Query(ctx.Ctx, finder, &lst, nil)
		//err = DB(ctx).Find(&lst).Error
		if err != nil {
			return nil, err
		}
	}

	ret := make(map[int64]*BusiGroup)
	for i := 0; i < len(lst); i++ {
		ret[lst[i].Id] = lst[i]
	}

	return ret, nil
}

func BusiGroupGetAll(ctx *ctx.Context) ([]*BusiGroup, error) {
	lst := make([]*BusiGroup, 0)
	finder := zorm.NewSelectFinder(BusiGroupTableName)
	err := zorm.Query(ctx.Ctx, finder, &lst, nil)
	//err := DB(ctx).Find(&lst).Error
	return lst, err
}

func BusiGroupGet(ctx *ctx.Context, where string, args ...interface{}) (*BusiGroup, error) {
	lst := make([]*BusiGroup, 0)
	finder := zorm.NewSelectFinder(BusiGroupTableName)
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

func BusiGroupGetById(ctx *ctx.Context, id int64) (*BusiGroup, error) {
	return BusiGroupGet(ctx, "id=?", id)
}

func BusiGroupExists(ctx *ctx.Context, where string, args ...interface{}) (bool, error) {
	finder := zorm.NewSelectFinder(BusiGroupTableName, "count(*)")
	AppendWhere(finder, where, args...)
	num, err := Count(ctx, finder)
	//num, err := Count(DB(ctx).Model(&BusiGroup{}).Where(where, args...))
	return num > 0, err
}

func (bg *BusiGroup) Del(ctx *ctx.Context) error {
	finder := zorm.NewSelectFinder(AlertMuteTableName, "count(*)").Append("WHERE group_id=?", bg.Id)
	has, err := Exists(ctx, finder)
	//has, err := Exists(DB(ctx).Model(&AlertMute{}).Where("group_id=?", bg.Id))
	if err != nil {
		return err
	}

	if has {
		return errors.New("Some alert mutes still in the BusiGroup")
	}
	finder = zorm.NewSelectFinder(AlertSubscribeTableName, "count(*)").Append("WHERE group_id=?", bg.Id)
	has, err = Exists(ctx, finder)
	//has, err = Exists(DB(ctx).Model(&AlertSubscribe{}).Where("group_id=?", bg.Id))
	if err != nil {
		return err
	}

	if has {
		return errors.New("Some alert subscribes still in the BusiGroup")
	}
	finder = zorm.NewSelectFinder(TargetTableName, "count(*)").Append("WHERE group_id=?", bg.Id)
	has, err = Exists(ctx, finder)
	//has, err = Exists(DB(ctx).Model(&Target{}).Where("group_id=?", bg.Id))
	if err != nil {
		return err
	}

	if has {
		return errors.New("Some targets still in the BusiGroup")
	}
	finder = zorm.NewSelectFinder(BoardTableName, "count(*)").Append("WHERE group_id=?", bg.Id)
	has, err = Exists(ctx, finder)
	//has, err = Exists(DB(ctx).Model(&Board{}).Where("group_id=?", bg.Id))
	if err != nil {
		return err
	}

	if has {
		return errors.New("Some dashboards still in the BusiGroup")
	}
	finder = zorm.NewSelectFinder(TaskTplTableName, "count(*)").Append("WHERE group_id=?", bg.Id)
	has, err = Exists(ctx, finder)
	//has, err = Exists(DB(ctx).Model(&TaskTpl{}).Where("group_id=?", bg.Id))
	if err != nil {
		return err
	}

	if has {
		return errors.New("Some recovery scripts still in the BusiGroup")
	}

	// hasCR, err := Exists(DB(ctx).Table("collect_rule").Where("group_id=?", bg.Id))
	// if err != nil {
	// 	return err
	// }

	// if hasCR {
	// 	return errors.New("Some collect rules still in the BusiGroup")
	// }
	finder = zorm.NewSelectFinder(AlertRuleTableName, "count(*)").Append("WHERE group_id=?", bg.Id)
	has, err = Exists(ctx, finder)
	//has, err = Exists(DB(ctx).Model(&AlertRule{}).Where("group_id=?", bg.Id))
	if err != nil {
		return err
	}

	if has {
		return errors.New("Some alert rules still in the BusiGroup")
	}

	/*
		return DB(ctx).Transaction(func(tx *zorm.DBDao) error {
			if err := tx.Where("busi_group_id=?", bg.Id).Delete(&BusiGroupMember{}).Error; err != nil {
				return err
			}

			if err := tx.Where("id=?", bg.Id).Delete(&BusiGroup{}).Error; err != nil {
				return err
			}

			// 这个需要好好斟酌一下，删掉BG，对应的活跃告警事件也一并删除
			// BG都删了，说明下面已经没有告警规则了，说明这些活跃告警永远都不会恢复了
			// 而且这些活跃告警已经没人关心了，既然是没人关心的，删了吧
			if err := tx.Where("group_id=?", bg.Id).Delete(&AlertCurEvent{}).Error; err != nil {
				return err
			}

			return nil
		})
	*/
	_, err = zorm.Transaction(ctx.Ctx, func(ctx context.Context) (interface{}, error) {
		f1 := zorm.NewDeleteFinder(BusiGroupMemberTableName).Append("WHERE busi_group_id=?", bg.Id)
		_, err := zorm.UpdateFinder(ctx, f1)
		if err != nil {
			return nil, err
		}
		f2 := zorm.NewDeleteFinder(BusiGroupTableName).Append("WHERE id=?", bg.Id)
		_, err = zorm.UpdateFinder(ctx, f2)
		if err != nil {
			return nil, err
		}
		f3 := zorm.NewDeleteFinder(AlertCurEventTableName).Append("WHERE group_id=?", bg.Id)
		return zorm.UpdateFinder(ctx, f3)
	})
	return err

}

func (bg *BusiGroup) AddMembers(ctx *ctx.Context, members []BusiGroupMember, username string) error {
	for i := 0; i < len(members); i++ {
		err := BusiGroupMemberAdd(ctx, members[i])
		if err != nil {
			return err
		}
	}

	/*
		return DB(ctx).Model(bg).Updates(map[string]interface{}{
			"update_at": time.Now().Unix(),
			"update_by": username,
		}).Error
	*/
	finder := zorm.NewUpdateFinder(BusiGroupTableName).Append("update_at=?,update_by=? WHERE id=?", time.Now().Unix(), username, bg.Id)
	return UpdateFinder(ctx, finder)
}

func (bg *BusiGroup) DelMembers(ctx *ctx.Context, members []BusiGroupMember, username string) error {
	for i := 0; i < len(members); i++ {
		num, err := BusiGroupMemberCount(ctx, "busi_group_id = ? and user_group_id <> ?", members[i].BusiGroupId, members[i].UserGroupId)
		if err != nil {
			return err
		}

		if num == 0 {
			// 说明这是最后一个user-group，如果再删了，就没人可以管理这个busi-group了
			return fmt.Errorf("the business group must retain at least one team")
		}

		err = BusiGroupMemberDel(ctx, "busi_group_id = ? and user_group_id = ?", members[i].BusiGroupId, members[i].UserGroupId)
		if err != nil {
			return err
		}
	}

	/*
		return DB(ctx).Model(bg).Updates(map[string]interface{}{
			"update_at": time.Now().Unix(),
			"update_by": username,
		}).Error
	*/
	finder := zorm.NewUpdateFinder(BusiGroupTableName).Append("update_at=?,update_by=? WHERE id=?", time.Now().Unix(), username, bg.Id)
	return UpdateFinder(ctx, finder)
}

func (bg *BusiGroup) Update(ctx *ctx.Context, name string, labelEnable int, labelValue string, updateBy string) error {
	if bg.Name == name && bg.LabelEnable == labelEnable && bg.LabelValue == labelValue {
		return nil
	}

	exists, err := BusiGroupExists(ctx, "name = ? and id <> ?", name, bg.Id)
	if err != nil {
		return fmt.Errorf("failed to count BusiGroup:%w", err)
	}

	if exists {
		return errors.New("BusiGroup already exists")
	}

	if labelEnable == 1 {
		exists, err = BusiGroupExists(ctx, "label_enable = 1 and label_value = ? and id <> ?", labelValue, bg.Id)
		if err != nil {
			return fmt.Errorf("failed to count BusiGroup:%w", err)
		}

		if exists {
			return errors.New("BusiGroup already exists")
		}
	} else {
		labelValue = ""
	}

	/*
		return DB(ctx).Model(bg).Updates(map[string]interface{}{
			"name":         name,
			"label_enable": labelEnable,
			"label_value":  labelValue,
			"update_at":    time.Now().Unix(),
			"update_by":    updateBy,
		}).Error
	*/
	finder := zorm.NewUpdateFinder(BusiGroupTableName).Append("name=?,label_enable=?,label_value=?,update_at=?,update_by=? WHERE id=?", name, labelEnable, labelValue, time.Now().Unix(), updateBy, bg.Id)
	return UpdateFinder(ctx, finder)

}

func BusiGroupAdd(ctx *ctx.Context, name string, labelEnable int, labelValue string, members []BusiGroupMember, creator string) error {
	exists, err := BusiGroupExists(ctx, "name=?", name)
	if err != nil {
		return fmt.Errorf("failed to count BusiGroup:%w", err)
	}

	if exists {
		return errors.New("BusiGroup already exists")
	}

	if labelEnable == 1 {
		exists, err = BusiGroupExists(ctx, "label_enable = 1 and label_value = ?", labelValue)
		if err != nil {
			return fmt.Errorf("failed to count BusiGroup:%w", err)
		}

		if exists {
			return errors.New("BusiGroup already exists")
		}
	} else {
		labelValue = ""
	}

	count := len(members)
	for i := 0; i < count; i++ {
		ug, err := UserGroupGet(ctx, "id=?", members[i].UserGroupId)
		if err != nil {
			return fmt.Errorf("failed to get UserGroup:%w", err)
		}

		if ug == nil {
			return errors.New("Some UserGroup id not exists")
		}
	}

	now := time.Now().Unix()
	obj := &BusiGroup{
		Name:        name,
		LabelEnable: labelEnable,
		LabelValue:  labelValue,
		CreateAt:    now,
		CreateBy:    creator,
		UpdateAt:    now,
		UpdateBy:    creator,
	}

	/*
		return DB(ctx).Transaction(func(tx *zorm.DBDao) error {
			if err := tx.Create(obj).Error; err != nil {
				return err
			}

			for i := 0; i < len(members); i++ {
				if err := tx.Create(&BusiGroupMember{
					BusiGroupId: obj.Id,
					UserGroupId: members[i].UserGroupId,
					PermFlag:    members[i].PermFlag,
				}).Error; err != nil {
					return err
				}
			}

			return nil
		})
	*/
	_, err = zorm.Transaction(ctx.Ctx, func(ctx context.Context) (interface{}, error) {

		_, err := zorm.Insert(ctx, obj)
		if err != nil {
			return nil, err
		}

		for i := 0; i < len(members); i++ {
			if _, err := zorm.Insert(ctx, &BusiGroupMember{
				BusiGroupId: obj.Id,
				UserGroupId: members[i].UserGroupId,
				PermFlag:    members[i].PermFlag,
			}); err != nil {
				return nil, err
			}
		}
		return nil, err
	})
	return err

}

func BusiGroupStatistics(ctx *ctx.Context) (*Statistics, error) {
	if !ctx.IsCenter {
		s, err := poster.GetByUrls[*Statistics](ctx, "/v1/n9e/statistic?name=busi_group")
		return s, err
	}

	return StatisticsGet(ctx, BusiGroupTableName)
	/*
		session := DB(ctx).Model(&BusiGroup{}).Select("count(*) as total", "max(update_at) as last_updated")

		var stats []*Statistics
		err := session.Find(&stats).Error
		if err != nil {
			return nil, err
		}

		return stats[0], nil
	*/
}
