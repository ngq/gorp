// Application scenarios:
// - Define the framework-agnostic HTTP request context contract.
// - Provide a reusable default implementation that providers can adapt onto concrete HTTP stacks.
// - Keep routing, binding, response output, and request metadata access consistent across providers.
//
// 适用场景：
// - 定义与具体框架无关的 HTTP 请求上下文契约。
// - 提供一个可复用的默认实现，供各类 provider 适配到具体 HTTP 框架。
// - 在不同 provider 之间统一路由、绑定、响应输出和请求元数据访问语义。
package transport

import (
	"context"
	"errors"
	"net/http"
)

// errBindFuncNotConfigured is returned by Bind/BindJSON/BindQuery when the
// underlying binding function has not been set, so callers cannot silently
// operate on a zero-valued target object.
//
// errBindFuncNotConfigured 在绑定函数未配置时由 Bind/BindJSON/BindQuery 返回，
// 避免调用方误以为绑定成功而对零值对象进行操作。
var errBindFuncNotConfigured = errors.New("bind function not configured")

// HTTPContext defines the transport-layer HTTP request context abstraction.
//
// Future improvement: Consider splitting into smaller interfaces for ISP compliance:
//
//	HTTPRequestReader  - Request(), GetHeader(), Param(), Query(), etc.
//	HTTPResponseWriter - JSON(), String(), XML(), Data(), Status(), etc.
//	HTTPMiddlewareContext - Get(), Set(), Abort(), Next(), etc.
//	HTTPContext composes the above + Bind().
//
// HTTPContext 定义 transport 层 HTTP 请求上下文抽象。
// 包含中间件所需的 Get/Set/Abort/Next 方法，支持认证中间件等场景。
//
// 未来改进：考虑拆分为更小的接口以符合接口隔离原则：
//
//	HTTPRequestReader  - Request(), GetHeader(), Param(), Query() 等
//	HTTPResponseWriter - JSON(), String(), XML(), Data(), Status() 等
//	HTTPMiddlewareContext - Get(), Set(), Abort(), Next() 等
//	HTTPContext 组合上述接口 + Bind()
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
	String(status int, body string)
	XML(status int, body any)
	Data(status int, contentType string, body []byte)
	Redirect(status int, location string)
	Status(code int)

	RoutePath() string
	ResponseStatus() int

	// Get retrieves a value stored in the context by key.
	// Used by middleware to pass data to handlers (e.g., user info from auth).
	//
	// Get 从上下文中按 key 获取存储的值。
	// 用于中间件向 handler 传递数据（如认证后的用户信息）。
	Get(key string) any

	// Set stores a key-value pair in the context.
	// Used by middleware to store data for downstream handlers.
	//
	// Set 在上下文中存储 key-value 对。
	// 用于中间件为下游 handler 存储数据。
	Set(key string, value any)

	// Abort aborts the request chain with the given status code.
	// Used by middleware to stop further processing (e.g., auth failure).
	//
	// Abort 以给定状态码中止请求链。
	// 用于中间件停止后续处理（如认证失败）。
	Abort(status int)

	// AbortWithJSON aborts the request chain and sends a JSON response.
	// Convenience method for auth middleware to return error details.
	//
	// AbortWithJSON 中止请求链并发送 JSON 响应。
	// 认证中间件返回错误详情的便捷方法。
	AbortWithJSON(status int, body any)

	// IsAborted returns whether the request chain has been aborted.
	// Used by handlers to check if middleware has stopped processing.
	//
	// IsAborted 返回请求链是否已被中止。
	// 用于 handler 检查中间件是否已停止处理。
	IsAborted() bool

	// Next continues to the next handler in the chain.
	// Used by middleware to pass control to the next handler.
	//
	// Next 继续执行链中的下一个 handler。
	// 用于中间件将控制权传递给下一个 handler。
	Next()
}

// HTTPHandler defines the transport-layer HTTP handler signature.
//
// HTTPHandler 定义 transport 层 HTTP 处理器签名。
type HTTPHandler func(HTTPContext)

// HTTPMiddleware defines the transport-layer HTTP middleware signature.
//
// HTTPMiddleware 定义 transport 层 HTTP 中间件签名。
type HTTPMiddleware func(next HTTPHandler) HTTPHandler

// DefaultHTTPContext is the reusable default implementation of HTTPContext.
//
// DefaultHTTPContext 是 HTTPContext 的可复用默认实现。
type DefaultHTTPContext struct {
	ctx            context.Context
	request        *http.Request
	paramFunc      func(string) string
	queryFunc      func(string) string
	defaultQueryFn func(string, string) string
	headerFunc     func(string) string
	setHeaderFunc  func(string, string)
	bindJSONFunc   func(any) error
	bindQueryFunc  func(any) error
	bindFunc       func(any) error
	jsonFunc       func(int, any)
	stringFunc     func(int, string)
	xmlFunc        func(int, any)
	dataFunc       func(int, string, []byte)
	redirectFunc   func(int, string)
	statusFunc     func(int)
	routePathFunc  func() string
	statusReadFunc func() int
	// 中间件相关函数
	getFunc       func(string) any
	setFunc       func(string, any)
	abortFunc     func(int)
	abortJSONFunc func(int, any)
	isAbortedFunc func() bool
	nextFunc      func()
}

// NewDefaultHTTPContext creates a default transport HTTP context.
//
// NewDefaultHTTPContext 创建一个默认 transport HTTP 上下文。
func NewDefaultHTTPContext(ctx context.Context, req *http.Request) *DefaultHTTPContext {
	return &DefaultHTTPContext{ctx: ctx, request: req}
}

// Context returns the bound request context.
//
// Context 返回已绑定的请求 context。
func (c *DefaultHTTPContext) Context() context.Context {
	if c == nil {
		return nil
	}
	return c.ctx
}

// SetContext replaces the current request context and syncs it back into the request when possible.
//
// SetContext 替换当前请求 context，并在可能时同步回 request。
func (c *DefaultHTTPContext) SetContext(ctx context.Context) {
	if c == nil {
		return
	}
	c.ctx = ctx
	if c.request != nil && ctx != nil {
		// Keep request.Context() and the stored context aligned so downstream code observes one source of truth.
		// 让 request.Context() 与内部保存的 context 保持一致，避免下游看到两套上下文状态。
		c.request = c.request.WithContext(ctx)
	}
}

// Request returns the underlying HTTP request.
//
// Request 返回底层 HTTP request。
func (c *DefaultHTTPContext) Request() *http.Request {
	if c == nil {
		return nil
	}
	return c.request
}

// SetRequest replaces the underlying HTTP request and refreshes the stored context from it.
//
// SetRequest 替换底层 HTTP request，并从中刷新内部保存的 context。
func (c *DefaultHTTPContext) SetRequest(req *http.Request) {
	if c == nil {
		return
	}
	c.request = req
	if req != nil {
		c.ctx = req.Context()
	}
}

// Param returns a route parameter by key.
//
// Param 按 key 返回路由参数。
func (c *DefaultHTTPContext) Param(key string) string {
	if c == nil || c.paramFunc == nil {
		return ""
	}
	return c.paramFunc(key)
}

// Query returns a query parameter by key.
//
// Query 按 key 返回查询参数。
func (c *DefaultHTTPContext) Query(key string) string {
	if c == nil || c.queryFunc == nil {
		return ""
	}
	return c.queryFunc(key)
}

// DefaultQuery returns a query parameter or a default value.
//
// DefaultQuery 返回查询参数，未命中时返回默认值。
func (c *DefaultHTTPContext) DefaultQuery(key, defaultValue string) string {
	if c == nil || c.defaultQueryFn == nil {
		return defaultValue
	}
	return c.defaultQueryFn(key, defaultValue)
}

// GetHeader reads a request header by key.
//
// GetHeader 按 key 读取请求头。
func (c *DefaultHTTPContext) GetHeader(key string) string {
	if c == nil || c.headerFunc == nil {
		return ""
	}
	return c.headerFunc(key)
}

// Header writes a response header.
//
// Header 写入响应头。
func (c *DefaultHTTPContext) Header(key, value string) {
	if c == nil || c.setHeaderFunc == nil {
		return
	}
	c.setHeaderFunc(key, value)
}

// BindJSON binds a JSON payload into the target object.
// Returns an error when the binding function is not configured, so callers
// cannot silently operate on a zero-valued target.
//
// BindJSON 将 JSON 载荷绑定到目标对象。
// 绑定函数未配置时返回错误，避免调用方误以为绑定成功。
func (c *DefaultHTTPContext) BindJSON(obj any) error {
	if c == nil {
		return errBindFuncNotConfigured
	}
	if c.bindJSONFunc == nil {
		return errBindFuncNotConfigured
	}
	return c.bindJSONFunc(obj)
}

// BindQuery binds query parameters into the target object.
// Returns an error when the binding function is not configured.
//
// BindQuery 将查询参数绑定到目标对象。
// 绑定函数未配置时返回错误。
func (c *DefaultHTTPContext) BindQuery(obj any) error {
	if c == nil {
		return errBindFuncNotConfigured
	}
	if c.bindQueryFunc == nil {
		return errBindFuncNotConfigured
	}
	return c.bindQueryFunc(obj)
}

// Bind binds a generic request payload into the target object.
// Returns an error when the binding function is not configured.
//
// Bind 将通用请求载荷绑定到目标对象。
// 绑定函数未配置时返回错误。
func (c *DefaultHTTPContext) Bind(obj any) error {
	if c == nil {
		return errBindFuncNotConfigured
	}
	if c.bindFunc == nil {
		return errBindFuncNotConfigured
	}
	return c.bindFunc(obj)
}

// JSON writes a JSON response.
//
// JSON 输出 JSON 响应。
func (c *DefaultHTTPContext) JSON(status int, body any) {
	if c == nil || c.jsonFunc == nil {
		return
	}
	c.jsonFunc(status, body)
}

// String writes a plain string response.
//
// String 输出纯字符串响应。
func (c *DefaultHTTPContext) String(status int, body string) {
	if c == nil || c.stringFunc == nil {
		return
	}
	c.stringFunc(status, body)
}

// XML writes an XML response.
//
// XML 输出 XML 响应。
func (c *DefaultHTTPContext) XML(status int, body any) {
	if c == nil || c.xmlFunc == nil {
		return
	}
	c.xmlFunc(status, body)
}

// Data writes a binary response.
//
// Data 输出二进制响应。
func (c *DefaultHTTPContext) Data(status int, contentType string, body []byte) {
	if c == nil || c.dataFunc == nil {
		return
	}
	c.dataFunc(status, contentType, body)
}

// Redirect writes a redirect response.
//
// Redirect 输出重定向响应。
func (c *DefaultHTTPContext) Redirect(status int, location string) {
	if c == nil || c.redirectFunc == nil {
		return
	}
	c.redirectFunc(status, location)
}

// Status sets the response status code.
//
// Status 设置响应状态码。
func (c *DefaultHTTPContext) Status(code int) {
	if c == nil || c.statusFunc == nil {
		return
	}
	c.statusFunc(code)
}

// RoutePath returns the current matched route path when available.
//
// RoutePath 返回当前命中的路由路径。
func (c *DefaultHTTPContext) RoutePath() string {
	if c == nil || c.routePathFunc == nil {
		return ""
	}
	return c.routePathFunc()
}

// ResponseStatus returns the current response status code when available.
//
// ResponseStatus 返回当前响应状态码。
func (c *DefaultHTTPContext) ResponseStatus() int {
	if c == nil || c.statusReadFunc == nil {
		return 0
	}
	return c.statusReadFunc()
}

// SetParamFunc sets the route parameter lookup function.
//
// SetParamFunc 设置路由参数读取函数。
func (c *DefaultHTTPContext) SetParamFunc(fn func(string) string) {
	if c == nil {
		return
	}
	c.paramFunc = fn
}

// SetQueryFunc sets the query parameter lookup function.
//
// SetQueryFunc 设置查询参数读取函数。
func (c *DefaultHTTPContext) SetQueryFunc(fn func(string) string) {
	if c == nil {
		return
	}
	c.queryFunc = fn
}

// SetDefaultQueryFunc sets the default-query resolution function.
//
// SetDefaultQueryFunc 设置带默认值的查询参数解析函数。
func (c *DefaultHTTPContext) SetDefaultQueryFunc(fn func(string, string) string) {
	if c == nil {
		return
	}
	c.defaultQueryFn = fn
}

// SetHeaderFuncs sets request-header read and response-header write functions.
//
// SetHeaderFuncs 设置请求头读取和响应头写入函数。
func (c *DefaultHTTPContext) SetHeaderFuncs(get func(string) string, set func(string, string)) {
	if c == nil {
		return
	}
	c.headerFunc = get
	c.setHeaderFunc = set
}

// SetBindFuncs sets request binding functions.
//
// SetBindFuncs 设置请求绑定函数。
func (c *DefaultHTTPContext) SetBindFuncs(bindJSON func(any) error, bindQuery func(any) error, bind func(any) error) {
	if c == nil {
		return
	}
	c.bindJSONFunc = bindJSON
	c.bindQueryFunc = bindQuery
	c.bindFunc = bind
}

// SetResponseFuncs sets the concrete response writer functions.
//
// SetResponseFuncs 设置具体的响应写出函数。
func (c *DefaultHTTPContext) SetResponseFuncs(
	json func(int, any),
	str func(int, string),
	xml func(int, any),
	data func(int, string, []byte),
	redirect func(int, string),
	status func(int),
	statusRead func() int,
) {
	if c == nil {
		return
	}
	c.jsonFunc = json
	c.stringFunc = str
	c.xmlFunc = xml
	c.dataFunc = data
	c.redirectFunc = redirect
	c.statusFunc = status
	c.statusReadFunc = statusRead
}

// SetRoutePathFunc sets the route-path resolution function.
//
// SetRoutePathFunc 设置路由路径解析函数。
func (c *DefaultHTTPContext) SetRoutePathFunc(fn func() string) {
	if c == nil {
		return
	}
	c.routePathFunc = fn
}

// Get retrieves a value stored in the context by key.
//
// Get 从上下文中按 key 获取存储的值。
func (c *DefaultHTTPContext) Get(key string) any {
	if c == nil || c.getFunc == nil {
		return nil
	}
	return c.getFunc(key)
}

// Set stores a key-value pair in the context.
//
// Set 在上下文中存储 key-value 对。
func (c *DefaultHTTPContext) Set(key string, value any) {
	if c == nil || c.setFunc == nil {
		return
	}
	c.setFunc(key, value)
}

// Abort aborts the request chain with the given status code.
//
// Abort 以给定状态码中止请求链。
func (c *DefaultHTTPContext) Abort(status int) {
	if c == nil || c.abortFunc == nil {
		return
	}
	c.abortFunc(status)
}

// AbortWithJSON aborts the request chain and sends a JSON response.
//
// AbortWithJSON 中止请求链并发送 JSON 响应。
func (c *DefaultHTTPContext) AbortWithJSON(status int, body any) {
	if c == nil || c.abortJSONFunc == nil {
		return
	}
	c.abortJSONFunc(status, body)
}

// IsAborted returns whether the request chain has been aborted.
//
// IsAborted 返回请求链是否已被中止。
func (c *DefaultHTTPContext) IsAborted() bool {
	if c == nil || c.isAbortedFunc == nil {
		return false
	}
	return c.isAbortedFunc()
}

// Next continues to the next handler in the chain.
//
// Next 继续执行链中的下一个 handler。
func (c *DefaultHTTPContext) Next() {
	if c == nil || c.nextFunc == nil {
		return
	}
	c.nextFunc()
}

// SetMiddlewareFuncs sets the middleware-related functions.
//
// SetMiddlewareFuncs 设置中间件相关函数。
func (c *DefaultHTTPContext) SetMiddlewareFuncs(
	get func(string) any,
	set func(string, any),
	abort func(int),
	abortJSON func(int, any),
	isAborted func() bool,
	next func(),
) {
	if c == nil {
		return
	}
	c.getFunc = get
	c.setFunc = set
	c.abortFunc = abort
	c.abortJSONFunc = abortJSON
	c.isAbortedFunc = isAborted
	c.nextFunc = next
}
