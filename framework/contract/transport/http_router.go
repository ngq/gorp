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

// HTTPRouter defines the transport-layer HTTP router abstraction.
//
// HTTPRouter 定义 transport 层 HTTP 路由抽象。
type HTTPRouter interface {
	Use(middleware ...HTTPMiddleware)
	Group(prefix string, middleware ...HTTPMiddleware) HTTPRouter

	Handle(method, path string, handler HTTPHandler)
	HandleFunc(method, path string, handlerFunc HTTPHandler)

	GET(path string, handler HTTPHandler)
	POST(path string, handler HTTPHandler)
	PUT(path string, handler HTTPHandler)
	DELETE(path string, handler HTTPHandler)

	Mount(path string, handler http.Handler)
}
