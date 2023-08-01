package models

import (
	"errors"

	"gitee.com/chunanyong/zorm"
	"github.com/ccfos/nightingale/v6/pkg/ctx"
)

const BoardPayloadTableName = "board_payload"

type BoardPayload struct {
	// 引入默认的struct,隔离IEntityStruct的方法改动
	zorm.EntityStruct
	Id      int64  `json:"id" column:"id"`
	Payload string `json:"payload" column:"payload"`
}

func (p *BoardPayload) GetTableName() string {
	return BoardPayloadTableName
}

func (p *BoardPayload) Update(ctx *ctx.Context, selectFields ...string) error {
	return Update(ctx, p, selectFields)
	//return DB(ctx).Model(p).Select(selectFields...).Updates(p).Error
}

func BoardPayloadGets(ctx *ctx.Context, ids []int64) ([]*BoardPayload, error) {
	if len(ids) == 0 {
		return nil, errors.New("empty ids")
	}
	arr := make([]*BoardPayload, 0)
	finder := zorm.NewSelectFinder(BoardPayloadTableName).Append("WHERE id in (?)", ids)
	err := zorm.Query(ctx.Ctx, finder, &arr, nil)
	//err := DB(ctx).Where("id in ?", ids).Find(&arr).Error
	return arr, err
}

func BoardPayloadGet(ctx *ctx.Context, id int64) (string, error) {
	payloads, err := BoardPayloadGets(ctx, []int64{id})
	if err != nil {
		return "", err
	}

	if len(payloads) == 0 {
		return "", nil
	}

	return payloads[0].Payload, nil
}

func BoardPayloadSave(ctx *ctx.Context, id int64, payload string) error {
	var bp BoardPayload
	finder := zorm.NewSelectFinder(BoardPayloadTableName).Append("WHERE id = ?", id)
	_, err := zorm.QueryRow(ctx.Ctx, finder, &bp)
	//err := DB(ctx).Where("id = ?", id).Find(&bp).Error
	if err != nil {
		return err
	}

	if bp.Id > 0 {
		// already exists
		bp.Payload = payload
		return bp.Update(ctx, "payload")
	}

	return Insert(ctx, &BoardPayload{
		Id:      id,
		Payload: payload,
	})
}
