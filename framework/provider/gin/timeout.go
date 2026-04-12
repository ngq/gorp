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
// - 为请求设置超时时间，超过指定时间后返回 504 Gateway Timeout；
// - 将带有超时的 context 注入到 gin.Context，业务代码可以使用；
// - 注意：此中间件只是设置超时 context，实际的请求取消需要业务代码配合。
func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 创建带有超时的 context
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		// 更新请求的 context
		c.Request = c.Request.WithContext(ctx)

		// 设置响应通道
		done := make(chan struct{})
		go func() {
			c.Next()
			close(done)
		}()

		select {
		case <-done:
			// 请求正常完成
		case <-ctx.Done():
			// 请求超时
			c.JSON(http.StatusGatewayTimeout, gin.H{
				"error": "request timeout",
			})
			c.Abort()
		}
	}
}

// TimeoutMiddlewareWithHandler 创建带自定义超时处理的中间件。
//
// 中文说明：
// - 与 TimeoutMiddleware 类似，但支持自定义超时处理函数；
// - 可以在超时时执行自定义逻辑（如记录日志、清理资源等）。
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
				c.JSON(http.StatusGatewayTimeout, gin.H{
					"error": "request timeout",
				})
			}
			c.Abort()
		}
	}
}

// RequestTimeout 从 gin.Context 获取请求超时时间。
//
// 中文说明：
// - 业务代码可以使用此函数检查请求是否即将超时；
// - 返回剩余时间，如果已经超时或 context 已取消，返回 0。
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
//
// 中文说明：
// - 业务代码可以在长时间操作中检查此状态；
// - 如果返回 true，应该尽快结束处理并返回。
func IsRequestCanceled(c *gin.Context) bool {
	ctx := c.Request.Context()
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}