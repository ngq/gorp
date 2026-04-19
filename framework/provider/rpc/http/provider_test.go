package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ngq/gorp/framework/contract"
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

func (cb *testCircuitBreaker) State(ctx context.Context, resource string) contract.CircuitBreakerState {
	return contract.CircuitBreakerStateClosed
}

func TestClientCall_UsesCircuitBreaker(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	cb := &testCircuitBreaker{}
	client := NewClient(
		&contract.RPCConfig{Mode: "http", BaseURL: server.URL, TimeoutMS: 1000},
		nil,
		nil,
		nil,
		nil,
		nil,
		cb,
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
