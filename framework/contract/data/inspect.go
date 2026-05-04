package data

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

type DBInspector interface {
	Ping(ctx context.Context) error
	Tables(ctx context.Context) ([]Table, error)
	Columns(ctx context.Context, table string) ([]Column, error)
}
