// Application scenarios:
// - Adapt the framework transport handler and middleware contracts onto Gin handlers.
// - Keep transport-level middleware reusable while still executing inside Gin's handler chain.
// - Mount standard net/http handlers through the same Gin provider surface.
//
// 适用场景：
// - 将框架 transport handler 与 middleware 契约适配到 Gin handler。
// - 让 transport 层中间件可复用，同时仍运行在 Gin 的处理链中。
// - 通过同一套 Gin provider 能力挂载标准 net/http handler。
package gin

import (
	"net/http"

	"github.com/gin-gonic/gin"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

// adaptMiddleware adapts a transport middleware into a Gin handler.
//
// adaptMiddleware 将 transport middleware 适配为 Gin handler。
func adaptMiddleware(middleware transportcontract.HTTPMiddleware) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if middleware == nil {
			ctx.Next()
			return
		}
		httpCtx := newHTTPContext(ctx)
		next := func(c transportcontract.HTTPContext) {
			if c != nil && c.Request() != nil {
				ctx.Request = c.Request()
			}
			ctx.Next()
			httpCtx.SetRequest(ctx.Request)
		}
		wrapped := middleware(next)
		if wrapped == nil {
			ctx.Next()
			return
		}
		wrapped(httpCtx)
		ctx.Request = httpCtx.Request()
	}
}

// adaptHandler adapts a transport handler into a Gin handler.
//
// adaptHandler 将 transport handler 适配为 Gin handler。
func adaptHandler(handler transportcontract.HTTPHandler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if handler == nil {
			return
		}
		handler(newHTTPContext(ctx))
	}
}

// wrapHTTPHandler adapts a standard http.Handler into a Gin handler.
//
// wrapHTTPHandler 将标准 http.Handler 适配为 Gin handler。
func wrapHTTPHandler(handler http.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		handler.ServeHTTP(ctx.Writer, ctx.Request)
	}
}
