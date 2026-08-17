// Application scenarios:
// - Stop long-running HTTP requests from occupying request resources forever.
// - Return a unified timeout response to upstream callers.
// - Provide a request deadline source for downstream business code.
//
// 适用场景：
// - 阻止长时间运行的 HTTP 请求长期占用请求资源。
// - 为上游调用方返回统一的超时响应。
// - 为下游业务代码提供请求 deadline 来源。
package middleware

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"net"
	"net/http"
	"reflect"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

// Timeout enforces a request deadline on the transport-level HTTP middleware chain.
//
// Timeout 在 transport 层 HTTP 中间件链上施加请求超时约束。
func Timeout(timeout time.Duration) transportcontract.Middleware {
	return func(next transportcontract.Handler) transportcontract.Handler {
		return func(c transportcontract.Context) {
			if timeout <= 0 {
				if next != nil {
					next(c)
				}
				return
			}

			ctx, cancel := context.WithTimeout(c.Context(), timeout)
			defer cancel()
			if req := c.Request(); req != nil {
				*req = *req.WithContext(ctx)
			}

			gc, ok := unwrapGinContext(c)
			if !ok {
				// A provider without response-writer isolation must execute
				// synchronously to avoid concurrent writes.
				if next != nil {
					next(c)
				}
				return
			}
			base := gc.Writer
			buffered := newTimeoutResponseWriter(base)
			gc.Writer = buffered

			done := make(chan struct{})
			go func() {
				defer close(done)
				if next != nil {
					next(c)
				}
			}()

			select {
			case <-done:
				buffered.commit(base)
			case <-ctx.Done():
				buffered.timeout()
				writeTimeoutResponse(base)
			}
		}
	}
}

// TimeoutMiddleware is the native Gin timeout middleware form.
//
// TimeoutMiddleware 是超时中间件的原生 Gin 形态。
func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		if timeout <= 0 {
			c.Next()
			return
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)

		base := c.Writer
		buffered := newTimeoutResponseWriter(base)
		worker := cloneGinContextForContinuation(c)
		worker.Writer = buffered
		c.Abort()
		done := make(chan struct{})
		go func() {
			worker.Next()
			close(done)
		}()

		select {
		case <-done:
			buffered.commit(base)
		case <-ctx.Done():
			buffered.timeout()
			writeTimeoutResponse(base)
			c.Abort()
		}
	}
}

// TimeoutMiddlewareWithHandler applies a timeout and delegates timeout output to a custom callback.
//
// TimeoutMiddlewareWithHandler 应用超时控制，并把超时后的输出交给自定义回调处理。
func TimeoutMiddlewareWithHandler(timeout time.Duration, onTimeout func(*gin.Context)) gin.HandlerFunc {
	return func(c *gin.Context) {
		if timeout <= 0 {
			c.Next()
			return
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)

		base := c.Writer
		buffered := newTimeoutResponseWriter(base)
		worker := cloneGinContextForContinuation(c)
		worker.Writer = buffered
		c.Abort()
		done := make(chan struct{})
		go func() {
			worker.Next()
			close(done)
		}()

		select {
		case <-done:
			buffered.commit(base)
		case <-ctx.Done():
			buffered.timeout()
			if onTimeout != nil {
				timeoutContext := c.Copy()
				timeoutContext.Writer = base
				onTimeout(timeoutContext)
			} else {
				writeTimeoutResponse(base)
			}
			c.Abort()
		}
	}
}

// cloneGinContextForContinuation preserves Gin's private middleware cursor;
// gin.Context.Copy cannot be used because it intentionally drops handlers.
func cloneGinContextForContinuation(ctx *gin.Context) *gin.Context {
	worker := reflect.New(reflect.TypeOf(ctx).Elem()).Interface().(*gin.Context)
	reflect.ValueOf(worker).Elem().Set(reflect.ValueOf(ctx).Elem())
	return worker
}

type timeoutResponseWriter struct {
	gin.ResponseWriter
	mu       sync.Mutex
	header   http.Header
	body     bytes.Buffer
	status   int
	timedOut bool
}

func newTimeoutResponseWriter(base gin.ResponseWriter) *timeoutResponseWriter {
	header := make(http.Header, len(base.Header()))
	for key, values := range base.Header() {
		header[key] = append([]string(nil), values...)
	}
	return &timeoutResponseWriter{ResponseWriter: base, header: header, status: http.StatusOK}
}

func (w *timeoutResponseWriter) Header() http.Header { return w.header }
func (w *timeoutResponseWriter) WriteHeader(code int) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if !w.timedOut && w.body.Len() == 0 {
		w.status = code
	}
}
func (w *timeoutResponseWriter) WriteHeaderNow() {}
func (w *timeoutResponseWriter) Write(data []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.timedOut {
		return len(data), nil
	}
	return w.body.Write(data)
}
func (w *timeoutResponseWriter) WriteString(value string) (int, error) {
	return w.Write([]byte(value))
}
func (w *timeoutResponseWriter) Status() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.status
}
func (w *timeoutResponseWriter) Size() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.body.Len()
}
func (w *timeoutResponseWriter) Written() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.body.Len() > 0 || w.status != http.StatusOK
}
func (w *timeoutResponseWriter) Flush()                   {}
func (w *timeoutResponseWriter) Pusher() http.Pusher      { return nil }
func (w *timeoutResponseWriter) CloseNotify() <-chan bool { return w.ResponseWriter.CloseNotify() }
func (w *timeoutResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, errors.New("timeout middleware does not support hijacking")
}
func (w *timeoutResponseWriter) timeout() {
	w.mu.Lock()
	w.timedOut = true
	w.status = http.StatusGatewayTimeout
	w.body.Reset()
	w.mu.Unlock()
}
func (w *timeoutResponseWriter) commit(dst gin.ResponseWriter) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.timedOut {
		return
	}
	copyTimeoutHeaders(dst.Header(), w.header)
	dst.WriteHeader(w.status)
	_, _ = dst.Write(w.body.Bytes())
}

func copyTimeoutHeaders(dst, src http.Header) {
	for key, values := range src {
		dst[key] = append([]string(nil), values...)
	}
}

func writeTimeoutResponse(writer http.ResponseWriter) {
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(http.StatusGatewayTimeout)
	_, _ = writer.Write([]byte(`{"code":1006,"message":"request timeout","data":null}`))
}

// RequestTimeout returns the remaining timeout budget of the current request.
//
// RequestTimeout 返回当前请求剩余的超时预算。
func RequestTimeout(c *gin.Context) time.Duration {
	ctx := c.Request.Context()
	deadline, ok := ctx.Deadline()
	if !ok {
		return 0
	}
	remaining := time.Until(deadline)
	if remaining < 0 {
		return 0
	}
	return remaining
}

// IsRequestCanceled reports whether the current request context has been canceled.
//
// IsRequestCanceled 判断当前请求上下文是否已经取消。
func IsRequestCanceled(c *gin.Context) bool {
	ctx := c.Request.Context()
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}
