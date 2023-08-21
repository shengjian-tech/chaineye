package ormx

import (
	"context"
	"strings"

	"gitee.com/chunanyong/zorm"
	// 引入数据库驱动
	_ "github.com/ClickHouse/clickhouse-go/v2"
	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/sijms/go-ora/v2"
	_ "github.com/taosdata/driver-go/v3/taosRestful"
	"github.com/toolkits/pkg/logger"
)

// DBConfig zorm.DataSourceConfig
type DBConfig struct {
	// DSN 数据库的连接字符串,parseTime=true会自动转换为time格式,默认查询出来的是[]byte数组.&loc=Local用于设置时区
	DSN string
	// DriverName 数据库驱动名称:mysql,postgres,oracle(go-ora),sqlserver,sqlite3,go_ibm_db,clickhouse,dm,kingbase,aci,taosSql|taosRestful 和Dialect对应
	// sql.Open(DriverName,DSN) DriverName就是驱动的sql.Open第一个字符串参数,根据驱动实际情况获取
	DriverName string
	// Dialect 数据库方言:mysql,postgresql,oracle,mssql,sqlite,db2,clickhouse,dm,kingbase,shentong,tdengine 和 DriverName 对应
	Dialect      string
	MaxLifetime  int
	MaxOpenConns int
	MaxIdleConns int
	// SlowSQLMillis 慢sql的时间阈值,单位毫秒.小于0是禁用SQL语句输出;等于0是只输出SQL语句,不计算执行时间;大于0是计算SQL执行时间,并且>=SlowSQLMillis值
	SlowSQLMillis int
	// DisableTransaction 禁用事务,默认false,如果设置了DisableTransaction=true,Transaction方法失效,不再要求有事务,为了处理某些数据库不支持事务,比如TDengine
	// 禁用事务应该有驱动伪造事务API,不应该有orm实现,clickhouse的驱动就是这样做的
	DisableTransaction bool
	// TDengineInsertsColumnName TDengine批量insert语句中是否有列名.默认false没有列名,插入值和数据库列顺序保持一致,减少语句长度
	TDengineInsertsColumnName bool
}

// New Create zorm.DBDao instance
func New(c DBConfig) (*zorm.DBDao, error) {
	dbDaoConfig := zorm.DataSourceConfig{
		DSN:                       c.DSN,
		DriverName:                c.DriverName,
		Dialect:                   c.Dialect,
		MaxOpenConns:              c.MaxOpenConns,
		MaxIdleConns:              c.MaxIdleConns,
		ConnMaxLifetimeSecond:     c.MaxLifetime,
		SlowSQLMillis:             c.SlowSQLMillis,
		DisableTransaction:        c.DisableTransaction,
		TDengineInsertsColumnName: c.TDengineInsertsColumnName,
	}
	db, err := zorm.NewDBDao(&dbDaoConfig)

	//注册达梦TEXT类型转string插件,dialectColumnType 值是 Dialect.字段类型 ,例如 dm.TEXT
	if strings.ToLower(c.Dialect) == "dm" {
		zorm.RegisterCustomDriverValueConver("dm.TEXT", CustomDMText{})
	}
	zorm.FuncLogError = zormFuncLogError // 记录异常日志的函数
	zorm.FuncLogPanic = zormFuncLogError
	zorm.FuncPrintSQL = zormPrintSQL // 打印sql的函数
	return db, err
}

func zormFuncLogError(ctx context.Context, err error) {
	//log.Output(LogCallDepth, fmt.Sprintln(err))
	logger.LogDepth(logger.ERROR, zorm.LogCallDepth, "zorm exec error:%v", err)
}

func zormPrintSQL(ctx context.Context, sqlstr string, args []interface{}, execSQLMillis int64) {
	if args != nil {
		logger.LogDepth(logger.INFO, zorm.LogCallDepth, "sql:%s,args:%v,execSQLMillis:%d", sqlstr, args, execSQLMillis)
	} else {
		logger.LogDepth(logger.INFO, zorm.LogCallDepth, "sql:%s,args: [] ,execSQLMillis:%d", sqlstr, execSQLMillis)
	}
}
