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
	"errors"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

type routeSurface interface {
	transportcontract.Router
}

type service struct {
	routeSurface
	srv             *http.Server
	engine          *gin.Engine
	log             observabilitycontract.Logger
	shutdownTimeout time.Duration
	mu              sync.Mutex
	listener        net.Listener
}

// service is the runtime HTTP service implementation built on top of Gin.
//
// service 是构建在 Gin 之上的运行时 HTTP 服务实现。
//
// Router returns the framework router facade backed by Gin.
//
// Router 返回由 Gin 驱动的框架 Router 门面。
func (s *service) Router() transportcontract.Router { return s.routeSurface }

// Server returns the underlying net/http server instance.
//
// Server 返回底层 net/http server 实例。
func (s *service) Server() *http.Server { return s.srv }

// Run starts serving HTTP traffic.
//
// Run 启动 HTTP 服务监听。
func (s *service) Run() error { return s.srv.ListenAndServe() }

// Start binds the address before returning, allowing lifecycle orchestration
// to observe port conflicts and roll back already-started services.
func (s *service) Start(context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.listener != nil {
		return nil
	}
	listener, err := net.Listen("tcp", s.srv.Addr)
	if err != nil {
		return err
	}
	s.listener = listener
	go func() {
		if err := s.srv.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) && !errors.Is(err, net.ErrClosed) {
			s.log.Error("http server stopped unexpectedly", observabilitycontract.Field{Key: "error", Value: err})
		}
	}()
	return nil
}

// Stop gracefully stops the HTTP server.
func (s *service) Stop(ctx context.Context) error {
	if s.shutdownTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, s.shutdownTimeout)
		defer cancel()
	}
	return s.Shutdown(ctx)
}

// Shutdown gracefully stops the HTTP server.
//
// Shutdown 优雅关闭 HTTP 服务。
func (s *service) Shutdown(ctx context.Context) error {
	s.mu.Lock()
	listener := s.listener
	s.mu.Unlock()
	err := s.srv.Shutdown(ctx)
	// Shutdown closes listeners already tracked by http.Server. Closing the
	// bound listener explicitly also covers the narrow Start/Serve handoff
	// window where Serve has not registered it yet.
	if listener != nil {
		_ = listener.Close()
	}
	s.mu.Lock()
	s.listener = nil
	s.mu.Unlock()
	return err
}

// GINEngine returns the underlying *gin.Engine for Gin-first usage.
// Implements transportcontract.GINEngineProvider.
//
// GINEngine 返回底层 *gin.Engine，供 Gin-first 使用。
// 实现 transportcontract.GINEngineProvider。
func (s *service) GINEngine() any { return s.engine }

// UseGlobal 注册全局级中间件，对所有路由生效。
// 语义等同于 Gin 的 engine.Use()，区别于 Router.Use() 的组级语义。
// 全局中间件在框架治理中间件之后、路由组中间件之前执行。
//
// UseGlobal registers global-level middleware that applies to all routes.
// Semantically equivalent to Gin's engine.Use(), distinct from Router.Use() which is group-level.
// Global middleware executes after framework governance middleware and before group-level middleware.
func (s *service) UseGlobal(middleware ...transportcontract.Middleware) {
	if s.engine == nil || len(middleware) == 0 {
		return
	}
	adapted := make([]gin.HandlerFunc, 0, len(middleware))
	for _, mw := range middleware {
		if mw == nil {
			continue
		}
		adapted = append(adapted, adaptMiddleware(mw))
	}
	if len(adapted) == 0 {
		return
	}
	s.engine.Use(adapted...)
}
