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
			worker := cloneGinContextForContinuation(gc)
			worker.Writer = buffered
			workerCtx := newContext(worker)

			var panicVal any
			done := make(chan struct{})
			go func() {
				defer close(done)
				defer func() {
					if r := recover(); r != nil {
						panicVal = r
					}
				}()
				if next != nil {
					next(workerCtx)
				}
			}()

			select {
			case <-done:
				if panicVal != nil {
					buffered.timeout()
					panic(panicVal)
				}
				buffered.commit(base)
			case <-ctx.Done():
				if buffered.timeout() {
					writeTimeoutResponse(base)
				}
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

		var panicVal any
		done := make(chan struct{})
		go func() {
			defer close(done)
			defer func() {
				if r := recover(); r != nil {
					panicVal = r
				}
			}()
			worker.Next()
		}()

		select {
		case <-done:
			if panicVal != nil {
				buffered.timeout()
				panic(panicVal)
			}
			buffered.commit(base)
		case <-ctx.Done():
			if buffered.timeout() {
				writeTimeoutResponse(base)
			}
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

		var panicVal any
		done := make(chan struct{})
		go func() {
			defer close(done)
			defer func() {
				if r := recover(); r != nil {
					panicVal = r
				}
			}()
			worker.Next()
		}()

		select {
		case <-done:
			if panicVal != nil {
				buffered.timeout()
				panic(panicVal)
			}
			buffered.commit(base)
		case <-ctx.Done():
			if !buffered.timeout() {
				c.Abort()
				return
			}
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
// Deep-copies Keys map and clones Request to prevent concurrent data races with sync.Pool reuse.
func cloneGinContextForContinuation(ctx *gin.Context) *gin.Context {
	worker := reflect.New(reflect.TypeOf(ctx).Elem()).Interface().(*gin.Context)
	reflect.ValueOf(worker).Elem().Set(reflect.ValueOf(ctx).Elem())
	if ctx.Keys != nil {
		copiedKeys := make(map[any]any, len(ctx.Keys))
		for k, v := range ctx.Keys {
			copiedKeys[k] = v
		}
		worker.Keys = copiedKeys
	}
	if ctx.Request != nil {
		worker.Request = ctx.Request.Clone(ctx.Request.Context())
	}
	return worker
}

const maxTimeoutResponseBytes = 10 << 20

type timeoutResponseWriter struct {
	gin.ResponseWriter
	mu        sync.Mutex
	header    http.Header
	body      bytes.Buffer
	status    int
	timedOut  bool
	overflow  bool
	streaming bool
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
	if w.streaming {
		return w.ResponseWriter.Write(data)
	}
	if w.body.Len()+len(data) > maxTimeoutResponseBytes {
		w.overflow = true
		return 0, http.ErrContentLength
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
func (w *timeoutResponseWriter) Flush() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.timedOut || w.overflow {
		return
	}
	copyTimeoutHeaders(w.ResponseWriter.Header(), w.header)
	w.ResponseWriter.WriteHeader(w.status)
	if w.body.Len() > 0 {
		_, _ = w.ResponseWriter.Write(w.body.Bytes())
		w.body.Reset()
	}
	w.ResponseWriter.Flush()
	w.streaming = true
}
func (w *timeoutResponseWriter) Pusher() http.Pusher      { return nil }
func (w *timeoutResponseWriter) CloseNotify() <-chan bool { return w.ResponseWriter.CloseNotify() }
func (w *timeoutResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, errors.New("timeout middleware does not support hijacking")
}
func (w *timeoutResponseWriter) timeout() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.streaming {
		w.timedOut = true
		return false
	}
	w.timedOut = true
	w.status = http.StatusGatewayTimeout
	w.body.Reset()
	return true
}
func (w *timeoutResponseWriter) commit(dst gin.ResponseWriter) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.timedOut || w.streaming {
		return
	}
	copyTimeoutHeaders(dst.Header(), w.header)
	if w.overflow {
		dst.Header().Set("Content-Type", "application/json; charset=utf-8")
		dst.WriteHeader(http.StatusInternalServerError)
		_, _ = dst.Write([]byte(`{"code":13,"reason":"Internal","message":"response exceeds timeout middleware buffer limit","data":null}`))
		return
	}
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
	// code 4 = httpx.CodeDeadlineExceeded（canonical，AIP-193），与 HTTP 504 对齐。
	// 旧值 1006 是 legacy 词汇（ServiceUnavailable/503），语义错位（FEAT-01）。
	_, _ = writer.Write([]byte(`{"code":4,"reason":"DeadlineExceeded","message":"request timeout","data":null}`))
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
