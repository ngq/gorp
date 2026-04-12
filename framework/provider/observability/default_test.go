package observability

import (
	"context"
	"testing"

	"github.com/ngq/gorp/framework/contract"
	"github.com/stretchr/testify/assert"
)

func TestNoopTracer(t *testing.T) {
	tracer := NewNoopTracer()

	// 测试 StartSpan
	ctx, span := tracer.StartSpan(context.Background(), "test-span")
	assert.NotNil(t, span)
	assert.NotNil(t, ctx)

	// 测试 SpanFromContext
	spanFromCtx := tracer.SpanFromContext(context.Background())
	assert.NotNil(t, spanFromCtx)

	// 测试 Inject
	carrier := &mockCarrier{}
	err := tracer.Inject(context.Background(), carrier)
	assert.NoError(t, err)

	// 测试 Extract
	ctx2, err := tracer.Extract(context.Background(), carrier)
	assert.NoError(t, err)
	assert.NotNil(t, ctx2)
}

func TestNoopSpan(t *testing.T) {
	span := &NoopSpan{}

	// 测试 End
	span.End()
	span.End(contract.SpanEndOption(func(c *contract.SpanEndConfig) {}))

	// 测试 AddEvent
	span.AddEvent("test-event", map[string]interface{}{"key": "value"})

	// 测试 SetTag
	span.SetTag("key", "value")

	// 测试 SetAttributes
	span.SetAttributes(map[string]interface{}{"key": "value"})

	// 测试 SetError
	span.SetError(assert.AnError)

	// 测试 SetStatus
	span.SetStatus(contract.SpanStatusCodeError, "test error")

	// 测试 SpanContext
	sc := span.SpanContext()
	assert.Equal(t, contract.SpanContext{}, sc)

	// 测试 IsRecording
	assert.False(t, span.IsRecording())

	// 测试 Context
	ctx := span.Context()
	assert.NotNil(t, ctx)
}

func TestDefaultObservability(t *testing.T) {
	obs := NewDefaultObservability(
		&PrometheusMetrics{},
		NewNoopTracer(),
		nil, // logger
		nil, // error reporter
	)

	assert.NotNil(t, obs.Metrics())
	assert.NotNil(t, obs.Tracer())
	assert.Nil(t, obs.Logger())
	assert.Nil(t, obs.ErrorReporter())
}

// mockCarrier 用于测试 TextMapCarrier
type mockCarrier struct {
	data map[string]string
}

func (m *mockCarrier) Get(key string) string {
	if m.data == nil {
		return ""
	}
	return m.data[key]
}

func (m *mockCarrier) Set(key string, value string) {
	if m.data == nil {
		m.data = make(map[string]string)
	}
	m.data[key] = value
}

func (m *mockCarrier) Keys() []string {
	if m.data == nil {
		return nil
	}
	keys := make([]string, 0, len(m.data))
	for k := range m.data {
		keys = append(keys, k)
	}
	return keys
}