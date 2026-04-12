package noop

import (
	"context"
	"testing"

	"github.com/ngq/gorp/framework/contract"
	"github.com/stretchr/testify/assert"
)

func TestNoopTracer(t *testing.T) {
	tracer := &noopTracer{}

	// 测试 StartSpan
	ctx, span := tracer.StartSpan(context.Background(), "test-span")
	assert.NotNil(t, ctx)
	assert.NotNil(t, span)

	// 测试 Span 操作
	span.End()
	span.AddEvent("test-event", nil)
	span.SetTag("key", "value")
	span.SetAttributes(map[string]any{"key": "value"})
	span.SetError(nil)
	span.SetStatus(contract.SpanStatusCodeOk, "")

	// 测试 SpanContext
	sc := span.SpanContext()
	assert.Empty(t, sc.TraceID)
	assert.Empty(t, sc.SpanID)

	// 测试 IsRecording
	assert.False(t, span.IsRecording())
}

func TestNoopTracerProvider(t *testing.T) {
	provider := &noopTracerProvider{}

	// 测试 Tracer
	tracer := provider.Tracer("test-service")
	assert.NotNil(t, tracer)

	// 测试 Shutdown
	err := provider.Shutdown(context.Background())
	assert.NoError(t, err)

	// 测试 ForceFlush
	err = provider.ForceFlush(context.Background())
	assert.NoError(t, err)
}

func TestNoopInjectExtract(t *testing.T) {
	tracer := &noopTracer{}
	carrier := &mockCarrier{data: make(map[string]string)}

	// 测试 Inject
	err := tracer.Inject(context.Background(), carrier)
	assert.NoError(t, err)

	// 测试 Extract
	ctx, err := tracer.Extract(context.Background(), carrier)
	assert.NoError(t, err)
	assert.NotNil(t, ctx)
}

func TestProvider(t *testing.T) {
	p := NewProvider()

	assert.Equal(t, "tracing.noop", p.Name())
	assert.True(t, p.IsDefer())
	assert.ElementsMatch(t, []string{
		contract.TracerKey,
		contract.TracerProviderKey,
	}, p.Provides())
}

// mockCarrier 用于测试的 mock carrier
type mockCarrier struct {
	data map[string]string
}

func (c *mockCarrier) Get(key string) string {
	return c.data[key]
}

func (c *mockCarrier) Set(key string, value string) {
	c.data[key] = value
}

func (c *mockCarrier) Keys() []string {
	keys := make([]string, 0, len(c.data))
	for k := range c.data {
		keys = append(keys, k)
	}
	return keys
}