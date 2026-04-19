package grpc

import (
	"context"
	"testing"

	"github.com/ngq/gorp/framework/contract"
	gogrpc "google.golang.org/grpc"
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

func TestCircuitBreakerUnaryInterceptor_UsesNormalizedResource(t *testing.T) {
	cb := &testCircuitBreaker{}
	client := NewClient(
		&contract.RPCConfig{Mode: "grpc", TimeoutMS: 1000},
		nil,
		nil,
		nil,
		nil,
		nil,
		cb,
	)

	interceptor := client.circuitBreakerUnaryInterceptor("user-service")
	err := interceptor(
		context.Background(),
		"/user.v1.UserService/GetUser",
		nil,
		nil,
		nil,
		func(ctx context.Context, method string, req, reply interface{}, cc *gogrpc.ClientConn, opts ...gogrpc.CallOption) error {
			return nil
		},
	)
	if err != nil {
		t.Fatalf("interceptor returned error: %v", err)
	}
	if !cb.called {
		t.Fatal("expected circuit breaker to wrap grpc invoke")
	}
	if cb.resource != "rpc.grpc.user-service.user.v1.UserService.GetUser" {
		t.Fatalf("unexpected resource %q", cb.resource)
	}
}
