package migrate

import (
	"context"
	"database/sql"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"testing/fstest"
	"time"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
)

func TestToGooseDialect(t *testing.T) {
	tests := []struct {
		driver  string
		want    goose.Dialect
		wantErr bool
	}{
		{"mysql", goose.DialectMySQL, false},
		{"MySQL", goose.DialectMySQL, false},
		{"postgres", goose.DialectPostgres, false},
		{"postgresql", goose.DialectPostgres, false},
		{"pgsql", goose.DialectPostgres, false},
		{"pgx", goose.DialectPostgres, false},
		{"sqlite", goose.DialectSQLite3, false},
		{"sqlite3", goose.DialectSQLite3, false},
		{"clickhouse", goose.DialectClickHouse, false},
		{"mssql", goose.DialectMSSQL, false},
		{"sqlserver", goose.DialectMSSQL, false},
		{"oracle", "", true},
		{"unknown", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.driver, func(t *testing.T) {
			got, err := ToGooseDialect(tt.driver)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}
		})
	}
}

func TestRunUpFS_And_RunDownFS_Lifecycle(t *testing.T) {
	// Prepare in-memory SQLite DB
	db, err := sql.Open("sqlite3", "file:test_lifecycle?mode=memory&cache=shared")
	require.NoError(t, err)
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Virtual FS with 2 migration files
	memFS := fstest.MapFS{
		"migrations/20261103000001_create_users.sql": &fstest.MapFile{
			Data: []byte(`-- +goose Up
CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT);
-- +goose Down
DROP TABLE users;
`),
		},
		"migrations/20261103000002_add_email.sql": &fstest.MapFile{
			Data: []byte(`-- +goose Up
ALTER TABLE users ADD COLUMN email TEXT;
-- +goose Down
ALTER TABLE users DROP COLUMN email;
`),
		},
	}


	// 1. Run migrations up
	err = RunUpFS(ctx, db, "sqlite3", memFS, "migrations", WithTableName("my_migration_versions"))
	require.NoError(t, err)

	// Verify table exists and has columns
	_, err = db.Exec("INSERT INTO users (id, name, email) VALUES (1, 'Alice', 'alice@example.com')")
	require.NoError(t, err)

	// 2. Check DB version
	ver, err := GetDBVersion(ctx, db, "sqlite3", WithTableName("my_migration_versions"))
	require.NoError(t, err)
	require.Equal(t, int64(20261103000002), ver)

	// 3. Check status
	status, err := GetStatus(ctx, db, "sqlite3", memFS, "migrations", WithTableName("my_migration_versions"))
	require.NoError(t, err)
	require.Len(t, status, 2)
	require.Equal(t, goose.StateApplied, status[0].State)
	require.Equal(t, goose.StateApplied, status[1].State)

	// 4. Run Down (rollback step 2)
	err = RunDownFS(ctx, db, "sqlite3", memFS, "migrations", WithTableName("my_migration_versions"))
	require.NoError(t, err)

	// Check DB version rolled back to step 1
	ver, err = GetDBVersion(ctx, db, "sqlite3", WithTableName("my_migration_versions"))
	require.NoError(t, err)
	require.Equal(t, int64(20261103000001), ver)
}

func TestConcurrentProvidersIsolation(t *testing.T) {
	// Verify that multiple providers running concurrently on separate in-memory DBs do not race
	var wg sync.WaitGroup
	workers := 5

	for i := range workers {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			db, err := sql.Open("sqlite3", ":memory:")
			require.NoError(t, err)
			defer db.Close()

			memFS := fstest.MapFS{
				"20261103000001_demo.sql": &fstest.MapFile{
					Data: []byte(`-- +goose Up
CREATE TABLE demo (id INTEGER PRIMARY KEY);
-- +goose Down
DROP TABLE demo;
`),
				},
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err = RunUpFS(ctx, db, "sqlite3", memFS, ".")
			require.NoError(t, err)

			ver, err := GetDBVersion(ctx, db, "sqlite3")
			require.NoError(t, err)
			require.Equal(t, int64(20261103000001), ver)
		}(i)
	}

	wg.Wait()
}

func TestGetDBVersion_ErrorPropagation(t *testing.T) {
	// 1. nil db should return clean error
	_, err := GetDBVersion(context.Background(), nil, "sqlite3")
	require.Error(t, err)
	require.Contains(t, err.Error(), "database connection cannot be nil")

	// 2. canceled context should return context error, NOT 0, nil
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // canceled before calling

	_, err = GetDBVersion(ctx, db, "sqlite3")
	require.Error(t, err)
	require.ErrorIs(t, err, context.Canceled)
}

func TestResolveFS_Paths(t *testing.T) {
	memFS := fstest.MapFS{
		"migrations/users/20261103000001_u.sql": &fstest.MapFile{Data: []byte("-- +goose Up")},
	}

	// Test relative ./ and trailing slash and Windows backslash
	tests := []string{
		"./migrations/users",
		"migrations/users/",
		`migrations\users`,
	}

	for _, p := range tests {
		sub, err := resolveFS(memFS, p)
		require.NoError(t, err, "failed on path: %s", p)
		entries, err := fs.Glob(sub, "*.sql")
		require.NoError(t, err)
		require.Len(t, entries, 1)
	}
}

func TestRunUp_And_RunDown_Directory(t *testing.T) {
	tempDir := t.TempDir()
	migDir := filepath.Join(tempDir, "migrations")
	require.NoError(t, os.MkdirAll(migDir, 0o755))

	sqlFile := filepath.Join(migDir, "20261103000001_create_posts.sql")
	content := `-- +goose Up
CREATE TABLE posts (id INTEGER PRIMARY KEY, title TEXT);
-- +goose Down
DROP TABLE posts;
`
	require.NoError(t, os.WriteFile(sqlFile, []byte(content), 0o644))

	db, err := sql.Open("sqlite3", filepath.Join(tempDir, "test_dir.db"))
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()

	// Test RunUp with directory path
	err = RunUp(ctx, db, "sqlite3", migDir)
	require.NoError(t, err)

	ver, err := GetDBVersion(ctx, db, "sqlite3")
	require.NoError(t, err)
	require.Equal(t, int64(20261103000001), ver)

	// Test RunDown with directory path
	err = RunDown(ctx, db, "sqlite3", migDir)
	require.NoError(t, err)

	ver, err = GetDBVersion(ctx, db, "sqlite3")
	require.NoError(t, err)
	require.Equal(t, int64(0), ver)
}
