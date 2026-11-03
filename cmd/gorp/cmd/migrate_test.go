package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSanitizeMigrationName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"add_user_table", "add_user_table"},
		{"add-user-table", "add_user_table"},
		{"add user table", "add_user_table"},
		{"  Add  User--Table_ ", "add_user_table"},
		{"../../etc/passwd", "etc_passwd"},
		{"`drop_table`;--", "drop_table"},
		{"", "migration"},
	}
	for _, tt := range tests {
		got := sanitizeMigrationName(tt.input)
		require.Equal(t, tt.want, got)
	}
}

func TestGenerateSQLTemplate_Dialects(t *testing.T) {
	t.Run("MySQL", func(t *testing.T) {
		tmpl := generateSQLTemplate("mysql", "users")
		require.Contains(t, tmpl, "-- +goose Up")
		require.Contains(t, tmpl, "AUTO_INCREMENT")
		require.Contains(t, tmpl, "ENGINE=InnoDB")
		require.Contains(t, tmpl, "-- +goose Down")
		require.Contains(t, tmpl, "DROP TABLE IF EXISTS `users`")
	})

	t.Run("PostgreSQL", func(t *testing.T) {
		tmpl := generateSQLTemplate("postgres", "orders")
		require.Contains(t, tmpl, "-- +goose Up")
		require.Contains(t, tmpl, "GENERATED ALWAYS AS IDENTITY PRIMARY KEY")
		require.Contains(t, tmpl, "TIMESTAMPTZ")
		require.NotContains(t, tmpl, "ENGINE=InnoDB")
		require.NotContains(t, tmpl, "`orders`")
		require.Contains(t, tmpl, "-- +goose Down")
		require.Contains(t, tmpl, `DROP TABLE IF EXISTS "orders"`)
	})

	t.Run("SQLite", func(t *testing.T) {
		tmpl := generateSQLTemplate("sqlite3", "products")
		require.Contains(t, tmpl, "-- +goose Up")
		require.Contains(t, tmpl, `CREATE TABLE IF NOT EXISTS "products"`)
		require.Contains(t, tmpl, "AUTOINCREMENT")
		require.Contains(t, tmpl, "-- +goose Down")
		require.Contains(t, tmpl, `DROP TABLE IF EXISTS "products"`)
	})
}

func TestMigrateCLI_FullLifecycle(t *testing.T) {
	tempDir := t.TempDir()
	migrationsDir := filepath.Join(tempDir, "migrations")
	dbFile := filepath.Join(tempDir, "test.db")
	configFile := filepath.Join(tempDir, "app.yaml")

	// 1. Write temp app.yaml pointing to sqlite
	cfgContent := `
database:
  driver: sqlite3
  dsn: "` + filepath.ToSlash(dbFile) + `"
`
	err := os.WriteFile(configFile, []byte(cfgContent), 0o644)
	require.NoError(t, err)

	// Save original flags
	origDir := migrateFlags.dir
	origCfg := migrateFlags.config
	origTable := migrateFlags.tableName
	defer func() {
		migrateFlags.dir = origDir
		migrateFlags.config = origCfg
		migrateFlags.tableName = origTable
	}()

	migrateFlags.dir = migrationsDir
	migrateFlags.config = configFile
	migrateFlags.tableName = "goose_db_version"

	// 2. Test create command
	var out bytes.Buffer
	migrateCreateCmd.SetOut(&out)
	err = runMigrateCreate(migrateCreateCmd, []string{"create_test_table", "sql"})
	require.NoError(t, err)
	require.Contains(t, out.String(), "成功创建迁移文件")

	// Verify file was created in migrationsDir
	files, err := os.ReadDir(migrationsDir)
	require.NoError(t, err)
	require.Len(t, files, 1)
	require.True(t, strings.HasSuffix(files[0].Name(), "_create_test_table.sql"))

	// 3. Test up command
	out.Reset()
	migrateUpCmd.SetOut(&out)
	err = runMigrateUp(migrateUpCmd, nil)
	require.NoError(t, err)
	require.Contains(t, out.String(), "已应用版本")

	// 4. Test status command
	out.Reset()
	migrateStatusCmd.SetOut(&out)
	err = runMigrateStatus(migrateStatusCmd, nil)
	require.NoError(t, err)
	require.Contains(t, out.String(), "Applied")

	// 5. Test version command
	out.Reset()
	migrateVersionCmd.SetOut(&out)
	err = runMigrateVersion(migrateVersionCmd, nil)
	require.NoError(t, err)
	require.Contains(t, out.String(), "当前数据库迁移版本:")

	// 6. Test redo command
	out.Reset()
	migrateRedoCmd.SetOut(&out)
	err = runMigrateRedo(migrateRedoCmd, nil)
	require.NoError(t, err)
	require.Contains(t, out.String(), "Redo 操作完成")

	// 7. Test down command
	out.Reset()
	migrateDownCmd.SetOut(&out)
	err = runMigrateDown(migrateDownCmd, nil)
	require.NoError(t, err)
	require.Contains(t, out.String(), "回滚完成")

	// 8. Test up again and then reset command
	out.Reset()
	migrateUpCmd.SetOut(&out)
	err = runMigrateUp(migrateUpCmd, nil)
	require.NoError(t, err)

	out.Reset()
	migrateResetCmd.SetOut(&out)
	err = runMigrateReset(migrateResetCmd, nil)
	require.NoError(t, err)
	require.Contains(t, out.String(), "Reset 重置完成")
}

func TestMigrateCLI_EmptyDirectoryAndNegativeVersions(t *testing.T) {
	tempDir := t.TempDir()
	emptyMigDir := filepath.Join(tempDir, "empty_migrations")
	require.NoError(t, os.MkdirAll(emptyMigDir, 0o755))

	dbFile := filepath.Join(tempDir, "empty_test.db")
	configFile := filepath.Join(tempDir, "app.yaml")
	cfgContent := `
database:
  driver: sqlite3
  dsn: "` + filepath.ToSlash(dbFile) + `"
`
	require.NoError(t, os.WriteFile(configFile, []byte(cfgContent), 0o644))

	origDir := migrateFlags.dir
	origCfg := migrateFlags.config
	defer func() {
		migrateFlags.dir = origDir
		migrateFlags.config = origCfg
	}()

	migrateFlags.dir = emptyMigDir
	migrateFlags.config = configFile

	// Status on empty directory should give friendly notice, not error
	var out bytes.Buffer
	migrateStatusCmd.SetOut(&out)
	err := runMigrateStatus(migrateStatusCmd, nil)
	require.NoError(t, err)
	require.Contains(t, out.String(), "未找到任何迁移文件")

	// Up on empty directory should give friendly notice, not error
	out.Reset()
	migrateUpCmd.SetOut(&out)
	err = runMigrateUp(migrateUpCmd, nil)
	require.NoError(t, err)
	require.Contains(t, out.String(), "未找到任何迁移文件")

	// UpTo with negative version should fail validation
	err = runMigrateUpTo(migrateUpToCmd, []string{"-5"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "版本号必须为非负整数")

	// DownTo with negative version should fail validation
	err = runMigrateDownTo(migrateDownToCmd, []string{"-1"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "版本号必须为非负整数")
}
