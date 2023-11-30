package models

import (
	"context"

	"gitee.com/chunanyong/zorm"
	"github.com/ccfos/nightingale/v6/pkg/ctx"
	"github.com/toolkits/pkg/slice"
)

const RoleOperationTableName = "role_operation"

type RoleOperation struct {
	// 引入默认的struct,隔离IEntityStruct的方法改动
	zorm.EntityStruct
	Id        int64  `column:"id"`
	RoleName  string `column:"role_name"`
	Operation string `column:"operation"`
}

func (ro *RoleOperation) GetTableName() string {
	return RoleOperationTableName
}

func (r *RoleOperation) DB2FE() error {
	return nil
}

func RoleHasOperation(ctx *ctx.Context, roles []string, operation string) (bool, error) {
	if len(roles) == 0 {
		return false, nil
	}
	finder := zorm.NewSelectFinder(RoleOperationTableName, "count(*)").Append("WHERE operation = ? and role_name in (?)", operation, roles)
	return Exists(ctx, finder)
	//return Exists(DB(ctx).Model(&RoleOperation{}).Where("operation = ? and role_name in ?", operation, roles))
}

func OperationsOfRole(ctx *ctx.Context, roles []string) ([]string, error) {
	finder := zorm.NewSelectFinder(RoleOperationTableName, "DISTINCT operation")
	//session := DB(ctx).Model(&RoleOperation{}).Select("distinct(operation) as operation")

	if !slice.ContainsString(roles, AdminRole) {
		//session = session.Where("role_name in ?", roles)
		finder.Append("WHERE role_name in (?)", roles)
	}

	ret := make([]string, 0)
	err := zorm.Query(ctx.Ctx, finder, &ret, nil)
	//err := session.Pluck("operation", &ret).Error
	return ret, err
}

func RoleOperationBind(ctx *ctx.Context, roleName string, operation []string) error {

	_, err := zorm.Transaction(ctx.Ctx, func(ctx context.Context) (interface{}, error) {
		f1 := zorm.NewDeleteFinder(RoleOperationTableName).Append("WHERE role_name = ?", roleName)
		_, err := zorm.UpdateFinder(ctx, f1)
		if err != nil {
			return nil, err
		}
		if len(operation) == 0 {
			return nil, err
		}
		ops := make([]zorm.IEntityStruct, 0)
		for _, op := range operation {
			ops = append(ops, &RoleOperation{
				RoleName:  roleName,
				Operation: op,
			})
		}
		_, err = zorm.InsertSlice(ctx, ops)

		return nil, err
	})
	return err

	/*
		tx := DB(ctx).Begin()

		if err := tx.Where("role_name = ?", roleName).Delete(&RoleOperation{}).Error; err != nil {
			tx.Rollback()
			return err
		}

		if len(operation) == 0 {
			return tx.Commit().Error
		}

		var ops []RoleOperation
		for _, op := range operation {
			ops = append(ops, RoleOperation{
				RoleName:  roleName,
				Operation: op,
			})
		}

		if err := tx.Create(&ops).Error; err != nil {
			tx.Rollback()
			return err
		}

		return tx.Commit().Error
	*/
}
