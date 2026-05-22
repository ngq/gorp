// Application scenarios:
// - Record HTTP access logs with stable request fields.
// - Reuse request-scoped logger context in downstream business code.
// - Provide a default request logging baseline for production services.
//
// 适用场景：
// - 记录带有稳定请求字段的 HTTP 访问日志。
// - 在下游业务代码中复用请求级 logger 上下文。
// - 为生产服务提供默认的请求日志基线。
package middleware

import (
	"time"

	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	supportcontract "github.com/ngq/gorp/framework/contract/support"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	frameworkbizlog "github.com/ngq/gorp/framework/log"
)

// LoggingMiddleware writes a request-scoped access log and stores the derived logger in context.
//
// LoggingMiddleware 输出请求级访问日志，并把派生后的 logger 写回上下文。
//
// Example:
//
//	router.Use(httpmiddleware.LoggingMiddleware(logger))
func LoggingMiddleware(base observabilitycontract.Logger) transportcontract.Middleware {
	return func(next transportcontract.Handler) transportcontract.Handler {
		return func(c transportcontract.Context) {
			logger := base
			fields := make([]observabilitycontract.Field, 0, 2)
			traceID := ""
			requestID := ""
			if tid, ok := supportcontract.FromTraceIDContext(c.Context()); ok && tid != "" {
				traceID = tid
				fields = append(fields, observabilitycontract.Field{Key: "trace_id", Value: tid})
			}
			if rid, ok := supportcontract.FromRequestIDContext(c.Context()); ok && rid != "" {
				requestID = rid
				fields = append(fields, observabilitycontract.Field{Key: "request_id", Value: rid})
			}
			if logger != nil && len(fields) > 0 {
				logger = logger.With(fields...)
			}
			c.Set("logger", logger)
			// Also update gin.Request.Context for context.Context value propagation
			if gc, ok := unwrapGinContext(c); ok && gc.Request != nil {
				gc.Request = gc.Request.WithContext(frameworkbizlog.WithContext(gc.Request.Context(), logger))
			}

			start := time.Now()
			if next != nil {
				next(c)
			}

			method := ""
			route := c.RoutePath()
			path := route
			status := c.ResponseStatus()
			if req := c.Request(); req != nil {
				method = req.Method
				if req.URL != nil {
					path = req.URL.Path
				}
				if route == "" && req.URL != nil {
					path = req.URL.Path
				}
			}
			if status == 0 {
				status = 200
			}
			logFields := []observabilitycontract.Field{
				{Key: "method", Value: method},
				{Key: "path", Value: path},
				{Key: "route", Value: route},
				{Key: "status", Value: status},
				{Key: "latency_ms", Value: time.Since(start).Milliseconds()},
			}
			if requestID != "" {
				logFields = append(logFields, observabilitycontract.Field{Key: "request_id", Value: requestID})
			}
			if traceID != "" {
				logFields = append(logFields, observabilitycontract.Field{Key: "trace_id", Value: traceID})
			}
			frameworkbizlog.Ctx(c.Context()).Info("http request", logFields...)
		}
	}
}
