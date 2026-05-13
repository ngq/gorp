// Package gin provides Gin-based HTTP server implementation for gorp framework.
// This file implements HTTP context adaptation wrapping Gin Context.
// Provides request/response primitives, bind helpers, and header access.
//
// Gin HTTP 服务包，提供基于 Gin 的 HTTP 服务器实现，用于 gorp 框架。
// 本文件实现 HTTP context 适配，包装 Gin Context。
// 提供请求/响应原语、绑定助手和头部访问。
package gin

import (
	"net/http"

	"github.com/gin-gonic/gin"
	supportcontract "github.com/ngq/gorp/framework/contract/support"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

// ginHTTPContext is the Gin-backed implementation of the framework HTTP context contract.
//
// ginHTTPContext 是基于 Gin 的框架 HTTPContext 实现。
type ginHTTPContext struct {
	*transportcontract.DefaultHTTPContext
	gin *gin.Context
}

// GinContext exposes the underlying Gin context.
//
// GinContext 暴露底层 Gin context。
func (c *ginHTTPContext) GinContext() *gin.Context {
	if c == nil {
		return nil
	}
	return c.gin
}

// newHTTPContext adapts a Gin context into the transport HTTP context contract.
//
// newHTTPContext 将 Gin context 适配为 transport HTTPContext 契约。
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
	// 设置中间件相关函数
	base.SetMiddlewareFuncs(
		func(key string) any {
			if ctx == nil {
				return nil
			}
			val, _ := ctx.Get(key)
			return val
		},
		func(key string, value any) {
			if ctx == nil {
				return
			}
			ctx.Set(key, value)
		},
		func(status int) {
			if ctx == nil {
				return
			}
			ctx.AbortWithStatus(status)
		},
		func(status int, body any) {
			if ctx == nil {
				return
			}
			ctx.AbortWithStatusJSON(status, body)
		},
		func() bool {
			if ctx == nil {
				return false
			}
			return ctx.IsAborted()
		},
		func() {
			if ctx == nil {
				return
			}
			ctx.Next()
		},
	)
	return &ginHTTPContext{DefaultHTTPContext: base, gin: ctx}
}

// unwrapGinContext extracts the raw Gin context from a transport HTTP context.
//
// unwrapGinContext 从 transport HTTPContext 中提取原始 Gin context。
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

// UnwrapGinContext extracts the raw Gin context from a transport HTTP context.
// This is the public API for users who need to access Gin-specific features.
//
// UnwrapGinContext 从 transport HTTPContext 中提取原始 Gin context。
// 这是公开 API，供需要访问 Gin 特有功能的用户使用。
//
// 使用示例：
//
//	func handler(c gorp.HTTPContext) {
//	    gc, ok := gin.UnwrapGinContext(c)
//	    if ok {
//	        // 使用原生 Gin context
//	        gc.HTML(200, "index.html", gin.H{})
//	    }
//	}
func UnwrapGinContext(c transportcontract.HTTPContext) (*gin.Context, bool) {
	return unwrapGinContext(c)
}

// writeGinResponseHeaders syncs request identity headers from request context into the Gin response.
//
// writeGinResponseHeaders 将请求上下文中的请求标识头同步到 Gin 响应。
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

// copyHeaders replaces the destination headers with headers copied from the source.
//
// copyHeaders 使用源头部内容覆盖目标头部集合。
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
