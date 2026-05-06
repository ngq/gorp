// Application scenarios:
// - Capture request and response payload samples for debugging, auditing, and risk-control workflows.
// - Inspect HTTP exchanges without forcing all business code to implement custom logging.
// - Keep request-response capture explicit and bounded to reduce privacy and performance risk.
//
// 适用场景：
// - 为调试、审计和风控流程捕获请求与响应样本。
// - 在不要求业务代码自行实现自定义日志的前提下观测 HTTP 交换内容。
// - 通过显式且有边界的方式进行请求响应捕获，降低隐私和性能风险。
package middleware

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	supportcontract "github.com/ngq/gorp/framework/contract/support"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

// HTTPExchangeCapture describes a captured HTTP request-response exchange.
//
// HTTPExchangeCapture 描述一次被捕获的 HTTP 请求响应交换。
type HTTPExchangeCapture struct {
	Method            string
	Path              string
	Route             string
	Status            int
	RequestID         string
	TraceID           string
	Tenant            string
	RequestHeaders    map[string]string
	ResponseHeaders   map[string]string
	RequestBody       []byte
	ResponseBody      []byte
	RequestTruncated  bool
	ResponseTruncated bool
}

// BodyDumpOptions controls request-response capture behavior.
//
// BodyDumpOptions 控制请求响应捕获行为。
type BodyDumpOptions struct {
	CaptureRequestHeaders  bool
	CaptureResponseHeaders bool
	CaptureRequestBody     bool
	CaptureResponseBody    bool
	MaxBodyBytes           int
	Skip                   func(transportcontract.HTTPContext) bool
	OnCapture              func(transportcontract.HTTPContext, *HTTPExchangeCapture)
}

// BodyDump captures request-response exchange details and forwards them to the configured callback.
//
// BodyDump 捕获请求响应交换细节，并转发给配置的回调。
//
// Example:
//
//	router.Use(httpmiddleware.BodyDump(httpmiddleware.BodyDumpOptions{
//	    CaptureRequestBody:  true,
//	    CaptureResponseBody: true,
//	    MaxBodyBytes:        4096,
//	    OnCapture: func(ctx gorp.HTTPContext, dump *httpmiddleware.HTTPExchangeCapture) {
//	        // consume captured exchange
//	    },
//	}))
func BodyDump(opts BodyDumpOptions) transportcontract.HTTPMiddleware {
	opts = normalizeBodyDumpOptions(opts)

	return func(next transportcontract.HTTPHandler) transportcontract.HTTPHandler {
		return func(c transportcontract.HTTPContext) {
			if c == nil {
				if next != nil {
					next(c)
				}
				return
			}
			if opts.Skip != nil && opts.Skip(c) {
				if next != nil {
					next(c)
				}
				return
			}

			dump := &HTTPExchangeCapture{
				Method:          requestMethod(c),
				Path:            requestPath(c),
				Route:           c.RoutePath(),
				RequestHeaders:  map[string]string{},
				ResponseHeaders: map[string]string{},
			}
			if requestID, ok := supportcontract.FromRequestIDContext(c.Context()); ok {
				dump.RequestID = requestID
			}
			if traceID, ok := supportcontract.FromTraceIDContext(c.Context()); ok {
				dump.TraceID = traceID
			}
			if tenant, ok := supportcontract.FromTenantContext(c.Context()); ok {
				dump.Tenant = tenant
			}
			if opts.CaptureRequestHeaders {
				dump.RequestHeaders = flattenHeaders(requestHeaders(c))
			}

			var requestCapture *captureReadCloser
			if req := c.Request(); req != nil && req.Body != nil && opts.CaptureRequestBody {
				requestCapture = newCaptureReadCloser(req.Body, opts.MaxBodyBytes)
				req.Body = requestCapture
				c.SetRequest(req)
			}

			var responseCapture *captureResponseWriter
			if gc, ok := unwrapGinContext(c); ok && gc != nil && (opts.CaptureResponseBody || opts.CaptureResponseHeaders) {
				responseCapture = newCaptureResponseWriter(gc.Writer, opts.CaptureResponseBody, opts.MaxBodyBytes)
				gc.Writer = responseCapture
				defer func() {
					gc.Writer = responseCapture.ResponseWriter
				}()
			}

			if next != nil {
				next(c)
			}

			dump.Status = c.ResponseStatus()
			if dump.Status == 0 {
				dump.Status = http.StatusOK
			}
			if requestCapture != nil {
				dump.RequestBody = requestCapture.Bytes()
				dump.RequestTruncated = requestCapture.Truncated()
			}
			if responseCapture != nil {
				if opts.CaptureResponseBody {
					dump.ResponseBody = responseCapture.Bytes()
					dump.ResponseTruncated = responseCapture.Truncated()
				}
				if opts.CaptureResponseHeaders {
					dump.ResponseHeaders = flattenHeaders(responseCapture.Header())
				}
			}
			if opts.OnCapture != nil {
				opts.OnCapture(c, dump)
			}
		}
	}
}

type captureReadCloser struct {
	io.ReadCloser
	maxBytes  int
	buf       bytes.Buffer
	truncated bool
}

func newCaptureReadCloser(rc io.ReadCloser, maxBytes int) *captureReadCloser {
	return &captureReadCloser{ReadCloser: rc, maxBytes: maxBytes}
}

func (r *captureReadCloser) Read(p []byte) (int, error) {
	n, err := r.ReadCloser.Read(p)
	if n > 0 {
		r.capture(p[:n])
	}
	return n, err
}

func (r *captureReadCloser) capture(data []byte) {
	if r.maxBytes <= 0 {
		return
	}
	remaining := r.maxBytes - r.buf.Len()
	if remaining <= 0 {
		r.truncated = true
		return
	}
	if len(data) > remaining {
		r.truncated = true
		data = data[:remaining]
	}
	_, _ = r.buf.Write(data)
}

func (r *captureReadCloser) Bytes() []byte {
	return append([]byte(nil), r.buf.Bytes()...)
}

func (r *captureReadCloser) Truncated() bool {
	return r.truncated
}

type captureResponseWriter struct {
	gin.ResponseWriter
	captureBody bool
	maxBytes    int
	buf         bytes.Buffer
	truncated   bool
}

func newCaptureResponseWriter(base gin.ResponseWriter, captureBody bool, maxBytes int) *captureResponseWriter {
	return &captureResponseWriter{
		ResponseWriter: base,
		captureBody:    captureBody,
		maxBytes:       maxBytes,
	}
}

func (w *captureResponseWriter) Write(data []byte) (int, error) {
	if w.captureBody {
		w.capture(data)
	}
	return w.ResponseWriter.Write(data)
}

func (w *captureResponseWriter) WriteString(s string) (int, error) {
	if w.captureBody {
		w.capture([]byte(s))
	}
	return w.ResponseWriter.WriteString(s)
}

func (w *captureResponseWriter) capture(data []byte) {
	if w.maxBytes <= 0 {
		return
	}
	remaining := w.maxBytes - w.buf.Len()
	if remaining <= 0 {
		w.truncated = true
		return
	}
	if len(data) > remaining {
		w.truncated = true
		data = data[:remaining]
	}
	_, _ = w.buf.Write(data)
}

func (w *captureResponseWriter) Bytes() []byte {
	return append([]byte(nil), w.buf.Bytes()...)
}

func (w *captureResponseWriter) Truncated() bool {
	return w.truncated
}

// normalizeBodyDumpOptions fills missing body dump options with safe defaults.
//
// normalizeBodyDumpOptions 用安全默认值补齐 body dump 选项。
func normalizeBodyDumpOptions(opts BodyDumpOptions) BodyDumpOptions {
	if opts.MaxBodyBytes <= 0 {
		opts.MaxBodyBytes = 4 << 10
	}
	return opts
}

// requestMethod returns the current HTTP method.
//
// requestMethod 返回当前 HTTP 方法。
func requestMethod(c transportcontract.HTTPContext) string {
	if c == nil || c.Request() == nil {
		return ""
	}
	return c.Request().Method
}

// requestPath returns the current URL path or route path fallback.
//
// requestPath 返回当前 URL 路径，必要时回退到 route path。
func requestPath(c transportcontract.HTTPContext) string {
	if c != nil && c.Request() != nil && c.Request().URL != nil {
		return c.Request().URL.Path
	}
	if c != nil {
		return c.RoutePath()
	}
	return ""
}

// requestHeaders returns the current request header set.
//
// requestHeaders 返回当前请求头集合。
func requestHeaders(c transportcontract.HTTPContext) http.Header {
	if c == nil || c.Request() == nil {
		return http.Header{}
	}
	return c.Request().Header
}

// flattenHeaders converts a header map into a single-value string map.
//
// flattenHeaders 将头部集合转换为单值字符串 map。
func flattenHeaders(header http.Header) map[string]string {
	if len(header) == 0 {
		return map[string]string{}
	}
	result := make(map[string]string, len(header))
	for key, values := range header {
		if len(values) == 0 {
			continue
		}
		result[key] = strings.Join(values, ",")
	}
	return result
}
