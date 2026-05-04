package transport

import (
	"context"
	"net/http"
)

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

func (c *DefaultHTTPContext) String(status int, body string) {
	if c == nil || c.stringFunc == nil {
		return
	}
	c.stringFunc(status, body)
}

func (c *DefaultHTTPContext) XML(status int, body any) {
	if c == nil || c.xmlFunc == nil {
		return
	}
	c.xmlFunc(status, body)
}

func (c *DefaultHTTPContext) Data(status int, contentType string, body []byte) {
	if c == nil || c.dataFunc == nil {
		return
	}
	c.dataFunc(status, contentType, body)
}

func (c *DefaultHTTPContext) Redirect(status int, location string) {
	if c == nil || c.redirectFunc == nil {
		return
	}
	c.redirectFunc(status, location)
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

func (c *DefaultHTTPContext) SetParamFunc(fn func(string) string) {
	if c == nil {
		return
	}
	c.paramFunc = fn
}

func (c *DefaultHTTPContext) SetQueryFunc(fn func(string) string) {
	if c == nil {
		return
	}
	c.queryFunc = fn
}

func (c *DefaultHTTPContext) SetDefaultQueryFunc(fn func(string, string) string) {
	if c == nil {
		return
	}
	c.defaultQueryFn = fn
}

func (c *DefaultHTTPContext) SetHeaderFuncs(get func(string) string, set func(string, string)) {
	if c == nil {
		return
	}
	c.headerFunc = get
	c.setHeaderFunc = set
}

func (c *DefaultHTTPContext) SetBindFuncs(bindJSON func(any) error, bindQuery func(any) error, bind func(any) error) {
	if c == nil {
		return
	}
	c.bindJSONFunc = bindJSON
	c.bindQueryFunc = bindQuery
	c.bindFunc = bind
}

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

func (c *DefaultHTTPContext) SetRoutePathFunc(fn func() string) {
	if c == nil {
		return
	}
	c.routePathFunc = fn
}
