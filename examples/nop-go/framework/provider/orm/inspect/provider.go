// Package inspect provides database schema inspection service.
// Supported drivers: sqlite/sqlite3, mysql, pgx (PostgreSQL).
// The service queries information_schema or PRAGMA tables to get schema metadata.
//
// 数据库 schema 检查服务包，提供表和列信息查询能力。
// 支持的驱动：sqlite/sqlite3, mysql, pgx (PostgreSQL)。
// 服务通过查询 information_schema 或 PRAGMA 表获取 schema 元数据。
// Eg:
//
//	// 注册 Provider（依赖 sqlx）
//	app.Register(inspect.NewProvider())
//
//	// 使用检查服务
//	inspector := c.MustMake(datacontract.DBInspectorKey).(datacontract.DBInspector)
//	tables, _ := inspector.Tables(ctx)
//	columns, _ := inspector.Columns(ctx, "users")
package inspect

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/ngq/gorp/framework/container"
	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"

	"github.com/jmoiron/sqlx"
)

// Provider registers the database inspector contract.
//
// Provider 注册数据库检查器契约。
type Provider struct{}

// NewProvider creates a new inspect provider instance.
//
// NewProvider 创建新的检查器 Provider 实例。
func NewProvider() *Provider { return &Provider{} }

// Name returns the provider name "orm.inspect".
//
// Name 返回 Provider 名称 "orm.inspect"。
func (p *Provider) Name() string { return "orm.inspect" }

// IsDefer returns false, inspector should be initialized immediately.
//
// IsDefer 返回 false，检查器应立即初始化。
func (p *Provider) IsDefer() bool { return false }

// Provides returns the DB inspector contract key.
//
// Provides 返回数据库检查器契约键。
func (p *Provider) Provides() []string { return []string{datacontract.DBInspectorKey} }

// DependsOn returns the keys this provider depends on.
// Inspect provider depends on SQLX for database connection.
//
// DependsOn 返回该 provider 依赖的 key。
// Inspect provider 依赖 SQLX 获取数据库连接。
func (p *Provider) DependsOn() []string { return []string{datacontract.SQLXKey} }

// Register binds the inspector service factory to the container.
// Note: This provider depends on SQLX provider being registered first.
//
// Register 将检查器服务工厂绑定到容器。
// 注意：此 Provider 依赖 SQLX Provider 先注册。
func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(datacontract.DBInspectorKey, func(c runtimecontract.Container) (any, error) {
		db, err := container.MakeWith[*sqlx.DB](c, datacontract.SQLXKey)
		if err != nil {
			return nil, err
		}
		driver := db.DriverName()
		return &Service{db: db, driver: driver}, nil
	}, true)
	return nil
}

// Boot is a no-op for inspect provider.
//
// Boot 检查器 Provider 无启动逻辑。
func (p *Provider) Boot(runtimecontract.Container) error { return nil }

// Service implements datacontract.DBInspector interface.
//
// Service 实现 datacontract.DBInspector 接口。
type Service struct {
	db *sqlx.DB // db is the underlying database connection.
	//
	// db 底层数据库连接。
	driver string // driver is the database driver name.
	//
	// driver 数据库驱动名称。
}

// Ping checks database connectivity.
//
// Ping 检查数据库连接是否正常。
func (s *Service) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

// Driver returns the database driver name.
//
// Driver 返回数据库驱动名称。
func (s *Service) Driver() string {
	return s.driver
}

// Tables returns all table names in the database.
// Core logic: Query driver-specific system tables (sqlite_master, information_schema.tables).
//
// Tables 返回数据库中所有表的名称。
// 核心逻辑：查询驱动特定的系统表（sqlite_master、information_schema.tables）。
func (s *Service) Tables(ctx context.Context) ([]datacontract.Table, error) {
	switch s.driver {
	case "sqlite", "sqlite3":
		var names []string
		err := s.db.SelectContext(ctx, &names, `SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name`)
		if err != nil {
			return nil, err
		}
		out := make([]datacontract.Table, 0, len(names))
		for _, n := range names {
			out = append(out, datacontract.Table{Name: n})
		}
		return out, nil
	case "mysql":
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
		out := make([]datacontract.Table, 0, len(names))
		for _, n := range names {
			out = append(out, datacontract.Table{Name: n})
		}
		return out, nil
	case "pgx", "postgres":
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
		out := make([]datacontract.Table, 0, len(names))
		for _, n := range names {
			out = append(out, datacontract.Table{Name: n})
		}
		return out, nil
	default:
		return nil, fmt.Errorf("tables not implemented for driver: %s", s.driver)
	}
}

// Columns returns all column info for a specific table.
// Core logic: Query driver-specific system tables (PRAGMA table_info, information_schema.columns).
//
// Columns 返回指定表的所有列信息。
// 核心逻辑：查询驱动特定的系统表（PRAGMA table_info、information_schema.columns）。
func (s *Service) Columns(ctx context.Context, table string) ([]datacontract.Column, error) {
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
		q := fmt.Sprintf("PRAGMA table_info(%s)", quoteSQLiteIdent(table))
		if err := s.db.SelectContext(ctx, &rows, q); err != nil {
			return nil, err
		}
		out := make([]datacontract.Column, 0, len(rows))
		for _, r := range rows {
			var d *string
			if r.Dflt.Valid {
				v := r.Dflt.String
				d = &v
			}
			out = append(out, datacontract.Column{
				Name:       r.Name,
				Type:       r.Type,
				NotNull:    r.NotNull == 1,
				PrimaryKey: r.PK == 1,
				DefaultVal: d,
				Comment:    "", // SQLite 原生不支持列注释，保留空字符串。
			})
		}
		return out, nil
	case "mysql":
		type row struct {
			Name       string         `db:"column_name"`
			Type       string         `db:"column_type"`
			Nullable   string         `db:"is_nullable"`
			Key        sql.NullString `db:"column_key"`
			DefaultVal sql.NullString `db:"column_default"`
			Extra      sql.NullString `db:"extra"`
			Ordinal    int            `db:"ordinal_position"`
			Comment    string         `db:"column_comment"`
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
				ordinal_position,
				column_comment
			FROM information_schema.columns
			WHERE table_schema = DATABASE()
			  AND table_name = ?
			ORDER BY ordinal_position
		`, table); err != nil {
			return nil, err
		}
		out := make([]datacontract.Column, 0, len(rows))
		for _, r := range rows {
			var d *string
			if r.DefaultVal.Valid {
				v := r.DefaultVal.String
				d = &v
			}
			out = append(out, datacontract.Column{
				Name:       r.Name,
				Type:       r.Type,
				NotNull:    r.Nullable == "NO",
				PrimaryKey: r.Key.Valid && r.Key.String == "PRI",
				DefaultVal: d,
				Comment:    r.Comment,
			})
		}
		return out, nil
	case "pgx", "postgres":
		type row struct {
			Name       string         `db:"column_name"`
			Type       string         `db:"column_type"`
			Nullable   string         `db:"is_nullable"`
			DefaultVal sql.NullString `db:"column_default"`
			Ordinal    int            `db:"ordinal_position"`
			IsPrimary  bool           `db:"is_primary"`
			Comment    string         `db:"column_comment"`
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
				) AS is_primary,
				COALESCE(col_description(cls.oid, c.ordinal_position::int), '') AS column_comment
			FROM information_schema.columns c
			JOIN pg_class cls ON cls.relname = c.table_name
			                  AND cls.relnamespace = (SELECT oid FROM pg_namespace WHERE nspname = c.table_schema)
			WHERE c.table_schema = current_schema()
			  AND c.table_name = $1
			ORDER BY c.ordinal_position
		`, table); err != nil {
			return nil, err
		}
		out := make([]datacontract.Column, 0, len(rows))
		for _, r := range rows {
			var d *string
			if r.DefaultVal.Valid {
				v := r.DefaultVal.String
				d = &v
			}
			out = append(out, datacontract.Column{
				Name:       r.Name,
				Type:       r.Type,
				NotNull:    r.Nullable == "NO",
				PrimaryKey: r.IsPrimary,
				DefaultVal: d,
				Comment:    r.Comment,
			})
		}
		return out, nil
	default:
		return nil, fmt.Errorf("columns not implemented for driver: %s", s.driver)
	}
}

// quoteSQLiteIdent quotes a SQLite identifier to handle special characters.
// Core logic: Escape double quotes by doubling them, then wrap in quotes.
//
// quoteSQLiteIdent 对 SQLite 标识符进行引号包裹，处理特殊字符。
// 核心逻辑：双引号转义为两个双引号，再用双引号包裹。
func quoteSQLiteIdent(name string) string {
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
