package models

import (
	"context"
	"strings"
	"time"

	"errors"

	"gitee.com/chunanyong/zorm"
	"github.com/ccfos/nightingale/v6/pkg/ctx"
	"github.com/toolkits/pkg/str"
)

const BoardTableName = "board"

type Board struct {
	// 引入默认的struct,隔离IEntityStruct的方法改动
	zorm.EntityStruct
	Id       int64  `json:"id" column:"id"`
	GroupId  int64  `json:"group_id" column:"group_id"`
	Name     string `json:"name" column:"name"`
	Ident    string `json:"ident" column:"ident"`
	Tags     string `json:"tags" column:"tags"`
	CreateAt int64  `json:"create_at" column:"create_at"`
	CreateBy string `json:"create_by" column:"create_by"`
	UpdateAt int64  `json:"update_at" column:"update_at"`
	UpdateBy string `json:"update_by" column:"update_by"`
	Configs  string `json:"configs"`
	Public   int    `json:"public" column:"is_public"`  // 0: false, 1: true
	BuiltIn  int    `json:"built_in" column:"built_in"` // 0: false, 1: true
	Hide     int    `json:"hide" column:"hide"`         // 0: false, 1: true
}

func (b *Board) GetTableName() string {
	return BoardTableName
}

func (b *Board) DB2FE() error {
	return nil
}

func (b *Board) Verify() error {
	if b.Name == "" {
		return errors.New("Name is blank")
	}

	if str.Dangerous(b.Name) {
		return errors.New("Name has invalid characters")
	}

	return nil
}

func (b *Board) CanRenameIdent(ctx *ctx.Context, ident string) (bool, error) {
	if ident == "" {
		return true, nil
	}

	finder := zorm.NewSelectFinder(BoardTableName, "count(*)").Append("WHERE ident=? and id <> ?", ident, b.Id)
	cnt, err := Count(ctx, finder)
	//cnt, err := Count(DB(ctx).Model(b).Where("ident=? and id <> ?", ident, b.Id))
	if err != nil {
		return false, err
	}

	return cnt == 0, nil
}

func (b *Board) Add(ctx *ctx.Context) error {
	if err := b.Verify(); err != nil {
		return err
	}

	if b.Ident != "" {
		// ident duplicate check
		finder := zorm.NewSelectFinder(BoardTableName, "count(*)").Append("WHERE ident=? ", b.Ident)
		cnt, err := Count(ctx, finder)
		//cnt, err := Count(DB(ctx).Model(b).Where("ident=?", b.Ident))
		if err != nil {
			return err
		}

		if cnt > 0 {
			return errors.New("Ident duplicate")
		}
	}

	now := time.Now().Unix()
	b.CreateAt = now
	b.UpdateAt = now

	return Insert(ctx, b)
}

func (b *Board) Update(ctx *ctx.Context, selectFields ...string) error {
	if err := b.Verify(); err != nil {
		return err
	}
	return Update(ctx, b, selectFields)
	//return DB(ctx).Model(b).Select(selectField, selectFields...).Updates(b).Error
}

func (b *Board) Del(ctx *ctx.Context) error {
	/*
		return DB(ctx).Transaction(func(tx *zorm.DBDao) error {
			if err := tx.Where("id=?", b.Id).Delete(&BoardPayload{}).Error; err != nil {
				return err
			}

			if err := tx.Where("id=?", b.Id).Delete(&Board{}).Error; err != nil {
				return err
			}

			return nil
		})
	*/
	_, err := zorm.Transaction(ctx.Ctx, func(ctx context.Context) (interface{}, error) {
		f1 := zorm.NewDeleteFinder(BoardPayloadTableName).Append("WHERE id=?", b.Id)
		_, err := zorm.UpdateFinder(ctx, f1)
		if err != nil {
			return nil, err
		}
		f2 := zorm.NewDeleteFinder(BoardTableName).Append("WHERE id=?", b.Id)
		return zorm.UpdateFinder(ctx, f2)

	})
	return err

}

func BoardGetByID(ctx *ctx.Context, id int64) (*Board, error) {
	lst := make([]Board, 0)
	finder := zorm.NewSelectFinder(BoardTableName).Append("WHERE id = ?", id)
	err := zorm.Query(ctx.Ctx, finder, &lst, nil)
	//err := DB(ctx).Where("id = ?", id).Find(&lst).Error
	if err != nil {
		return nil, err
	}

	if len(lst) == 0 {
		return nil, nil
	}

	return &lst[0], nil
}

// BoardGet for detail page
func BoardGet(ctx *ctx.Context, where string, args ...interface{}) (*Board, error) {
	lst := make([]Board, 0)
	finder := zorm.NewSelectFinder(BoardTableName)
	AppendWhere(finder, where, args...)
	err := zorm.Query(ctx.Ctx, finder, &lst, nil)
	//err := DB(ctx).Where(where, args...).Find(&lst).Error
	if err != nil {
		return nil, err
	}

	if len(lst) == 0 {
		return nil, nil
	}

	payload, err := BoardPayloadGet(ctx, lst[0].Id)
	if err != nil {
		return nil, err
	}

	lst[0].Configs = payload

	return &lst[0], nil
}

func BoardCount(ctx *ctx.Context, where string, args ...interface{}) (num int64, err error) {
	finder := zorm.NewSelectFinder(BoardTableName, "count(*)")
	AppendWhere(finder, where, args...)
	return Count(ctx, finder)
	//return Count(DB(ctx).Model(&Board{}).Where(where, args...))
}

func BoardExists(ctx *ctx.Context, where string, args ...interface{}) (bool, error) {
	num, err := BoardCount(ctx, where, args...)
	return num > 0, err
}

// BoardGets for list page
func BoardGetsByGroupId(ctx *ctx.Context, groupId int64, query string) ([]Board, error) {
	finder := zorm.NewSelectFinder(BoardTableName).Append("WHERE group_id=?", groupId)
	//session := DB(ctx).Where("group_id=?", groupId).Order("name")

	arr := strings.Fields(query)
	if len(arr) > 0 {
		for i := 0; i < len(arr); i++ {
			if strings.HasPrefix(arr[i], "-") {
				q := "%" + arr[i][1:] + "%"
				//session = session.Where("name not like ? and tags not like ?", q, q)
				finder.Append("and name not like ? and tags not like ?", q, q)
			} else {
				q := "%" + arr[i] + "%"
				//session = session.Where("(name like ? or tags like ?)", q, q)
				finder.Append("and (name like ? or tags like ?)", q, q)
			}
		}
	}
	finder.Append("order by name asc")
	lst := make([]Board, 0)
	err := zorm.Query(ctx.Ctx, finder, &lst, nil)
	//err := session.Find(&objs).Error
	return lst, err
}

func BoardGets(ctx *ctx.Context, query, where string, args ...interface{}) ([]Board, error) {
	finder := zorm.NewSelectFinder(BoardTableName).Append("WHERE 1=1")
	//session := DB(ctx).Order("name")
	if where != "" {
		//session = session.Where(where, args...)
		finder.Append("and "+where, args...)
	}

	arr := strings.Fields(query)
	if len(arr) > 0 {
		for i := 0; i < len(arr); i++ {
			if strings.HasPrefix(arr[i], "-") {
				q := "%" + arr[i][1:] + "%"
				//session = session.Where("name not like ? and tags not like ?", q, q)
				finder.Append("and name not like ? and tags not like ?", q, q)
			} else {
				q := "%" + arr[i] + "%"
				//session = session.Where("(name like ? or tags like ?)", q, q)
				finder.Append("and (name like ? or tags like ?)", q, q)
			}
		}
	}

	finder.Append("order by name asc")
	lst := make([]Board, 0)
	err := zorm.Query(ctx.Ctx, finder, &lst, nil)
	//err := session.Find(&objs).Error
	return lst, err
}

func BoardSetHide(ctx *ctx.Context, ids []int64) error {
	/*
		return DB(ctx).Transaction(func(tx *zorm.DBDao) error {
			if err := tx.Model(&Board{}).Where("built_in = 1").Update("hide", 0).Error; err != nil {
				return err
			}

			if err := tx.Model(&Board{}).Where("id in (?) and built_in = 1", ids).Update("hide", 1).Error; err != nil {
				return err
			}
			return nil
		})
	*/
	_, err := zorm.Transaction(ctx.Ctx, func(ctx context.Context) (interface{}, error) {
		f1 := zorm.NewUpdateFinder(BoardTableName).Append("hide=0 WHERE built_in=1")
		_, err := zorm.UpdateFinder(ctx, f1)
		if err != nil {
			return nil, err
		}
		f2 := zorm.NewUpdateFinder(BoardTableName).Append("hide=1 WHERE id in (?) and built_in=1", ids)
		return zorm.UpdateFinder(ctx, f2)

	})
	return err
}
