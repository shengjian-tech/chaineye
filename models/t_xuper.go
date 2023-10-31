package models

import (
	"gitee.com/chunanyong/zorm"
)

//XuperStructTableName 表名常量,方便直接调用
const XuperStructTableName = "t_xuper"

// XuperStruct
type XuperStruct struct {
	//引入默认的struct,隔离IEntityStruct的方法改动
	zorm.EntityStruct

	// 主键
	Id string `column:"id"`

	//BlockHeight 区块高度
	BlockHeight int64 `column:"block_height"`

	//BlockHash 区块id  即区块hash
	BlockHash string `column:"block_hash"`

	//BlockTxCount 区块内的交易总数
	BlockTxCount int64 `column:"block_tx_count"`

	//TotalTxCount 当前区块链上总交易数
	TotalTxCount int64 `column:"total_tx_count"`

	//Timestamp 当前区块产生时间
	Timestamp int64 `column:"timestamp"`

	//ChainName 当前链名 默认 xuper
	ChainName string `column:"chain_name"`

	// 区块链网络名
	RootNet string `column:"rootnet"`

	//------------------数据库字段结束,自定义字段写在下面---------------//
	//如果查询的字段在column tag中没有找到,就会根据名称(不区分大小写,支持 _ 转驼峰)映射到struct的属性上

}

//GetTableName 获取表名称
//IEntityStruct 接口的方法,实体类需要实现!!!
func (entity *XuperStruct) GetTableName() string {
	return XuperStructTableName
}

//GetPKColumnName 获取数据库表的主键字段名称.因为要兼容Map,只能是数据库的字段名称
//不支持联合主键,变通认为无主键,业务控制实现(艰难取舍)
//如果没有主键,也需要实现这个方法, return "" 即可
//IEntityStruct 接口的方法,实体类需要实现!!!
func (entity *XuperStruct) GetPKColumnName() string {
	//如果没有主键
	//return ""
	return "id"
}
