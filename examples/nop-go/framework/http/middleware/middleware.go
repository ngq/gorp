// Application scenarios:
// - Add unified request identity to every HTTP request.
// - Expose request_id and trace_id to downstream handlers, logs, and responders.
// - Provide a stable request identity source for idempotency, audit, and tracing correlation.
//
// 适用场景：
// - 为每一个 HTTP 请求补齐统一请求标识。
// - 向下游处理器、日志和响应输出暴露 request_id 与 trace_id。
// - 为幂等、审计和链路追踪关联提供稳定的请求身份来源。
package middleware

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/gin-gonic/gin"
	supportcontract "github.com/ngq/gorp/framework/contract/support"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

const (
	requestIDHeader  = "X-Request-Id"
	traceIDHeader    = "X-Trace-Id"
	traceIDHeaderAlt = "X-Trace-Id"
)

// RequestIdentity injects request_id and trace_id into the current HTTP request context.
//
// RequestIdentity 将 request_id 与 trace_id 注入当前 HTTP 请求上下文。
//
// Example:
//
//	router.Use(httpmiddleware.RequestIdentity())
func RequestIdentity() transportcontract.Middleware {
	return func(next transportcontract.Handler) transportcontract.Handler {
		return func(c transportcontract.Context) {
			rid := resolveRequestID(c)
			c.SetHeader(requestIDHeader, rid)
			c.Set("request_id", rid)

			tid := resolveTraceID(c, rid)
			c.SetHeader(traceIDHeader, tid)
			c.Set("trace_id", tid)

			// Also update gin.Request.Context for context.Context value propagation
			if gc, ok := unwrapGinContext(c); ok && gc.Request != nil {
				ctx := gc.Request.Context()
				ctx = supportcontract.NewRequestIDContext(ctx, rid)
				ctx = supportcontract.NewTraceIDContext(ctx, tid)
				gc.Request = gc.Request.WithContext(ctx)
			}

			if next != nil {
				next(c)
			}
		}
	}
}

// GetTraceID reads the trace id from a Gin request context.
//
// GetTraceID 从 Gin 请求上下文中读取 trace id。
func GetTraceID(c *gin.Context) string {
	if c != nil && c.Request != nil {
		if tid, ok := supportcontract.FromTraceIDContext(c.Request.Context()); ok {
			return tid
		}
	}
	return ""
}

// GetRequestID reads the request id from a Gin request context.
//
// GetRequestID 从 Gin 请求上下文中读取 request id。
func GetRequestID(c *gin.Context) string {
	if c != nil && c.Request != nil {
		if rid, ok := supportcontract.FromRequestIDContext(c.Request.Context()); ok {
			return rid
		}
	}
	return ""
}

// resolveRequestID reads or generates a request id for the current request.
//
// resolveRequestID 为当前请求读取或生成 request id。
func resolveRequestID(c transportcontract.Context) string {
	if c == nil {
		return generateIdentityID()
	}
	rid := c.GetHeader(requestIDHeader)
	if rid == "" {
		rid = generateIdentityID()
	}
	return rid
}

// resolveTraceID reads or derives a trace id for the current request.
//
// resolveTraceID 为当前请求读取或派生 trace id。
func resolveTraceID(c transportcontract.Context, rid string) string {
	if c != nil {
		tid := c.GetHeader(traceIDHeader)
		if tid == "" {
			tid = c.GetHeader(traceIDHeaderAlt)
		}
		if tid != "" {
			return tid
		}
	}

	if rid != "" {
		return rid
	}

	return generateIdentityID()
}

// generateIdentityID creates a random identity string for request and trace correlation.
//
// generateIdentityID 生成用于请求与链路关联的随机身份标识。
func generateIdentityID() string {
	buf := make([]byte, 16)
	_, _ = rand.Read(buf)
	return hex.EncodeToString(buf)
}
