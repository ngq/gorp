package baseconfigsource_test

import (
	"context"
	"testing"
	"time"

	"github.com/ngq/gorp/contrib/internal/baseconfigsource"
	datacontract "github.com/ngq/gorp/framework/contract/data"
	"github.com/stretchr/testify/require"
)

type mockConfig struct {
	data map[string]any
}

func (m *mockConfig) Env() string        { return "" }
func (m *mockConfig) Get(key string) any { return m.data[key] }
func (m *mockConfig) GetString(key string) string {
	if v, ok := m.data[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
func (m *mockConfig) GetInt(key string) int {
	if v, ok := m.data[key]; ok {
		if n, ok := v.(int); ok {
			return n
		}
	}
	return 0
}
func (m *mockConfig) GetBool(key string) bool {
	if v, ok := m.data[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}
func (m *mockConfig) GetFloat(key string) float64     { return 0 }
func (m *mockConfig) Unmarshal(_ string, _ any) error { return nil }
func (m *mockConfig) Watch(_ context.Context, _ string) (datacontract.ConfigWatcher, error) {
	return nil, nil
}
func (m *mockConfig) Reload(_ context.Context) error { return nil }

func TestGetStringFallback_PrimaryPath(t *testing.T) {
	cfg := &mockConfig{data: map[string]any{
		"configsource.apollo.app_id": "primary-id",
		"config.apollo.app_id":       "fallback-id",
	}}
	result := baseconfigsource.GetStringFallback(cfg, "apollo", "app_id")
	require.Equal(t, "primary-id", result)
}

func TestGetStringFallback_FallbackPath(t *testing.T) {
	cfg := &mockConfig{data: map[string]any{
		"config.apollo.app_id": "fallback-id",
	}}
	result := baseconfigsource.GetStringFallback(cfg, "apollo", "app_id")
	require.Equal(t, "fallback-id", result)
}

func TestGetStringFallback_Empty(t *testing.T) {
	cfg := &mockConfig{data: map[string]any{}}
	result := baseconfigsource.GetStringFallback(cfg, "apollo", "app_id")
	require.Equal(t, "", result)
}

func TestGetIntFallback(t *testing.T) {
	cfg := &mockConfig{data: map[string]any{
		"configsource.nacos.port": 8848,
	}}
	result := baseconfigsource.GetIntFallback(cfg, "nacos", "port")
	require.Equal(t, 8848, result)
}

func TestGetDurationSecondsFallback(t *testing.T) {
	cfg := &mockConfig{data: map[string]any{
		"configsource.apollo.poll_interval_seconds": 10,
	}}
	result := baseconfigsource.GetDurationSecondsFallback(cfg, "apollo", "poll_interval_seconds")
	require.Equal(t, 10*time.Second, result)
}

func TestGetDurationMillisFallback(t *testing.T) {
	cfg := &mockConfig{data: map[string]any{
		"configsource.apollo.watch_retry_interval_ms": 500,
	}}
	result := baseconfigsource.GetDurationMillisFallback(cfg, "apollo", "watch_retry_interval_ms")
	require.Equal(t, 500*time.Millisecond, result)
}

func TestGetBoolFallback(t *testing.T) {
	cfg := &mockConfig{data: map[string]any{
		"configsource.kubernetes.in_cluster": true,
	}}
	val, ok := baseconfigsource.GetBoolFallback(cfg, "kubernetes", "in_cluster")
	require.True(t, ok)
	require.True(t, val)
}
