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
	"context"
	"net/http"
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

			ctx, cancel := context.WithTimeout(c, timeout)
			defer cancel()

			done := make(chan struct{})
			go func() {
				defer close(done)
				if next != nil {
					next(c)
				}
			}()

			select {
			case <-done:
			case <-ctx.Done():
				// Wait for the handler to finish to avoid race condition
				// 等待 handler 完成以避免 race condition
				<-done
				if gc, ok := unwrapGinContext(c); ok {
					writeGinResponseHeaders(gc)
					resp := Response{
						Code:    CodeServiceUnavailable,
						Message: "request timeout",
						Data:    nil,
					}
					gc.JSON(http.StatusGatewayTimeout, resp)
					gc.Abort()
					return
				}
				c.JSON(http.StatusGatewayTimeout, map[string]any{
					"code":    CodeServiceUnavailable,
					"message": "request timeout",
				})
			}
		}
	}
}

// TimeoutMiddleware is the native Gin timeout middleware form.
//
// TimeoutMiddleware 是超时中间件的原生 Gin 形态。
func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)

		done := make(chan struct{})
		go func() {
			c.Next()
			close(done)
		}()

		select {
		case <-done:
		case <-ctx.Done():
			// Wait for the handler to finish to avoid race condition
			// 等待 handler 完成以避免 race condition
			<-done
			writeGinResponseHeaders(c)
			resp := Response{
				Code:    CodeServiceUnavailable,
				Message: "request timeout",
				Data:    nil,
			}
			c.JSON(http.StatusGatewayTimeout, resp)
			c.Abort()
		}
	}
}

// TimeoutMiddlewareWithHandler applies a timeout and delegates timeout output to a custom callback.
//
// TimeoutMiddlewareWithHandler 应用超时控制，并把超时后的输出交给自定义回调处理。
func TimeoutMiddlewareWithHandler(timeout time.Duration, onTimeout func(*gin.Context)) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)

		done := make(chan struct{})
		go func() {
			c.Next()
			close(done)
		}()

		select {
		case <-done:
		case <-ctx.Done():
			// Wait for the handler to finish to avoid race condition
			// 等待 handler 完成以避免 race condition
			<-done
			if onTimeout != nil {
				onTimeout(c)
			} else {
				writeGinResponseHeaders(c)
				resp := Response{
					Code:    CodeServiceUnavailable,
					Message: "request timeout",
					Data:    nil,
				}
				c.JSON(http.StatusGatewayTimeout, resp)
			}
			c.Abort()
		}
	}
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
