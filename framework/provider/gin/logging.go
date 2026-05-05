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
			c.SetContext(frameworkbizlog.WithContext(c.Context(), logger))

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
				observabilitycontract.Field{Key: "method", Value: method},
				observabilitycontract.Field{Key: "path", Value: path},
				observabilitycontract.Field{Key: "route", Value: route},
				observabilitycontract.Field{Key: "status", Value: status},
				observabilitycontract.Field{Key: "latency_ms", Value: time.Since(start).Milliseconds()},
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
