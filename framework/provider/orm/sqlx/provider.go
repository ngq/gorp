package sqlx

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/ngq/gorp/framework/contract"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

// Provider 把 SQLX 数据库连接注册进容器。
//
// 中文说明：
// - 与 orm.gorm 并存，面向更适合手写 SQL 的场景。
// - 对外暴露 contract.SQLXKey，供列表查询、报表查询等直接使用。
// - 从 framework 抽离视角看，这里不再默认补 sqlite/demo.db；
//   provider 只消费已经明确提供的 database 配置。
type Provider struct{}

// NewProvider 创建 sqlx provider。
func NewProvider() *Provider { return &Provider{} }

// Name 返回 provider 名称。
func (p *Provider) Name() string { return "orm.sqlx" }

// IsDefer 表示 sqlx provider 不走延迟加载。
func (p *Provider) IsDefer() bool {
	return false
}

// Provides 返回 sqlx provider 暴露的能力 key。
func (p *Provider) Provides() []string { return []string{contract.SQLXKey} }

// Register 绑定统一 SQLX 数据库连接。
func (p *Provider) Register(c contract.Container) error {
	c.Bind(contract.SQLXKey, func(c contract.Container) (any, error) {
		cfgAny, err := c.Make(contract.ConfigKey)
		if err != nil {
			return nil, err
		}
		cfg := cfgAny.(contract.Config)

		var dbc contract.DBConfig
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

// Boot sqlx provider 无额外启动逻辑。
func (p *Provider) Boot(contract.Container) error { return nil }

// normalizeDriver 把配置中的 driver 名归一化为 sqlx.Open 可识别的 driver 名称。
//
// 中文说明：
// - framework 对外沿用统一 database.driver 语义；
// - 这里负责把业务更常见的 postgres/postgresql/pgsql 归一到 pgx。
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
