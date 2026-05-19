// Application scenarios:
// - Bridge the transport HTTP contract onto Gin without leaking Gin details into mainline middleware.
// - Keep request context, request object, and response helpers synchronized between transport and Gin.
// - Support unified middleware behavior while still allowing Gin-specific escape hatches.
//
// 适用场景：
// - 在不把 Gin 细节泄漏到主线中间件的前提下，将 transport HTTP 契约桥接到 Gin。
// - 在 transport 与 Gin 之间同步 request context、request 对象和响应助手。
// - 在保持统一中间件行为的同时，保留必要的 Gin 专属下沉能力。
package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	supportcontract "github.com/ngq/gorp/framework/contract/support"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

// ginContext implements Context by delegating to gin.Context.
// This is used by bridge middleware to wrap Gin context.
type ginContext struct {
	gin *gin.Context
}

// Compilation checks: ensure ginContext implements all transport interfaces.
var (
	_ transportcontract.RequestContext    = (*ginContext)(nil)
	_ transportcontract.BindingContext    = (*ginContext)(nil)
	_ transportcontract.ResponseContext   = (*ginContext)(nil)
	_ transportcontract.MiddlewareContext = (*ginContext)(nil)
	_ transportcontract.RouteContext      = (*ginContext)(nil)
	_ transportcontract.Context           = (*ginContext)(nil)
)

// GinContext exposes the underlying Gin context when the middleware is running on Gin.
//
// GinContext 在中间件运行于 Gin 时暴露底层 Gin 上下文。
func (c *ginContext) GinContext() *gin.Context {
	return c.gin
}

// ========== context.Context delegation ==========

func (c *ginContext) Deadline() (deadline time.Time, ok bool) {
	return c.gin.Request.Context().Deadline()
}

func (c *ginContext) Done() <-chan struct{} {
	return c.gin.Request.Context().Done()
}

func (c *ginContext) Err() error {
	return c.gin.Request.Context().Err()
}

func (c *ginContext) Value(key any) any {
	return c.gin.Request.Context().Value(key)
}

// ========== Request/Response ==========

func (c *ginContext) Request() *http.Request {
	return c.gin.Request
}

func (c *ginContext) Response() http.ResponseWriter {
	return c.gin.Writer
}

// ========== Params/Query ==========

func (c *ginContext) Param(key string) string {
	return c.gin.Param(key)
}

func (c *ginContext) Query(key string) string {
	return c.gin.Query(key)
}

func (c *ginContext) DefaultQuery(key, defaultValue string) string {
	return c.gin.DefaultQuery(key, defaultValue)
}

// ========== Headers ==========

func (c *ginContext) GetHeader(key string) string {
	return c.gin.GetHeader(key)
}

func (c *ginContext) SetHeader(key, value string) {
	c.gin.Header(key, value)
}

// ========== Binding ==========

func (c *ginContext) Bind(obj any) error {
	return c.gin.ShouldBind(obj)
}

func (c *ginContext) BindJSON(obj any) error {
	return c.gin.ShouldBindJSON(obj)
}

func (c *ginContext) BindQuery(obj any) error {
	return c.gin.ShouldBindQuery(obj)
}

// ========== Response ==========

func (c *ginContext) JSON(status int, body any) {
	c.gin.JSON(status, body)
}

func (c *ginContext) String(status int, body string) {
	c.gin.String(status, body)
}

func (c *ginContext) XML(status int, body any) {
	c.gin.XML(status, body)
}

func (c *ginContext) Data(status int, contentType string, body []byte) {
	c.gin.Data(status, contentType, body)
}

func (c *ginContext) Redirect(status int, location string) {
	c.gin.Redirect(status, location)
}

func (c *ginContext) Status(code int) {
	c.gin.Status(code)
}

// ========== Route info ==========

func (c *ginContext) RoutePath() string {
	return c.gin.FullPath()
}

func (c *ginContext) ResponseStatus() int {
	return c.gin.Writer.Status()
}

// ========== Middleware support ==========

func (c *ginContext) Get(key string) any {
	val, _ := c.gin.Get(key)
	return val
}

func (c *ginContext) Set(key string, value any) {
	c.gin.Set(key, value)
}

func (c *ginContext) Abort(status int) {
	c.gin.AbortWithStatus(status)
}

func (c *ginContext) AbortWithJSON(status int, body any) {
	c.gin.AbortWithStatusJSON(status, body)
}

func (c *ginContext) IsAborted() bool {
	return c.gin.IsAborted()
}

func (c *ginContext) Next() {
	c.gin.Next()
}

// newContext creates a new gin-backed Context.
func newContext(ctx *gin.Context) transportcontract.Context {
	return &ginContext{gin: ctx}
}

// unwrapGinContext extracts the raw gin.Context from a transport Context.
func unwrapGinContext(c transportcontract.Context) (*gin.Context, bool) {
	type ginContextProvider interface {
		GinContext() *gin.Context
	}
	provider, ok := c.(ginContextProvider)
	if !ok {
		return nil, false
	}
	gc := provider.GinContext()
	if gc == nil {
		return nil, false
	}
	return gc, true
}

// writeGinResponseHeaders syncs request identity headers from request context into Gin response headers.
//
// writeGinResponseHeaders 将请求上下文中的请求标识头同步到 Gin 响应头。
func writeGinResponseHeaders(c *gin.Context) {
	if c == nil || c.Request == nil {
		return
	}
	if requestID, ok := supportcontract.FromRequestIDContext(c.Request.Context()); ok && requestID != "" {
		c.Header("X-Request-Id", requestID)
	}
	if traceID, ok := supportcontract.FromTraceIDContext(c.Request.Context()); ok && traceID != "" {
		c.Header("X-Trace-Id", traceID)
	}
}

// copyHeaders replaces the destination header set with all values from the source header set.
//
// copyHeaders 使用源头部集合完整替换目标头部集合。
func copyHeaders(dst, src http.Header) {
	for k := range dst {
		dst.Del(k)
	}
	for k, values := range src {
		for _, value := range values {
			dst.Add(k, value)
		}
	}
}
