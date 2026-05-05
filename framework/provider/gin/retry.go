package gin

import (
	"bytes"
	"bufio"
	"context"
	"io"
	"math/rand"
	"net/http"
	"net"
	"reflect"
	"time"
	"unsafe"

	"github.com/gin-gonic/gin"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
)

// RetryMiddleware is an advanced Gin-provider extension.
// It is intentionally not part of the default HTTP middleware mainline.
func RetryMiddleware(retry resiliencecontract.Retry, policy resiliencecontract.RetryPolicy) gin.HandlerFunc {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	return func(c *gin.Context) {
		method := c.Request.Method
		if !isIdempotentMethod(method) {
			c.Next()
			return
		}

		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		baseWriter := c.Writer
		startIndex := ginContextIndex(c)
		baseErrorsLen := len(c.Errors)

		var lastErr error
		var lastStatusCode int

		for attempt := 0; attempt < policy.MaxAttempts; attempt++ {
			writer := newRetryResponseWriter(baseWriter)
			if len(bodyBytes) > 0 {
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
			c.Errors = c.Errors[:baseErrorsLen]
			setGinContextIndex(c, startIndex)

			c.Writer = writer
			c.Next()

			statusCode := writer.Status()
			lastStatusCode = statusCode
			if statusCode < 500 {
				writer.FlushTo(baseWriter)
				return
			}
			if !isRetryableHTTPStatus(statusCode, policy.RetryableCodes) {
				writer.FlushTo(baseWriter)
				return
			}

			lastErr = resiliencecontract.NewError(statusCode, resiliencecontract.ErrorReasonInternal, "request failed")
			if attempt == policy.MaxAttempts-1 {
				break
			}

			select {
			case <-c.Request.Context().Done():
				c.Writer = baseWriter
				writeGinResponseHeaders(c)
				resp := Response{
					Code:    CodeServiceUnavailable,
					Message: "request timeout during retry",
					Data: map[string]any{
						"attempts": attempt + 1,
						"reason":   "TIMEOUT",
					},
				}
				c.AbortWithStatusJSON(http.StatusGatewayTimeout, resp)
				return
			default:
			}

			jitter := rng.Float64()
			delay := policy.CalculateDelay(attempt, jitter)
			select {
			case <-c.Request.Context().Done():
				c.Writer = baseWriter
				writeGinResponseHeaders(c)
				resp := Response{
					Code:    CodeServiceUnavailable,
					Message: "request timeout during retry",
					Data: map[string]any{
						"attempts": attempt + 1,
						"reason":   "TIMEOUT",
					},
				}
				c.AbortWithStatusJSON(http.StatusGatewayTimeout, resp)
				return
			case <-time.After(delay):
			}
		}

		c.Writer = baseWriter
		if lastErr != nil {
			writeGinResponseHeaders(c)
			resp := Response{
				Code:    CodeServiceUnavailable,
				Message: lastErr.Error(),
				Data: map[string]any{
					"attempts": policy.MaxAttempts,
					"reason":   "SERVICE_UNAVAILABLE",
				},
			}
			c.JSON(lastStatusCode, resp)
		}
	}
}

// RetryAllMethodsMiddleware is a provider-specific escape hatch.
// Use it only when the caller has already accepted the replay risk of non-idempotent requests.
func RetryAllMethodsMiddleware(retry resiliencecontract.Retry, policy resiliencecontract.RetryPolicy) gin.HandlerFunc {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	return func(c *gin.Context) {
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		baseWriter := c.Writer
		startIndex := ginContextIndex(c)
		baseErrorsLen := len(c.Errors)

		var lastErr error
		var lastStatusCode int

		for attempt := 0; attempt < policy.MaxAttempts; attempt++ {
			writer := newRetryResponseWriter(baseWriter)
			if len(bodyBytes) > 0 {
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
			c.Errors = c.Errors[:baseErrorsLen]
			setGinContextIndex(c, startIndex)

			c.Writer = writer
			c.Next()

			statusCode := writer.Status()
			lastStatusCode = statusCode
			if statusCode < 500 {
				writer.FlushTo(baseWriter)
				return
			}
			if !isRetryableHTTPStatus(statusCode, policy.RetryableCodes) {
				writer.FlushTo(baseWriter)
				return
			}

			lastErr = resiliencecontract.NewError(statusCode, resiliencecontract.ErrorReasonInternal, "request failed")
			if attempt == policy.MaxAttempts-1 {
				break
			}

			select {
			case <-c.Request.Context().Done():
				c.Writer = baseWriter
				writeGinResponseHeaders(c)
				resp := Response{
					Code:    CodeServiceUnavailable,
					Message: "request timeout during retry",
					Data: map[string]any{
						"attempts": attempt + 1,
						"reason":   "TIMEOUT",
					},
				}
				c.AbortWithStatusJSON(http.StatusGatewayTimeout, resp)
				return
			default:
			}

			jitter := rng.Float64()
			delay := policy.CalculateDelay(attempt, jitter)
			select {
			case <-c.Request.Context().Done():
				c.Writer = baseWriter
				writeGinResponseHeaders(c)
				resp := Response{
					Code:    CodeServiceUnavailable,
					Message: "request timeout during retry",
					Data: map[string]any{
						"attempts": attempt + 1,
						"reason":   "TIMEOUT",
					},
				}
				c.AbortWithStatusJSON(http.StatusGatewayTimeout, resp)
				return
			case <-time.After(delay):
			}
		}

		c.Writer = baseWriter
		if lastErr != nil {
			writeGinResponseHeaders(c)
			resp := Response{
				Code:    CodeServiceUnavailable,
				Message: lastErr.Error(),
				Data: map[string]any{
					"attempts": policy.MaxAttempts,
					"reason":   "SERVICE_UNAVAILABLE",
				},
			}
			c.JSON(lastStatusCode, resp)
		}
	}
}

func isIdempotentMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
		return true
	default:
		return false
	}
}

func isRetryableHTTPStatus(status int, retryableCodes []int) bool {
	for _, code := range retryableCodes {
		if status == code {
			return true
		}
	}
	return status == 502 || status == 503 || status == 504
}

type retryResponseWriter struct {
	gin.ResponseWriter
	header     http.Header
	body       bytes.Buffer
	statusCode int
	size       int
}

func newRetryResponseWriter(base gin.ResponseWriter) *retryResponseWriter {
	return &retryResponseWriter{
		ResponseWriter: base,
		header:         make(http.Header),
		statusCode:     http.StatusOK,
		size:           -1,
	}
}

func (w *retryResponseWriter) Header() http.Header {
	return w.header
}

func (w *retryResponseWriter) Write(data []byte) (int, error) {
	w.WriteHeaderNow()
	n, err := w.body.Write(data)
	w.size += n
	return n, err
}

func (w *retryResponseWriter) WriteString(s string) (int, error) {
	w.WriteHeaderNow()
	n, err := w.body.WriteString(s)
	w.size += n
	return n, err
}

func (w *retryResponseWriter) WriteHeader(code int) {
	if code > 0 && w.statusCode != code {
		w.statusCode = code
	}
}

func (w *retryResponseWriter) Status() int {
	return w.statusCode
}

func (w *retryResponseWriter) Size() int {
	return w.size
}

func (w *retryResponseWriter) Written() bool {
	return w.size != -1
}

func (w *retryResponseWriter) WriteHeaderNow() {
	if !w.Written() {
		w.size = 0
	}
}

func (w *retryResponseWriter) Flush() {}

func (w *retryResponseWriter) FlushTo(dst gin.ResponseWriter) {
	if dst == nil {
		return
	}
	copyHeaders(dst.Header(), w.header)
	dst.WriteHeader(w.statusCode)
	if w.body.Len() > 0 {
		_, _ = dst.Write(w.body.Bytes())
	}
}

func (w *retryResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := w.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, http.ErrNotSupported
}

func (w *retryResponseWriter) CloseNotify() <-chan bool {
	return w.ResponseWriter.CloseNotify()
}

func (w *retryResponseWriter) Pusher() http.Pusher {
	return w.ResponseWriter.Pusher()
}

func DoWithRetry(c *gin.Context, retry resiliencecontract.Retry, fn func(ctx context.Context) (any, error)) (any, error) {
	return retry.DoWithResult(c.Request.Context(), func() (any, error) {
		return fn(c.Request.Context())
	})
}

func copyHeaders(dst, src http.Header) {
	for k := range dst {
		dst.Del(k)
	}
	for k, values := range src {
		for _, value := range values {
			dst.Add(k, value)
		}
	}
}

func ginContextIndex(c *gin.Context) int8 {
	if c == nil {
		return -1
	}
	value := reflect.ValueOf(c).Elem().FieldByName("index")
	return *(*int8)(unsafe.Pointer(value.UnsafeAddr()))
}

func setGinContextIndex(c *gin.Context, index int8) {
	if c == nil {
		return
	}
	value := reflect.ValueOf(c).Elem().FieldByName("index")
	*(*int8)(unsafe.Pointer(value.UnsafeAddr())) = index
}
