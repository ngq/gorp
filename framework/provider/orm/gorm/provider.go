package gorm

import (
	"database/sql"
	"fmt"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite"
)

type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string { return "orm.gorm" }

func (p *Provider) IsDefer() bool { return false }

func (p *Provider) Provides() []string { return []string{datacontract.GormKey} }

func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(datacontract.GormKey, func(c runtimecontract.Container) (any, error) {
		cfgAny, err := c.Make(datacontract.ConfigKey)
		if err != nil {
			return nil, err
		}
		cfg := cfgAny.(datacontract.Config)

		var dbc datacontract.DBConfig
		if err := cfg.Unmarshal("database", &dbc); err != nil {
			return nil, err
		}
		if dbc.Driver == "" {
			return nil, fmt.Errorf("database.driver is required")
		}
		if dbc.DSN == "" {
			return nil, fmt.Errorf("database.dsn is required")
		}

		logAny, err := c.Make(observabilitycontract.LogKey)
		if err != nil {
			return nil, err
		}
		logger := logAny.(observabilitycontract.Logger)

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

		return db, nil
	}, true)
	return nil
}

func (p *Provider) Boot(runtimecontract.Container) error { return nil }
