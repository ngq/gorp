// Package sqlx provides SQLX ORM integration for gorp framework.
// SQLX is a lighter alternative to GORM, suitable for scenarios requiring raw SQL.
// Supported drivers: sqlite/sqlite3, mysql, postgres/postgresql/pgsql (via pgx).
// Configuration via config.yaml:
//
// SQLX ORM 包，提供基于 jmoiron/sqlx 的轻量级数据库操作能力。
// SQLX 是比 GORM 更轻量的选择，适用于需要原生 SQL 的场景。
// 支持的驱动：sqlite/sqlite3, mysql, postgres/postgresql/pgsql（通过 pgx）。
// 通过 config.yaml 配置：
//
//	database:
//	  driver: mysql
//	  dsn: "user:password@tcp(localhost:3306)/dbname?charset=utf8mb4"
//	  max_open_conns: 100
//	  max_idle_conns: 10
//	  conn_max_lifetime: "30m"
//
// Eg:
//
//	// 注册 Provider
//	app.Register(sqlx.NewProvider())
//
//	// 使用 SQLX
//	db := c.MustMake(datacontract.SQLXKey).(*sqlx.DB)
//	db.Select(&users, "SELECT * FROM users")
package sqlx

import (
	"database/sql"
	"fmt"
	"time"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

// Provider registers the SQLX DB contract.
//
// Provider 注册 SQLX 数据库契约。
type Provider struct{}

// NewProvider creates a new SQLX provider instance.
//
// NewProvider 创建新的 SQLX Provider 实例。
func NewProvider() *Provider { return &Provider{} }

// Name returns the provider name "orm.sqlx".
//
// Name 返回 Provider 名称 "orm.sqlx"。
func (p *Provider) Name() string { return "orm.sqlx" }

// IsDefer returns false, SQLX should be initialized immediately for DB connection.
//
// IsDefer 返回 false，SQLX 应立即初始化以建立数据库连接。
func (p *Provider) IsDefer() bool { return false }

// Provides returns the SQLX contract key.
//
// Provides 返回 SQLX 契约键。
func (p *Provider) Provides() []string { return []string{datacontract.SQLXKey} }

// Register binds the SQLX DB factory to the container.
// Core logic: Parse config, normalize driver, open connection, apply pool settings.
//
// Register 将 SQLX 数据库工厂绑定到容器。
// 核心逻辑：解析配置、标准化驱动名称、打开连接、应用连接池设置。
func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(datacontract.SQLXKey, func(c runtimecontract.Container) (any, error) {
		cfgAny, err := c.Make(datacontract.ConfigKey)
		if err != nil {
			return nil, err
		}
		cfg := cfgAny.(datacontract.Config)

		var dbc datacontract.DBConfig
		if err := cfg.Unmarshal("database", &dbc); err != nil {
			return nil, err
		}

		driver := dbc.Driver
		dsn := dbc.DSN
		if driver == "" {
			return nil, fmt.Errorf("database.driver is required")
		}
		if dsn == "" {
			return nil, fmt.Errorf("database.dsn is required")
		}

		if normalized, err := normalizeDriver(driver); err != nil {
			return nil, err
		} else {
			driver = normalized
		}

		if driver == "pgx" {
			_ = stdlib.Driver{}
		}

		db, err := sqlx.Open(driver, dsn)
		if err != nil {
			return nil, err
		}
		if dbc.MaxOpenConns > 0 {
			db.SetMaxOpenConns(dbc.MaxOpenConns)
		}
		if dbc.MaxIdleConns > 0 {
			db.SetMaxIdleConns(dbc.MaxIdleConns)
		}
		if dbc.ConnMaxLifetime != "" {
			d, err := time.ParseDuration(dbc.ConnMaxLifetime)
			if err != nil {
				return nil, fmt.Errorf("invalid database.conn_max_lifetime %q: %w", dbc.ConnMaxLifetime, err)
			}
			db.SetConnMaxLifetime(d)
		}
		if dbc.ConnMaxIdleTime != "" {
			d, err := time.ParseDuration(dbc.ConnMaxIdleTime)
			if err != nil {
				return nil, fmt.Errorf("invalid database.conn_max_idletime %q: %w", dbc.ConnMaxIdleTime, err)
			}
			db.SetConnMaxIdleTime(d)
		}
		if err := db.Ping(); err != nil {
			return nil, err
		}
		return db, nil
	}, true)
	return nil
}

// Boot is a no-op for SQLX provider.
//
// Boot SQLX Provider 无启动逻辑。
func (p *Provider) Boot(runtimecontract.Container) error { return nil }

// normalizeDriver converts driver aliases to standard driver names.
// Mapping: sqlite/sqlite3 -> sqlite, postgres/postgresql/pgsql -> pgx.
//
// normalizeDriver 将驱动别名转换为标准驱动名称。
// 映射：sqlite/sqlite3 -> sqlite, postgres/postgresql/pgsql -> pgx。
func normalizeDriver(driver string) (string, error) {
	switch driver {
	case "sqlite", "sqlite3":
		return "sqlite", nil
	case "mysql":
		return "mysql", nil
	case "postgres", "postgresql", "pgsql":
		return "pgx", nil
	default:
		return "", fmt.Errorf("unknown db driver: %s", driver)
	}
}

var _ = sql.ErrNoRows