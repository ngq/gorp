package contract

import "context"

const DBInspectorKey = "framework.orm.inspector"

type Table struct {
	Name string
}

type Column struct {
	Name       string
	Type       string
	NotNull    bool
	PrimaryKey bool
	DefaultVal *string
}

// DBInspector provides schema inspection APIs for code generation.
//
// This is primarily used by `gorp model *` commands.
type DBInspector interface {
	Ping(ctx context.Context) error
	Tables(ctx context.Context) ([]Table, error)
	Columns(ctx context.Context, table string) ([]Column, error)
}
