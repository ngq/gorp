// Package fuzz_test provides fuzz tests for metadata operations.
//
// 适用场景：
// - 验证 Metadata 对各种边界输入的处理稳定性。
// - 发现潜在的 panic、内存泄漏或异常行为。
package transport

import (
	"testing"
)

// FuzzMetadata_Set fuzzes the Set operation with arbitrary key-value pairs.
//
// FuzzMetadata_Set 对 Set 操作进行模糊测试，使用任意键值对。
func FuzzMetadata_Set(f *testing.F) {
	// Seed corpus with typical inputs
	f.Add("key", "value")
	f.Add("", "value")  // empty key
	f.Add("key", "")    // empty value
	f.Add("X-Request-Id", "12345-abcde")
	f.Add("中文key", "中文value")

	f.Fuzz(func(t *testing.T, key, value string) {
		md := NewMetadata()
		md.Set(key, value)

		// Verify Set doesn't panic and Get returns expected value
		if key != "" {
			got := md.Get(key)
			// Get should return the value (case-insensitive)
			if got != value && got != "" {
				t.Errorf("Get(%q) = %q, want %q or empty", key, got, value)
			}
		}
	})
}

// FuzzMetadata_Add fuzzes the Add operation with multiple values for same key.
//
// FuzzMetadata_Add 对 Add 操作进行模糊测试，验证多值添加。
func FuzzMetadata_Add(f *testing.F) {
	f.Add("key", "value1", "value2")
	f.Add("", "value1", "value2")
	f.Add("key", "", "value2")

	f.Fuzz(func(t *testing.T, key, value1, value2 string) {
		md := NewMetadata()
		md.Add(key, value1)
		md.Add(key, value2)

		if key != "" {
			values := md.Values(key)
			if len(values) < 1 {
				t.Errorf("Values(%q) returned empty, expected at least 1", key)
			}
		}
	})
}

// FuzzMetadata_Get fuzzes Get with arbitrary keys.
//
// FuzzMetadata_Get 对 Get 操作进行模糊测试。
func FuzzMetadata_Get(f *testing.F) {
	f.Add("missing-key")
	f.Add("")
	f.Add("X-Custom-Header")

	f.Fuzz(func(t *testing.T, key string) {
		md := NewMetadata()
		// Get on empty metadata should return empty string
		got := md.Get(key)
		if got != "" {
			t.Errorf("Get on empty metadata returned %q, expected empty", got)
		}

		// Set a value and verify Get
		md.Set("test-key", "test-value")
		_ = md.Get(key) // Should not panic
	})
}

// FuzzMetadata_Clone fuzzes Clone with various metadata states.
//
// FuzzMetadata_Clone 对 Clone 操作进行模糊测试。
func FuzzMetadata_Clone(f *testing.F) {
	f.Add("key1", "value1", "key2", "value2")

	f.Fuzz(func(t *testing.T, k1, v1, k2, v2 string) {
		md := NewMetadata()
		md.Set(k1, v1)
		md.Set(k2, v2)

		cloned := md.Clone()
		if cloned == nil {
			t.Fatal("Clone returned nil")
		}

		// Verify clone is independent
		md.Set(k1, "modified")
		if k1 != "" && cloned.Get(k1) != v1 && cloned.Get(k1) != "" {
			t.Errorf("Clone not independent: cloned.Get(%q) = %q, want %q", k1, cloned.Get(k1), v1)
		}
	})
}

// FuzzMetadata_Range fuzzes Range with various iteration patterns.
//
// FuzzMetadata_Range 对 Range 操作进行模糊测试。
func FuzzMetadata_Range(f *testing.F) {
	f.Add("key1", "value1", "key2", "value2", true)
	f.Add("key1", "value1", "key2", "value2", false)

	f.Fuzz(func(t *testing.T, k1, v1, k2, v2 string, continueIter bool) {
		md := NewMetadata()
		md.Set(k1, v1)
		md.Set(k2, v2)

		count := 0
		md.Range(func(key string, values []string) bool {
			count++
			return continueIter
		})

		// If continueIter is true, should iterate all keys
		// If false, should stop after first
		if continueIter && count < 1 {
			t.Errorf("Range with continue=true visited %d keys, expected at least 1", count)
		}
	})
}

// FuzzNewMetadataFromMap fuzzes creating metadata from arbitrary maps.
//
// FuzzNewMetadataFromMap 对从 map 创建 metadata 进行模糊测试。
func FuzzNewMetadataFromMap(f *testing.F) {
	// Seed with typical map structures (simplified for fuzz)
	f.Add("key1", "value1", "key2", "value2a", "value2b")

	f.Fuzz(func(t *testing.T, k1, v1, k2, v2a, v2b string) {
		m := map[string][]string{
			k1: {v1},
			k2: {v2a, v2b},
		}

		md := NewMetadataFromMap(m)
		if md == nil {
			t.Fatal("NewMetadataFromMap returned nil")
		}

		// Verify values are accessible
		if k1 != "" {
			_ = md.Get(k1)
		}
		if k2 != "" {
			values := md.Values(k2)
			if len(values) < 1 {
				t.Errorf("Values(%q) empty after NewMetadataFromMap", k2)
			}
		}
	})
}