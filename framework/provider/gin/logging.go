package gin

import (
	"time"

	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	supportcontract "github.com/ngq/gorp/framework/contract/support"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	frameworkbizlog "github.com/ngq/gorp/framework/log"
)

func LoggingMiddleware(base observabilitycontract.Logger) transportcontract.HTTPMiddleware {
	return func(next transportcontract.HTTPHandler) transportcontract.HTTPHandler {
		return func(c transportcontract.HTTPContext) {
			logger := base
			fields := make([]observabilitycontract.Field, 0, 2)
			if traceID, ok := supportcontract.FromTraceIDContext(c.Context()); ok && traceID != "" {
				fields = append(fields, observabilitycontract.Field{Key: "trace_id", Value: traceID})
			}
			if requestID, ok := supportcontract.FromRequestIDContext(c.Context()); ok && requestID != "" {
				fields = append(fields, observabilitycontract.Field{Key: "request_id", Value: requestID})
			}
			if logger != nil && len(fields) > 0 {
				logger = logger.With(fields...)
			}
			c.SetContext(frameworkbizlog.WithContext(c.Context(), logger))

			start := time.Now()
			if next != nil {
				next(c)
			}

			method := ""
			path := c.RoutePath()
			status := c.ResponseStatus()
			if req := c.Request(); req != nil {
				method = req.Method
				if path == "" && req.URL != nil {
					path = req.URL.Path
				}
			}
			if status == 0 {
				status = 200
			}
			frameworkbizlog.Ctx(c.Context()).Info("http request",
				observabilitycontract.Field{Key: "method", Value: method},
				observabilitycontract.Field{Key: "path", Value: path},
				observabilitycontract.Field{Key: "status", Value: status},
				observabilitycontract.Field{Key: "latency_ms", Value: time.Since(start).Milliseconds()},
			)
		}
	}
}
