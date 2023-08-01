package idents

import (
	"context"
	"sync"
	"time"

	"gitee.com/chunanyong/zorm"
	"github.com/ccfos/nightingale/v6/models"
	"github.com/ccfos/nightingale/v6/pkg/ctx"
	"github.com/ccfos/nightingale/v6/pkg/poster"

	"github.com/toolkits/pkg/logger"
	"github.com/toolkits/pkg/slice"
)

type Set struct {
	sync.Mutex
	items map[string]struct{}
	ctx   *ctx.Context
}

func New(ctx *ctx.Context) *Set {
	set := &Set{
		items: make(map[string]struct{}),
		ctx:   ctx,
	}

	set.Init()
	return set
}

func (s *Set) Init() {
	go s.LoopPersist()
}

func (s *Set) MSet(items map[string]struct{}) {
	s.Lock()
	defer s.Unlock()
	for ident := range items {
		s.items[ident] = struct{}{}
	}
}

func (s *Set) LoopPersist() {
	for {
		time.Sleep(time.Second)
		s.persist()
	}
}

func (s *Set) persist() {
	var items map[string]struct{}

	s.Lock()
	if len(s.items) == 0 {
		s.Unlock()
		return
	}

	items = s.items
	s.items = make(map[string]struct{})
	s.Unlock()

	s.updateTimestamp(items)
}

func (s *Set) updateTimestamp(items map[string]struct{}) {
	lst := make([]string, 0, 100)
	now := time.Now().Unix()
	num := 0
	for ident := range items {
		lst = append(lst, ident)
		num++
		if num == 100 {
			if err := s.UpdateTargets(lst, now); err != nil {
				logger.Errorf("failed to update targets: %v", err)
			}
			lst = lst[:0]
			num = 0
		}
	}

	if err := s.UpdateTargets(lst, now); err != nil {
		logger.Errorf("failed to update targets: %v", err)
	}
}

type TargetUpdate struct {
	Lst []string `json:"lst"`
	Now int64    `json:"now"`
}

func (s *Set) UpdateTargets(lst []string, now int64) error {
	if !s.ctx.IsCenter {
		t := TargetUpdate{
			Lst: lst,
			Now: now,
		}
		err := poster.PostByUrls(s.ctx, "/v1/n9e/target-update", t)
		return err
	}

	count := int64(len(lst))
	if count == 0 {
		return nil
	}

	finder := zorm.NewUpdateFinder(models.TargetTableName).Append("update_at = ? WHERE ident in (?)", now, lst)
	rowsAffected, err := zorm.Transaction(s.ctx.Ctx, func(ctx context.Context) (interface{}, error) {
		return zorm.UpdateFinder(ctx, finder)
	})

	//ret := s.ctx.DB.Table("target").Where("ident in ?", lst).Update("update_at", now)
	if err != nil {
		return err
	}

	if rowsAffected == count {
		return nil
	}
	// there are some idents not found in db, so insert them
	var exists []string
	f := zorm.NewSelectFinder(models.TargetTableName, "ident").Append("WHERE ident in (?)", lst)
	err = zorm.Query(s.ctx.Ctx, f, &exists, nil)
	//err := s.ctx.DB.Table("target").Where("ident in ?", lst).Pluck("ident", &exists).Error
	if err != nil {
		return err
	}

	news := slice.SubString(lst, exists)
	data := make([]zorm.IEntityMap, 0, len(news))
	for i := 0; i < len(news); i++ {
		entity := zorm.NewEntityMap(models.TargetTableName)
		entity.Set("ident", news[i])
		entity.Set("update_at", now)
		data = append(data, entity)
		/*
			err = s.ctx.DB.Exec("INSERT INTO target(ident, update_at) VALUES(?, ?)", news[i], now).Error
			if err != nil {
				logger.Error("failed to insert target:", news[i], "error:", err)
			}
		*/
	}

	if len(data) < 1 {
		return nil
	}
	_, err = zorm.Transaction(s.ctx.Ctx, func(ctx context.Context) (interface{}, error) {
		return zorm.InsertEntityMapSlice(ctx, data)
	})
	if err != nil {
		logger.Error("failed to insert target:, error:", err)
	}
	return err
}
