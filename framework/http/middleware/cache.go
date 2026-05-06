// Application scenarios:
// - Add cache directives to read-heavy HTTP APIs.
// - Avoid re-sending unchanged payloads when the client already has the current entity.
// - Build a lightweight conditional-request layer for REST-style endpoints.
//
// 适用场景：
// - 为读多写少的 HTTP API 添加缓存指令。
// - 当客户端已持有最新实体时避免重复发送未变更内容。
// - 为 REST 风格接口建立轻量级条件请求层。
package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

// CacheControl writes a fixed Cache-Control header for the current response.
//
// CacheControl 为当前响应写入固定的 Cache-Control 头。
func CacheControl(value string) transportcontract.HTTPMiddleware {
	value = strings.TrimSpace(value)
	return func(next transportcontract.HTTPHandler) transportcontract.HTTPHandler {
		return func(c transportcontract.HTTPContext) {
			if c != nil && value != "" {
				c.Header("Cache-Control", value)
			}
			if next != nil {
				next(c)
			}
		}
	}
}

// ETag computes a weak ETag for successful GET/HEAD responses and handles If-None-Match.
//
// ETag 为成功的 GET/HEAD 响应计算弱 ETag，并处理 If-None-Match。
func ETag() transportcontract.HTTPMiddleware {
	return func(next transportcontract.HTTPHandler) transportcontract.HTTPHandler {
		return func(c transportcontract.HTTPContext) {
			if next == nil {
				return
			}

			req := c.Request()
			if req == nil || (req.Method != http.MethodGet && req.Method != http.MethodHead) {
				next(c)
				return
			}

			gc, ok := unwrapGinContext(c)
			if !ok || gc == nil {
				next(c)
				return
			}

			recorder := newETagResponseWriter(gc.Writer)
			gc.Writer = recorder
			next(c)
			gc.Writer = recorder.ResponseWriter

			status := recorder.Status()
			if status == 0 {
				status = http.StatusOK
			}
			if status < 200 || status >= 300 {
				return
			}
			if len(recorder.body) == 0 {
				return
			}

			etag := weakETag(recorder.body)
			if etag == "" {
				recorder.FlushTo(gc.Writer)
				return
			}

			if matchesIfNoneMatch(req.Header.Get("If-None-Match"), etag) {
				clearEntityHeaders(gc.Writer.Header())
				gc.Header("ETag", etag)
				gc.Status(http.StatusNotModified)
				return
			}

			recorder.header.Set("ETag", etag)
			recorder.FlushTo(gc.Writer)
		}
	}
}

type etagResponseWriter struct {
	gin.ResponseWriter
	header     http.Header
	body       []byte
	statusCode int
	written    bool
	size       int
}

// newETagResponseWriter creates a response recorder used for ETag calculation.
//
// newETagResponseWriter 创建一个用于 ETag 计算的响应记录器。
func newETagResponseWriter(base gin.ResponseWriter) *etagResponseWriter {
	return &etagResponseWriter{
		ResponseWriter: base,
		header:         make(http.Header),
		statusCode:     http.StatusOK,
		size:           -1,
	}
}

// Header returns the staged response headers.
//
// Header 返回暂存中的响应头。
func (w *etagResponseWriter) Header() http.Header {
	return w.header
}

// WriteHeader records the response status code.
//
// WriteHeader 记录响应状态码。
func (w *etagResponseWriter) WriteHeader(code int) {
	if code > 0 {
		w.statusCode = code
	}
}

// WriteHeaderNow marks the staged response as written.
//
// WriteHeaderNow 标记暂存响应已经开始写出。
func (w *etagResponseWriter) WriteHeaderNow() {
	if !w.written {
		w.written = true
		w.size = 0
	}
}

// Write buffers response bytes for later ETag processing.
//
// Write 缓存响应字节，供后续 ETag 处理使用。
func (w *etagResponseWriter) Write(data []byte) (int, error) {
	w.WriteHeaderNow()
	w.body = append(w.body, data...)
	w.size += len(data)
	return len(data), nil
}

// WriteString buffers response text for later ETag processing.
//
// WriteString 缓存响应文本，供后续 ETag 处理使用。
func (w *etagResponseWriter) WriteString(s string) (int, error) {
	return w.Write([]byte(s))
}

// Status returns the staged response status code.
//
// Status 返回暂存响应状态码。
func (w *etagResponseWriter) Status() int {
	return w.statusCode
}

// Size returns the staged response size.
//
// Size 返回暂存响应大小。
func (w *etagResponseWriter) Size() int {
	return w.size
}

// Written reports whether the staged response has started writing.
//
// Written 返回暂存响应是否已经开始写出。
func (w *etagResponseWriter) Written() bool {
	return w.written
}

// Flush is a no-op because the response is fully buffered.
//
// Flush 在这里为空操作，因为响应会先被完整缓冲。
func (w *etagResponseWriter) Flush() {}

// FlushTo writes the staged response into the destination writer.
//
// FlushTo 将暂存响应写入目标 writer。
func (w *etagResponseWriter) FlushTo(dst gin.ResponseWriter) {
	if dst == nil {
		return
	}
	copyHeaders(dst.Header(), w.header)
	dst.WriteHeader(w.statusCode)
	if len(w.body) > 0 {
		_, _ = dst.Write(w.body)
	}
}

// weakETag builds a weak ETag value from response bytes.
//
// weakETag 基于响应字节构造弱 ETag。
func weakETag(body []byte) string {
	sum := sha256.Sum256(body)
	return `W/"` + hex.EncodeToString(sum[:]) + `"`
}

// matchesIfNoneMatch checks whether the request header matches the generated ETag.
//
// matchesIfNoneMatch 判断请求头是否命中生成出的 ETag。
func matchesIfNoneMatch(header string, etag string) bool {
	header = strings.TrimSpace(header)
	if header == "" || etag == "" {
		return false
	}
	if header == "*" {
		return true
	}
	for _, item := range strings.Split(header, ",") {
		if strings.TrimSpace(item) == etag {
			return true
		}
	}
	return false
}

// clearEntityHeaders removes entity headers before returning 304.
//
// clearEntityHeaders 在返回 304 前清理实体相关响应头。
func clearEntityHeaders(header http.Header) {
	header.Del("Content-Length")
	header.Del("Content-Type")
	header.Del("Content-Encoding")
}
