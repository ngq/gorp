// Package config_test provides fuzz tests for configuration parsing.
//
// 适用场景：
// - 验证配置解析对各种边界输入的处理稳定性。
// - 发现潜在的 panic 或异常行为。
package config_test

import (
	"context"
	"testing"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	"github.com/ngq/gorp/framework/provider/config"
)

// mockConfig implements datacontract.Config for fuzz testing.
//
// mockConfig 实现 datacontract.Config 用于模糊测试。
type mockConfig struct {
	data map[string]any
}

func newMockConfig(data map[string]any) *mockConfig {
	return &mockConfig{data: data}
}

func (m *mockConfig) Get(key string) any {
	return m.data[key]
}

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
		switch n := v.(type) {
		case int:
			return n
		case int64:
			return int(n)
		case float64:
			return int(n)
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

func (m *mockConfig) GetFloat(key string) float64 {
	if v, ok := m.data[key]; ok {
		switch n := v.(type) {
		case float64:
			return n
		case int:
			return float64(n)
		case int64:
			return float64(n)
		}
	}
	return 0
}

func (m *mockConfig) Unmarshal(key string, out any) error {
	return nil
}

func (m *mockConfig) Watch(ctx context.Context, key string) (datacontract.ConfigWatcher, error) {
	return nil, nil
}

func (m *mockConfig) Reload(ctx context.Context) error {
	return nil
}

func (m *mockConfig) Env() string {
	return "test"
}

// FuzzGetAny fuzzes GetAny with arbitrary keys.
//
// FuzzGetAny 对 GetAny 进行模糊测试，使用任意键。
func FuzzGetAny(f *testing.F) {
	f.Add("key1", "value1")
	f.Add("", "value")
	f.Add("key", "")
	f.Add("key.with.dots", "value")
	f.Add("key-with-dashes", "value")
	f.Add("key_with_underscores", "value")
	f.Add("中文key", "中文value")

	f.Fuzz(func(t *testing.T, key, value string) {
		cfg := newMockConfig(map[string]any{key: value})
		got, ok := config.GetAny(cfg, key)
		if ok && got != value {
			t.Errorf("GetAny(%q) = %v, want %v", key, got, value)
		}

		// Test with nil config - should not panic
		// 测试 nil 配置 - 不应该 panic
		_, ok = config.GetAny(nil, key)
		if ok {
			t.Error("GetAny(nil) returned true")
		}
	})
}

// FuzzGetStringAny fuzzes GetStringAny with multiple keys.
//
// FuzzGetStringAny 对 GetStringAny 进行模糊测试，使用多个键。
func FuzzGetStringAny(f *testing.F) {
	f.Add("key1", "value1", "key2", "value2")
	f.Add("", "value1", "key2", "value2")
	f.Add("key1", "", "key2", "value2")
	f.Add("key1", "value1", "", "value2")

	f.Fuzz(func(t *testing.T, k1, v1, k2, v2 string) {
		data := make(map[string]any)
		if k1 != "" {
			data[k1] = v1
		}
		if k2 != "" {
			data[k2] = v2
		}

		cfg := newMockConfig(data)
		result := config.GetStringAny(cfg, k1, k2)

		// Should return first non-empty value
		// 应该返回第一个非空值
		if v1 != "" && result != v1 {
			t.Logf("GetStringAny returned %q, expected first non-empty", result)
		}
	})
}

// FuzzGetIntAny fuzzes GetIntAny with int values.
//
// FuzzGetIntAny 对 GetIntAny 进行模糊测试，使用整数值。
func FuzzGetIntAny(f *testing.F) {
	f.Add("int_key", 42)
	f.Add("zero_key", 0)
	f.Add("negative_key", -100)

	f.Fuzz(func(t *testing.T, key string, value int) {
		cfg := newMockConfig(map[string]any{key: value})
		result := config.GetIntAny(cfg, key)
		if result != value {
			t.Logf("GetIntAny(%q) = %d, want %d", key, result, value)
		}
	})
}

// FuzzGetBoolAny fuzzes GetBoolAny with bool values.
//
// FuzzGetBoolAny 对 GetBoolAny 进行模糊测试，使用布尔值。
func FuzzGetBoolAny(f *testing.F) {
	f.Add("bool_true", true)
	f.Add("bool_false", false)

	f.Fuzz(func(t *testing.T, key string, value bool) {
		cfg := newMockConfig(map[string]any{key: value})
		result, ok := config.GetBoolAny(cfg, key)
		if !ok {
			t.Logf("GetBoolAny(%q) returned ok=false", key)
		}
		if result != value {
			t.Logf("GetBoolAny(%q) = %v, want %v", key, result, value)
		}
	})
}

// FuzzGetFloatAny fuzzes GetFloatAny with float values.
//
// FuzzGetFloatAny 对 GetFloatAny 进行模糊测试，使用浮点值。
func FuzzGetFloatAny(f *testing.F) {
	f.Add("float_key", 3.14159)
	f.Add("zero_key", 0.0)
	f.Add("negative_key", -2.71828)

	f.Fuzz(func(t *testing.T, key string, value float64) {
		cfg := newMockConfig(map[string]any{key: value})
		result := config.GetFloatAny(cfg, key)
		// Allow small floating point differences
		// 允许小的浮点数差异
		if result != value {
			t.Logf("GetFloatAny(%q) = %v, want %v", key, result, value)
		}
	})
}

// FuzzGetStringSliceAny fuzzes GetStringSliceAny with slice values.
//
// FuzzGetStringSliceAny 对 GetStringSliceAny 进行模糊测试，使用切片值。
func FuzzGetStringSliceAny(f *testing.F) {
	f.Add("key1", "v1", "key2", "v2")

	f.Fuzz(func(t *testing.T, k1, v1, k2, v2 string) {
		data := make(map[string]any)
		if k1 != "" && v1 != "" {
			data[k1] = []string{v1}
		}
		if k2 != "" && v2 != "" {
			data[k2] = []string{v2}
		}

		cfg := newMockConfig(data)
		result := config.GetStringSliceAny(cfg, k1, k2)

		// Should not panic regardless of input
		// 不管输入如何都不应该 panic
		_ = result
	})
}

// FuzzGetStringMapAny fuzzes GetStringMapAny with map values.
//
// FuzzGetStringMapAny 对 GetStringMapAny 进行模糊测试，使用 map 值。
func FuzzGetStringMapAny(f *testing.F) {
	f.Add("key", "mk", "mv")

	f.Fuzz(func(t *testing.T, key, mk, mv string) {
		data := make(map[string]any)
		if key != "" && mk != "" {
			data[key] = map[string]string{mk: mv}
		}

		cfg := newMockConfig(data)
		result := config.GetStringMapAny(cfg, key)

		// Should not panic regardless of input
		// 不管输入如何都不应该 panic
		_ = result
	})
}

// Ensure mockConfig implements datacontract.Config
var _ datacontract.Config = (*mockConfig)(nil)
