// Package gin provides Gin-based HTTP server implementation for gorp framework.
// This file implements the HTTP router interface wrapping Gin RouterGroup.
// Supports route registration, middleware mounting, and child group creation.
//
// Gin HTTP 服务包，提供基于 Gin 的 HTTP 服务器实现，用于 gorp 框架。
// 本文件实现 HTTP 路由接口，包装 Gin RouterGroup。
// 支持路由注册、中间件挂载和子路由组创建。
package gin

import (
	"net/http"

	"github.com/gin-gonic/gin"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

type router struct {
	group *gin.RouterGroup
}

// newRouter wraps a Gin router group as a framework router.
//
// newRouter 将 Gin router group 包装为框架 router。
func newRouter(group *gin.RouterGroup) transportcontract.Router {
	return &router{group: group}
}

// Use registers middleware on the current router group.
//
// Use 在当前 router group 上注册中间件。
func (r *router) Use(middleware ...transportcontract.Middleware) {
	if r == nil || r.group == nil || len(middleware) == 0 {
		return
	}
	adapted := make([]gin.HandlerFunc, 0, len(middleware))
	for _, mw := range middleware {
		if mw == nil {
			continue
		}
		adapted = append(adapted, adaptMiddleware(mw))
	}
	if len(adapted) == 0 {
		return
	}
	r.group.Use(adapted...)
}

// Group creates a child router group with optional middleware.
//
// Group 创建一个可选挂载中间件的子路由组。
func (r *router) Group(prefix string, middleware ...transportcontract.Middleware) transportcontract.Router {
	if r == nil || r.group == nil {
		return &router{}
	}
	adapted := make([]gin.HandlerFunc, 0, len(middleware))
	for _, mw := range middleware {
		if mw == nil {
			continue
		}
		adapted = append(adapted, adaptMiddleware(mw))
	}
	return newRouter(r.group.Group(prefix, adapted...))
}

// Handle registers a route handler for the given method and path.
//
// Handle 为指定 method 和 path 注册路由处理器。
func (r *router) Handle(method, path string, handler transportcontract.Handler) {
	if r == nil || r.group == nil || handler == nil {
		return
	}
	r.group.Handle(method, path, adaptHandler(handler))
}

// HandleFunc is a function-style alias for Handle.
//
// HandleFunc 是 Handle 的函数式别名。
func (r *router) HandleFunc(method, path string, handlerFunc transportcontract.Handler) {
	if handlerFunc == nil {
		return
	}
	r.Handle(method, path, handlerFunc)
}

// GET registers a GET route handler.
//
// GET 注册 GET 路由处理器。
func (r *router) GET(path string, handler transportcontract.Handler) {
	r.Handle(http.MethodGet, path, handler)
}

// POST registers a POST route handler.
//
// POST 注册 POST 路由处理器。
func (r *router) POST(path string, handler transportcontract.Handler) {
	r.Handle(http.MethodPost, path, handler)
}

// PUT registers a PUT route handler.
//
// PUT 注册 PUT 路由处理器。
func (r *router) PUT(path string, handler transportcontract.Handler) {
	r.Handle(http.MethodPut, path, handler)
}

// DELETE registers a DELETE route handler.
//
// DELETE 注册 DELETE 路由处理器。
func (r *router) DELETE(path string, handler transportcontract.Handler) {
	r.Handle(http.MethodDelete, path, handler)
}

// Mount exposes a standard http.Handler on the given path.
//
// Mount 在指定路径挂载一个标准 http.Handler。
func (r *router) Mount(path string, handler http.Handler) {
	if handler == nil {
		return
	}
	h := wrapHandler(handler)
	r.group.Handle(http.MethodGet, path, h)
	r.group.Handle(http.MethodHead, path, h)
}
