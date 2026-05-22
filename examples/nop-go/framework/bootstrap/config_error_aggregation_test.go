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
// 多重错误聚合测试
// ---------------------------------------------------------------------------

// TestValidateCriticalConfigAggregatesMultipleErrors 验证多个校验错误被聚合到一条错误消息中。
//
// TestValidateCriticalConfigAggregatesMultipleErrors verifies multiple validation errors
// are aggregated into a single error message.
func TestValidateCriticalConfigAggregatesMultipleErrors(t *testing.T) {
	cfg := newMapConfigStub()
	cfg.setSection("app", map[string]any{})                                     // address 缺失
	cfg.setSection("log", map[string]any{"level": "verbose", "format": "yaml"}) // level 和 format 都无效

	err := ValidateCriticalConfig(cfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "config: app.address is required")
	require.Contains(t, err.Error(), "config: log.level must be one of [debug info warn error]")
	require.Contains(t, err.Error(), "config: log.format must be one of [console json]")
}
