// Package gin provides Gin-based HTTP server implementation for gorp framework.
// This file implements the HTTP service interface with Gin engine and net/http server.
// Supports application lifecycle operations: Run, Shutdown, Router, Server access.
//
// Gin HTTP 服务包，提供基于 Gin 的 HTTP 服务器实现，用于 gorp 框架。
// 本文件实现 HTTP 服务接口，包含 Gin engine 和 net/http server。
// 支持应用生命周期操作：Run、Shutdown、Router、Server 访问。
package gin

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

type service struct {
	srv    *http.Server
	engine *gin.Engine
	router transportcontract.HTTPRouter
	log    observabilitycontract.Logger
}

// service is the runtime HTTP service implementation built on top of Gin.
//
// service 是构建在 Gin 之上的运行时 HTTP 服务实现。
//
// Router returns the framework HTTP router facade backed by Gin.
//
// Router 返回由 Gin 驱动的框架 HTTPRouter 门面。
func (s *service) Router() transportcontract.HTTPRouter { return s.router }

// Server returns the underlying net/http server instance.
//
// Server 返回底层 net/http server 实例。
func (s *service) Server() *http.Server { return s.srv }

// Run starts serving HTTP traffic.
//
// Run 启动 HTTP 服务监听。
func (s *service) Run() error { return s.srv.ListenAndServe() }

// Shutdown gracefully stops the HTTP server.
//
// Shutdown 优雅关闭 HTTP 服务。
func (s *service) Shutdown(ctx context.Context) error { return s.srv.Shutdown(ctx) }

// GINEngine returns the underlying *gin.Engine for Gin-first usage.
// Implements transportcontract.GINEngineProvider.
//
// GINEngine 返回底层 *gin.Engine，供 Gin-first 使用。
// 实现 transportcontract.GINEngineProvider。
func (s *service) GINEngine() any { return s.engine }
