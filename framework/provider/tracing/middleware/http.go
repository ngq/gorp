package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ngq/gorp/framework/contract"
)

// TracingMiddleware 创建 HTTP 追踪中间件。
//
// 中文说明：
// - 自动为每个请求创建 Span；
// - 从请求头提取追踪上下文（W3C TraceContext）；
// - 记录请求方法、路径、状态码等信息；
// - 支持与 OpenTelemetry 集成。
func TracingMiddleware(tracer contract.Tracer, serviceName string) contract.HTTPMiddleware {
	return func(c contract.HTTPContext, next contract.HTTPNext) {
		carrier := &httpHeaderCarrier{c.Request().Header}
		ctx, err := tracer.Extract(c.Context(), carrier)
		if err != nil {
			ctx = c.Context()
		}

		spanName := fmt.Sprintf("HTTP %s %s", c.Request().Method, c.RoutePath())
		if spanName == fmt.Sprintf("HTTP %s ", c.Request().Method) {
			spanName = fmt.Sprintf("HTTP %s %s", c.Request().Method, c.Request().URL.Path)
		}

		ctx, span := tracer.StartSpan(ctx, spanName,
			WithSpanKind(contract.SpanKindServer),
			WithAttributes(map[string]any{
				"http.method":     c.Request().Method,
				"http.url":        c.Request().URL.String(),
				"http.host":       c.Request().Host,
				"http.scheme":     c.Request().URL.Scheme,
				"http.target":     c.Request().URL.Path,
				"http.user_agent": c.Request().UserAgent(),
				"http.route":      c.RoutePath(),
				"service.name":    serviceName,
			}),
		)
		defer span.End()

		if traceID := span.SpanContext().TraceID; traceID != "" {
			ctx = contract.NewTraceIDContext(ctx, traceID)
		}
		c.SetContext(ctx)

		startTime := time.Now()
		if next != nil {
			next()
		}

		latency := time.Since(startTime)
		statusCode := c.ResponseStatus()

		span.SetAttributes(map[string]any{
			"http.status_code":  statusCode,
			"http.latency_ms":   latency.Milliseconds(),
		})

		if statusCode >= 400 {
			span.SetStatus(contract.SpanStatusCodeError, http.StatusText(statusCode))
		} else {
			span.SetStatus(contract.SpanStatusCodeOk, "")
		}

		traceID := span.SpanContext().TraceID
		if traceID != "" {
			c.Header("X-Trace-ID", traceID)
		}
	}
}

// httpHeaderCarrier 实现 TextMapCarrier 接口。
type httpHeaderCarrier struct {
	header http.Header
}

func (c *httpHeaderCarrier) Get(key string) string {
	return c.header.Get(key)
}

func (c *httpHeaderCarrier) Set(key string, value string) {
	c.header.Set(key, value)
}

func (c *httpHeaderCarrier) Keys() []string {
	keys := make([]string, 0, len(c.header))
	for k := range c.header {
		keys = append(keys, k)
	}
	return keys
}

// WithSpanKind 设置 Span 类型的选项。
func WithSpanKind(kind contract.SpanKind) contract.SpanOption {
	return func(cfg *contract.SpanConfig) {
		cfg.Kind = kind
	}
}

// WithAttributes 设置 Span 属性的选项。
func WithAttributes(attrs map[string]any) contract.SpanOption {
	return func(cfg *contract.SpanConfig) {
		if cfg.Attributes == nil {
			cfg.Attributes = make(map[string]any)
		}
		for k, v := range attrs {
			cfg.Attributes[k] = v
		}
	}
}

// TracingMiddlewareFromContainer 从容器创建追踪中间件。
//
// 中文说明：
// - 自动从容器获取 Tracer；
// - 当前默认主线下，这里先返回一个空 framework middleware 占位；
// - 真正从容器解析 tracer 的路径应由 provider 装配层统一承担，而不是继续暴露 Gin-only helper。
func TracingMiddlewareFromContainer(serviceName string) contract.HTTPMiddleware {
	_ = serviceName
	return func(c contract.HTTPContext, next contract.HTTPNext) {
		if next != nil {
			next()
		}
	}
}

// ParseTraceParent 解析 W3C traceparent 头。
//
// 中文说明：
// - 格式：version-traceid-spanid-flags；
// - 例如：00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01。
func ParseTraceParent(traceparent string) (traceID, spanID string, flags string, ok bool) {
	parts := strings.Split(traceparent, "-")
	if len(parts) != 4 {
		return "", "", "", false
	}
	if parts[0] != "00" {
		return "", "", "", false
	}
	return parts[1], parts[2], parts[3], true
}
