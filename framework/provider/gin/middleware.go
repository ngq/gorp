package gin

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/gin-gonic/gin"
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
// - 设置到响应头和 gin.Context，供后续中间件和业务代码使用。
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader(requestIDHeader)
		if rid == "" {
			buf := make([]byte, 16)
			_, _ = rand.Read(buf)
			rid = hex.EncodeToString(buf)
		}
		c.Writer.Header().Set(requestIDHeader, rid)
		c.Set("request_id", rid)
		c.Next()
	}
}

// TraceID 为每个请求生成或透传唯一 Trace ID。
//
// 中文说明：
// - 优先从请求头读取 X-Trace-Id（支持分布式链路追踪场景）；
// - 如果不存在则生成一个 32 字符的随机 ID；
// - 设置到响应头和 gin.Context；
// - 与 Request ID 区分：Trace ID 用于跨服务追踪，Request ID 用于单次请求标识。
// - 如果没有单独设置 Trace ID，默认使用 Request ID 作为 Trace ID。
func TraceID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 尝试从多个可能的请求头读取 trace id
		tid := c.GetHeader(traceIDHeader)
		if tid == "" {
			tid = c.GetHeader(traceIDHeaderAlt)
		}

		// 如果没有 trace id，使用 request id 或生成新的
		if tid == "" {
			rid, exists := c.Get("request_id")
			if exists {
				tid = rid.(string)
			} else {
				buf := make([]byte, 16)
				_, _ = rand.Read(buf)
				tid = hex.EncodeToString(buf)
			}
		}

		c.Writer.Header().Set(traceIDHeader, tid)
		c.Set("trace_id", tid)
		c.Next()
	}
}

// GetTraceID 从 gin.Context 获取 Trace ID。
//
// 中文说明：
// - 供业务代码和日志中间件使用；
// - 如果不存在则返回空字符串。
func GetTraceID(c *gin.Context) string {
	if tid, exists := c.Get("trace_id"); exists {
		return tid.(string)
	}
	return ""
}

// GetRequestID 从 gin.Context 获取 Request ID。
//
// 中文说明：
// - 供业务代码和日志中间件使用；
// - 如果不存在则返回空字符串。
func GetRequestID(c *gin.Context) string {
	if rid, exists := c.Get("request_id"); exists {
		return rid.(string)
	}
	return ""
}
