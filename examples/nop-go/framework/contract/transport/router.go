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
type Router interface {
	Use(middleware ...Middleware)
	Group(prefix string, middleware ...Middleware) Router

	Handle(method, path string, handler Handler)
	HandleFunc(method, path string, handlerFunc Handler)

	GET(path string, handler Handler)
	POST(path string, handler Handler)
	PUT(path string, handler Handler)
	DELETE(path string, handler Handler)

	Mount(path string, handler http.Handler)
}
