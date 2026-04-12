package gorm

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/ngq/gorp/framework/contract"
)

// applySQLDBPool 把框架统一配置写入 database/sql 连接池。
//
// 中文说明：
// - GORM 最终仍然落在 *sql.DB 之上，因此连接池参数要配置到 sql.DB。
// - 这里专门拆成独立函数，避免 provider.go 中驱动选择逻辑和连接池逻辑混在一起。
// - 对 sqlite/mysql/postgres 等驱动都可以复用。
func applySQLDBPool(sqlDB *sql.DB, dbc contract.DBConfig) error {
	if sqlDB == nil {
		return nil
	}
	if dbc.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(dbc.MaxOpenConns)
	}
	if dbc.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(dbc.MaxIdleConns)
	}
	if dbc.ConnMaxLifetime != "" {
		d, err := time.ParseDuration(dbc.ConnMaxLifetime)
		if err != nil {
			return fmt.Errorf("invalid database.conn_max_lifetime %q: %w", dbc.ConnMaxLifetime, err)
		}
		sqlDB.SetConnMaxLifetime(d)
	}
	if dbc.ConnMaxIdleTime != "" {
		d, err := time.ParseDuration(dbc.ConnMaxIdleTime)
		if err != nil {
			return fmt.Errorf("invalid database.conn_max_idletime %q: %w", dbc.ConnMaxIdleTime, err)
		}
		sqlDB.SetConnMaxIdleTime(d)
	}
	return nil
}
