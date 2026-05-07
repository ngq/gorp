// Application scenarios:
// - Provide a shared outbound RPC governance chain model used by HTTP and gRPC clients.
// - Keep timeout, retry, and future client-side governance logic provider-neutral.
// - Let transport-specific RPC providers reuse one middleware assembly pattern.
//
// 适用场景：
// - 为 HTTP 和 gRPC 客户端提供共享的出站 RPC 治理链模型。
// - 让 timeout、retry 以及后续 client 侧治理逻辑保持 provider-neutral。
// - 让不同 transport 的 RPC provider 复用同一套中间件装配模式。
package governance

import (
	"context"
	"time"

	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

// Chain composes outbound RPC middleware in declaration order.
//
// Chain 按声明顺序组合出站 RPC 中间件。
func Chain(middleware ...transportcontract.RPCClientMiddleware) transportcontract.RPCClientMiddleware {
	return func(next transportcontract.RPCInvoker) transportcontract.RPCInvoker {
		for i := len(middleware) - 1; i >= 0; i-- {
			if middleware[i] == nil {
				continue
			}
			next = middleware[i](next)
		}
		return next
	}
}

// Apply wraps one invoker with zero or more outbound RPC middleware.
//
// Apply 使用零个或多个出站 RPC 中间件包装一次调用。
func Apply(invoker transportcontract.RPCInvoker, middleware ...transportcontract.RPCClientMiddleware) transportcontract.RPCInvoker {
	if invoker == nil {
		return nil
	}
	return Chain(middleware...)(invoker)
}

// TimeoutMiddleware enforces one timeout around the outbound invocation.
//
// TimeoutMiddleware 为出站调用包一层超时控制。
func TimeoutMiddleware(timeout time.Duration) transportcontract.RPCClientMiddleware {
	return func(next transportcontract.RPCInvoker) transportcontract.RPCInvoker {
		if next == nil || timeout <= 0 {
			return next
		}
		return func(ctx context.Context, service, method string, req, resp any) error {
			timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			return next(timeoutCtx, service, method, req, resp)
		}
	}
}

// RetryMiddleware retries the outbound invocation through the shared retry capability when available.
//
// RetryMiddleware 在可用时通过统一 Retry 能力重试出站调用。
func RetryMiddleware(retry resiliencecontract.Retry) transportcontract.RPCClientMiddleware {
	return func(next transportcontract.RPCInvoker) transportcontract.RPCInvoker {
		if next == nil || retry == nil {
			return next
		}
		return func(ctx context.Context, service, method string, req, resp any) error {
			return retry.Do(ctx, func() error {
				return next(ctx, service, method, req, resp)
			})
		}
	}
}
