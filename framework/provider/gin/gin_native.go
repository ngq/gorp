// Package gin provides Gin-based HTTP server implementation for gorp framework.
// This file provides helper functions for Gin-first users to access native Gin capabilities.
// These functions allow users to use *gin.Engine and *gin.RouterGroup directly
// while remaining within the framework's governance boundary.
//
// Gin HTTP 服务包，提供基于 Gin 的 HTTP 服务器实现，用于 gorp 框架。
// 本文件为 Gin-first 用户提供访问原生 Gin 能力的辅助函数。
// 这些函数允许用户直接使用 *gin.Engine 和 *gin.RouterGroup，
// 同时保持在框架的治理边界内。
package gin

import (
	"github.com/gin-gonic/gin"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

// NativeEngine extracts the underlying *gin.Engine from an HTTP service.
// Returns the engine and true if the HTTP implementation is Gin-backed.
// Returns nil and false otherwise.
//
// NativeEngine 从 HTTP 服务中提取底层 *gin.Engine。
// 如果 HTTP 实现基于 Gin，返回 engine 和 true。
// 否则返回 nil 和 false。
//
// Example:
//
//	httpSvc := c.MustMake(transportcontract.HTTPKey).(transportcontract.HTTP)
//	engine, ok := gin.NativeEngine(httpSvc)
//	if ok {
//	    engine.Use(gin.AdaptMiddleware(httpmiddleware.RequestIdentity()))
//	}
func NativeEngine(httpSvc transportcontract.HTTP) (*gin.Engine, bool) {
	provider, ok := httpSvc.(transportcontract.GINEngineProvider)
	if !ok {
		return nil, false
	}
	engineAny := provider.GINEngine()
	if engineAny == nil {
		return nil, false
	}
	engine, ok := engineAny.(*gin.Engine)
	return engine, ok
}

// NativeRouterGroup extracts the root *gin.RouterGroup from an HTTP service.
// Returns the router group and true if the HTTP implementation is Gin-backed.
// Returns nil and false otherwise.
//
// NativeRouterGroup 从 HTTP 服务中提取根 *gin.RouterGroup。
// 如果 HTTP 实现基于 Gin，返回 routerGroup 和 true。
// 否则返回 nil 和 false。
func NativeRouterGroup(httpSvc transportcontract.HTTP) (*gin.RouterGroup, bool) {
	engine, ok := NativeEngine(httpSvc)
	if !ok || engine == nil {
		return nil, false
	}
	return &engine.RouterGroup, true
}

// MustEngine directly extracts *gin.Engine from the container without requiring
// the user to first obtain the HTTP service.
//
// MustEngine 直接从容器提取 *gin.Engine，无需用户先获取 HTTP 服务。
//
// Example:
//
//	engine, ok := gin.MustEngine(c)
//	if ok {
//	    engine.Use(gin.AdaptMiddleware(httpmiddleware.RequestIdentity()))
//	}
func MustEngine(c runtimecontract.Container) (*gin.Engine, bool) {
	httpSvc, err := c.Make(transportcontract.HTTPKey)
	if err != nil {
		return nil, false
	}
	return NativeEngine(httpSvc.(transportcontract.HTTP))
}
