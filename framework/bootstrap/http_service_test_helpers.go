// Package bootstrap_test provides integration and boundary tests for HTTP service runtime bootstrapping.
//
// 适用场景：
// - 验证 HTTP Service runtime 的初始化、governance override 和 provider 注册行为。
// - 验证 pprof / governance inspect 端点的路由注册与响应格式。
// - 验证 governance summary 与 diagnostic 的构建与格式化逻辑。
package bootstrap

import (
	"context"
	"mime/multipart"
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

func (r *recordingRouter) Use(middleware ...transportcontract.Middleware) {}
func (r *recordingRouter) Group(prefix string, middleware ...transportcontract.Middleware) transportcontract.Router {
	return r
}
func (r *recordingRouter) Handle(method, path string, handler transportcontract.Handler, middleware ...transportcontract.Middleware) {
}
func (r *recordingRouter) HandleFunc(method, path string, handlerFunc transportcontract.Handler, middleware ...transportcontract.Middleware) {
}
func (r *recordingRouter) GET(path string, handler transportcontract.Handler, middleware ...transportcontract.Middleware) {
	r.gets = append(r.gets, path)
}
func (r *recordingRouter) POST(path string, handler transportcontract.Handler, middleware ...transportcontract.Middleware) {
}
func (r *recordingRouter) PUT(path string, handler transportcontract.Handler, middleware ...transportcontract.Middleware) {
}
func (r *recordingRouter) DELETE(path string, handler transportcontract.Handler, middleware ...transportcontract.Middleware) {
}
func (r *recordingRouter) PATCH(path string, handler transportcontract.Handler, middleware ...transportcontract.Middleware) {
}
func (r *recordingRouter) HEAD(path string, handler transportcontract.Handler, middleware ...transportcontract.Middleware) {
}
func (r *recordingRouter) OPTIONS(path string, handler transportcontract.Handler, middleware ...transportcontract.Middleware) {
}
func (r *recordingRouter) ANY(path string, handler transportcontract.Handler, middleware ...transportcontract.Middleware) {
}
func (r *recordingRouter) Static(relativePath, root string)                 {}
func (r *recordingRouter) StaticFile(relativePath, filePath string)         {}
func (r *recordingRouter) StaticFS(relativePath string, fs http.FileSystem) {}
func (r *recordingRouter) NoRoute(handler transportcontract.Handler)        {}
func (r *recordingRouter) NoMethod(handler transportcontract.Handler)       {}
func (r *recordingRouter) Routes() []transportcontract.RouteInfo            { return nil }
func (r *recordingRouter) Mount(path string, handler http.Handler) {
	r.mounted = append(r.mounted, path)
}

// testContext implements Context by delegating to gin.Context.
type testContext struct {
	gin *gin.Context
}

var (
	_ transportcontract.RequestContext    = (*testContext)(nil)
	_ transportcontract.BindingContext    = (*testContext)(nil)
	_ transportcontract.ResponseContext   = (*testContext)(nil)
	_ transportcontract.MiddlewareContext = (*testContext)(nil)
	_ transportcontract.RouteContext      = (*testContext)(nil)
	_ transportcontract.Context           = (*testContext)(nil)
)

func (c *testContext) Context() context.Context {
	return c.gin.Request.Context()
}

func (c *testContext) Request() *http.Request {
	return c.gin.Request
}

func (c *testContext) Response() http.ResponseWriter {
	return c.gin.Writer
}

func (c *testContext) Param(key string) string {
	return c.gin.Param(key)
}

func (c *testContext) Query(key string) string {
	return c.gin.Query(key)
}

func (c *testContext) DefaultQuery(key, defaultValue string) string {
	return c.gin.DefaultQuery(key, defaultValue)
}

func (c *testContext) GetHeader(key string) string {
	return c.gin.GetHeader(key)
}

func (c *testContext) SetHeader(key, value string) {
	c.gin.Header(key, value)
}
func (c *testContext) Cookie(name string) (string, error) { return c.gin.Cookie(name) }

func (c *testContext) Bind(obj any) error {
	return c.gin.ShouldBind(obj)
}

func (c *testContext) BindJSON(obj any) error {
	return c.gin.ShouldBindJSON(obj)
}

func (c *testContext) BindQuery(obj any) error {
	return c.gin.ShouldBindQuery(obj)
}
func (c *testContext) BindURI(obj any) error    { return c.gin.ShouldBindUri(obj) }
func (c *testContext) BindHeader(obj any) error { return c.gin.ShouldBindHeader(obj) }
func (c *testContext) BindForm(obj any) error   { return c.gin.ShouldBind(obj) }

func (c *testContext) JSON(status int, body any) {
	c.gin.JSON(status, body)
}

func (c *testContext) String(status int, body string) {
	c.gin.String(status, body)
}

func (c *testContext) XML(status int, body any) {
	c.gin.XML(status, body)
}

func (c *testContext) Data(status int, contentType string, body []byte) {
	c.gin.Data(status, contentType, body)
}

func (c *testContext) Redirect(status int, location string) {
	c.gin.Redirect(status, location)
}

func (c *testContext) Status(code int) {
	c.gin.Status(code)
}
func (c *testContext) File(path string)                     { c.gin.File(path) }
func (c *testContext) FileAttachment(path, filename string) { c.gin.FileAttachment(path, filename) }
func (c *testContext) SetCookie(cookie *http.Cookie) {
	if cookie != nil {
		http.SetCookie(c.gin.Writer, cookie)
	}
}
func (c *testContext) DeleteCookie(name, path, domain string) {}
func (c *testContext) Stream(contentType string, write func(transportcontract.StreamWriter) error) error {
	return nil
}
func (c *testContext) SSE(event transportcontract.SSEEvent) error { return nil }

func (c *testContext) RoutePath() string {
	return c.gin.FullPath()
}

func (c *testContext) ResponseStatus() int {
	return c.gin.Writer.Status()
}

func (c *testContext) Get(key string) (any, bool) {
	return c.gin.Get(key)
}

func (c *testContext) Set(key string, value any) {
	c.gin.Set(key, value)
}

func (c *testContext) Abort(status int) {
	c.gin.AbortWithStatus(status)
}

func (c *testContext) AbortWithJSON(status int, body any) {
	c.gin.AbortWithStatusJSON(status, body)
}

func (c *testContext) IsAborted() bool {
	return c.gin.IsAborted()
}

func (c *testContext) Next() {
	c.gin.Next()
}

func (c *testContext) DefaultIntQuery(key string, defaultValue int) int {
	return defaultValue
}

func (c *testContext) Int64Param(key string) (int64, error) {
	return 0, nil
}

func (c *testContext) FormFile(name string) (multipart.File, *multipart.FileHeader, error) {
	return nil, nil, http.ErrNoCookie
}

func (c *testContext) SaveUploadedFile(file *multipart.FileHeader, dst string) error {
	return nil
}

func newTestContext(c *gin.Context) transportcontract.Context {
	return &testContext{gin: c}
}

type ginTestRouter struct {
	engine *gin.Engine
}

func (r *ginTestRouter) Use(middleware ...transportcontract.Middleware) {
	adapted := make([]gin.HandlerFunc, 0, len(middleware))
	for _, mw := range middleware {
		mw := mw
		adapted = append(adapted, func(c *gin.Context) {
			httpCtx := newTestContext(c)
			if wrapped := mw(func(inner transportcontract.Context) {
				if inner != nil {
					c.Next()
				}
			}); wrapped != nil {
				wrapped(httpCtx)
			}
		})
	}
	r.engine.Use(adapted...)
}

func (r *ginTestRouter) Group(prefix string, middleware ...transportcontract.Middleware) transportcontract.Router {
	group := r.engine.Group(prefix)
	wrapped := &ginGroupTestRouter{group: group}
	wrapped.Use(middleware...)
	return wrapped
}

func (r *ginTestRouter) Handle(method, path string, handler transportcontract.Handler, middleware ...transportcontract.Middleware) {
	handlers := make([]gin.HandlerFunc, 0, len(middleware)+1)
	for _, mw := range middleware {
		if mw != nil {
			handlers = append(handlers, func(c *gin.Context) {
				httpCtx := newTestContext(c)
				if wrapped := mw(func(inner transportcontract.Context) {
					if inner != nil {
						c.Next()
					}
				}); wrapped != nil {
					wrapped(httpCtx)
				}
			})
		}
	}
	handlers = append(handlers, func(c *gin.Context) {
		httpCtx := newTestContext(c)
		handler(httpCtx)
	})
	r.engine.Handle(method, path, handlers...)
}

func (r *ginTestRouter) HandleFunc(method, path string, handlerFunc transportcontract.Handler, middleware ...transportcontract.Middleware) {
	r.Handle(method, path, handlerFunc, middleware...)
}

func (r *ginTestRouter) GET(path string, handler transportcontract.Handler, middleware ...transportcontract.Middleware) {
	r.Handle(http.MethodGet, path, handler, middleware...)
}

func (r *ginTestRouter) POST(path string, handler transportcontract.Handler, middleware ...transportcontract.Middleware) {
	r.Handle(http.MethodPost, path, handler, middleware...)
}

func (r *ginTestRouter) PUT(path string, handler transportcontract.Handler, middleware ...transportcontract.Middleware) {
	r.Handle(http.MethodPut, path, handler, middleware...)
}

func (r *ginTestRouter) DELETE(path string, handler transportcontract.Handler, middleware ...transportcontract.Middleware) {
	r.Handle(http.MethodDelete, path, handler, middleware...)
}
func (r *ginTestRouter) PATCH(path string, handler transportcontract.Handler, middleware ...transportcontract.Middleware) {
	r.Handle(http.MethodPatch, path, handler, middleware...)
}
func (r *ginTestRouter) HEAD(path string, handler transportcontract.Handler, middleware ...transportcontract.Middleware) {
	r.Handle(http.MethodHead, path, handler, middleware...)
}
func (r *ginTestRouter) OPTIONS(path string, handler transportcontract.Handler, middleware ...transportcontract.Middleware) {
	r.Handle(http.MethodOptions, path, handler, middleware...)
}
func (r *ginTestRouter) ANY(path string, handler transportcontract.Handler, middleware ...transportcontract.Middleware) {
	r.engine.Any(path, func(c *gin.Context) { handler(newTestContext(c)) })
}
func (r *ginTestRouter) Static(relativePath, root string) { r.engine.Static(relativePath, root) }
func (r *ginTestRouter) StaticFile(relativePath, filePath string) {
	r.engine.StaticFile(relativePath, filePath)
}
func (r *ginTestRouter) StaticFS(relativePath string, fs http.FileSystem) {
	r.engine.StaticFS(relativePath, fs)
}
func (r *ginTestRouter) NoRoute(handler transportcontract.Handler) {
	r.engine.NoRoute(func(c *gin.Context) { handler(newTestContext(c)) })
}
func (r *ginTestRouter) NoMethod(handler transportcontract.Handler) {
	r.engine.HandleMethodNotAllowed = true
	r.engine.NoMethod(func(c *gin.Context) { handler(newTestContext(c)) })
}
func (r *ginTestRouter) Routes() []transportcontract.RouteInfo { return nil }

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

func (r *ginGroupTestRouter) Use(middleware ...transportcontract.Middleware) {
	adapted := make([]gin.HandlerFunc, 0, len(middleware))
	for _, mw := range middleware {
		mw := mw
		adapted = append(adapted, func(c *gin.Context) {
			httpCtx := newTestContext(c)
			if wrapped := mw(func(inner transportcontract.Context) {
				if inner != nil {
					c.Next()
				}
			}); wrapped != nil {
				wrapped(httpCtx)
			}
		})
	}
	r.group.Use(adapted...)
}

func (r *ginGroupTestRouter) Group(prefix string, middleware ...transportcontract.Middleware) transportcontract.Router {
	group := &ginGroupTestRouter{group: r.group.Group(prefix)}
	group.Use(middleware...)
	return group
}

func (r *ginGroupTestRouter) Handle(method, path string, handler transportcontract.Handler, middleware ...transportcontract.Middleware) {
	handlers := make([]gin.HandlerFunc, 0, len(middleware)+1)
	for _, mw := range middleware {
		if mw != nil {
			handlers = append(handlers, func(c *gin.Context) {
				httpCtx := newTestContext(c)
				if wrapped := mw(func(inner transportcontract.Context) {
					if inner != nil {
						c.Next()
					}
				}); wrapped != nil {
					wrapped(httpCtx)
				}
			})
		}
	}
	handlers = append(handlers, func(c *gin.Context) {
		httpCtx := newTestContext(c)
		handler(httpCtx)
	})
	r.group.Handle(method, path, handlers...)
}

func (r *ginGroupTestRouter) HandleFunc(method, path string, handlerFunc transportcontract.Handler, middleware ...transportcontract.Middleware) {
	r.Handle(method, path, handlerFunc, middleware...)
}

func (r *ginGroupTestRouter) GET(path string, handler transportcontract.Handler, middleware ...transportcontract.Middleware) {
	r.Handle(http.MethodGet, path, handler, middleware...)
}

func (r *ginGroupTestRouter) POST(path string, handler transportcontract.Handler, middleware ...transportcontract.Middleware) {
	r.Handle(http.MethodPost, path, handler, middleware...)
}

func (r *ginGroupTestRouter) PUT(path string, handler transportcontract.Handler, middleware ...transportcontract.Middleware) {
	r.Handle(http.MethodPut, path, handler, middleware...)
}

func (r *ginGroupTestRouter) DELETE(path string, handler transportcontract.Handler, middleware ...transportcontract.Middleware) {
	r.Handle(http.MethodDelete, path, handler, middleware...)
}
func (r *ginGroupTestRouter) PATCH(path string, handler transportcontract.Handler, middleware ...transportcontract.Middleware) {
	r.Handle(http.MethodPatch, path, handler, middleware...)
}
func (r *ginGroupTestRouter) HEAD(path string, handler transportcontract.Handler, middleware ...transportcontract.Middleware) {
	r.Handle(http.MethodHead, path, handler, middleware...)
}
func (r *ginGroupTestRouter) OPTIONS(path string, handler transportcontract.Handler, middleware ...transportcontract.Middleware) {
	r.Handle(http.MethodOptions, path, handler, middleware...)
}
func (r *ginGroupTestRouter) ANY(path string, handler transportcontract.Handler, middleware ...transportcontract.Middleware) {
	r.group.Any(path, func(c *gin.Context) { handler(newTestContext(c)) })
}
func (r *ginGroupTestRouter) Static(relativePath, root string) { r.group.Static(relativePath, root) }
func (r *ginGroupTestRouter) StaticFile(relativePath, filePath string) {
	r.group.StaticFile(relativePath, filePath)
}
func (r *ginGroupTestRouter) StaticFS(relativePath string, fs http.FileSystem) {
	r.group.StaticFS(relativePath, fs)
}
func (r *ginGroupTestRouter) NoRoute(handler transportcontract.Handler)  {}
func (r *ginGroupTestRouter) NoMethod(handler transportcontract.Handler) {}
func (r *ginGroupTestRouter) Routes() []transportcontract.RouteInfo      { return nil }

func (r *ginGroupTestRouter) Mount(path string, handler http.Handler) {
	h := func(c *gin.Context) {
		handler.ServeHTTP(c.Writer, c.Request)
	}
	r.group.Handle(http.MethodGet, path, h)
	r.group.Handle(http.MethodHead, path, h)
}
