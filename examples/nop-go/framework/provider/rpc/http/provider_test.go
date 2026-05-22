// Package http_test provides unit tests for HTTP RPC transport and handler registration.
//
// 适用场景：
// - 验证 HTTP RPC server 的启动、路由注册和中间件链路。
// - 确保 HTTP 场景下的 discovery、tracing 集成正确。
package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	discoverycontract "github.com/ngq/gorp/framework/contract/discovery"
	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	supportcontract "github.com/ngq/gorp/framework/contract/support"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

type testCircuitBreaker struct {
	resource string
	called   bool
}

func (cb *testCircuitBreaker) Allow(ctx context.Context, resource string) error   { return nil }
func (cb *testCircuitBreaker) RecordSuccess(ctx context.Context, resource string) {}
func (cb *testCircuitBreaker) RecordFailure(ctx context.Context, resource string, err error) {
}
func (cb *testCircuitBreaker) Do(ctx context.Context, resource string, fn func() error) error {
	cb.resource = resource
	cb.called = true
	return fn()
}
func (cb *testCircuitBreaker) State(ctx context.Context, resource string) resiliencecontract.CircuitBreakerState {
	return resiliencecontract.CircuitBreakerStateClosed
}

type captureRetry struct {
	resource string
	calls    int
}

func (r *captureRetry) Do(ctx context.Context, fn func() error) error {
	return r.DoForResource(ctx, "", fn)
}

func (r *captureRetry) DoForResource(ctx context.Context, resource string, fn func() error) error {
	r.resource = resource
	r.calls++
	return fn()
}

func (r *captureRetry) DoWithResult(ctx context.Context, fn func() (any, error)) (any, error) {
	return fn()
}

func (r *captureRetry) IsRetryable(err error) bool { return false }

type tokenIssuer struct{ token string }

func (i tokenIssuer) GenerateToken(ctx context.Context, targetService string) (string, error) {
	return i.token, nil
}

type headerTracer struct{}

func (headerTracer) StartSpan(ctx context.Context, name string, opts ...observabilitycontract.SpanOption) (context.Context, observabilitycontract.Span) {
	return ctx, noopSpan{}
}
func (headerTracer) SpanFromContext(ctx context.Context) observabilitycontract.Span {
	return noopSpan{}
}
func (headerTracer) Inject(ctx context.Context, carrier observabilitycontract.TextMapCarrier) error {
	carrier.Set("X-Test-Trace", "trace-from-tracer")
	return nil
}
func (headerTracer) Extract(ctx context.Context, carrier observabilitycontract.TextMapCarrier) (context.Context, error) {
	return ctx, nil
}

type noopSpan struct{}

func (noopSpan) End(options ...observabilitycontract.SpanEndOption)                      {}
func (noopSpan) AddEvent(name string, attributes map[string]interface{})                 {}
func (noopSpan) SetTag(key string, value interface{})                                    {}
func (noopSpan) SetAttributes(attributes map[string]interface{})                         {}
func (noopSpan) SetError(err error)                                                      {}
func (noopSpan) SetStatus(code observabilitycontract.SpanStatusCode, description string) {}
func (noopSpan) SpanContext() observabilitycontract.SpanContext {
	return observabilitycontract.SpanContext{}
}
func (noopSpan) IsRecording() bool        { return false }
func (noopSpan) Context() context.Context { return context.Background() }

type fakeMetadataPropagator struct{}

func (fakeMetadataPropagator) Inject(ctx context.Context, carrier transportcontract.MetadataCarrier) {
	carrier.Set("X-MD-Region", "cn-hz")
}
func (fakeMetadataPropagator) Extract(ctx context.Context, carrier transportcontract.MetadataCarrier) context.Context {
	return ctx
}

type stubRegistry struct {
	instances []transportcontract.ServiceInstance
}

func (s stubRegistry) Register(ctx context.Context, name, addr string, meta map[string]string) error {
	return nil
}
func (s stubRegistry) Deregister(ctx context.Context, name, addr string) error { return nil }
func (s stubRegistry) Discover(ctx context.Context, name string) ([]transportcontract.ServiceInstance, error) {
	return s.instances, nil
}
func (s stubRegistry) Close() error { return nil }

type captureSelector struct {
	doneInfo discoverycontract.DoneInfo
}

func (s *captureSelector) Select(ctx context.Context, instances []transportcontract.ServiceInstance, opts ...discoverycontract.SelectOption) (transportcontract.ServiceInstance, discoverycontract.DoneFunc, error) {
	return instances[0], func(ctx context.Context, info discoverycontract.DoneInfo) {
		s.doneInfo = info
	}, nil
}

// TestClientCallUsesCircuitBreaker 验证 HTTP 客户端调用使用熔断器包装。
//
// 中文说明：
// - 资源名格式为 "rpc.http.{service}.{path}"。
// - 响应体正确解码为 map[string]bool。
func TestClientCallUsesCircuitBreaker(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	cb := &testCircuitBreaker{}
	client := NewClient(
		&transportcontract.RPCConfig{Mode: "http", BaseURL: server.URL, TimeoutMS: 1000},
		nil, nil, nil, nil, nil,
		cb,
		nil,
	)

	var resp map[string]bool
	if err := client.Call(context.Background(), "user-service", "/api/user/get", map[string]string{"id": "1"}, &resp); err != nil {
		t.Fatalf("Call returned error: %v", err)
	}
	if !cb.called {
		t.Fatal("expected circuit breaker to wrap http call")
	}
	if cb.resource != "rpc.http.user-service.api.user.get" {
		t.Fatalf("unexpected resource %q", cb.resource)
	}
	if !resp["ok"] {
		t.Fatalf("expected response body to be decoded, got %#v", resp)
	}
}

// TestClientCallPropagatesTraceIDFromContext 验证 HTTP 客户端从 context 传播 trace ID。
//
// 中文说明：
// - NewTraceIDContext 设置的 trace ID 会随请求传播到下游服务。
func TestClientCallPropagatesTraceIDFromContext(t *testing.T) {
	var gotTraceID string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotTraceID = r.Header.Get("X-Trace-ID")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	client := NewClient(
		&transportcontract.RPCConfig{Mode: "http", BaseURL: server.URL, TimeoutMS: 1000},
		nil, nil, nil, nil, nil, nil, nil,
	)

	ctx := supportcontract.NewTraceIDContext(context.Background(), "trace-123")
	var resp map[string]bool
	if err := client.Call(ctx, "user-service", "/api/user/get", map[string]string{"id": "1"}, &resp); err != nil {
		t.Fatalf("Call returned error: %v", err)
	}
	if gotTraceID != "trace-123" {
		t.Fatalf("expected trace id trace-123, got %q", gotTraceID)
	}
}

// TestClientCallInjectsServiceAuthTracingAndMetadata 验证 HTTP 客户端注入 service token、tracing 和 metadata。
//
// 中文说明：
// - service token、tracing header 和 metadata 均正确注入到下游 HTTP 请求。
func TestClientCallInjectsServiceAuthTracingAndMetadata(t *testing.T) {
	var gotToken, gotTrace, gotMetadata string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotToken = r.Header.Get("X-Service-Token")
		gotTrace = r.Header.Get("X-Test-Trace")
		gotMetadata = r.Header.Get("X-MD-Region")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	client := NewClient(
		&transportcontract.RPCConfig{Mode: "http", BaseURL: server.URL, TimeoutMS: 1000},
		nil, nil,
		fakeMetadataPropagator{},
		tokenIssuer{token: "svc-token"},
		headerTracer{},
		nil,
		nil,
	)

	var resp map[string]bool
	if err := client.Call(context.Background(), "billing-service", "/pay", map[string]string{"id": "1"}, &resp); err != nil {
		t.Fatalf("Call returned error: %v", err)
	}
	if gotToken != "svc-token" {
		t.Fatalf("expected service token header, got %q", gotToken)
	}
	if gotTrace != "trace-from-tracer" {
		t.Fatalf("expected trace header from tracer, got %q", gotTrace)
	}
	if gotMetadata != "cn-hz" {
		t.Fatalf("expected propagated metadata header, got %q", gotMetadata)
	}
}

// TestClientCallRetryUsesResourceAwarePath 验证 HTTP 客户端使用资源感知的路径作为重试键。
//
// 中文说明：
// - retry.DoForResource 使用标准化后的资源名 "rpc.http.{service}.{path}"。
func TestClientCallRetryUsesResourceAwarePath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	retry := &captureRetry{}
	client := NewClient(
		&transportcontract.RPCConfig{Mode: "http", BaseURL: server.URL, TimeoutMS: 1000},
		nil, nil, nil, nil, nil, nil, retry,
	)

	var resp map[string]bool
	if err := client.Call(context.Background(), "user-service", "/orders/create", map[string]string{"id": "1"}, &resp); err != nil {
		t.Fatalf("Call returned error: %v", err)
	}
	if retry.calls != 1 {
		t.Fatalf("expected retry wrapper invoked once, got %d", retry.calls)
	}
	if retry.resource != "rpc.http.user-service.orders.create" {
		t.Fatalf("expected resource-aware retry key, got %q", retry.resource)
	}
}

// TestClientCallReportsLatencyToSelector 验证 HTTP 客户端向 selector 报告延迟和字节数。
//
// 中文说明：
// - DoneInfo 包含 Latency、BytesSent、BytesReceived 等指标。
func TestClientCallReportsLatencyToSelector(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	selector := &captureSelector{}
	client := NewClient(
		&transportcontract.RPCConfig{Mode: "http", TimeoutMS: 1000},
		stubRegistry{instances: []transportcontract.ServiceInstance{{ID: "1", Address: server.URL, Healthy: true}}},
		selector,
		nil, nil, nil, nil, nil,
	)

	var resp map[string]bool
	if err := client.Call(context.Background(), "user-service", "/slow", map[string]string{"id": "1"}, &resp); err != nil {
		t.Fatalf("Call returned error: %v", err)
	}
	if selector.doneInfo.Latency <= 0 {
		t.Fatalf("expected selector latency feedback > 0, got %s", selector.doneInfo.Latency)
	}
	if !selector.doneInfo.BytesSent || !selector.doneInfo.BytesReceived {
		t.Fatalf("expected selector bytes flags true, got %+v", selector.doneInfo)
	}
}
