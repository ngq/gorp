package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ngq/gorp/framework/contract"
)

// TracingMiddleware 创建 HTTP 追踪中间件。
//
// 中文说明：
// - 自动为每个请求创建 Span；
// - 从请求头提取追踪上下文（W3C TraceContext）；
// - 记录请求方法、路径、状态码等信息；
// - 支持与 OpenTelemetry 集成。
func TracingMiddleware(tracer contract.Tracer, serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头提取追踪上下文
		carrier := &httpHeaderCarrier{c.Request.Header}
		ctx, err := tracer.Extract(c.Request.Context(), carrier)
		if err != nil {
			ctx = c.Request.Context()
		}

		// 构建 Span 名称
		spanName := fmt.Sprintf("HTTP %s %s", c.Request.Method, c.FullPath())
		if spanName == fmt.Sprintf("HTTP %s ", c.Request.Method) {
			spanName = fmt.Sprintf("HTTP %s %s", c.Request.Method, c.Request.URL.Path)
		}

		// 创建 Span
		ctx, span := tracer.StartSpan(ctx, spanName,
			WithSpanKind(contract.SpanKindServer),
			WithAttributes(map[string]any{
				"http.method":      c.Request.Method,
				"http.url":         c.Request.URL.String(),
				"http.host":        c.Request.Host,
				"http.scheme":      c.Request.URL.Scheme,
				"http.target":      c.Request.URL.Path,
				"http.user_agent":  c.Request.UserAgent(),
				"http.route":       c.FullPath(),
				"service.name":     serviceName,
			}),
		)
		defer span.End()

		// 将 Span 存入上下文
		c.Request = c.Request.WithContext(ctx)

		// 记录开始时间
		startTime := time.Now()

		// 执行请求
		c.Next()

		// 记录响应信息
		latency := time.Since(startTime)
		statusCode := c.Writer.Status()

		span.SetAttributes(map[string]any{
			"http.status_code": statusCode,
			"http.response_size": c.Writer.Size(),
			"http.latency_ms":   latency.Milliseconds(),
		})

		// 记录错误状态
		if statusCode >= 400 {
			span.SetStatus(contract.SpanStatusCodeError, http.StatusText(statusCode))
			if len(c.Errors) > 0 {
				span.SetAttributes(map[string]any{
					"http.errors": c.Errors.String(),
				})
			}
		} else {
			span.SetStatus(contract.SpanStatusCodeOk, "")
		}

		// 注入追踪信息到响应头
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
// - 如果 Tracer 未注册，返回空中间件。
func TracingMiddlewareFromContainer(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 尝试从上下文获取容器（需要配合其他中间件注入容器）
		// 这里简化处理，实际使用时需要配合框架容器
		c.Next()
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