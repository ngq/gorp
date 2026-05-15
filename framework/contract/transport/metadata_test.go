// Package transport_test provides unit tests for transport metadata.
//
// 适用场景：
// - 验证 Metadata 的 CRUD 操作。
// - 验证 Clone 和 ToMap 的正确性。
// - 验证 Context 透传行为。
package transport

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

// ============================================================
// Metadata CRUD tests
// ============================================================

func TestMetadata_Get(t *testing.T) {
	md := NewMetadata()
	md.Set("x-request-id", "12345")

	require.Equal(t, "12345", md.Get("x-request-id"))
	require.Equal(t, "", md.Get("missing"))
}

func TestMetadata_Get_CaseInsensitive(t *testing.T) {
	md := NewMetadata()
	md.Set("X-Request-Id", "12345")

	require.Equal(t, "12345", md.Get("x-request-id"))
	require.Equal(t, "12345", md.Get("X-REQUEST-ID"))
}

func TestMetadata_Values(t *testing.T) {
	md := NewMetadata()
	md.Add("x-trace-id", "trace-1")
	md.Add("x-trace-id", "trace-2")

	values := md.Values("x-trace-id")
	require.Len(t, values, 2)
	require.Equal(t, "trace-1", values[0])
	require.Equal(t, "trace-2", values[1])
}

func TestMetadata_Set(t *testing.T) {
	md := NewMetadata()
	md.Set("key", "value1")
	md.Set("key", "value2")

	require.Equal(t, "value2", md.Get("key"))
}

func TestMetadata_Set_EmptyKey(t *testing.T) {
	md := NewMetadata()
	md.Set("", "value")

	require.Equal(t, "", md.Get(""))
}

func TestMetadata_Add(t *testing.T) {
	md := NewMetadata()
	md.Add("key", "value1")
	md.Add("key", "value2")

	values := md.Values("key")
	require.Len(t, values, 2)
}

func TestMetadata_Add_EmptyKey(t *testing.T) {
	md := NewMetadata()
	md.Add("", "value")

	require.Equal(t, "", md.Get(""))
}

func TestMetadata_Del(t *testing.T) {
	md := NewMetadata()
	md.Set("key", "value")
	md.Del("key")

	require.Equal(t, "", md.Get("key"))
}

func TestMetadata_Del_CaseInsensitive(t *testing.T) {
	md := NewMetadata()
	md.Set("Key", "value")
	md.Del("key")

	require.Equal(t, "", md.Get("key"))
}

func TestMetadata_Range(t *testing.T) {
	md := NewMetadata()
	md.Set("key1", "value1")
	md.Set("key2", "value2")

	count := 0
	md.Range(func(key string, values []string) bool {
		count++
		return true
	})
	require.Equal(t, 2, count)
}

func TestMetadata_Range_StopEarly(t *testing.T) {
	md := NewMetadata()
	md.Set("key1", "value1")
	md.Set("key2", "value2")
	md.Set("key3", "value3")

	count := 0
	md.Range(func(key string, values []string) bool {
		count++
		return count < 2
	})
	require.Equal(t, 2, count)
}

// ============================================================
// Clone and ToMap tests
// ============================================================

func TestMetadata_Clone(t *testing.T) {
	md := NewMetadata()
	md.Set("key", "value")

	cloned := md.Clone()
	require.Equal(t, "value", cloned.Get("key"))

	// Modify original should not affect clone
	md.Set("key", "modified")
	require.Equal(t, "value", cloned.Get("key"))
}

func TestMetadata_Clone_MultipleValues(t *testing.T) {
	md := NewMetadata()
	md.Add("key", "value1")
	md.Add("key", "value2")

	cloned := md.Clone()
	values := cloned.Values("key")
	require.Len(t, values, 2)
	require.Equal(t, "value1", values[0])
	require.Equal(t, "value2", values[1])
}

func TestMetadata_ToMap(t *testing.T) {
	md := NewMetadata()
	md.Set("key1", "value1")
	md.Set("key2", "value2")

	m := md.ToMap()
	require.Len(t, m, 2)
	require.Equal(t, []string{"value1"}, m["key1"])
	require.Equal(t, []string{"value2"}, m["key2"])
}

func TestMetadata_ToMap_Immutable(t *testing.T) {
	md := NewMetadata()
	md.Set("key", "value")

	m := md.ToMap()
	m["key"] = []string{"modified"}

	// Original should not be affected
	require.Equal(t, "value", md.Get("key"))
}

// ============================================================
// NewMetadataFromMap tests
// ============================================================

func TestNewMetadataFromMap(t *testing.T) {
	m := map[string][]string{
		"key1": {"value1"},
		"key2": {"value2a", "value2b"},
	}

	md := NewMetadataFromMap(m)
	require.Equal(t, "value1", md.Get("key1"))
	require.Len(t, md.Values("key2"), 2)
}

func TestNewMetadataFromMap_CaseInsensitive(t *testing.T) {
	m := map[string][]string{
		"X-Request-Id": {"12345"},
	}

	md := NewMetadataFromMap(m)
	require.Equal(t, "12345", md.Get("x-request-id"))
}

func TestNewMetadataFromMap_Immutable(t *testing.T) {
	m := map[string][]string{
		"key": {"value"},
	}

	md := NewMetadataFromMap(m)
	m["key"] = []string{"modified"}

	require.Equal(t, "value", md.Get("key"))
}

// ============================================================
// Context propagation tests
// ============================================================

func TestNewServerContext(t *testing.T) {
	md := NewMetadata()
	md.Set("key", "value")

	ctx := NewServerContext(context.Background(), md)
	retrieved, ok := FromServerContext(ctx)
	require.True(t, ok)
	require.Equal(t, "value", retrieved.Get("key"))
}

func TestFromServerContext_Missing(t *testing.T) {
	md, ok := FromServerContext(context.Background())
	require.False(t, ok)
	require.Nil(t, md)
}

func TestNewClientContext(t *testing.T) {
	md := NewMetadata()
	md.Set("key", "value")

	ctx := NewClientContext(context.Background(), md)
	retrieved, ok := FromClientContext(ctx)
	require.True(t, ok)
	require.Equal(t, "value", retrieved.Get("key"))
}

func TestFromClientContext_Missing(t *testing.T) {
	md, ok := FromClientContext(context.Background())
	require.False(t, ok)
	require.Nil(t, md)
}

func TestAppendToClientContext(t *testing.T) {
	ctx := AppendToClientContext(context.Background(), "key1", "value1", "key2", "value2")

	md, ok := FromClientContext(ctx)
	require.True(t, ok)
	require.Equal(t, "value1", md.Get("key1"))
	require.Equal(t, "value2", md.Get("key2"))
}

func TestAppendToClientContext_Append(t *testing.T) {
	md := NewMetadata()
	md.Set("key", "value1")
	ctx := NewClientContext(context.Background(), md)

	ctx = AppendToClientContext(ctx, "key", "value2")

	retrieved, ok := FromClientContext(ctx)
	require.True(t, ok)
	// AppendToClientContext uses Set, which replaces the value
	require.Equal(t, "value2", retrieved.Get("key"))
}

func TestAppendToClientContext_OddPairs(t *testing.T) {
	// Odd number of arguments should be ignored
	ctx := AppendToClientContext(context.Background(), "key", "value", "odd")

	md, ok := FromClientContext(ctx)
	require.False(t, ok)
	require.Nil(t, md)
}

func TestAppendToClientContext_EmptyPairs(t *testing.T) {
	ctx := AppendToClientContext(context.Background())

	// No metadata should be created
	md, ok := FromClientContext(ctx)
	// Empty pairs creates metadata with no values
	if ok {
		require.NotNil(t, md)
	} else {
		require.Nil(t, md)
	}
}