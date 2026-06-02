// Application scenarios:
// - Define the framework-agnostic HTTP router contract.
// - Let provider adapters expose route registration and middleware composition through one shared shape.
// - Keep business route declaration independent from Gin or other concrete router implementations.
//
// 适用场景：
// - 定义与具体框架无关的 HTTP 路由契约。
// - 让 provider 适配层通过统一形态暴露路由注册与中间件组合能力。
// - 让业务路由声明不依赖 Gin 或其他具体路由实现。
package transport

import "net/http"

// Router defines the HTTP router abstraction.
//
// Router 定义 HTTP 路由抽象。
//
// 路由注册方法支持接口级中间件：
//   - 不传 middleware 时行为与原来一致：GET("/users", listUsers)
//   - 传入 middleware 时只对该路由端点生效：GET("/users", listUsers, authMW, rateLimitMW)
//   - 执行顺序：middleware 按参数顺序依次执行，最后执行 handler
//
// 中间件层级体系：
//   - 全局级：HTTP.UseGlobal(middleware) — 对所有路由生效，等同 Gin engine.Use()
//   - 组级：Router.Use(middleware) / Router.Group(prefix, middleware) — 对该组下所有路由生效
//   - 接口级：Router.GET(path, handler, middleware...) — 只对该路由端点生效
type Router interface {
	Use(middleware ...Middleware)
	Group(prefix string, middleware ...Middleware) Router

	// Handle 为指定 method 和 path 注册路由处理器，可选挂载接口级中间件。
	// middleware 只对该路由端点生效，不影响其他路由。
	//
	// Handle registers a route handler for the given method and path,
	// with optional interface-level middleware that applies only to this route.
	Handle(method, path string, handler Handler, middleware ...Middleware)

	// HandleFunc 是 Handle 的函数式别名，支持可选接口级中间件。
	//
	// HandleFunc is a function-style alias for Handle with optional interface-level middleware.
	HandleFunc(method, path string, handlerFunc Handler, middleware ...Middleware)

	// GET 注册 GET 路由处理器，可选挂载接口级中间件。
	// 示例：GET("/users", listUsers) 或 GET("/users", listUsers, authMW)
	//
	// GET registers a GET route handler with optional interface-level middleware.
	GET(path string, handler Handler, middleware ...Middleware)

	// POST 注册 POST 路由处理器，可选挂载接口级中间件。
	//
	// POST registers a POST route handler with optional interface-level middleware.
	POST(path string, handler Handler, middleware ...Middleware)

	// PUT 注册 PUT 路由处理器，可选挂载接口级中间件。
	//
	// PUT registers a PUT route handler with optional interface-level middleware.
	PUT(path string, handler Handler, middleware ...Middleware)

	// DELETE 注册 DELETE 路由处理器，可选挂载接口级中间件。
	//
	// DELETE registers a DELETE route handler with optional interface-level middleware.
	DELETE(path string, handler Handler, middleware ...Middleware)

	Mount(path string, handler http.Handler)
}
