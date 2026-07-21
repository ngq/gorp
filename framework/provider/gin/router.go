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
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

type router struct {
	group    *gin.RouterGroup
	engine   *gin.Engine
	registry *routeRegistry
	prefix   string
}

type routeRegistry struct {
	mu     sync.RWMutex
	routes []transportcontract.RouteInfo
}

// newRouter wraps a Gin router group as a framework router.
//
// newRouter 将 Gin router group 包装为框架 router。
func newRouter(group *gin.RouterGroup, engines ...*gin.Engine) transportcontract.Router {
	var engine *gin.Engine
	if len(engines) > 0 {
		engine = engines[0]
	}
	prefix := ""
	if group != nil && group.BasePath() != "/" {
		prefix = group.BasePath()
	}
	return &router{group: group, engine: engine, registry: &routeRegistry{}, prefix: prefix}
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
	if r == nil {
		return &router{registry: &routeRegistry{}, prefix: joinRoutePath("", prefix)}
	}
	if r.group == nil {
		return &router{engine: r.engine, registry: r.registry, prefix: joinRoutePath(r.prefix, prefix)}
	}
	adapted := make([]gin.HandlerFunc, 0, len(middleware))
	for _, mw := range middleware {
		if mw == nil {
			continue
		}
		adapted = append(adapted, adaptMiddleware(mw))
	}
	return &router{
		group:    r.group.Group(prefix, adapted...),
		engine:   r.engine,
		registry: r.registry,
		prefix:   joinRoutePath(r.prefix, prefix),
	}
}

// Handle 为指定 method 和 path 注册路由处理器，可选挂载接口级中间件。
// 中间件只对该路由端点生效，不影响其他路由。
// 执行顺序：中间件按参数顺序依次执行，最后执行 handler。
//
// Handle registers a route handler for the given method and path,
// with optional interface-level middleware that applies only to this route.
// Middleware executes in argument order, then the handler runs last.
func (r *router) Handle(method, path string, handler transportcontract.Handler, middleware ...transportcontract.Middleware) {
	if r == nil || r.group == nil || handler == nil {
		return
	}
	handlers := make([]gin.HandlerFunc, 0, len(middleware)+1)
	for _, mw := range middleware {
		if mw != nil {
			handlers = append(handlers, adaptMiddleware(mw))
		}
	}
	handlers = append(handlers, adaptHandler(handler))
	r.group.Handle(method, path, handlers...)
	r.registry.add(transportcontract.RouteInfo{Method: strings.ToUpper(method), Path: joinRoutePath(r.prefix, path)})
}

// HandleFunc 是 Handle 的函数式别名，支持可选接口级中间件。
//
// HandleFunc is a function-style alias for Handle with optional interface-level middleware.
func (r *router) HandleFunc(method, path string, handlerFunc transportcontract.Handler, middleware ...transportcontract.Middleware) {
	if handlerFunc == nil {
		return
	}
	r.Handle(method, path, handlerFunc, middleware...)
}

// GET 注册 GET 路由处理器，可选挂载接口级中间件。
//
// GET registers a GET route handler with optional interface-level middleware.
func (r *router) GET(path string, handler transportcontract.Handler, middleware ...transportcontract.Middleware) {
	r.Handle(http.MethodGet, path, handler, middleware...)
}

// POST 注册 POST 路由处理器，可选挂载接口级中间件。
//
// POST registers a POST route handler with optional interface-level middleware.
func (r *router) POST(path string, handler transportcontract.Handler, middleware ...transportcontract.Middleware) {
	r.Handle(http.MethodPost, path, handler, middleware...)
}

// PUT 注册 PUT 路由处理器，可选挂载接口级中间件。
//
// PUT registers a PUT route handler with optional interface-level middleware.
func (r *router) PUT(path string, handler transportcontract.Handler, middleware ...transportcontract.Middleware) {
	r.Handle(http.MethodPut, path, handler, middleware...)
}

// DELETE 注册 DELETE 路由处理器，可选挂载接口级中间件。
//
// DELETE registers a DELETE route handler with optional interface-level middleware.
func (r *router) DELETE(path string, handler transportcontract.Handler, middleware ...transportcontract.Middleware) {
	r.Handle(http.MethodDelete, path, handler, middleware...)
}

func (r *router) PATCH(path string, handler transportcontract.Handler, middleware ...transportcontract.Middleware) {
	r.Handle(http.MethodPatch, path, handler, middleware...)
}

func (r *router) HEAD(path string, handler transportcontract.Handler, middleware ...transportcontract.Middleware) {
	r.Handle(http.MethodHead, path, handler, middleware...)
}

func (r *router) OPTIONS(path string, handler transportcontract.Handler, middleware ...transportcontract.Middleware) {
	r.Handle(http.MethodOptions, path, handler, middleware...)
}

func (r *router) ANY(path string, handler transportcontract.Handler, middleware ...transportcontract.Middleware) {
	for _, method := range []string{
		http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch,
		http.MethodHead, http.MethodOptions, http.MethodDelete, http.MethodConnect, http.MethodTrace,
	} {
		r.Handle(method, path, handler, middleware...)
	}
}

// Mount exposes a standard http.Handler on the given path.
//
// Mount 在指定路径挂载一个标准 http.Handler。
func (r *router) Mount(path string, handler http.Handler) {
	if r == nil || r.group == nil || handler == nil {
		return
	}
	h := wrapHandler(handler)
	r.group.Handle(http.MethodGet, path, h)
	r.registry.add(transportcontract.RouteInfo{Method: http.MethodGet, Path: joinRoutePath(r.prefix, path)})
	r.group.Handle(http.MethodHead, path, h)
	r.registry.add(transportcontract.RouteInfo{Method: http.MethodHead, Path: joinRoutePath(r.prefix, path)})
}

func (r *router) Static(relativePath, root string) {
	if r == nil || r.group == nil {
		return
	}
	r.group.Static(relativePath, root)
	r.recordStatic(relativePath, true)
}

func (r *router) StaticFile(relativePath, filePath string) {
	if r == nil || r.group == nil {
		return
	}
	r.group.StaticFile(relativePath, filePath)
	r.recordStatic(relativePath, false)
}

func (r *router) StaticFS(relativePath string, fs http.FileSystem) {
	if r == nil || r.group == nil {
		return
	}
	r.group.StaticFS(relativePath, fs)
	r.recordStatic(relativePath, true)
}

func (r *router) recordStatic(relativePath string, wildcard bool) {
	path := relativePath
	if wildcard {
		path = joinRoutePath(path, "/*filepath")
	}
	r.registry.add(transportcontract.RouteInfo{Method: http.MethodGet, Path: joinRoutePath(r.prefix, path)})
	r.registry.add(transportcontract.RouteInfo{Method: http.MethodHead, Path: joinRoutePath(r.prefix, path)})
}

func (r *router) NoRoute(handler transportcontract.Handler) {
	if r == nil || r.engine == nil || handler == nil {
		return
	}
	r.engine.NoRoute(adaptHandler(handler))
}

func (r *router) NoMethod(handler transportcontract.Handler) {
	if r == nil || r.engine == nil || handler == nil {
		return
	}
	r.engine.HandleMethodNotAllowed = true
	r.engine.NoMethod(adaptHandler(handler))
}

func (r *router) Routes() []transportcontract.RouteInfo {
	if r == nil || r.registry == nil {
		return []transportcontract.RouteInfo{}
	}
	return r.registry.snapshot()
}

func (r *routeRegistry) add(route transportcontract.RouteInfo) {
	if r == nil {
		return
	}
	r.mu.Lock()
	r.routes = append(r.routes, route)
	r.mu.Unlock()
}

func (r *routeRegistry) snapshot() []transportcontract.RouteInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]transportcontract.RouteInfo(nil), r.routes...)
}

func joinRoutePath(prefix, path string) string {
	prefix = strings.TrimSpace(prefix)
	path = strings.TrimSpace(path)
	if prefix == "" || prefix == "/" {
		if path == "" || path == "/" {
			return "/"
		}
		return "/" + strings.TrimLeft(path, "/")
	}
	if path == "" || path == "/" {
		return "/" + strings.Trim(prefix, "/")
	}
	return "/" + strings.Trim(prefix, "/") + "/" + strings.TrimLeft(path, "/")
}
