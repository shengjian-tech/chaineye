package models

import "gitee.com/chunanyong/zorm"

//XuperStructTableName 表名常量,方便直接调用
const XuperDayTxStructTableName = "t_xuper_day_tx"

// XuperStruct
type XuperDayTx struct {
	//引入默认的struct,隔离IEntityStruct的方法改动
	zorm.EntityStruct

	// 主键
	Id string `column:"id"`

	// 日期
	Day string `column:"block_day"`

	TotalTxCount int64 `column:"total_tx_count"`

	DayTxCount int64 `column:"day_tx_count"`

	//ChainName 当前链名 默认 xuper
	ChainName string `column:"chain_name"`

	// 区块链网络名
	RootNet string `column:"rootnet"`

	//------------------数据库字段结束,自定义字段写在下面---------------//
	//如果查询的字段在column tag中没有找到,就会根据名称(不区分大小写,支持 _ 转驼峰)映射到struct的属性上

}

//GetTableName 获取表名称
//IEntityStruct 接口的方法,实体类需要实现!!!
func (entity *XuperDayTx) GetTableName() string {
	return XuperDayTxStructTableName
}

//GetPKColumnName 获取数据库表的主键字段名称.因为要兼容Map,只能是数据库的字段名称
//不支持联合主键,变通认为无主键,业务控制实现(艰难取舍)
//如果没有主键,也需要实现这个方法, return "" 即可
//IEntityStruct 接口的方法,实体类需要实现!!!
func (entity *XuperDayTx) GetPKColumnName() string {
	//如果没有主键
	//return ""
	return "id"
}
