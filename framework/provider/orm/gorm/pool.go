// Package gorm provides connection pool configuration utilities.
// The pool settings are applied to the underlying sql.DB for all database drivers.
//
// 本文件提供数据库连接池配置工具函数。
// 连接池设置应用于所有数据库驱动的底层 sql.DB。
package gorm

import (
	"database/sql"
	"fmt"
	"time"

	datacontract "github.com/ngq/gorp/framework/contract/data"
)

// applySQLDBPool applies connection pool settings to sql.DB.
// Settings include: MaxOpenConns, MaxIdleConns, ConnMaxLifetime, ConnMaxIdleTime.
// Core logic: Apply each setting only if it's configured (>0 or non-empty string).
//
// applySQLDBPool 将连接池设置应用到 sql.DB。
// 设置包括：最大打开连接数、最大空闲连接数、连接最大生命周期、连接最大空闲时间。
// 核心逻辑：仅对配置了的设置进行应用（>0 或非空字符串）。
func applySQLDBPool(sqlDB *sql.DB, dbc datacontract.DBConfig) error {
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
