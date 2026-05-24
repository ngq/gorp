// Application scenarios:
// - Define the framework-agnostic HTTP request context contract.
// - Let provider adapters implement this interface by delegating to concrete HTTP frameworks.
// - Keep routing, binding, response output, and request metadata access consistent across providers.
//
// 适用场景：
// - 定义与具体框架无关的 HTTP 请求上下文契约。
// - 让 provider 适配层通过委托模式实现此接口。
// - 在不同 provider 之间统一路由、绑定、响应输出和请求元数据访问语义。
package transport

import (
	"context"
	"mime/multipart"
	"net/http"
)

// RequestContext 提供请求元数据访问。
// 注意：与 context.Context 解耦 —— 业务/中间件如需获取标准 context.Context，
// 必须通过 c.Context() 方法显式获取，禁止把 RequestContext 直接当作 context.Context 使用。
// 这样做的根本原因：若直接嵌入 context.Context，provider 实现（如 ginContext）会与
// Request().Context() 形成双向引用，标准库遍历 valueCtx 链时触发无限递归导致栈溢出。
// 详见 docs/design/bug-list.md BUG-001。
type RequestContext interface {
	// Context 返回标准 context.Context。
	// 用于 context.WithTimeout/context.WithCancel/context.WithValue 等场景。
	// 实现层应返回 Request().Context()，确保只返回纯标准 context，不再回包到 RequestContext。
	Context() context.Context

	// Request/Response
	Request() *http.Request
	Response() http.ResponseWriter

	// Route params
	Param(key string) string

	// Query params
	Query(key string) string
	DefaultQuery(key, defaultValue string) string
	DefaultIntQuery(key string, defaultValue int) int

	// Typed param parsing
	Int64Param(key string) (int64, error)

	// File upload
	FormFile(name string) (multipart.File, *multipart.FileHeader, error)
	SaveUploadedFile(file *multipart.FileHeader, dst string) error

	// Headers
	GetHeader(key string) string
	SetHeader(key, value string)
}

// BindingContext 提供请求绑定能力。
// 用于需要解析请求体的 Handler。
type BindingContext interface {
	Bind(obj any) error
	BindJSON(obj any) error
	BindQuery(obj any) error
}

// ResponseContext 提供响应输出能力。
// 用于需要返回响应的 Handler。
type ResponseContext interface {
	JSON(status int, body any)
	String(status int, body string)
	XML(status int, body any)
	Data(status int, contentType string, body []byte)
	Redirect(status int, location string)
	Status(code int)
}

// MiddlewareContext 提供中间件控制流能力。
// 用于中间件实现。
// Get 返回 (value, exists)，与 Gin 原生语义对齐，
// 可区分"key 不存在"和"key 存在但值为 nil"。
type MiddlewareContext interface {
	Get(key string) (any, bool)
	Set(key string, value any)
	Abort(status int)
	AbortWithJSON(status int, body any)
	IsAborted() bool
	Next()
}

// RouteContext 提供路由信息访问。
// 用于需要知道当前路由信息的场景。
type RouteContext interface {
	RoutePath() string
	ResponseStatus() int
}

// Context is the aggregate interface composing all sub-interfaces.
// It provides the full HTTP request context abstraction.
// Providers implement this interface by delegating to their underlying context.
//
// Context 是聚合接口，组合所有子接口。
// 提供完整的 HTTP 请求上下文抽象。
// Provider 通过委托到底层 context 实现此接口。
type Context interface {
	RequestContext
	BindingContext
	ResponseContext
	MiddlewareContext
	RouteContext
}

// Handler defines the HTTP handler signature.
//
// Handler 定义 HTTP 处理器签名。
type Handler func(Context)

// Middleware defines the HTTP middleware signature.
//
// Middleware 定义 HTTP 中间件签名。
type Middleware func(next Handler) Handler

// MiddlewareFunc is a business-friendly helper signature used by Middleware.
//
// MiddlewareFunc 是 Middleware 使用的业务友好辅助签名。
type MiddlewareFunc func(ctx Context, next Handler)
