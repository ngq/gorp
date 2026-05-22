// Package governance_test provides unit tests for RPC governance preset and interceptor chain.
//
// 适用场景：
// - 验证 RPC governance preset 的重试、超时、熔断等拦截器链配置。
// - 确保 preset 组装和参数传递正确。
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
	if len(order) != 8 {
		t.Fatalf("expected 8 outbound governance stages (including loadshedding), got %d", len(order))
	}
	expected := []string{"selector", "timeout", "tracing", "metadata", "serviceauth", "loadshedding", "breaker", "retry"}
	for i := range expected {
		if order[i] != expected[i] {
			t.Fatalf("expected stable order %v, got %v", expected, order)
		}
	}
	if order[0] != "selector" || order[len(order)-1] != "retry" {
		t.Fatalf("unexpected order %v", order)
	}
}

// TestDefaultClientPresetOptionsZeroValueRemainsStable verifies DefaultClientPresetOptions zero value is stable.
//
// TestDefaultClientPresetOptionsZeroValueRemainsStable 验证 DefaultClientPresetOptions 零值保持稳定。
func TestDefaultClientPresetOptionsZeroValueRemainsStable(t *testing.T) {
	var opts DefaultClientPresetOptions
	if opts.Timeout != 0 {
		t.Fatalf("expected zero timeout default, got %s", opts.Timeout)
	}
}

// TestChainSkipsNilAndPreservesDeclaredOrder verifies Chain skips nil interceptors and preserves order.
//
// TestChainSkipsNilAndPreservesDeclaredOrder 验证 Chain 跳过 nil 拦截器并保持声明顺序。
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

// TestApplyReturnsNilWhenInvokerIsNil verifies Apply returns nil when the base invoker is nil.
//
// TestApplyReturnsNilWhenInvokerIsNil 验证 Apply 在基础 invoker 为 nil 时返回 nil。
func TestApplyReturnsNilWhenInvokerIsNil(t *testing.T) {
	if Apply(nil, TimeoutMiddleware(time.Second)) != nil {
		t.Fatal("expected nil invoker to stay nil")
	}
}

// TestTimeoutMiddlewareNoopWhenDisabled verifies TimeoutMiddleware is a noop when timeout is zero.
//
// TestTimeoutMiddlewareNoopWhenDisabled 验证 TimeoutMiddleware 在超时为零时是空操作。
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

// TestRetryMiddlewareWithResourceUsesNormalizedResource verifies RetryMiddleware normalizes resource keys.
//
// TestRetryMiddlewareWithResourceUsesNormalizedResource 验证 RetryMiddleware 使用规范化的资源键。
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
