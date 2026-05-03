package gin

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/gin-gonic/gin"
	"github.com/ngq/gorp/framework/contract"
)

const (
	// requestIDHeader 是 Request ID 请求头名称
	requestIDHeader = "X-Request-Id"
	// traceIDHeader 是 Trace ID 请求头名称（兼容多种命名）
	traceIDHeader = "X-Trace-Id"
	// traceIDHeaderAlt 是 Trace ID 备选请求头名称
	traceIDHeaderAlt = "X-Trace-Id"
)

// RequestID 为每个请求生成或透传唯一 Request ID。
//
// 中文说明：
// - 优先从请求头读取 X-Request-Id；
// - 如果不存在则生成一个 32 字符的随机 ID；
// - 设置到响应头和 request context，供后续中间件和业务代码使用。
func RequestID() contract.HTTPMiddleware {
	return func(c contract.HTTPContext, next contract.HTTPNext) {
		rid := c.GetHeader(requestIDHeader)
		if rid == "" {
			buf := make([]byte, 16)
			_, _ = rand.Read(buf)
			rid = hex.EncodeToString(buf)
		}
		c.Header(requestIDHeader, rid)
		ctx := contract.NewRequestIDContext(c.Context(), rid)
		c.SetContext(ctx)
		if next != nil {
			next()
		}
	}
}

// TraceID 为每个请求生成或透传唯一 Trace ID。
//
// 中文说明：
// - 优先从请求头读取 X-Trace-Id（支持分布式链路追踪场景）；
// - 如果不存在则优先复用 request id；
// - 设置到响应头和 request context；
// - 与 Request ID 区分：Trace ID 用于跨服务追踪，Request ID 用于单次请求标识。
func TraceID() contract.HTTPMiddleware {
	return func(c contract.HTTPContext, next contract.HTTPNext) {
		tid := c.GetHeader(traceIDHeader)
		if tid == "" {
			tid = c.GetHeader(traceIDHeaderAlt)
		}
		if tid == "" {
			if rid, ok := contract.FromRequestIDContext(c.Context()); ok && rid != "" {
				tid = rid
			} else {
				buf := make([]byte, 16)
				_, _ = rand.Read(buf)
				tid = hex.EncodeToString(buf)
			}
		}
		c.Header(traceIDHeader, tid)
		ctx := contract.NewTraceIDContext(c.Context(), tid)
		c.SetContext(ctx)
		if next != nil {
			next()
		}
	}
}

// GetTraceID 从 gin.Context 获取 Trace ID。
//
// 中文说明：
// - 这是 provider 内部兼容读取入口；
// - 默认业务主线应优先通过 contract.FromTraceIDContext 读取。
func GetTraceID(c *gin.Context) string {
	if c != nil && c.Request != nil {
		if tid, ok := contract.FromTraceIDContext(c.Request.Context()); ok {
			return tid
		}
	}
	return ""
}

// GetRequestID 从 gin.Context 获取 Request ID。
//
// 中文说明：
// - 这是 provider 内部兼容读取入口；
// - 默认业务主线应优先通过 contract.FromRequestIDContext 读取。
func GetRequestID(c *gin.Context) string {
	if c != nil && c.Request != nil {
		if rid, ok := contract.FromRequestIDContext(c.Request.Context()); ok {
			return rid
		}
	}
	return ""
}
