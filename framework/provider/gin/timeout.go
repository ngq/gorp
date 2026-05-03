package gin

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// TimeoutMiddleware 创建请求超时中间件。
//
// 中文说明：
// - 这是 Gin provider 扩展层中间件，不属于默认 framework 主线契约；
// - 为请求设置超时时间，超过指定时间后返回 504 Gateway Timeout；
// - 将带有超时的 context 注入到 gin.Context，业务代码可以继续复用。
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

// TimeoutMiddlewareWithHandler 创建带自定义超时处理的中间件。
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

// RequestTimeout 从 gin.Context 获取请求超时时间。
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

// IsRequestCanceled 检查请求是否已被取消。
func IsRequestCanceled(c *gin.Context) bool {
	ctx := c.Request.Context()
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}
