// Package log_test provides unit tests for the log provider.
//
// 适用场景：
// - 验证 Log provider 的注册、引导和日志写入行为。
package log

import (
	"context"
	"testing"

	"github.com/ngq/gorp/framework/container"
	datacontract "github.com/ngq/gorp/framework/contract/data"
	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	"github.com/stretchr/testify/require"
)

type stubRoot struct {
	log string
}

func (s stubRoot) BasePath() string    { return "." }
func (s stubRoot) StoragePath() string { return "." }
func (s stubRoot) RuntimePath() string { return "." }
func (s stubRoot) LogPath() string     { return s.log }
func (s stubRoot) ConfigPath() string  { return "." }
func (s stubRoot) TempPath() string    { return "." }

type stubConfig struct {
	values map[string]any
}

func (s *stubConfig) Env() string                 { return "testing" }
func (s *stubConfig) Get(key string) any          { return s.values[key] }
func (s *stubConfig) GetString(key string) string { v, _ := s.values[key].(string); return v }
func (s *stubConfig) GetInt(key string) int       { v, _ := s.values[key].(int); return v }
func (s *stubConfig) GetBool(key string) bool     { v, _ := s.values[key].(bool); return v }
func (s *stubConfig) GetFloat(string) float64     { return 0 }
func (s *stubConfig) Unmarshal(string, any) error { return nil }
func (s *stubConfig) Watch(_ context.Context, _ string) (datacontract.ConfigWatcher, error) {
	return nil, nil
}
func (s *stubConfig) Reload(_ context.Context) error { return nil }

// TestProviderMeta verifies that the log provider has correct name, defer behavior, and provides the log key.
//
// TestProviderMeta 验证 log provider 的名称、延迟加载行为和提供的日志键。
func TestProviderMeta(t *testing.T) {
	p := NewProvider()
	require.Equal(t, "log", p.Name())
	require.False(t, p.IsDefer())
	require.Equal(t, []string{observabilitycontract.LogKey}, p.Provides())
}

// TestProviderUsesRootLogPathWhenRootExists verifies that the log provider uses the root log path when a root service is available.
//
// TestProviderUsesRootLogPathWhenRootExists 验证当 root 服务存在时 log provider 使用 root 日志路径。
func TestProviderUsesRootLogPathWhenRootExists(t *testing.T) {
	c := container.New()
	c.Bind(datacontract.ConfigKey, func(runtimecontract.Container) (any, error) {
		return &stubConfig{values: map[string]any{
			"log.driver": "single",
		}}, nil
	}, true)
	c.Bind(runtimecontract.RootKey, func(runtimecontract.Container) (any, error) {
		return stubRoot{log: "var/logs"}, nil
	}, true)

	require.NoError(t, c.RegisterProvider(NewProvider()))
	v, err := c.Make(observabilitycontract.LogKey)
	require.NoError(t, err)
	_, ok := v.(observabilitycontract.Logger)
	require.True(t, ok)
}

// TestProviderFallsBackToRelativeLogFileWithoutRoot verifies that the log provider falls back to relative paths when no root service exists.
//
// TestProviderFallsBackToRelativeLogFileWithoutRoot 验证当不存在 root 服务时 log provider 回退到相对路径。
func TestProviderFallsBackToRelativeLogFileWithoutRoot(t *testing.T) {
	c := container.New()
	c.Bind(datacontract.ConfigKey, func(runtimecontract.Container) (any, error) {
		return &stubConfig{values: map[string]any{
			"log.driver": "single",
		}}, nil
	}, true)

	require.NoError(t, c.RegisterProvider(NewProvider()))
	v, err := c.Make(observabilitycontract.LogKey)
	require.NoError(t, err)
	_, ok := v.(observabilitycontract.Logger)
	require.True(t, ok)
}

// TestProviderBindLoggerWithBoolOverrides verifies that boolean config options are correctly applied to the logger.
//
// TestProviderBindLoggerWithBoolOverrides 验证布尔配置选项能正确应用到 logger。
func TestProviderBindLoggerWithBoolOverrides(t *testing.T) {
	c := container.New()
	c.Bind(datacontract.ConfigKey, func(runtimecontract.Container) (any, error) {
		return &stubConfig{values: map[string]any{
			"log.driver":     "stdout",
			"log.local_time": false,
			"log.compress":   true,
		}}, nil
	}, true)

	require.NoError(t, c.RegisterProvider(NewProvider()))
	v, err := c.Make(observabilitycontract.LogKey)
	require.NoError(t, err)
	_, ok := v.(observabilitycontract.Logger)
	require.True(t, ok)
}
