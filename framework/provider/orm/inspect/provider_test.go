// Package inspect_test provides unit tests for the ORM schema inspector provider.
//
// 适用场景：
// - 验证 ORM schema inspector provider 的注册与表结构检查行为。
package inspect

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

// TestService_MySQLTablesAndColumns verifies that the inspector service
// can correctly query tables and columns from a MySQL database using SQL mock.
//
// TestService_MySQLTablesAndColumns 验证 inspector service 能正确查询 MySQL 数据库的表和列信息。
func TestService_MySQLTablesAndColumns(t *testing.T) {
	rawDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer rawDB.Close()

	sx := sqlx.NewDb(rawDB, "mysql")
	svc := &Service{db: sx, driver: "mysql"}
	ctx := context.Background()

	mock.ExpectQuery(`SELECT table_name`).
		WillReturnRows(sqlmock.NewRows([]string{"table_name"}).
			AddRow("roles").
			AddRow("users"))

	tables, err := svc.Tables(ctx)
	require.NoError(t, err)
	require.Equal(t, []string{"roles", "users"}, []string{tables[0].Name, tables[1].Name})

	mock.ExpectQuery(`SELECT\s+column_name,`).
		WithArgs("users").
		WillReturnRows(sqlmock.NewRows([]string{"column_name", "column_type", "is_nullable", "column_key", "column_default", "extra", "ordinal_position"}).
			AddRow("id", "bigint unsigned", "NO", "PRI", nil, "auto_increment", 1).
			AddRow("name", "varchar(255)", "NO", "", nil, "", 2).
			AddRow("email", "varchar(255)", "YES", "", "guest@example.com", "", 3))

	cols, err := svc.Columns(ctx, "users")
	require.NoError(t, err)
	require.Len(t, cols, 3)
	require.Equal(t, "id", cols[0].Name)
	require.True(t, cols[0].PrimaryKey)
	require.True(t, cols[0].NotNull)
	require.Equal(t, "name", cols[1].Name)
	require.True(t, cols[1].NotNull)
	require.Equal(t, "email", cols[2].Name)
	require.False(t, cols[2].NotNull)
	require.NotNil(t, cols[2].DefaultVal)
	require.Equal(t, "guest@example.com", *cols[2].DefaultVal)

	require.NoError(t, mock.ExpectationsWereMet())
}

// TestService_PostgresTablesAndColumns verifies that the inspector service
// can correctly query tables and columns from a PostgreSQL database using SQL mock.
//
// TestService_PostgresTablesAndColumns 验证 inspector service 能正确查询 PostgreSQL 数据库的表和列信息。
func TestService_PostgresTablesAndColumns(t *testing.T) {
	rawDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer rawDB.Close()

	sx := sqlx.NewDb(rawDB, "pgx")
	svc := &Service{db: sx, driver: "pgx"}
	ctx := context.Background()

	mock.ExpectQuery(`SELECT table_name`).
		WillReturnRows(sqlmock.NewRows([]string{"table_name"}).
			AddRow("accounts").
			AddRow("users"))

	tables, err := svc.Tables(ctx)
	require.NoError(t, err)
	require.Equal(t, []string{"accounts", "users"}, []string{tables[0].Name, tables[1].Name})

	mock.ExpectQuery(`SELECT\s+c\.column_name,`).
		WithArgs("users").
		WillReturnRows(sqlmock.NewRows([]string{"column_name", "column_type", "is_nullable", "column_default", "ordinal_position", "is_primary"}).
			AddRow("id", "int8", "NO", "nextval('users_id_seq'::regclass)", 1, true).
			AddRow("email", "text", "NO", nil, 2, false).
			AddRow("nickname", "text", "YES", "'guest'::text", 3, false))

	cols, err := svc.Columns(ctx, "users")
	require.NoError(t, err)
	require.Len(t, cols, 3)
	require.Equal(t, "id", cols[0].Name)
	require.True(t, cols[0].PrimaryKey)
	require.True(t, cols[0].NotNull)
	require.Equal(t, "nickname", cols[2].Name)
	require.False(t, cols[2].NotNull)
	require.NotNil(t, cols[2].DefaultVal)
	require.Equal(t, "'guest'::text", *cols[2].DefaultVal)

	require.NoError(t, mock.ExpectationsWereMet())
}
