package gorm

import (
	"database/sql"
	"fmt"
	"time"

	datacontract "github.com/ngq/gorp/framework/contract/data"
)

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
