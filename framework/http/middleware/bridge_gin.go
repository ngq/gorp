// Application scenarios:
// - Bridge the transport HTTP contract onto Gin without leaking Gin details into mainline middleware.
// - Keep request context, request object, and response helpers synchronized between transport and Gin.
// - Support unified middleware behavior while still allowing Gin-specific escape hatches.
//
// 适用场景：
// - 在不把 Gin 细节泄漏到主线中间件的前提下，将 transport HTTP 契约桥接到 Gin。
// - 在 transport 与 Gin 之间同步 request context、request 对象和响应助手。
// - 在保持统一中间件行为的同时，保留必要的 Gin 专属下探能力。
package middleware

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	supportcontract "github.com/ngq/gorp/framework/contract/support"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

type ginHTTPContext struct {
	*transportcontract.DefaultHTTPContext
	gin *gin.Context
}

// SetContext updates both the transport context and the underlying Gin request context.
//
// SetContext 同步更新 transport context 与底层 Gin request context。
func (c *ginHTTPContext) SetContext(ctx context.Context) {
	if c == nil {
		return
	}
	c.DefaultHTTPContext.SetContext(ctx)
	if c.gin != nil && c.gin.Request != nil && ctx != nil {
		c.gin.Request = c.gin.Request.WithContext(ctx)
		c.DefaultHTTPContext.SetRequest(c.gin.Request)
	}
}

// SetRequest updates both the transport request and the underlying Gin request pointer.
//
// SetRequest 同步更新 transport request 与底层 Gin request 指针。
func (c *ginHTTPContext) SetRequest(req *http.Request) {
	if c == nil {
		return
	}
	c.DefaultHTTPContext.SetRequest(req)
	if c.gin != nil && req != nil {
		c.gin.Request = req
	}
}

// GinContext exposes the underlying Gin context when the middleware is running on Gin.
//
// GinContext 在中间件运行于 Gin 时暴露底层 Gin 上下文。
func (c *ginHTTPContext) GinContext() *gin.Context {
	if c == nil {
		return nil
	}
	return c.gin
}

func newHTTPContext(ctx *gin.Context) transportcontract.HTTPContext {
	base := transportcontract.NewDefaultHTTPContext(nil, nil)
	if ctx != nil {
		base.SetRequest(ctx.Request)
	}
	base.SetParamFunc(func(key string) string {
		if ctx == nil {
			return ""
		}
		return ctx.Param(key)
	})
	base.SetQueryFunc(func(key string) string {
		if ctx == nil {
			return ""
		}
		return ctx.Query(key)
	})
	base.SetDefaultQueryFunc(func(key, defaultValue string) string {
		if ctx == nil {
			return defaultValue
		}
		return ctx.DefaultQuery(key, defaultValue)
	})
	base.SetHeaderFuncs(func(key string) string {
		if ctx == nil {
			return ""
		}
		return ctx.GetHeader(key)
	}, func(key, value string) {
		if ctx == nil {
			return
		}
		ctx.Header(key, value)
	})
	base.SetBindFuncs(func(obj any) error {
		if ctx == nil {
			return nil
		}
		return ctx.ShouldBindJSON(obj)
	}, func(obj any) error {
		if ctx == nil {
			return nil
		}
		return ctx.ShouldBindQuery(obj)
	}, func(obj any) error {
		if ctx == nil {
			return nil
		}
		return ctx.ShouldBind(obj)
	})
	base.SetResponseFuncs(func(status int, body any) {
		if ctx == nil {
			return
		}
		ctx.JSON(status, body)
	}, func(status int, body string) {
		if ctx == nil {
			return
		}
		ctx.String(status, body)
	}, func(status int, body any) {
		if ctx == nil {
			return
		}
		ctx.XML(status, body)
	}, func(status int, contentType string, body []byte) {
		if ctx == nil {
			return
		}
		ctx.Data(status, contentType, body)
	}, func(status int, location string) {
		if ctx == nil {
			return
		}
		ctx.Redirect(status, location)
	}, func(code int) {
		if ctx == nil {
			return
		}
		ctx.Status(code)
	}, func() int {
		if ctx == nil {
			return 0
		}
		return ctx.Writer.Status()
	})
	base.SetRoutePathFunc(func() string {
		if ctx == nil {
			return ""
		}
		return ctx.FullPath()
	})
	return &ginHTTPContext{DefaultHTTPContext: base, gin: ctx}
}

func unwrapGinContext(c transportcontract.HTTPContext) (*gin.Context, bool) {
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
