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
// 有效配置测试
// ---------------------------------------------------------------------------

// TestValidateCriticalConfigPassesWithValidConfig 验证完整有效的配置能通过校验。
//
// TestValidateCriticalConfigPassesWithValidConfig verifies that a complete valid config passes validation.
func TestValidateCriticalConfigPassesWithValidConfig(t *testing.T) {
	cfg := newMapConfigStub()
	cfg.setSection("app", map[string]any{"address": ":8080"})
	cfg.setSection("log", map[string]any{"level": "info", "format": "console"})

	err := ValidateCriticalConfig(cfg)
	require.NoError(t, err)
}

// TestValidateCriticalConfigPassesWithAllSections 验证所有节都存在且有效时通过校验。
//
// TestValidateCriticalConfigPassesWithAllSections verifies validation passes when all sections are present and valid.
func TestValidateCriticalConfigPassesWithAllSections(t *testing.T) {
	cfg := newMapConfigStub()
	cfg.setSection("app", map[string]any{"address": ":8080"})
	cfg.setSection("server", map[string]any{}) // server.http 通过嵌套需要特殊处理
	cfg.setSection("log", map[string]any{"level": "debug", "format": "json"})
	cfg.setSection("database", map[string]any{"driver": "mysql", "dsn": "user:pass@tcp(localhost:3306)/db"})

	err := ValidateCriticalConfig(cfg)
	require.NoError(t, err)
}

// TestValidateCriticalConfigPassesWithJsonFormat 验证 json 日志格式通过校验。
//
// TestValidateCriticalConfigPassesWithJsonFormat verifies json log format passes validation.
func TestValidateCriticalConfigPassesWithJsonFormat(t *testing.T) {
	cfg := newMapConfigStub()
	cfg.setSection("app", map[string]any{"address": ":9090"})
	cfg.setSection("log", map[string]any{"level": "warn", "format": "json"})

	err := ValidateCriticalConfig(cfg)
	require.NoError(t, err)
}

// ---------------------------------------------------------------------------
// 各种日志级别和格式组合测试
// ---------------------------------------------------------------------------

// TestValidateCriticalConfigAcceptsAllLogLevels 验证所有合法日志级别均通过校验。
//
// TestValidateCriticalConfigAcceptsAllLogLevels verifies all valid log levels pass validation.
func TestValidateCriticalConfigAcceptsAllLogLevels(t *testing.T) {
	for _, level := range []string{"debug", "info", "warn", "error"} {
		cfg := newMapConfigStub()
		cfg.setSection("app", map[string]any{"address": ":8080"})
		cfg.setSection("log", map[string]any{"level": level, "format": "console"})

		err := ValidateCriticalConfig(cfg)
		require.NoError(t, err, "log.level=%s should be valid", level)
	}
}

// TestValidateCriticalConfigAcceptsAllDatabaseDrivers 验证所有合法数据库驱动均通过校验。
//
// TestValidateCriticalConfigAcceptsAllDatabaseDrivers verifies all valid database drivers pass validation.
func TestValidateCriticalConfigAcceptsAllDatabaseDrivers(t *testing.T) {
	for _, driver := range []string{"sqlite", "mysql", "postgres", "pgx"} {
		cfg := newMapConfigStub()
		cfg.setSection("app", map[string]any{"address": ":8080"})
		cfg.setSection("log", map[string]any{"level": "info", "format": "console"})
		cfg.setSection("database", map[string]any{"driver": driver, "dsn": "test-dsn"})

		err := ValidateCriticalConfig(cfg)
		require.NoError(t, err, "database.driver=%s should be valid", driver)
	}
}
