package model

import (
	"Goshop/utils/enum"
	"Goshop/utils/redis"
	"Goshop/utils/sql_factory"
	"Goshop/utils/yml_config"
	"database/sql"
	"log"
	"strings"
)

var (
	rds = redis.GetRedisClient()
)

// 创建一个数据库基类工厂
func CreateBaseSqlFactory(sqlType string) (res *BaseModel) {
	var dbType string
	var sqlDriverRead *sql.DB

	sqlType = strings.ToLower(strings.Replace(sqlType, " ", "", -1))
	sqlDriverWrite := sql_factory.GetOneSqlClient(sqlType, "Write")

	switch sqlType {
	case "mysql":
		dbType = "Mysql"
	case "sqlserver", "mssql":
		dbType = "SqlServer"
	case "postgre", "postgres", "postgresql":
		dbType = "PostgreSql"
	default:
		log.Println(enum.ErrorsDbDriverNotExists + sqlType)
		return nil
	}
	// 配置项是否开启读写分离
	isOpenReadDb := yml_config.CreateYamlFactory().GetInt(dbType + ".IsOpenReadDb")
	//开启读写分离配置，就继续初始化一个 Read 数据库连接
	if isOpenReadDb == 1 {
		sqlDriverRead = sql_factory.GetOneSqlClient(sqlType, "Read")
	} else {
		// 没有开启读写分离，那么 Read 数据库连接就是 Write 连接
		sqlDriverRead = sqlDriverWrite
	}
	return &BaseModel{dbDriverWrite: sqlDriverWrite, dbDriverRead: sqlDriverRead}

}

// 定义一个数据库操作的基本结构体
type BaseModel struct {
	dbDriverWrite *sql.DB
	dbDriverRead  *sql.DB
	stm           *sql.Stmt
}

//  执行类: 新增、更新、删除，  适合一次性执行完成就结束的操作
func (b *BaseModel) ExecuteSql(sql string, args ...interface{}) int64 {
	if stm, err := b.dbDriverWrite.Prepare(sql); err == nil {
		if res, err := stm.Exec(args...); err == nil {
			if affectNum, err := res.RowsAffected(); err == nil {
				return affectNum
			} else {
				log.Println(enum.ErrorsDbExecuteRunFail, err)
			}
		} else {
			log.Println(enum.ErrorsDbPrepareRunFail, err)
		}
	}
	return -1

}

//  查询类: select， 适合一次性查询完成就结束的操作
func (b *BaseModel) QuerySql(sql string, args ...interface{}) *sql.Rows {
	if stm, err := b.dbDriverRead.Prepare(sql); err == nil {
		if Rows, err := stm.Query(args...); err == nil {
			return Rows
		} else {
			log.Println(enum.ErrorsDbQueryRunFail, err)
		}
	} else {
		log.Println(enum.ErrorsDbPrepareRunFail, err)
	}
	return nil

}
func (b *BaseModel) QueryRow(sql string, args ...interface{}) *sql.Row {
	if stm, err := b.dbDriverRead.Prepare(sql); err == nil {
		return stm.QueryRow(args...)
	} else {
		log.Println(enum.ErrorsDbQueryRowRunFail, err)
		return nil
	}
}

//  预处理，主要针对有sql语句需要批量循环执行的场景，就必须独立预编译
// 批量执行sql，查询类和执行类其实不是很明确，这里我们直接定位在 Write 库，就能做到两者兼容
func (b *BaseModel) PrepareSql(sql string) bool {
	if stm, err := b.dbDriverWrite.Prepare(sql); err == nil {
		b.stm = stm
		return true
	} else {
		log.Println(enum.ErrorsDbPrepareRunFail, err)
		return false
	}
}

// 适合一次性预编译sql之后，批量操作sql，避免mysql产生大量的预编译sql无法释放
func (b *BaseModel) ExecuteSqlForMultiple(args ...interface{}) int64 {
	if res, err := b.stm.Exec(args...); err == nil {
		if affectNum, err := res.RowsAffected(); err == nil {
			return affectNum
		} else {
			log.Println(enum.ErrorsDbGetEffectiveRowsFail, err)
		}
	} else {
		log.Println(enum.ErrorsDbExecuteForMultipleFail, err)
	}
	return -1
}

// 适合预一次性预编译sql之后，批量操作sql，避免mysql产生大量的预编译sql无法释放
func (b *BaseModel) QuerySqlForMultiple(args ...interface{}) *sql.Rows {
	if Rows, err := b.stm.Query(args...); err == nil {
		return Rows
	} else {
		log.Println(enum.ErrorsDbQueryRunFail, err)
	}
	return nil
}

// 开启事物一个事务（Tx）,返回 *sql.Tx， 提交 调用  Commit ， 回滚调用 Rollback
func (b *BaseModel) BeginTx() *sql.Tx {
	if tx, err := b.dbDriverWrite.Begin(); err == nil {
		return tx
	} else {
		log.Println(enum.ErrorsDbTransactionBeginFail + err.Error())
	}
	return nil
}
