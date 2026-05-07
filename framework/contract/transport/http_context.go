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
	"net/http"
)

// HTTPContext defines the transport-layer HTTP request context abstraction.
//
// HTTPContext 定义 transport 层 HTTP 请求上下文抽象。
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
//
// BindJSON 将 JSON 载荷绑定到目标对象。
func (c *DefaultHTTPContext) BindJSON(obj any) error {
	if c == nil || c.bindJSONFunc == nil {
		return nil
	}
	return c.bindJSONFunc(obj)
}

// BindQuery binds query parameters into the target object.
//
// BindQuery 将查询参数绑定到目标对象。
func (c *DefaultHTTPContext) BindQuery(obj any) error {
	if c == nil || c.bindQueryFunc == nil {
		return nil
	}
	return c.bindQueryFunc(obj)
}

// Bind binds a generic request payload into the target object.
//
// Bind 将通用请求载荷绑定到目标对象。
func (c *DefaultHTTPContext) Bind(obj any) error {
	if c == nil || c.bindFunc == nil {
		return nil
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
