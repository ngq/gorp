package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	supportcontract "github.com/ngq/gorp/framework/contract/support"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

func TracingMiddleware(tracer observabilitycontract.Tracer, serviceName string) transportcontract.HTTPMiddleware {
	return func(next transportcontract.HTTPHandler) transportcontract.HTTPHandler {
		return func(c transportcontract.HTTPContext) {
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
				WithSpanKind(observabilitycontract.SpanKindServer),
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
				ctx = supportcontract.NewTraceIDContext(ctx, traceID)
			}
			c.SetContext(ctx)

			startTime := time.Now()
			if next != nil {
				next(c)
			}

			latency := time.Since(startTime)
			statusCode := c.ResponseStatus()

			span.SetAttributes(map[string]any{
				"http.status_code": statusCode,
				"http.latency_ms":  latency.Milliseconds(),
			})

			if statusCode >= 400 {
				span.SetStatus(observabilitycontract.SpanStatusCodeError, http.StatusText(statusCode))
			} else {
				span.SetStatus(observabilitycontract.SpanStatusCodeOk, "")
			}

			traceID := span.SpanContext().TraceID
			if traceID != "" {
				c.Header("X-Trace-ID", traceID)
			}
		}
	}
}

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

func WithSpanKind(kind observabilitycontract.SpanKind) observabilitycontract.SpanOption {
	return func(cfg *observabilitycontract.SpanConfig) {
		cfg.Kind = kind
	}
}

func WithAttributes(attrs map[string]any) observabilitycontract.SpanOption {
	return func(cfg *observabilitycontract.SpanConfig) {
		if cfg.Attributes == nil {
			cfg.Attributes = make(map[string]any)
		}
		for k, v := range attrs {
			cfg.Attributes[k] = v
		}
	}
}

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
