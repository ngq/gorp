package data

import (
	"context"
	"database/sql"

	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

const (
	GormKey             = "framework.orm.gorm"
	SQLXKey             = "framework.orm.sqlx"
	ORMBackendKey       = "framework.orm.backend"
	DBRuntimeKey        = "framework.orm.runtime"
	EntClientKey        = "framework.orm.ent"
	EntClientFactoryKey = "framework.orm.ent.factory"
	MigratorKey         = "framework.orm.migrator"
	SQLExecutorKey      = "framework.orm.sql_executor"
)

type DBConfig struct {
	Driver  string `mapstructure:"driver"`
	Backend string `mapstructure:"backend"`
	DSN     string `mapstructure:"dsn"`

	MaxOpenConns int `mapstructure:"max_open_conns"`
	MaxIdleConns int `mapstructure:"max_idle_conns"`

	ConnMaxLifetime string `mapstructure:"conn_max_lifetime"`
	ConnMaxIdleTime string `mapstructure:"conn_max_idletime"`
}

type RuntimeBackend string

const (
	RuntimeBackendGorm RuntimeBackend = "gorm"
	RuntimeBackendSQLX RuntimeBackend = "sqlx"
	RuntimeBackendEnt  RuntimeBackend = "ent"
)

func NormalizeBackendName(name string) RuntimeBackend {
	switch name {
	case "", string(RuntimeBackendGorm):
		return RuntimeBackendGorm
	case string(RuntimeBackendSQLX):
		return RuntimeBackendSQLX
	case string(RuntimeBackendEnt):
		return RuntimeBackendEnt
	default:
		return RuntimeBackendGorm
	}
}

type EntClientFactory interface {
	CreateEntClient(c runtimecontract.Container) (any, error)
}

type Migrator interface {
	AutoMigrate(dst ...any) error
}

type SQLExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type GormDB interface {
	WithContext(ctx context.Context) any
}

type SQLX interface {
	ExecContext(ctx context.Context, query string, args ...any) (any, error)
}
