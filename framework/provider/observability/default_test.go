// Package observability_test provides unit tests for the observability default provider.
//
// 适用场景：
// - 验证可观测性默认 provider 的注册与指标/tracing 行为。
package observability

import (
	"context"
	"testing"
	"time"

	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	"github.com/stretchr/testify/assert"
)

// TestNoopTracer verifies noop tracer creates spans and handles inject/extract operations.
//
// TestNoopTracer 验证 noop tracer 创建 span 并正确处理 inject/extract 操作。
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

// TestNoopSpan verifies noop span operations are safe no-ops.
//
// TestNoopSpan 验证 noop span 的各项操作均为安全的无操作。
func TestNoopSpan(t *testing.T) {
	span := &NoopSpan{}

	// 测试 End
	span.End()
	span.End(observabilitycontract.SpanEndOption(func(c *observabilitycontract.SpanEndConfig) {}))

	// 测试 AddEvent
	span.AddEvent("test-event", map[string]interface{}{"key": "value"})

	// 测试 SetTag
	span.SetTag("key", "value")

	// 测试 SetAttributes
	span.SetAttributes(map[string]interface{}{"key": "value"})

	// 测试 SetError
	span.SetError(assert.AnError)

	// 测试 SetStatus
	span.SetStatus(observabilitycontract.SpanStatusCodeError, "test error")

	// 测试 SpanContext
	sc := span.SpanContext()
	assert.Equal(t, observabilitycontract.SpanContext{}, sc)

	// 测试 IsRecording
	assert.False(t, span.IsRecording())

	// 测试 Context
	ctx := span.Context()
	assert.NotNil(t, ctx)
}

// TestDefaultObservability verifies default observability provider initialization with components.
//
// TestDefaultObservability 验证可观测性 provider 的初始化，包含 metrics、tracer、logger 和 error reporter。
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

// TestPrometheusTracer_RecordsSpans 验证内置 tracer 生成真实 span 并记录。
func TestPrometheusTracer_RecordsSpans(t *testing.T) {
	tr := NewPrometheusTracer()
	_, span := tr.StartSpan(context.Background(), "GET /orders",
		WithSpanKindForTest(observabilitycontract.SpanKindServer))
	span.SetAttributes(map[string]interface{}{"http.method": "GET"})
	span.SetStatus(observabilitycontract.SpanStatusCodeOk, "")
	time.Sleep(time.Millisecond) // 保证时长可测量
	span.End()

	sc := span.SpanContext()
	if sc.TraceID == "" || len(sc.TraceID) != 32 {
		t.Fatalf("expected 32-char trace id, got %q", sc.TraceID)
	}
	if sc.SpanID == "" || len(sc.SpanID) != 16 {
		t.Fatalf("expected 16-char span id, got %q", sc.SpanID)
	}
	if !span.IsRecording() {
		t.Fatal("expected recording span")
	}

	recorded := tr.RecordedSpans()
	if len(recorded) != 1 {
		t.Fatalf("expected 1 recorded span, got %d", len(recorded))
	}
	if recorded[0].Name != "GET /orders" {
		t.Fatalf("expected span name GET /orders, got %q", recorded[0].Name)
	}
	if recorded[0].Attributes["http.method"] != "GET" {
		t.Fatalf("expected recorded attribute, got %v", recorded[0].Attributes)
	}
	if recorded[0].Duration <= 0 {
		t.Fatalf("expected positive duration, got %v", recorded[0].Duration)
	}
}

// TestPrometheusTracer_ChildInheritsTraceID 验证子 span 继承父 traceID。
func TestPrometheusTracer_ChildInheritsTraceID(t *testing.T) {
	tr := NewPrometheusTracer()
	ctx, parent := tr.StartSpan(context.Background(), "parent")
	parentCtx := parent.SpanContext()

	_, child := tr.StartSpan(ctx, "child")
	childCtx := child.SpanContext()
	if childCtx.TraceID != parentCtx.TraceID {
		t.Fatalf("expected child trace %s to equal parent trace %s", childCtx.TraceID, parentCtx.TraceID)
	}
	if childCtx.SpanID == parentCtx.SpanID {
		t.Fatal("child span id must differ from parent span id")
	}

	parent.End()
	child.End()
	recorded := tr.RecordedSpans()
	if len(recorded) != 2 {
		t.Fatalf("expected 2 recorded spans, got %d", len(recorded))
	}
	// 两条 span 都应以 parent 为父（child 的父 span id = parent 的 span id）。
	var childRec *RecordedSpan
	for _, r := range recorded {
		if r.Name == "child" {
			childRec = r
		}
	}
	if childRec == nil {
		t.Fatal("child span not recorded")
	}
	if childRec.ParentSpanID != parentCtx.SpanID {
		t.Fatalf("expected child parent %s to equal parent span %s", childRec.ParentSpanID, parentCtx.SpanID)
	}
}

// TestPrometheusTracer_InjectExtractTraceParent 验证 W3C traceparent 跨服务传播。
func TestPrometheusTracer_InjectExtractTraceParent(t *testing.T) {
	tr := NewPrometheusTracer()
	_, span := tr.StartSpan(context.Background(), "service-a")
	sc := span.SpanContext()

	carrier := &mockCarrier{}
	if err := tr.Inject(span.Context(), carrier); err != nil {
		t.Fatalf("inject failed: %v", err)
	}
	if tp := carrier.data["traceparent"]; tp == "" {
		t.Fatal("expected traceparent header after inject")
	}

	// 服务 B 侧 Extract 出远端父 span，其子 span 应继承同一 traceID。
	ctx, err := tr.Extract(context.Background(), carrier)
	if err != nil {
		t.Fatalf("extract failed: %v", err)
	}
	_, child := tr.StartSpan(ctx, "service-b")
	if child.SpanContext().TraceID != sc.TraceID {
		t.Fatalf("expected propagated trace %s, got %s", sc.TraceID, child.SpanContext().TraceID)
	}
	child.End()
	span.End()
}

// TestPrometheusTracer_RingBufferBound 验证环形缓冲上限。
func TestPrometheusTracer_RingBufferBound(t *testing.T) {
	tr := NewPrometheusTracer()
	for i := 0; i < maxRecordedSpans+100; i++ {
		_, s := tr.StartSpan(context.Background(), "span")
		s.End()
	}
	if got := len(tr.RecordedSpans()); got > maxRecordedSpans {
		t.Fatalf("expected at most %d recorded spans, got %d", maxRecordedSpans, got)
	}
}

// WithSpanKindForTest 构造测试用 SpanOption（default.go 的 WithSpanKind 在
// tracing/middleware 包，这里本地构造避免跨包依赖）。
func WithSpanKindForTest(kind observabilitycontract.SpanKind) observabilitycontract.SpanOption {
	return func(cfg *observabilitycontract.SpanConfig) { cfg.Kind = kind }
}
