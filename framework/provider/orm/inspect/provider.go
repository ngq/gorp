package inspect

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/ngq/gorp/framework/contract"

	"github.com/jmoiron/sqlx"
)

// Provider 把数据库结构探查器注册进容器。
//
// 中文说明：
// - 该服务主要供 `gorp model test/gen/api` 这类脚手架命令使用。
// - 当前实现底层依赖 SQLX，并根据 driver 提供表/列元信息读取能力。
type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string       { return "orm.inspect" }
func (p *Provider) IsDefer() bool      { return false }
func (p *Provider) Provides() []string { return []string{contract.DBInspectorKey} }

func (p *Provider) Register(c contract.Container) error {
	c.Bind(contract.DBInspectorKey, func(c contract.Container) (any, error) {
		v, err := c.Make(contract.SQLXKey)
		if err != nil {
			return nil, err
		}
		db := v.(*sqlx.DB)
		driver := db.DriverName()
		return &Service{db: db, driver: driver}, nil
	}, true)
	return nil
}

func (p *Provider) Boot(contract.Container) error { return nil }

type Service struct {
	db     *sqlx.DB
	driver string
}

func (s *Service) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

func (s *Service) Tables(ctx context.Context) ([]contract.Table, error) {
	switch s.driver {
	case "sqlite", "sqlite3":
		var names []string
		err := s.db.SelectContext(ctx, &names, `SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name`)
		if err != nil {
			return nil, err
		}
		out := make([]contract.Table, 0, len(names))
		for _, n := range names {
			out = append(out, contract.Table{Name: n})
		}
		return out, nil
	case "mysql":
		// 中文说明：
		// - MySQL 这里直接读取当前连接对应 schema（DATABASE()）下的基表。
		// - 先过滤掉 VIEW，避免 `model gen/api` 把视图也误当成普通数据表。
		var names []string
		err := s.db.SelectContext(ctx, &names, `
			SELECT table_name
			FROM information_schema.tables
			WHERE table_schema = DATABASE()
			  AND table_type = 'BASE TABLE'
			ORDER BY table_name
		`)
		if err != nil {
			return nil, err
		}
		out := make([]contract.Table, 0, len(names))
		for _, n := range names {
			out = append(out, contract.Table{Name: n})
		}
		return out, nil
	case "pgx":
		// 中文说明：
		// - PostgreSQL 这里读取当前 search_path 对应 schema 中的普通表。
		// - 默认采用 current_schema()，与大多数单 schema 项目一致，避免把系统 schema 混入结果。
		var names []string
		err := s.db.SelectContext(ctx, &names, `
			SELECT table_name
			FROM information_schema.tables
			WHERE table_schema = current_schema()
			  AND table_type = 'BASE TABLE'
			ORDER BY table_name
		`)
		if err != nil {
			return nil, err
		}
		out := make([]contract.Table, 0, len(names))
		for _, n := range names {
			out = append(out, contract.Table{Name: n})
		}
		return out, nil
	default:
		return nil, fmt.Errorf("tables not implemented for driver: %s", s.driver)
	}
}

func (s *Service) Columns(ctx context.Context, table string) ([]contract.Column, error) {
	switch s.driver {
	case "sqlite", "sqlite3":
		type row struct {
			CID     int            `db:"cid"`
			Name    string         `db:"name"`
			Type    string         `db:"type"`
			NotNull int            `db:"notnull"`
			Dflt    sql.NullString `db:"dflt_value"`
			PK      int            `db:"pk"`
		}
		rows := make([]row, 0)
		// NOTE: PRAGMA table_info doesn't support placeholders; quote the table name.
		q := fmt.Sprintf("PRAGMA table_info(%s)", quoteSQLiteIdent(table))
		if err := s.db.SelectContext(ctx, &rows, q); err != nil {
			return nil, err
		}
		out := make([]contract.Column, 0, len(rows))
		for _, r := range rows {
			var d *string
			if r.Dflt.Valid {
				v := r.Dflt.String
				d = &v
			}
			out = append(out, contract.Column{
				Name:       r.Name,
				Type:       r.Type,
				NotNull:    r.NotNull == 1,
				PrimaryKey: r.PK == 1,
				DefaultVal: d,
			})
		}
		return out, nil
	case "mysql":
		type row struct {
			Name          string         `db:"column_name"`
			Type          string         `db:"column_type"`
			Nullable      string         `db:"is_nullable"`
			Key           sql.NullString `db:"column_key"`
			DefaultVal    sql.NullString `db:"column_default"`
			Extra         sql.NullString `db:"extra"`
			Ordinal       int            `db:"ordinal_position"`
		}
		rows := make([]row, 0)
		if err := s.db.SelectContext(ctx, &rows, `
			SELECT
				column_name,
				column_type,
				is_nullable,
				column_key,
				column_default,
				extra,
				ordinal_position
			FROM information_schema.columns
			WHERE table_schema = DATABASE()
			  AND table_name = ?
			ORDER BY ordinal_position
		`, table); err != nil {
			return nil, err
		}
		out := make([]contract.Column, 0, len(rows))
		for _, r := range rows {
			var d *string
			if r.DefaultVal.Valid {
				v := r.DefaultVal.String
				d = &v
			}
			out = append(out, contract.Column{
				Name:       r.Name,
				Type:       r.Type,
				NotNull:    r.Nullable == "NO",
				PrimaryKey: r.Key.Valid && r.Key.String == "PRI",
				DefaultVal: d,
			})
		}
		return out, nil
	case "pgx":
		type row struct {
			Name       string         `db:"column_name"`
			Type       string         `db:"column_type"`
			Nullable   string         `db:"is_nullable"`
			DefaultVal sql.NullString `db:"column_default"`
			Ordinal    int            `db:"ordinal_position"`
			IsPrimary  bool           `db:"is_primary"`
		}
		rows := make([]row, 0)
		if err := s.db.SelectContext(ctx, &rows, `
			SELECT
				c.column_name,
				c.udt_name AS column_type,
				c.is_nullable,
				c.column_default,
				c.ordinal_position,
				EXISTS (
					SELECT 1
					FROM information_schema.table_constraints tc
					JOIN information_schema.key_column_usage kcu
					  ON tc.constraint_name = kcu.constraint_name
					 AND tc.table_schema = kcu.table_schema
					 AND tc.table_name = kcu.table_name
					WHERE tc.constraint_type = 'PRIMARY KEY'
					  AND tc.table_schema = c.table_schema
					  AND tc.table_name = c.table_name
					  AND kcu.column_name = c.column_name
				) AS is_primary
			FROM information_schema.columns c
			WHERE c.table_schema = current_schema()
			  AND c.table_name = $1
			ORDER BY c.ordinal_position
		`, table); err != nil {
			return nil, err
		}
		out := make([]contract.Column, 0, len(rows))
		for _, r := range rows {
			var d *string
			if r.DefaultVal.Valid {
				v := r.DefaultVal.String
				d = &v
			}
			out = append(out, contract.Column{
				Name:       r.Name,
				Type:       r.Type,
				NotNull:    r.Nullable == "NO",
				PrimaryKey: r.IsPrimary,
				DefaultVal: d,
			})
		}
		return out, nil
	default:
		return nil, fmt.Errorf("columns not implemented for driver: %s", s.driver)
	}
}

func quoteSQLiteIdent(name string) string {
	// SQLite identifiers can be quoted with double quotes.
	// We also escape embedded quotes.
	escaped := ""
	for _, r := range name {
		if r == '"' {
			escaped += "\"\""
			continue
		}
		escaped += string(r)
	}
	return "\"" + escaped + "\""
}
