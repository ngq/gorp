package contract

import (
	"context"
	"net/http"
)

// DefaultHTTPContext 是一个可被 provider 复用的最小 HTTP 上下文实现。
//
// 中文说明：
// - 它承载 framework 默认 HTTP 主线的通用请求/响应读写行为；
// - provider 可以在内部组合或嵌入它，减少重复实现同一批基础方法；
// - 该实现本身不依赖具体 Web 框架，只要求外部提供必要的函数适配底层行为。
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
	statusFunc     func(int)
	routePathFunc  func() string
	statusReadFunc func() int
}

// NewDefaultHTTPContext 创建可配置的默认 HTTPContext。
//
// 中文说明：
// - 由 provider 在适配层注入底层实现细节；
// - framework 其余层只依赖 HTTPContext 接口，不依赖具体 provider。
func NewDefaultHTTPContext(ctx context.Context, req *http.Request) *DefaultHTTPContext {
	return &DefaultHTTPContext{ctx: ctx, request: req}
}

func (c *DefaultHTTPContext) Context() context.Context {
	if c == nil {
		return nil
	}
	return c.ctx
}

func (c *DefaultHTTPContext) SetContext(ctx context.Context) {
	if c == nil {
		return
	}
	c.ctx = ctx
	if c.request != nil && ctx != nil {
		c.request = c.request.WithContext(ctx)
	}
}

func (c *DefaultHTTPContext) Request() *http.Request {
	if c == nil {
		return nil
	}
	return c.request
}

func (c *DefaultHTTPContext) SetRequest(req *http.Request) {
	if c == nil {
		return
	}
	c.request = req
	if req != nil {
		c.ctx = req.Context()
	}
}

func (c *DefaultHTTPContext) Param(key string) string {
	if c == nil || c.paramFunc == nil {
		return ""
	}
	return c.paramFunc(key)
}

func (c *DefaultHTTPContext) Query(key string) string {
	if c == nil || c.queryFunc == nil {
		return ""
	}
	return c.queryFunc(key)
}

func (c *DefaultHTTPContext) DefaultQuery(key, defaultValue string) string {
	if c == nil || c.defaultQueryFn == nil {
		return defaultValue
	}
	return c.defaultQueryFn(key, defaultValue)
}

func (c *DefaultHTTPContext) GetHeader(key string) string {
	if c == nil || c.headerFunc == nil {
		return ""
	}
	return c.headerFunc(key)
}

func (c *DefaultHTTPContext) Header(key, value string) {
	if c == nil || c.setHeaderFunc == nil {
		return
	}
	c.setHeaderFunc(key, value)
}

func (c *DefaultHTTPContext) BindJSON(obj any) error {
	if c == nil || c.bindJSONFunc == nil {
		return nil
	}
	return c.bindJSONFunc(obj)
}

func (c *DefaultHTTPContext) BindQuery(obj any) error {
	if c == nil || c.bindQueryFunc == nil {
		return nil
	}
	return c.bindQueryFunc(obj)
}

func (c *DefaultHTTPContext) Bind(obj any) error {
	if c == nil || c.bindFunc == nil {
		return nil
	}
	return c.bindFunc(obj)
}

func (c *DefaultHTTPContext) JSON(status int, body any) {
	if c == nil || c.jsonFunc == nil {
		return
	}
	c.jsonFunc(status, body)
}

func (c *DefaultHTTPContext) Status(code int) {
	if c == nil || c.statusFunc == nil {
		return
	}
	c.statusFunc(code)
}

func (c *DefaultHTTPContext) RoutePath() string {
	if c == nil || c.routePathFunc == nil {
		return ""
	}
	return c.routePathFunc()
}

func (c *DefaultHTTPContext) ResponseStatus() int {
	if c == nil || c.statusReadFunc == nil {
		return 0
	}
	return c.statusReadFunc()
}

// SetParamFunc 配置 path 参数读取函数。
func (c *DefaultHTTPContext) SetParamFunc(fn func(string) string) {
	if c == nil {
		return
	}
	c.paramFunc = fn
}

// SetQueryFunc 配置 query 读取函数。
func (c *DefaultHTTPContext) SetQueryFunc(fn func(string) string) {
	if c == nil {
		return
	}
	c.queryFunc = fn
}

// SetDefaultQueryFunc 配置带默认值的 query 读取函数。
func (c *DefaultHTTPContext) SetDefaultQueryFunc(fn func(string, string) string) {
	if c == nil {
		return
	}
	c.defaultQueryFn = fn
}

// SetHeaderFuncs 配置 header 读取与写入函数。
func (c *DefaultHTTPContext) SetHeaderFuncs(get func(string) string, set func(string, string)) {
	if c == nil {
		return
	}
	c.headerFunc = get
	c.setHeaderFunc = set
}

// SetBindFuncs 配置绑定函数。
func (c *DefaultHTTPContext) SetBindFuncs(bindJSON func(any) error, bindQuery func(any) error, bind func(any) error) {
	if c == nil {
		return
	}
	c.bindJSONFunc = bindJSON
	c.bindQueryFunc = bindQuery
	c.bindFunc = bind
}

// SetResponseFuncs 配置响应函数。
func (c *DefaultHTTPContext) SetResponseFuncs(json func(int, any), status func(int), statusRead func() int) {
	if c == nil {
		return
	}
	c.jsonFunc = json
	c.statusFunc = status
	c.statusReadFunc = statusRead
}

// SetRoutePathFunc 配置路由路径读取函数。
func (c *DefaultHTTPContext) SetRoutePathFunc(fn func() string) {
	if c == nil {
		return
	}
	c.routePathFunc = fn
}
