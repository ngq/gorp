package contract

import (
	"context"
	"testing"
)

func TestMetadata_Get_Set(t *testing.T) {
	md := NewMetadata()
	md.Set("X-Request-Id", "12345")

	// 测试大小写不敏感
	if md.Get("x-request-id") != "12345" {
		t.Errorf("expected 12345, got: %s", md.Get("x-request-id"))
	}
	if md.Get("X-REQUEST-ID") != "12345" {
		t.Errorf("expected 12345 for uppercase key")
	}
}

func TestMetadata_Add_Values(t *testing.T) {
	md := NewMetadata()
	md.Add("x-forwarded-for", "192.168.1.1")
	md.Add("x-forwarded-for", "192.168.1.2")

	values := md.Values("x-forwarded-for")
	if len(values) != 2 {
		t.Errorf("expected 2 values, got: %d", len(values))
	}
	if values[0] != "192.168.1.1" || values[1] != "192.168.1.2" {
		t.Errorf("unexpected values: %v", values)
	}
}

func TestMetadata_Del(t *testing.T) {
	md := NewMetadata()
	md.Set("x-trace-id", "abc123")
	md.Del("x-trace-id")

	if md.Get("x-trace-id") != "" {
		t.Errorf("expected empty after delete")
	}
}

func TestMetadata_Clone(t *testing.T) {
	md := NewMetadata()
	md.Set("x-key", "value")

	cloned := md.Clone()
	cloned.Set("x-key", "modified")

	// 原始 metadata 不应被修改
	if md.Get("x-key") != "value" {
		t.Errorf("original metadata was modified")
	}
}

func TestMetadata_ToMap(t *testing.T) {
	md := NewMetadata()
	md.Set("x-key-1", "value1")
	md.Add("x-key-2", "value2a")
	md.Add("x-key-2", "value2b")

	m := md.ToMap()
	if m["x-key-1"][0] != "value1" {
		t.Errorf("unexpected map value")
	}
	if len(m["x-key-2"]) != 2 {
		t.Errorf("expected 2 values for x-key-2")
	}
}

func TestMetadata_Range(t *testing.T) {
	md := NewMetadata()
	md.Set("x-key-1", "value1")
	md.Set("x-key-2", "value2")

	count := 0
	md.Range(func(key string, values []string) bool {
		count++
		return true
	})

	if count != 2 {
		t.Errorf("expected 2 iterations, got: %d", count)
	}
}

func TestMetadata_Range_Stop(t *testing.T) {
	md := NewMetadata()
	md.Set("x-key-1", "value1")
	md.Set("x-key-2", "value2")

	count := 0
	md.Range(func(key string, values []string) bool {
		count++
		return false // 停止遍历
	})

	if count != 1 {
		t.Errorf("expected 1 iteration, got: %d", count)
	}
}

func TestNewMetadataFromMap(t *testing.T) {
	original := map[string][]string{
		"X-Key": {"value1", "value2"},
	}

	md := NewMetadataFromMap(original)

	// 验证 key 被转为小写
	if md.Get("x-key") != "value1" {
		t.Errorf("expected value1, got: %s", md.Get("x-key"))
	}

	// 验证原始 map 不受影响
	if len(original["X-Key"]) != 2 {
		t.Errorf("original map was modified")
	}
}

func TestServerContext(t *testing.T) {
	ctx := context.Background()
	md := NewMetadata()
	md.Set("x-request-id", "test-123")

	ctx = NewServerContext(ctx, md)

	retrieved, ok := FromServerContext(ctx)
	if !ok {
		t.Error("expected to find server metadata")
	}
	if retrieved.Get("x-request-id") != "test-123" {
		t.Errorf("unexpected metadata value")
	}
}

func TestClientContext(t *testing.T) {
	ctx := context.Background()
	md := NewMetadata()
	md.Set("x-client-id", "client-456")

	ctx = NewClientContext(ctx, md)

	retrieved, ok := FromClientContext(ctx)
	if !ok {
		t.Error("expected to find client metadata")
	}
	if retrieved.Get("x-client-id") != "client-456" {
		t.Errorf("unexpected metadata value")
	}
}

func TestAppendToClientContext(t *testing.T) {
	ctx := context.Background()

	// 追加 metadata
	ctx = AppendToClientContext(ctx, "x-key-1", "value1", "x-key-2", "value2")

	md, ok := FromClientContext(ctx)
	if !ok {
		t.Error("expected to find client metadata")
	}
	if md.Get("x-key-1") != "value1" {
		t.Errorf("expected value1 for x-key-1")
	}
	if md.Get("x-key-2") != "value2" {
		t.Errorf("expected value2 for x-key-2")
	}
}

func TestAppendToClientContext_OddPairs(t *testing.T) {
	ctx := context.Background()

	// 奇数参数应被忽略，不应 panic
	ctx = AppendToClientContext(ctx, "x-key-1", "value1", "x-key-2")

	md, ok := FromClientContext(ctx)
	if ok && md != nil {
		// 应该没有创建 metadata 或只有第一个键值对
	}
}

func TestMapMetadata_EmptyKey(t *testing.T) {
	md := NewMetadata()

	// 空 key 应被忽略
	md.Set("", "value")
	md.Add("", "value")

	// 不应 panic
}