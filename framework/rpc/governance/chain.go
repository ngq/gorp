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
	"strings"
	"time"

	discoverycontract "github.com/ngq/gorp/framework/contract/discovery"
	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	securitycontract "github.com/ngq/gorp/framework/contract/security"
	supportcontract "github.com/ngq/gorp/framework/contract/support"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	"google.golang.org/grpc/metadata"
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
	return RetryMiddlewareWithResource(retry, nil)
}

// RetryMiddlewareWithResource retries the outbound invocation with a normalized resource name.
//
// RetryMiddlewareWithResource 使用归一化资源名对出站调用执行重试。
func RetryMiddlewareWithResource(retry resiliencecontract.Retry, resource func(service, method string) string) transportcontract.RPCClientMiddleware {
	return func(next transportcontract.RPCInvoker) transportcontract.RPCInvoker {
		if next == nil || retry == nil {
			return next
		}
		return func(ctx context.Context, service, method string, req, resp any) error {
			resourceName := normalizeRetryResource(service, method)
			if resource != nil {
				resourceName = resource(service, method)
			}
			return retry.DoForResource(ctx, resourceName, func() error {
				return next(ctx, service, method, req, resp)
			})
		}
	}
}

func normalizeRetryResource(service, method string) string {
	parts := []string{"rpc"}
	if service = sanitizeRetrySegment(service); service != "" {
		parts = append(parts, service)
	}
	if method = sanitizeRetrySegment(method); method != "" {
		parts = append(parts, method)
	}
	return strings.Join(parts, ".")
}

func sanitizeRetrySegment(segment string) string {
	segment = strings.TrimSpace(segment)
	segment = strings.Trim(segment, "/")
	if segment == "" {
		return ""
	}
	replacer := strings.NewReplacer("/", ".", " ", "_", ":", ".")
	return replacer.Replace(segment)
}

// CircuitBreakerMiddleware wraps the outbound invocation with circuit breaker protection.
// Uses the provided resource function to generate the resource key for breaker state tracking.
//
// CircuitBreakerMiddleware 使用熔断器保护包装出站调用。
// 使用提供的 resource 函数生成熔断器状态跟踪的资源键。
func CircuitBreakerMiddleware(cb resiliencecontract.CircuitBreaker, resource func(service, method string) string) transportcontract.RPCClientMiddleware {
	return func(next transportcontract.RPCInvoker) transportcontract.RPCInvoker {
		if next == nil || cb == nil {
			return next
		}
		return func(ctx context.Context, service, method string, req, resp any) error {
			resourceName := normalizeRetryResource(service, method)
			if resource != nil {
				resourceName = resource(service, method)
			}
			return cb.Do(ctx, resourceName, func() error {
				return next(ctx, service, method, req, resp)
			})
		}
	}
}

// TracingMiddleware creates a span for the outbound RPC invocation.
// Records service and method as span attributes, and captures errors on span end.
//
// TracingMiddleware 为出站 RPC 调用创建 span。
// 将 service 和 method 记录为 span attributes，并在结束时捕获错误。
func TracingMiddleware(tracer observabilitycontract.Tracer, serviceName string) transportcontract.RPCClientMiddleware {
	return func(next transportcontract.RPCInvoker) transportcontract.RPCInvoker {
		if next == nil || tracer == nil {
			return next
		}
		return func(ctx context.Context, service, method string, req, resp any) error {
			spanName := service + "/" + method
			ctx, span := tracer.StartSpan(ctx, spanName, func(cfg *observabilitycontract.SpanConfig) {
				cfg.Kind = observabilitycontract.SpanKindClient
			})
			span.SetAttributes(map[string]interface{}{
				"rpc.service":  service,
				"rpc.method":   method,
				"rpc.system":   "grpc",
				"rpc.target":   serviceName,
			})
			err := next(ctx, service, method, req, resp)
			if err != nil {
				span.SetError(err)
				span.SetStatus(observabilitycontract.SpanStatusCodeError, err.Error())
			} else {
				span.SetStatus(observabilitycontract.SpanStatusCodeOk, "")
			}
			span.End()
			return err
		}
	}
}

// MetadataMiddleware propagates metadata from context to the outgoing RPC call.
// Uses the MetadataPropagator to inject context metadata into the gRPC outgoing context.
//
// MetadataMiddleware 将 metadata 从 context 传播到出站 RPC 调用。
// 使用 MetadataPropagator 将 context metadata 注入到 gRPC outgoing context。
func MetadataMiddleware(propagator transportcontract.MetadataPropagator) transportcontract.RPCClientMiddleware {
	return func(next transportcontract.RPCInvoker) transportcontract.RPCInvoker {
		if next == nil || propagator == nil {
			return next
		}
		return func(ctx context.Context, service, method string, req, resp any) error {
			md, ok := metadata.FromOutgoingContext(ctx)
			if !ok {
				md = metadata.New(nil)
			}
			carrier := NewGRPCMetadataCarrier(md)
			propagator.Inject(ctx, carrier)
			ctx = metadata.NewOutgoingContext(ctx, md)
			return next(ctx, service, method, req, resp)
		}
	}
}

// ServiceAuthMiddleware injects service authentication token into outgoing RPC metadata.
// Uses the ServiceTokenIssuer to generate a token for the target service.
//
// ServiceAuthMiddleware 将服务认证令牌注入到出站 RPC metadata。
// 使用 ServiceTokenIssuer 为目标服务生成令牌。
func ServiceAuthMiddleware(issuer securitycontract.ServiceTokenIssuer) transportcontract.RPCClientMiddleware {
	return func(next transportcontract.RPCInvoker) transportcontract.RPCInvoker {
		if next == nil || issuer == nil {
			return next
		}
		return func(ctx context.Context, service, method string, req, resp any) error {
			token, err := issuer.GenerateToken(ctx, service)
			if err == nil && token != "" {
				md, ok := metadata.FromOutgoingContext(ctx)
				if !ok {
					md = metadata.New(nil)
				}
				md.Set("x-service-token", token)
				ctx = metadata.NewOutgoingContext(ctx, md)
			}
			return next(ctx, service, method, req, resp)
		}
	}
}

// TraceIDMiddleware injects trace ID into outgoing RPC metadata for distributed tracing correlation.
// Uses supportcontract.FromTraceIDContext to extract trace ID from context.
//
// TraceIDMiddleware 将 trace ID 注入到出站 RPC metadata，用于分布式追踪关联。
// 使用 supportcontract.FromTraceIDContext 从 context 提取 trace ID。
func TraceIDMiddleware() transportcontract.RPCClientMiddleware {
	return func(next transportcontract.RPCInvoker) transportcontract.RPCInvoker {
		if next == nil {
			return next
		}
		return func(ctx context.Context, service, method string, req, resp any) error {
			if traceID, ok := supportcontract.FromTraceIDContext(ctx); ok && traceID != "" {
				md, ok := metadata.FromOutgoingContext(ctx)
				if !ok {
					md = metadata.New(nil)
				}
				md.Set("x-trace-id", traceID)
				ctx = metadata.NewOutgoingContext(ctx, md)
			}
			return next(ctx, service, method, req, resp)
		}
	}
}

// SelectorMiddleware selects a service instance before making the outbound RPC call.
// Uses the Selector to choose an instance from the registry and updates the service parameter.
// Calls the DoneFunc after invocation to report feedback to the selector.
//
// SelectorMiddleware 在出站 RPC 调用前选择服务实例。
// 使用 Selector 从 registry 选择实例并更新 service 参数。
// 在调用后调用 DoneFunc 向选择器报告反馈。
func SelectorMiddleware(selector discoverycontract.Selector, registry transportcontract.ServiceRegistry) transportcontract.RPCClientMiddleware {
	return func(next transportcontract.RPCInvoker) transportcontract.RPCInvoker {
		if next == nil || selector == nil || registry == nil {
			return next
		}
		return func(ctx context.Context, service, method string, req, resp any) error {
			instances, err := registry.Discover(ctx, service)
			if err != nil || len(instances) == 0 {
				return next(ctx, service, method, req, resp)
			}
			instance, done, err := selector.Select(ctx, instances)
			if err != nil {
				return next(ctx, service, method, req, resp)
			}
			// Replace service with the actual instance address for subsequent middleware
			// 将 service 替换为实际实例地址，供后续中间件使用
			targetAddr := instance.Address
			if done != nil {
				startedAt := time.Now()
				callErr := next(ctx, targetAddr, method, req, resp)
				latency := time.Since(startedAt)
				if latency <= 0 {
					latency = time.Nanosecond
				}
				done(ctx, discoverycontract.DoneInfo{
					Err:           callErr,
					BytesSent:     true,
					BytesReceived: callErr == nil,
					Latency:       latency,
				})
				return callErr
			}
			return next(ctx, targetAddr, method, req, resp)
		}
	}
}

// LoadSheddingMiddleware wraps the outbound invocation with load-shedding protection.
// Uses the LoadShedder to check if the request should be rejected due to system overload.
// This middleware protects the client from making requests when the system is under heavy load.
//
// LoadSheddingMiddleware 使用过载保护包装出站调用。
// 使用 LoadShedder 检查请求是否应因系统过载而被拒绝。
// 此中间件保护客户端在系统负载过高时不发送请求。
func LoadSheddingMiddleware(ls resiliencecontract.LoadShedder, resource func(service, method string) string) transportcontract.RPCClientMiddleware {
	return func(next transportcontract.RPCInvoker) transportcontract.RPCInvoker {
		if next == nil || ls == nil {
			return next
		}
		return func(ctx context.Context, service, method string, req, resp any) error {
			resourceName := normalizeRetryResource(service, method)
			if resource != nil {
				resourceName = resource(service, method)
			}
			// Check if request should be rejected due to overload
			// 检查请求是否应因过载而被拒绝
			if err := ls.Allow(ctx, resourceName); err != nil {
				return err
			}
			// Execute the call and report completion
			// 执行调用并报告完成
			callErr := next(ctx, service, method, req, resp)
			ls.Done(ctx, resourceName, callErr)
			return callErr
		}
	}
}
