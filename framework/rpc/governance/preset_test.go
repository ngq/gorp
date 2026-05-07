package governance

import (
	"context"
	"testing"
	"time"

	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

type captureRetry struct {
	resource string
}

func (r *captureRetry) Do(ctx context.Context, fn func() error) error { return fn() }
func (r *captureRetry) DoForResource(ctx context.Context, resource string, fn func() error) error {
	r.resource = resource
	return fn()
}
func (r *captureRetry) DoWithResult(ctx context.Context, fn func() (any, error)) (any, error) {
	return fn()
}
func (r *captureRetry) IsRetryable(err error) bool { return false }

// TestDefaultClientPresetOrderStable verifies the outbound RPC governance order remains stable.
//
// TestDefaultClientPresetOrderStable 验证出站 RPC 治理顺序保持稳定。
func TestDefaultClientPresetOrderStable(t *testing.T) {
	order := DefaultClientPresetOrder()
	if len(order) != 7 {
		t.Fatalf("expected 7 outbound governance stages, got %d", len(order))
	}
	expected := []string{"selector", "timeout", "tracing", "metadata", "serviceauth", "breaker", "retry"}
	for i := range expected {
		if order[i] != expected[i] {
			t.Fatalf("expected stable order %v, got %v", expected, order)
		}
	}
	if order[0] != "selector" || order[len(order)-1] != "retry" {
		t.Fatalf("unexpected order %v", order)
	}
}

func TestDefaultClientPresetOptionsZeroValueRemainsStable(t *testing.T) {
	var opts DefaultClientPresetOptions
	if opts.Timeout != 0 {
		t.Fatalf("expected zero timeout default, got %s", opts.Timeout)
	}
}

func TestChainSkipsNilAndPreservesDeclaredOrder(t *testing.T) {
	var calls []string
	first := func(next transportcontract.RPCInvoker) transportcontract.RPCInvoker {
		return func(ctx context.Context, service, method string, req, resp any) error {
			calls = append(calls, "first-before")
			err := next(ctx, service, method, req, resp)
			calls = append(calls, "first-after")
			return err
		}
	}
	second := func(next transportcontract.RPCInvoker) transportcontract.RPCInvoker {
		return func(ctx context.Context, service, method string, req, resp any) error {
			calls = append(calls, "second-before")
			err := next(ctx, service, method, req, resp)
			calls = append(calls, "second-after")
			return err
		}
	}

	invoker := Chain(first, nil, second)(func(ctx context.Context, service, method string, req, resp any) error {
		calls = append(calls, "handler")
		return nil
	})
	if err := invoker(context.Background(), "svc", "method", nil, nil); err != nil {
		t.Fatalf("unexpected invoke error: %v", err)
	}
	expected := []string{"first-before", "second-before", "handler", "second-after", "first-after"}
	for i := range expected {
		if calls[i] != expected[i] {
			t.Fatalf("expected order %v, got %v", expected, calls)
		}
	}
}

func TestApplyReturnsNilWhenInvokerIsNil(t *testing.T) {
	if Apply(nil, TimeoutMiddleware(time.Second)) != nil {
		t.Fatal("expected nil invoker to stay nil")
	}
}

func TestTimeoutMiddlewareNoopWhenDisabled(t *testing.T) {
	next := func(ctx context.Context, service, method string, req, resp any) error { return nil }
	wrapped := TimeoutMiddleware(0)(next)
	if wrapped == nil {
		t.Fatal("expected wrapped invoker")
	}
	if err := wrapped(context.Background(), "svc", "method", nil, nil); err != nil {
		t.Fatalf("unexpected invoke error: %v", err)
	}
}

func TestRetryMiddlewareWithResourceUsesNormalizedResource(t *testing.T) {
	var _ resiliencecontract.Retry = (*captureRetry)(nil)

	retry := &captureRetry{}
	invoker := RetryMiddlewareWithResource(retry, func(service, method string) string {
		return "rpc.grpc." + service + "." + method
	})(func(ctx context.Context, service, method string, req, resp any) error {
		return nil
	})
	if err := invoker(context.Background(), "user-service", "GetUser", nil, nil); err != nil {
		t.Fatalf("unexpected invoke error: %v", err)
	}
	if retry.resource != "rpc.grpc.user-service.GetUser" {
		t.Fatalf("expected resource-aware retry key, got %q", retry.resource)
	}
}
