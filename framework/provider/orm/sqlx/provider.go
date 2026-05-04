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

type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string { return "orm.sqlx" }

func (p *Provider) IsDefer() bool { return false }

func (p *Provider) Provides() []string { return []string{datacontract.SQLXKey} }

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

func (p *Provider) Boot(runtimecontract.Container) error { return nil }

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
