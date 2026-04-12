package propagator

import (
	"context"
	"testing"

	"github.com/ngq/gorp/framework/contract"
)

// mockCarrier 实现 MetadataCarrier 用于测试。
type mockCarrier struct {
	data map[string][]string
}

func newMockCarrier() *mockCarrier {
	return &mockCarrier{data: make(map[string][]string)}
}

func (c *mockCarrier) Get(key string) string {
	if v, ok := c.data[key]; ok && len(v) > 0 {
		return v[0]
	}
	return ""
}

func (c *mockCarrier) Set(key, value string) {
	c.data[key] = []string{value}
}

func (c *mockCarrier) Add(key, value string) {
	c.data[key] = append(c.data[key], value)
}

func (c *mockCarrier) Keys() []string {
	keys := make([]string, 0, len(c.data))
	for k := range c.data {
		keys = append(keys, k)
	}
	return keys
}

func (c *mockCarrier) Values(key string) []string {
	return c.data[key]
}

func TestDefaultPropagator_Extract(t *testing.T) {
	prop := NewDefaultPropagator([]string{"x-md-"}, nil)

	// 创建 carrier 模拟 HTTP Header
	carrier := newMockCarrier()
	carrier.Set("x-md-trace-id", "trace-123")
	carrier.Set("x-md-request-id", "req-456")
	carrier.Set("authorization", "bearer token") // 不匹配前缀，不应被提取

	ctx := prop.Extract(context.Background(), carrier)

	md, ok := contract.FromServerContext(ctx)
	if !ok {
		t.Error("expected to find server metadata")
	}

	// 验证匹配前缀的 key 被提取
	if md.Get("x-md-trace-id") != "trace-123" {
		t.Errorf("expected trace-123, got: %s", md.Get("x-md-trace-id"))
	}

	// 验证不匹配前缀的 key 未被提取
	if md.Get("authorization") != "" {
		t.Error("authorization should not be extracted")
	}
}

func TestDefaultPropagator_Inject(t *testing.T) {
	prop := NewDefaultPropagator([]string{"x-md-"}, map[string]string{
		"x-md-app": "test-app",
	})

	// 创建 server context
	md := contract.NewMetadata()
	md.Set("x-md-trace-id", "trace-123")
	md.Set("authorization", "bearer token") // 不匹配前缀
	ctx := contract.NewServerContext(context.Background(), md)

	// 注入到 carrier
	carrier := newMockCarrier()
	prop.Inject(ctx, carrier)

	// 验证常量 metadata 被注入
	if carrier.Get("x-md-app") != "test-app" {
		t.Error("constant metadata should be injected")
	}

	// 验证匹配前缀的 metadata 被注入
	if carrier.Get("x-md-trace-id") != "trace-123" {
		t.Error("matching prefix metadata should be injected")
	}
}

func TestDefaultPropagator_InjectClientMetadata(t *testing.T) {
	prop := NewDefaultPropagator(nil, nil)

	// 创建 client context
	md := contract.NewMetadata()
	md.Set("x-client-id", "client-456")
	ctx := contract.NewClientContext(context.Background(), md)

	// 注入到 carrier
	carrier := newMockCarrier()
	prop.Inject(ctx, carrier)

	// 验证 client metadata 被注入
	if carrier.Get("x-client-id") != "client-456" {
		t.Error("client metadata should be injected")
	}
}

func TestDefaultPropagator_MatchPrefix(t *testing.T) {
	// nil 前缀列表使用默认前缀 "x-md-"
	prop := NewDefaultPropagator(nil, nil)
	if !prop.matchPrefix("x-md-trace-id") {
		t.Error("nil prefix should use default x-md- prefix")
	}

	// 有前缀时应匹配
	prop = NewDefaultPropagator([]string{"x-md-"}, nil)
	if !prop.matchPrefix("x-md-trace-id") {
		t.Error("x-md- prefix should match x-md-trace-id")
	}
	if prop.matchPrefix("authorization") {
		t.Error("authorization should not match x-md- prefix")
	}

	// 空字符串前缀匹配所有 key
	prop = NewDefaultPropagator([]string{""}, nil)
	if !prop.matchPrefix("any-key") {
		t.Error("empty string prefix should match all keys")
	}
}

func TestNoopPropagator(t *testing.T) {
	prop := NewNoopPropagator()

	// Extract 应返回原 context
	ctx := context.Background()
	result := prop.Extract(ctx, newMockCarrier())
	if result != ctx {
		t.Error("noop extract should return original context")
	}

	// Inject 应不修改 carrier
	carrier := newMockCarrier()
	carrier.Set("existing", "value")
	prop.Inject(ctx, carrier)

	if carrier.Get("existing") != "value" {
		t.Error("noop inject should not modify carrier")
	}
}