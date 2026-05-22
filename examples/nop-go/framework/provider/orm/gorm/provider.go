// Package gorm provides GORM ORM integration for gorp framework.
// Supported drivers: sqlite/sqlite3, mysql, postgres/postgresql/pgsql.
// Configuration via config.yaml:
//
// GORM ORM 包，提供基于 GORM 的数据库 ORM 能力。
// 支持的驱动：sqlite/sqlite3, mysql, postgres/postgresql/pgsql。
// 通过 config.yaml 配置：
//
//	database:
//	  driver: mysql
//	  dsn: "user:password@tcp(localhost:3306)/dbname?charset=utf8mb4"
//	  max_open_conns: 100
//	  max_idle_conns: 10
//	  conn_max_lifetime: "30m"
//	  conn_max_idletime: "10m"
//
// Eg:
//
//	// 注册 Provider
//	app.Register(gorm.NewProvider())
//
//	// 使用 GORM
//	db := c.MustMake(datacontract.GormKey).(*gorm.DB)
//	db.Find(&users)
package gorm

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/ngq/gorp/framework/container"
	datacontract "github.com/ngq/gorp/framework/contract/data"
	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite"
)

// closerFunc 将 func() error 适配为 io.Closer 接口
type closerFunc func() error

func (f closerFunc) Close() error { return f() }

// Provider registers the GORM DB contract.
//
// Provider 注册 GORM 数据库契约。
type Provider struct{}

// NewProvider creates a new GORM provider instance.
//
// NewProvider 创建新的 GORM Provider 实例。
func NewProvider() *Provider { return &Provider{} }

// Name returns the provider name "orm.gorm".
//
// Name 返回 Provider 名称 "orm.gorm"。
func (p *Provider) Name() string { return "orm.gorm" }

// IsDefer returns false, GORM should be initialized immediately for DB connection.
//
// IsDefer 返回 false，GORM 应立即初始化以建立数据库连接。
func (p *Provider) IsDefer() bool { return false }

// Provides returns the GORM contract key.
//
// Provides 返回 GORM 契约键。
func (p *Provider) Provides() []string { return []string{datacontract.GormKey} }

// DependsOn returns the keys this provider depends on.
// GORM provider depends on Config for database configuration.
//
// DependsOn 返回该 provider 依赖的 key。
// GORM provider 依赖 Config 获取数据库配置。
func (p *Provider) DependsOn() []string { return []string{datacontract.ConfigKey} }

// Register binds the GORM DB factory to the container.
// Core logic: Parse config, create dialector, open GORM, apply pool settings, start metrics collector.
//
// Register 将 GORM 数据库工厂绑定到容器。
// 核心逻辑：解析配置、创建 dialector、打开 GORM、应用连接池设置、启动指标采集器。
func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(datacontract.GormKey, func(c runtimecontract.Container) (any, error) {
		cfg, err := container.MakeWith[datacontract.Config](c, datacontract.ConfigKey)
		if err != nil {
			return nil, err
		}

		var dbc datacontract.DBConfig
		if err := cfg.Unmarshal("database", &dbc); err != nil {
			return nil, err
		}
		if dbc.Driver == "" {
			return nil, errors.New("database.driver is required")
		}
		if dbc.DSN == "" {
			return nil, errors.New("database.dsn is required")
		}

		logger, err := container.MakeWith[observabilitycontract.Logger](c, observabilitycontract.LogKey)
		if err != nil {
			return nil, err
		}

		var (
			dialector gorm.Dialector
			sqlDB     *sql.DB
		)
		switch dbc.Driver {
		case "sqlite", "sqlite3":
			conn, err := sql.Open("sqlite", dbc.DSN)
			if err != nil {
				return nil, err
			}
			sqlDB = conn
			if err := applySQLDBPool(sqlDB, dbc); err != nil {
				return nil, err
			}
			dialector = sqlite.Dialector{Conn: conn}
		case "mysql":
			dialector = mysql.Open(dbc.DSN)
		case "postgres", "postgresql", "pgsql":
			dialector = postgres.Open(dbc.DSN)
		default:
			return nil, fmt.Errorf("unknown db driver: %s", dbc.Driver)
		}

		db, err := gorm.Open(dialector, &gorm.Config{Logger: newGormLogger(logger)})
		if err != nil {
			return nil, err
		}

		if sqlDB == nil {
			sqlDB, err = db.DB()
			if err != nil {
				return nil, err
			}
		}
		if err := applySQLDBPool(sqlDB, dbc); err != nil {
			return nil, err
		}
		if err := sqlDB.Ping(); err != nil {
			return nil, err
		}

		collector := NewDBMetricsCollector(sqlDB, dbc.Driver)
		collector.StartCollection()
		GormQueryCallback(db, dbc.Driver)

		// 注册 closer：关闭 DBMetricsCollector 和 sql.DB 连接
		c.RegisterCloser("gorm.db", closerFunc(func() error {
			collector.Stop()
			return sqlDB.Close()
		}))

		return db, nil
	}, true)
	return nil
}

// Boot is a no-op for GORM provider.
//
// Boot GORM Provider 无启动逻辑（初始化在 Register 中完成）。
func (p *Provider) Boot(runtimecontract.Container) error { return nil }
