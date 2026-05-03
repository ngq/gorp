package contract

import (
	"context"
	"net/http"
)

// HTTPContext 是 framework 默认 HTTP 处理上下文。
//
// 中文说明：
// - 这是默认 HTTP 业务主线暴露给 handler / middleware 的最小公共面；
// - 只覆盖当前 starter、生成器、基础端点、统一响应真正需要的读写能力；
// - 目标是让默认开发路径不再直接依赖 Gin，同时又避免在 contract 层重新发明一整套重型 Web API。
// - provider 可以在内部把它适配到底层具体实现，但默认业务代码不应再感知 provider 专属上下文类型。
type HTTPContext interface {
	Context() context.Context
	SetContext(ctx context.Context)

	Request() *http.Request
	SetRequest(req *http.Request)

	Param(key string) string
	Query(key string) string
	DefaultQuery(key, defaultValue string) string
	GetHeader(key string) string
	Header(key, value string)

	BindJSON(obj any) error
	BindQuery(obj any) error
	Bind(obj any) error

	JSON(status int, body any)
	Status(code int)

	RoutePath() string
	ResponseStatus() int
}

// HTTPHandler 是 framework 默认 HTTP handler 契约。
//
// 中文说明：
// - 默认业务 handler 统一接收 HTTPContext；
// - 这样 starter、facade、模板、生成器都能围绕同一条 latest-only 主线组织；
// - provider 内部如仍使用 Gin 或其他实现，应在 provider 边界完成适配。
type HTTPHandler func(HTTPContext)

// HTTPNext 表示 middleware 中继续执行后续链路的回调。
type HTTPNext func()

// HTTPMiddleware 是 framework 默认 HTTP middleware 契约。
//
// 中文说明：
// - middleware 不再直接暴露 Gin 类型；
// - 通过 `ctx + next` 的最小形态表达链式处理，便于 provider 在内部适配到底层实现；
// - 默认主线下，starter / bootstrap / 模板 / 业务模块都应优先依赖它，而不是 Gin middleware。
type HTTPMiddleware func(ctx HTTPContext, next HTTPNext)

// HTTPRouter 是最小 HTTP 路由契约。
//
// 中文说明：
// - 当前主线目标已经升级为“latest-only、无桥接态、默认可生产开发”；
// - 因此这里不再接受把 Gin 类型继续暴露为默认 handler / middleware 形态；
// - `Mount(path, http.Handler)` 专门留给 metrics / pprof / file server 这类原生 handler。
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
