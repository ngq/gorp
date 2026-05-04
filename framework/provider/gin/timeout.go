package gin

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

func Timeout(timeout time.Duration) transportcontract.HTTPMiddleware {
	return func(next transportcontract.HTTPHandler) transportcontract.HTTPHandler {
		return func(c transportcontract.HTTPContext) {
			if timeout <= 0 {
				if next != nil {
					next(c)
				}
				return
			}

			ctx, cancel := context.WithTimeout(c.Context(), timeout)
			defer cancel()
			c.SetContext(ctx)

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

func IsRequestCanceled(c *gin.Context) bool {
	ctx := c.Request.Context()
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}
