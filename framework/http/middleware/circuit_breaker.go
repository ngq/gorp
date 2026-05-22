// Application scenarios:
// - Stop unhealthy routes from continuously receiving traffic after repeated failures.
// - Fail fast on known-bad routes or dependencies and protect the rest of the service.
// - Align HTTP protection behavior with the framework's unified circuit-breaker capability.
//
// 适用场景：
// - 在连续失败后阻止不健康路由继续接收流量。
// - 对已知异常路由或依赖快速失败，保护其余服务能力。
// - 让 HTTP 保护行为与框架统一熔断能力保持一致。
package middleware

import (
	"fmt"
	"net/http"
	"strings"

	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

// CircuitBreaker protects HTTP requests with the configured circuit breaker.
//
// CircuitBreaker 使用给定的熔断器保护 HTTP 请求。
//
// 使用方式：
//
//	cb, _ := container.Make[resiliencecontract.CircuitBreaker](c, resiliencecontract.CircuitBreakerKey)
//	router.Use(middleware.CircuitBreaker(cb, "external-api"))
func CircuitBreaker(cb resiliencecontract.CircuitBreaker, resource string) transportcontract.Middleware {
	return func(next transportcontract.Handler) transportcontract.Handler {
		return func(c transportcontract.Context) {
			if cb == nil {
				if next != nil {
					next(c)
				}
				return
			}

			target := circuitBreakerResource(c, resource)
			if err := cb.Allow(c.Context(), target); err != nil {
				respondCircuitBreakerOpen(c)
				return
			}

			defer func() {
				if rec := recover(); rec != nil {
					cb.RecordFailure(c.Context(), target, fmt.Errorf("panic: %v", rec))
					panic(rec)
				}

				status := c.ResponseStatus()
				if status >= http.StatusInternalServerError {
					cb.RecordFailure(c.Context(), target, resiliencecontract.ServiceUnavailable(http.StatusText(status)))
					return
				}
				cb.RecordSuccess(c.Context(), target)
			}()

			if next != nil {
				next(c)
			}
		}
	}
}

// CircuitBreakerFromContainer 自动从容器获取熔断器并应用熔断中间件。
// 无需手动从容器获取 circuit breaker，直接使用即可。
//
// 使用方式：
//
//	router.Use(middleware.CircuitBreakerFromContainer(container, "external-api"))
//
// 如果容器中未注册 CircuitBreaker，中间件会静默跳过（不报错）。
func CircuitBreakerFromContainer(container runtimecontract.Container, resource string) transportcontract.Middleware {
	return func(next transportcontract.Handler) transportcontract.Handler {
		return func(c transportcontract.Context) {
			// 从容器获取熔断器
			if container == nil || !container.IsBind(resiliencecontract.CircuitBreakerKey) {
				if next != nil {
					next(c)
				}
				return
			}

			cbAny, err := container.Make(resiliencecontract.CircuitBreakerKey)
			if err != nil {
				if next != nil {
					next(c)
				}
				return
			}

			cb, ok := cbAny.(resiliencecontract.CircuitBreaker)
			if !ok || cb == nil {
				if next != nil {
					next(c)
				}
				return
			}

			// 获取熔断资源
			target := circuitBreakerResource(c, resource)
			if err := cb.Allow(c.Context(), target); err != nil {
				respondCircuitBreakerOpen(c)
				return
			}

			defer func() {
				if rec := recover(); rec != nil {
					cb.RecordFailure(c.Context(), target, fmt.Errorf("panic: %v", rec))
					panic(rec)
				}

				status := c.ResponseStatus()
				if status >= http.StatusInternalServerError {
					cb.RecordFailure(c.Context(), target, resiliencecontract.ServiceUnavailable(http.StatusText(status)))
					return
				}
				cb.RecordSuccess(c.Context(), target)
			}()

			if next != nil {
				next(c)
			}
		}
	}
}

// circuitBreakerResource resolves the breaker resource key for the current request.
//
// circuitBreakerResource 解析当前请求对应的熔断资源 key。
func circuitBreakerResource(c transportcontract.Context, resource string) string {
	if strings.TrimSpace(resource) != "" {
		return resource
	}

	route := ""
	method := ""
	if c != nil {
		route = c.RoutePath()
		if req := c.Request(); req != nil {
			method = req.Method
			if route == "" && req.URL != nil {
				route = req.URL.Path
			}
		}
	}

	parts := []string{"http"}
	if method != "" {
		parts = append(parts, strings.ToLower(strings.TrimSpace(method)))
	}
	if route != "" {
		parts = append(parts, sanitizeCircuitBreakerHTTPRoute(route))
	}
	return strings.Join(parts, ".")
}

// sanitizeCircuitBreakerHTTPRoute normalizes a route into a breaker-friendly key segment.
//
// sanitizeCircuitBreakerHTTPRoute 将路由归一化为适合熔断资源名的 key 片段。
func sanitizeCircuitBreakerHTTPRoute(route string) string {
	route = strings.TrimSpace(route)
	route = strings.Trim(route, "/")
	if route == "" {
		return "root"
	}
	replacer := strings.NewReplacer("/", ".", ":", "", "-", "_", " ", "_")
	return replacer.Replace(route)
}

// respondCircuitBreakerOpen writes the unified service-unavailable response for open breakers.
//
// respondCircuitBreakerOpen 在熔断器打开时输出统一的服务不可用响应。
func respondCircuitBreakerOpen(c transportcontract.Context) {
	if gc, ok := unwrapGinContext(c); ok {
		writeGinResponseHeaders(gc)
		resp := Response{
			Code:    CodeServiceUnavailable,
			Message: "circuit breaker is open",
		}
		gc.JSON(http.StatusServiceUnavailable, resp)
		gc.Abort()
		return
	}

	c.JSON(http.StatusServiceUnavailable, map[string]any{
		"code":    CodeServiceUnavailable,
		"message": "circuit breaker is open",
	})
}
