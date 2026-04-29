package gorm

import (
	"database/sql"
	"fmt"

	"github.com/ngq/gorp/framework/contract"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite"
)

// Provider 把 GORM 数据库连接注册进容器。
//
// 中文说明：
// - 对外暴露 contract.GormKey，业务层通过容器获取 *gorm.DB。
// - 这里负责把统一 database 配置翻译成不同驱动的 dialector。
// - 同时会接入框架自己的日志实现与连接池参数。
// - 从 framework 抽离视角看，这里不再偷偷发明 sqlite/demo.db 默认值；
//   provider 只消费已经明确提供的 database 配置。
type Provider struct{}

// NewProvider 创建 gorm provider。
func NewProvider() *Provider { return &Provider{} }

// Name 返回 provider 名称。
func (p *Provider) Name() string { return "orm.gorm" }

// IsDefer 表示 gorm provider 不走延迟加载。
func (p *Provider) IsDefer() bool {
	return false
}

// Provides 返回 gorm provider 暴露的能力 key。
func (p *Provider) Provides() []string { return []string{contract.GormKey} }

// Register 绑定统一 GORM 数据库连接。
func (p *Provider) Register(c contract.Container) error {
	c.Bind(contract.GormKey, func(c contract.Container) (any, error) {
		cfgAny, err := c.Make(contract.ConfigKey)
		if err != nil {
			return nil, err
		}
		cfg := cfgAny.(contract.Config)

		var dbc contract.DBConfig
		if err := cfg.Unmarshal("database", &dbc); err != nil {
			return nil, err
		}
		if dbc.Driver == "" {
			return nil, fmt.Errorf("database.driver is required")
		}
		if dbc.DSN == "" {
			return nil, fmt.Errorf("database.dsn is required")
		}

		logAny, err := c.Make(contract.LogKey)
		if err != nil {
			return nil, err
		}
		logger := logAny.(contract.Logger)

		var (
			dialector gorm.Dialector
			sqlDB    *sql.DB
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

		// 启动数据库指标收集
		// 中文说明：
		// - 收集连接池指标（Open/InUse/Idle/Wait）；
		// - 通过 GORM callback 记录查询耗时；
		// - 指标通过 /metrics 端点暴露给 Prometheus。
		collector := NewDBMetricsCollector(sqlDB, dbc.Driver)
		collector.StartCollection()
		GormQueryCallback(db, dbc.Driver)

		return db, nil
	}, true)
	return nil
}

// Boot gorm provider 无额外启动逻辑。
func (p *Provider) Boot(contract.Container) error { return nil }
