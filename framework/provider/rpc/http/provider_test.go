package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	supportcontract "github.com/ngq/gorp/framework/contract/support"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

type testCircuitBreaker struct {
	resource string
	called   bool
}

func (cb *testCircuitBreaker) Allow(ctx context.Context, resource string) error {
	return nil
}

func (cb *testCircuitBreaker) RecordSuccess(ctx context.Context, resource string) {}

func (cb *testCircuitBreaker) RecordFailure(ctx context.Context, resource string, err error) {}

func (cb *testCircuitBreaker) Do(ctx context.Context, resource string, fn func() error) error {
	cb.resource = resource
	cb.called = true
	return fn()
}

func (cb *testCircuitBreaker) State(ctx context.Context, resource string) resiliencecontract.CircuitBreakerState {
	return resiliencecontract.CircuitBreakerStateClosed
}

func TestClientCall_UsesCircuitBreaker(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	cb := &testCircuitBreaker{}
	client := NewClient(
		&transportcontract.RPCConfig{Mode: "http", BaseURL: server.URL, TimeoutMS: 1000},
		nil,
		nil,
		nil,
		nil,
		nil,
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

func TestClientCall_PropagatesTraceIDFromContext(t *testing.T) {
	var gotTraceID string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotTraceID = r.Header.Get("X-Trace-ID")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	client := NewClient(
		&transportcontract.RPCConfig{Mode: "http", BaseURL: server.URL, TimeoutMS: 1000},
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
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
