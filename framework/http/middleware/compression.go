// Application scenarios:
// - Reduce response payload size for JSON, text, and similar content.
// - Improve bandwidth efficiency for external HTTP APIs.
// - Add a simple, transparent response compression layer for clients that support gzip.
//
// 适用场景：
// - 压缩 JSON、文本等响应内容，减少返回体积。
// - 提升对外 HTTP API 的带宽利用效率。
// - 为支持 gzip 的客户端提供简单透明的响应压缩层。
package middleware

import (
	"bufio"
	"compress/gzip"
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

// Compression enables gzip response compression for clients that support it.
//
// Compression 为支持 gzip 的客户端开启响应压缩。
func Compression() transportcontract.Middleware {
	return func(next transportcontract.Handler) transportcontract.Handler {
		return func(c transportcontract.Context) {
			gc, ok := unwrapGinContext(c)
			if !ok || gc == nil || !acceptsGzip(c) {
				if next != nil {
					next(c)
				}
				return
			}

			gzw := gzip.NewWriter(gc.Writer)
			defer gzw.Close()

			writer := &gzipResponseWriter{
				ResponseWriter: gc.Writer,
				Writer:         gzw,
			}
			gc.Writer = writer
			gc.Header("Content-Encoding", "gzip")
			gc.Header("Vary", appendVaryHeader(gc.Writer.Header().Get("Vary"), "Accept-Encoding"))
			gc.Header("Content-Length", "")

			if next != nil {
				next(c)
			}

			_ = gzw.Close()
			gc.Writer = writer.ResponseWriter
		}
	}
}

type gzipResponseWriter struct {
	gin.ResponseWriter
	Writer      *gzip.Writer
	wroteHeader bool
}

// WriteHeader writes the response status and clears Content-Length for gzip output.
//
// WriteHeader 写入响应状态码，并清除 gzip 输出下无效的 Content-Length。
func (w *gzipResponseWriter) WriteHeader(code int) {
	w.Header().Del("Content-Length")
	w.wroteHeader = true
	w.ResponseWriter.WriteHeader(code)
}

// Write compresses response bytes through the gzip writer.
//
// Write 通过 gzip writer 压缩响应字节。
func (w *gzipResponseWriter) Write(data []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.Writer.Write(data)
}

// WriteString compresses response text through the gzip writer.
//
// WriteString 通过 gzip writer 压缩响应文本。
func (w *gzipResponseWriter) WriteString(s string) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.Writer.Write([]byte(s))
}

// Hijack forwards hijack support when the underlying writer supports it.
//
// Hijack 在底层 writer 支持时透传 hijack 能力。
func (w *gzipResponseWriter) Hijack() (net.Conn, *bufioReadWriter, error) {
	type hijacker interface {
		Hijack() (net.Conn, *bufioReadWriter, error)
	}
	if h, ok := w.ResponseWriter.(hijacker); ok {
		return h.Hijack()
	}
	return nil, nil, http.ErrNotSupported
}

type bufioReadWriter = bufio.ReadWriter

// acceptsGzip reports whether the client accepts gzip encoding.
//
// acceptsGzip 判断客户端是否接受 gzip 编码。
func acceptsGzip(c transportcontract.Context) bool {
	if c == nil {
		return false
	}
	encoding := strings.ToLower(strings.TrimSpace(c.GetHeader("Accept-Encoding")))
	return strings.Contains(encoding, "gzip")
}

// appendVaryHeader appends a Vary token only when it is not already present.
//
// appendVaryHeader 仅在尚未存在时追加 Vary 项。
func appendVaryHeader(current string, value string) string {
	if current == "" {
		return value
	}
	items := strings.Split(current, ",")
	for _, item := range items {
		if strings.EqualFold(strings.TrimSpace(item), value) {
			return current
		}
	}
	return current + ", " + value
}
