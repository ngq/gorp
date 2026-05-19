// Package gin provides Gin-based HTTP server implementation for gorp framework.
// This file implements Context interface by delegating to gin.Context.
//
// Gin HTTP 服务包，提供基于 Gin 的 HTTP 服务器实现，用于 gorp 框架。
// 本文件通过委托给 gin.Context 实现 Context 接口。
package gin

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

// ginContext implements Context by delegating to gin.Context.
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

// newContext creates a new gin-backed Context.
func newContext(ctx *gin.Context) transportcontract.Context {
	return &ginContext{gin: ctx}
}

// GinContext exposes the underlying gin.Context.
// This is useful when you need to access Gin-specific features.
//
// GinContext 暴露底层 gin.Context。
// 当需要访问 Gin 特有功能时使用。
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

// UnwrapContext extracts the raw gin.Context from a transport Context.
// This is the public API for users who need to access Gin-specific features.
//
// UnwrapContext 从 transport Context 中提取原始 gin.Context。
// 这是公开 API，供需要访问 Gin 特有功能的用户使用。
//
// Example:
//
//	func handler(c gorp.Context) {
//	    gc, ok := gin.UnwrapContext(c)
//	    if ok {
//	        // Use native Gin context
//	        gc.HTML(200, "index.html", gin.H{})
//	    }
//	}
func UnwrapContext(c transportcontract.Context) (*gin.Context, bool) {
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
