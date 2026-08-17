// Package gin provides Gin-based HTTP server implementation for gorp framework.
// This file implements handler and middleware adaptation from framework contracts to Gin handlers.
// Converts transport Handler/Middleware into Gin.HandlerFunc.
//
// Gin HTTP 服务包，提供基于 Gin 的 HTTP 服务器实现，用于 gorp 框架。
// 本文件实现 handler 和 middleware 适配，从框架契约转为 Gin handler。
// 将 transport Handler/Middleware 转换为 Gin.HandlerFunc。
package gin

import (
	"net/http"
	"reflect"
	"sync/atomic"

	"github.com/gin-gonic/gin"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

// NewTestEngine creates a gin.Engine configured for testing with ContextWithFallback enabled.
// This ensures gin.Context.Value() properly delegates to Request.Context().Value() for
// non-string keys, which is required for context.Context value propagation.
//
// NewTestEngine 创建用于测试的 gin.Engine，启用 ContextWithFallback。
// 确保 gin.Context.Value() 正确委托到 Request.Context().Value() 处理非字符串 key，
// 这是 context.Context 值传播的必要设置。
//
// Example:
//
//	engine := NewTestEngine()
//	engine.Use(YourMiddleware())
//	engine.GET("/test", handler)
func NewTestEngine() *gin.Engine {
	engine := gin.New()
	engine.ContextWithFallback = true
	return engine
}

// AdaptMiddleware adapts a framework Middleware into a Gin.HandlerFunc.
// This allows gorp governance middleware to be used directly on *gin.Engine or *gin.RouterGroup.
//
// AdaptMiddleware 将框架 Middleware 适配为 Gin.HandlerFunc。
// 允许 gorp 治理 middleware 直接在 *gin.Engine 或 *gin.RouterGroup 上使用。
//
// Example:
//
//	engine := gin.New()
//	engine.Use(ginprovider.AdaptMiddleware(httpmiddleware.RequestIdentity()))
func AdaptMiddleware(middleware transportcontract.Middleware) gin.HandlerFunc {
	return adaptMiddleware(middleware)
}

// AdaptHandler adapts a framework Handler into a Gin.HandlerFunc.
// This allows gorp handler functions to be used directly on Gin router groups.
//
// AdaptHandler 将框架 Handler 适配为 Gin.HandlerFunc。
// 允许 gorp handler 函数直接在 Gin 路由组上使用。
func AdaptHandler(handler transportcontract.Handler) gin.HandlerFunc {
	return adaptHandler(handler)
}

// adaptMiddleware adapts a transport middleware into a Gin handler.
//
// adaptMiddleware 将 transport middleware 适配为 Gin handler。
func adaptMiddleware(middleware transportcontract.Middleware) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if middleware == nil {
			ctx.Next()
			return
		}
		// The contract middleware owns continuation of the remaining Gin chain.
		// Keep that continuation on a private execution cursor so middleware such
		// as Timeout can invoke next from a worker without racing Gin's outer cursor.
		worker := cloneContextForContinuation(ctx)
		ctx.Abort()
		httpCtx := newContext(worker)
		// next 会把最终 Request 先写入原子槽再启动 worker 链；wrapped 返回后
		// 异步中间件（如 Timeout）可能仍在 worker goroutine 中执行 next，
		// 直接读 worker.Request 与该写构成 data race。
		var nextReq atomic.Pointer[http.Request]
		next := func(c transportcontract.Context) {
			if c != nil && c.Request() != nil {
				nextReq.Store(c.Request())
				worker.Request = c.Request()
			}
			worker.Next()
		}
		wrapped := middleware(next)
		if wrapped == nil {
			ctx.Next()
			return
		}
		wrapped(httpCtx)
		if req := nextReq.Load(); req != nil {
			ctx.Request = req
		}
	}
}

// cloneContextForContinuation snapshots Gin's private middleware cursor. Gin's
// public Copy intentionally removes the handler chain, so it cannot continue it.
func cloneContextForContinuation(ctx *gin.Context) *gin.Context {
	worker := reflect.New(reflect.TypeOf(ctx).Elem()).Interface().(*gin.Context)
	reflect.ValueOf(worker).Elem().Set(reflect.ValueOf(ctx).Elem())
	return worker
}

// adaptHandler adapts a transport handler into a Gin handler.
//
// adaptHandler 将 transport handler 适配为 Gin handler。
func adaptHandler(handler transportcontract.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if handler == nil {
			return
		}
		handler(newContext(ctx))
	}
}

// wrapHandler adapts a standard http.Handler into a Gin handler.
//
// wrapHandler 将标准 http.Handler 适配为 Gin handler。
func wrapHandler(handler http.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		handler.ServeHTTP(ctx.Writer, ctx.Request)
	}
}
