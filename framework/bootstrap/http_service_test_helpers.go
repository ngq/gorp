// Package bootstrap_test provides integration and boundary tests for HTTP service runtime bootstrapping.
//
// 适用场景：
// - 验证 HTTP Service runtime 的初始化、governance override 和 provider 注册行为。
// - 验证 pprof / governance inspect 端点的路由注册与响应格式。
// - 验证 governance summary 与 diagnostic 的构建与格式化逻辑。
package bootstrap

import (
	"net/http"

	"github.com/gin-gonic/gin"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

// =============================================================================
// 测试辅助函数
// =============================================================================

type recordingRouter struct {
	mounted []string
	gets    []string
}

func (r *recordingRouter) Use(middleware ...transportcontract.HTTPMiddleware) {}
func (r *recordingRouter) Group(prefix string, middleware ...transportcontract.HTTPMiddleware) transportcontract.HTTPRouter {
	return r
}
func (r *recordingRouter) Handle(method, path string, handler transportcontract.HTTPHandler) {}
func (r *recordingRouter) HandleFunc(method, path string, handlerFunc transportcontract.HTTPHandler) {
}
func (r *recordingRouter) GET(path string, handler transportcontract.HTTPHandler) {
	r.gets = append(r.gets, path)
}
func (r *recordingRouter) POST(path string, handler transportcontract.HTTPHandler)   {}
func (r *recordingRouter) PUT(path string, handler transportcontract.HTTPHandler)    {}
func (r *recordingRouter) DELETE(path string, handler transportcontract.HTTPHandler) {}
func (r *recordingRouter) Mount(path string, handler http.Handler) {
	r.mounted = append(r.mounted, path)
}

type ginTestRouter struct {
	engine *gin.Engine
}

func (r *ginTestRouter) Use(middleware ...transportcontract.HTTPMiddleware) {
	adapted := make([]gin.HandlerFunc, 0, len(middleware))
	for _, mw := range middleware {
		mw := mw
		adapted = append(adapted, func(c *gin.Context) {
			httpCtx := transportcontract.NewDefaultHTTPContext(c.Request.Context(), c.Request)
			httpCtx.SetParamFunc(c.Param)
			httpCtx.SetQueryFunc(c.Query)
			httpCtx.SetDefaultQueryFunc(c.DefaultQuery)
			httpCtx.SetHeaderFuncs(c.GetHeader, c.Header)
			httpCtx.SetBindFuncs(c.ShouldBindJSON, c.ShouldBindQuery, c.ShouldBind)
			httpCtx.SetResponseFuncs(c.JSON, func(code int, body string) { c.String(code, body) }, c.XML, c.Data, c.Redirect, c.Status, func() int { return c.Writer.Status() })
			httpCtx.SetRoutePathFunc(c.FullPath)
			if wrapped := mw(func(inner transportcontract.HTTPContext) {
				if inner != nil && inner.Request() != nil {
					c.Request = inner.Request()
				}
				c.Next()
			}); wrapped != nil {
				wrapped(httpCtx)
			}
		})
	}
	r.engine.Use(adapted...)
}
func (r *ginTestRouter) Group(prefix string, middleware ...transportcontract.HTTPMiddleware) transportcontract.HTTPRouter {
	group := r.engine.Group(prefix)
	wrapped := &ginGroupTestRouter{group: group}
	wrapped.Use(middleware...)
	return wrapped
}
func (r *ginTestRouter) Handle(method, path string, handler transportcontract.HTTPHandler) {
	r.engine.Handle(method, path, func(c *gin.Context) {
		httpCtx := transportcontract.NewDefaultHTTPContext(c.Request.Context(), c.Request)
		httpCtx.SetParamFunc(c.Param)
		httpCtx.SetQueryFunc(c.Query)
		httpCtx.SetDefaultQueryFunc(c.DefaultQuery)
		httpCtx.SetHeaderFuncs(c.GetHeader, c.Header)
		httpCtx.SetBindFuncs(c.ShouldBindJSON, c.ShouldBindQuery, c.ShouldBind)
		httpCtx.SetResponseFuncs(c.JSON, func(code int, body string) { c.String(code, body) }, c.XML, c.Data, c.Redirect, c.Status, func() int { return c.Writer.Status() })
		httpCtx.SetRoutePathFunc(c.FullPath)
		handler(httpCtx)
	})
}
func (r *ginTestRouter) HandleFunc(method, path string, handlerFunc transportcontract.HTTPHandler) {
	r.Handle(method, path, handlerFunc)
}
func (r *ginTestRouter) GET(path string, handler transportcontract.HTTPHandler) {
	r.Handle(http.MethodGet, path, handler)
}
func (r *ginTestRouter) POST(path string, handler transportcontract.HTTPHandler) {
	r.Handle(http.MethodPost, path, handler)
}
func (r *ginTestRouter) PUT(path string, handler transportcontract.HTTPHandler) {
	r.Handle(http.MethodPut, path, handler)
}
func (r *ginTestRouter) DELETE(path string, handler transportcontract.HTTPHandler) {
	r.Handle(http.MethodDelete, path, handler)
}
func (r *ginTestRouter) Mount(path string, handler http.Handler) {
	h := func(c *gin.Context) {
		handler.ServeHTTP(c.Writer, c.Request)
	}
	r.engine.Handle(http.MethodGet, path, h)
	r.engine.Handle(http.MethodHead, path, h)
}

type ginGroupTestRouter struct {
	group *gin.RouterGroup
}

func (r *ginGroupTestRouter) Use(middleware ...transportcontract.HTTPMiddleware) {
	adapted := make([]gin.HandlerFunc, 0, len(middleware))
	for _, mw := range middleware {
		mw := mw
		adapted = append(adapted, func(c *gin.Context) {
			httpCtx := transportcontract.NewDefaultHTTPContext(c.Request.Context(), c.Request)
			httpCtx.SetParamFunc(c.Param)
			httpCtx.SetQueryFunc(c.Query)
			httpCtx.SetDefaultQueryFunc(c.DefaultQuery)
			httpCtx.SetHeaderFuncs(c.GetHeader, c.Header)
			httpCtx.SetBindFuncs(c.ShouldBindJSON, c.ShouldBindQuery, c.ShouldBind)
			httpCtx.SetResponseFuncs(c.JSON, func(code int, body string) { c.String(code, body) }, c.XML, c.Data, c.Redirect, c.Status, func() int { return c.Writer.Status() })
			httpCtx.SetRoutePathFunc(c.FullPath)
			if wrapped := mw(func(inner transportcontract.HTTPContext) {
				if inner != nil && inner.Request() != nil {
					c.Request = inner.Request()
				}
				c.Next()
			}); wrapped != nil {
				wrapped(httpCtx)
			}
		})
	}
	r.group.Use(adapted...)
}
func (r *ginGroupTestRouter) Group(prefix string, middleware ...transportcontract.HTTPMiddleware) transportcontract.HTTPRouter {
	group := &ginGroupTestRouter{group: r.group.Group(prefix)}
	group.Use(middleware...)
	return group
}
func (r *ginGroupTestRouter) Handle(method, path string, handler transportcontract.HTTPHandler) {
	r.group.Handle(method, path, func(c *gin.Context) {
		httpCtx := transportcontract.NewDefaultHTTPContext(c.Request.Context(), c.Request)
		httpCtx.SetParamFunc(c.Param)
		httpCtx.SetQueryFunc(c.Query)
		httpCtx.SetDefaultQueryFunc(c.DefaultQuery)
		httpCtx.SetHeaderFuncs(c.GetHeader, c.Header)
		httpCtx.SetBindFuncs(c.ShouldBindJSON, c.ShouldBindQuery, c.ShouldBind)
		httpCtx.SetResponseFuncs(c.JSON, func(code int, body string) { c.String(code, body) }, c.XML, c.Data, c.Redirect, c.Status, func() int { return c.Writer.Status() })
		httpCtx.SetRoutePathFunc(c.FullPath)
		handler(httpCtx)
	})
}
func (r *ginGroupTestRouter) HandleFunc(method, path string, handlerFunc transportcontract.HTTPHandler) {
	r.Handle(method, path, handlerFunc)
}
func (r *ginGroupTestRouter) GET(path string, handler transportcontract.HTTPHandler) {
	r.group.Handle(http.MethodGet, path, func(c *gin.Context) {
		httpCtx := transportcontract.NewDefaultHTTPContext(c.Request.Context(), c.Request)
		httpCtx.SetParamFunc(c.Param)
		httpCtx.SetQueryFunc(c.Query)
		httpCtx.SetDefaultQueryFunc(c.DefaultQuery)
		httpCtx.SetHeaderFuncs(c.GetHeader, c.Header)
		httpCtx.SetBindFuncs(c.ShouldBindJSON, c.ShouldBindQuery, c.ShouldBind)
		httpCtx.SetResponseFuncs(c.JSON, func(code int, body string) { c.String(code, body) }, c.XML, c.Data, c.Redirect, c.Status, func() int { return c.Writer.Status() })
		httpCtx.SetRoutePathFunc(c.FullPath)
		handler(httpCtx)
	})
}
func (r *ginGroupTestRouter) POST(path string, handler transportcontract.HTTPHandler) {
	r.Handle(http.MethodPost, path, handler)
}
func (r *ginGroupTestRouter) PUT(path string, handler transportcontract.HTTPHandler) {
	r.Handle(http.MethodPut, path, handler)
}
func (r *ginGroupTestRouter) DELETE(path string, handler transportcontract.HTTPHandler) {
	r.Handle(http.MethodDelete, path, handler)
}
func (r *ginGroupTestRouter) Mount(path string, handler http.Handler) {
	h := func(c *gin.Context) {
		handler.ServeHTTP(c.Writer, c.Request)
	}
	r.group.Handle(http.MethodGet, path, h)
	r.group.Handle(http.MethodHead, path, h)
}
