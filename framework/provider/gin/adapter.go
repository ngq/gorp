// Package gin provides Gin-based HTTP server implementation for gorp framework.
// This file implements handler and middleware adaptation from framework contracts to Gin handlers.
// Converts transport HTTPHandler/HTTPMiddleware into Gin.HandlerFunc.
//
// Gin HTTP 服务包，提供基于 Gin 的 HTTP 服务器实现，用于 gorp 框架。
// 本文件实现 handler 和 middleware 适配，从框架契约转为 Gin handler。
// 将 transport HTTPHandler/HTTPMiddleware 转换为 Gin.HandlerFunc。
package gin

import (
	"net/http"

	"github.com/gin-gonic/gin"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

// AdaptMiddleware adapts a framework HTTPMiddleware into a Gin.HandlerFunc.
// This allows gorp governance middleware to be used directly on *gin.Engine or *gin.RouterGroup.
//
// AdaptMiddleware 将框架 HTTPMiddleware 适配为 Gin.HandlerFunc。
// 允许 gorp 治理 middleware 直接在 *gin.Engine 或 *gin.RouterGroup 上使用。
//
// Example:
//
//	engine := gin.New()
//	engine.Use(ginprovider.AdaptMiddleware(httpmiddleware.RequestIdentity()))
func AdaptMiddleware(middleware transportcontract.HTTPMiddleware) gin.HandlerFunc {
	return adaptMiddleware(middleware)
}

// AdaptHandler adapts a framework HTTPHandler into a Gin.HandlerFunc.
// This allows gorp handler functions to be used directly on Gin router groups.
//
// AdaptHandler 将框架 HTTPHandler 适配为 Gin.HandlerFunc。
// 允许 gorp handler 函数直接在 Gin 路由组上使用。
func AdaptHandler(handler transportcontract.HTTPHandler) gin.HandlerFunc {
	return adaptHandler(handler)
}

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
