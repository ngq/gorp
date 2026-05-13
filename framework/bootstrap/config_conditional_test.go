// Package bootstrap_test provides unit tests for config validation and schema checking.
//
// 适用场景：
// - 验证 ValidateCriticalConfig 在各种配置场景下的行为。
// - 有效配置通过校验；缺失必填字段报错；条件校验正确跳过或触发；错误消息格式清晰可读。
package bootstrap

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// 条件校验测试
// ---------------------------------------------------------------------------

// TestValidateCriticalConfigSkipsDatabaseWhenAbsent 验证 database 节缺失时跳过校验。
//
// TestValidateCriticalConfigSkipsDatabaseWhenAbsent verifies database validation is skipped when section is absent.
func TestValidateCriticalConfigSkipsDatabaseWhenAbsent(t *testing.T) {
	cfg := newMapConfigStub()
	cfg.setSection("app", map[string]any{"address": ":8080"})
	cfg.setSection("log", map[string]any{"level": "info", "format": "console"})
	// 不设置 database 节

	err := ValidateCriticalConfig(cfg)
	require.NoError(t, err)
}

// TestValidateCriticalConfigSkipsDatabaseWhenEmpty 验证 database 节存在但 driver/dsn 都为空时跳过校验。
//
// TestValidateCriticalConfigSkipsDatabaseWhenEmpty verifies database validation is skipped
// when section exists but both driver and dsn are empty.
func TestValidateCriticalConfigSkipsDatabaseWhenEmpty(t *testing.T) {
	cfg := newMapConfigStub()
	cfg.setSection("app", map[string]any{"address": ":8080"})
	cfg.setSection("log", map[string]any{"level": "info", "format": "console"})
	cfg.setSection("database", map[string]any{}) // driver 和 dsn 都为空

	err := ValidateCriticalConfig(cfg)
	require.NoError(t, err)
}

// TestValidateCriticalConfigFailsWithDatabaseDriverButNoDSN 验证 database.driver 存在但 dsn 缺失时报错。
//
// TestValidateCriticalConfigFailsWithDatabaseDriverButNoDSN verifies error when database.driver
// is set but database.dsn is missing.
func TestValidateCriticalConfigFailsWithDatabaseDriverButNoDSN(t *testing.T) {
	cfg := newMapConfigStub()
	cfg.setSection("app", map[string]any{"address": ":8080"})
	cfg.setSection("log", map[string]any{"level": "info", "format": "console"})
	cfg.setSection("database", map[string]any{"driver": "mysql"}) // dsn 为空

	err := ValidateCriticalConfig(cfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "config: database.dsn is required")
}

// TestValidateCriticalConfigFailsWithDatabaseDSNButNoDriver 验证 database.dsn 存在但 driver 缺失时报错。
//
// TestValidateCriticalConfigFailsWithDatabaseDSNButNoDriver verifies error when database.dsn
// is set but database.driver is missing.
func TestValidateCriticalConfigFailsWithDatabaseDSNButNoDriver(t *testing.T) {
	cfg := newMapConfigStub()
	cfg.setSection("app", map[string]any{"address": ":8080"})
	cfg.setSection("log", map[string]any{"level": "info", "format": "console"})
	cfg.setSection("database", map[string]any{"dsn": "user:pass@tcp(localhost:3306)/db"}) // driver 为空

	err := ValidateCriticalConfig(cfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "config: database.driver is required")
}

// TestValidateCriticalConfigFailsWithInvalidDatabaseDriver 验证无效的 database.driver 值报错。
//
// TestValidateCriticalConfigFailsWithInvalidDatabaseDriver verifies error with invalid database.driver value.
func TestValidateCriticalConfigFailsWithInvalidDatabaseDriver(t *testing.T) {
	cfg := newMapConfigStub()
	cfg.setSection("app", map[string]any{"address": ":8080"})
	cfg.setSection("log", map[string]any{"level": "info", "format": "console"})
	cfg.setSection("database", map[string]any{"driver": "oracle", "dsn": "user:pass@localhost/db"}) // "oracle" 不在允许列表

	err := ValidateCriticalConfig(cfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "config: database.driver must be one of [sqlite mysql postgres pgx]")
}

// TestValidateCriticalConfigPassesWithValidDatabase 验证有效的 database 配置通过校验。
//
// TestValidateCriticalConfigPassesWithValidDatabase verifies valid database config passes validation.
func TestValidateCriticalConfigPassesWithValidDatabase(t *testing.T) {
	cfg := newMapConfigStub()
	cfg.setSection("app", map[string]any{"address": ":8080"})
	cfg.setSection("log", map[string]any{"level": "info", "format": "console"})
	cfg.setSection("database", map[string]any{"driver": "sqlite", "dsn": "file:test.db"})

	err := ValidateCriticalConfig(cfg)
	require.NoError(t, err)
}

// ---------------------------------------------------------------------------
// server.http 条件校验测试
// ---------------------------------------------------------------------------

// TestValidateCriticalConfigSkipsServerHTTPWhenAbsent 验证 server.http 节缺失时跳过校验。
//
// TestValidateCriticalConfigSkipsServerHTTPWhenAbsent verifies server.http validation is skipped when absent.
func TestValidateCriticalConfigSkipsServerHTTPWhenAbsent(t *testing.T) {
	cfg := newMapConfigStub()
	cfg.setSection("app", map[string]any{"address": ":8080"})
	cfg.setSection("log", map[string]any{"level": "info", "format": "console"})
	// 不设置 server.http

	err := ValidateCriticalConfig(cfg)
	require.NoError(t, err)
}
