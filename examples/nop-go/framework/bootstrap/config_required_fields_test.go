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
// 缺失必填字段测试
// ---------------------------------------------------------------------------

// TestValidateCriticalConfigFailsWithoutAppAddress 验证缺失 app.address 时报错。
//
// TestValidateCriticalConfigFailsWithoutAppAddress verifies error when app.address is missing.
func TestValidateCriticalConfigFailsWithoutAppAddress(t *testing.T) {
	cfg := newMapConfigStub()
	cfg.setSection("app", map[string]any{}) // address 为空
	cfg.setSection("log", map[string]any{"level": "info", "format": "console"})

	err := ValidateCriticalConfig(cfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "config: app.address is required")
}

// TestValidateCriticalConfigFailsWithoutLogLevel 验证缺失 log.level 时报错。
//
// TestValidateCriticalConfigFailsWithoutLogLevel verifies error when log.level is missing.
func TestValidateCriticalConfigFailsWithoutLogLevel(t *testing.T) {
	cfg := newMapConfigStub()
	cfg.setSection("app", map[string]any{"address": ":8080"})
	cfg.setSection("log", map[string]any{"format": "console"}) // level 为空

	err := ValidateCriticalConfig(cfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "config: log.level is required")
}

// TestValidateCriticalConfigFailsWithoutLogFormat 验证缺失 log.format 时报错。
//
// TestValidateCriticalConfigFailsWithoutLogFormat verifies error when log.format is missing.
func TestValidateCriticalConfigFailsWithoutLogFormat(t *testing.T) {
	cfg := newMapConfigStub()
	cfg.setSection("app", map[string]any{"address": ":8080"})
	cfg.setSection("log", map[string]any{"level": "info"}) // format 为空

	err := ValidateCriticalConfig(cfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "config: log.format is required")
}

// TestValidateCriticalConfigFailsWithInvalidLogLevel 验证无效的 log.level 值报错。
//
// TestValidateCriticalConfigFailsWithInvalidLogLevel verifies error with invalid log.level value.
func TestValidateCriticalConfigFailsWithInvalidLogLevel(t *testing.T) {
	cfg := newMapConfigStub()
	cfg.setSection("app", map[string]any{"address": ":8080"})
	cfg.setSection("log", map[string]any{"level": "verbose", "format": "console"}) // "verbose" 不在允许列表

	err := ValidateCriticalConfig(cfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "config: log.level must be one of [debug info warn error]")
}

// TestValidateCriticalConfigFailsWithInvalidLogFormat 验证无效的 log.format 值报错。
//
// TestValidateCriticalConfigFailsWithInvalidLogFormat verifies error with invalid log.format value.
func TestValidateCriticalConfigFailsWithInvalidLogFormat(t *testing.T) {
	cfg := newMapConfigStub()
	cfg.setSection("app", map[string]any{"address": ":8080"})
	cfg.setSection("log", map[string]any{"level": "info", "format": "yaml"}) // "yaml" 不在允许列表

	err := ValidateCriticalConfig(cfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "config: log.format must be one of [console json]")
}

// ---------------------------------------------------------------------------
// 完全空配置测试
// ---------------------------------------------------------------------------

// TestValidateCriticalConfigFailsWithEmptyConfig 验证完全空的配置会报出必填字段缺失错误。
//
// TestValidateCriticalConfigFailsWithEmptyConfig verifies that an entirely empty config
// produces errors about missing required fields.
func TestValidateCriticalConfigFailsWithEmptyConfig(t *testing.T) {
	cfg := newMapConfigStub()

	err := ValidateCriticalConfig(cfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "config: app.address is required")
	require.Contains(t, err.Error(), "config: log.level is required")
	require.Contains(t, err.Error(), "config: log.format is required")
}
