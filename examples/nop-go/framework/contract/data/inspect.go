// Application scenarios:
// - Define lightweight schema inspection contracts for databases used by the framework.
// - Expose table and column metadata in a backend-neutral way.
// - Support health checks and schema introspection tooling without binding to one ORM.
//
// 适用场景：
// - 定义框架使用的轻量数据库结构探查契约。
// - 以与后端无关的方式暴露表和列元数据。
// - 在不绑定单一 ORM 的情况下支持健康检查和结构探查工具。
package data

import "context"

// DBInspectorKey is the container key for the database inspector capability.
//
// DBInspectorKey 是数据库探查能力的容器键。
const DBInspectorKey = "framework.orm.inspector"

// Table describes a database table name.
//
// Table 描述数据库表名。
type Table struct {
	Name string
}

// Column describes basic database column metadata.
//
// Column 描述数据库列的基础元数据。
type Column struct {
	Name       string
	Type       string
	NotNull    bool
	PrimaryKey bool
	DefaultVal *string
	Comment    string // Column comment from database schema.
	//
	// Column comment 数据库表的列注释。
}

// DBInspector defines database inspection operations.
//
// DBInspector 定义数据库探查操作。
type DBInspector interface {
	// Ping checks whether the database connection is available.
	//
	// Ping 检查数据库连接是否可用。
	Ping(ctx context.Context) error

	// Tables lists visible tables in the current database.
	//
	// Tables 列出当前数据库中可见的表。
	Tables(ctx context.Context) ([]Table, error)

	// Columns lists visible columns of a given table.
	//
	// Columns 列出指定表中可见的列。
	Columns(ctx context.Context, table string) ([]Column, error)
}
