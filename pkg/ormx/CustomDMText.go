package ormx

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"reflect"
	"strconv"

	"gitee.com/chunanyong/dm"
)

// CustomDMText 实现ICustomDriverValueConver接口,扩展自定义类型,例如 达梦数据库TEXT类型,映射出来的是dm.DmClob类型,无法使用string类型直接接收
type CustomDMText struct{}

// GetDriverValue 根据数据库列类型,返回driver.Value的实例,struct属性类型
// map接收或者字段不存在,无法获取到structFieldType,会传入nil
func (dmtext CustomDMText) GetDriverValue(ctx context.Context, columnType *sql.ColumnType, structFieldType *reflect.Type) (driver.Value, error) {
	// 如果需要使用structFieldType,需要先判断是否为nil
	// if structFieldType != nil {
	// }

	return &dm.DmClob{}, nil
}

// ConverDriverValue 数据库列类型,GetDriverValue返回的driver.Value的临时接收值,struct属性类型
// map接收或者字段不存在,无法获取到structFieldType,会传入nil
// 返回符合接收类型值的指针,指针,指针!!!!
func (dmtext CustomDMText) ConverDriverValue(ctx context.Context, columnType *sql.ColumnType, tempDriverValue driver.Value, structFieldType *reflect.Type) (interface{}, error) {
	// 如果需要使用structFieldType,需要先判断是否为nil
	// if structFieldType != nil {
	// }

	// 类型转换
	dmClob, isok := tempDriverValue.(*dm.DmClob)
	if !isok {
		return tempDriverValue, errors.New("->ConverDriverValue-->转换至*dm.DmClob类型失败")
	}
	if dmClob == nil || !dmClob.Valid {
		return new(string), nil
	}
	// 获取长度
	dmlen, errLength := dmClob.GetLength()
	if errLength != nil {
		return dmClob, errLength
	}

	// int64转成int类型
	strInt64 := strconv.FormatInt(dmlen, 10)
	dmlenInt, errAtoi := strconv.Atoi(strInt64)
	if errAtoi != nil {
		return dmClob, errAtoi
	}

	// 读取字符串
	str, errReadString := dmClob.ReadString(1, dmlenInt)

	// 处理空字符串或NULL造成的EOF错误
	if errReadString == io.EOF {
		return new(string), nil
	}

	return &str, errReadString
}
