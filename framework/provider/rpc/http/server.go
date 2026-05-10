// Package http provides HTTP RPC server for gorp framework.
// This file implements the RPCServer contract with HTTP transport,
// registering RPC handlers as HTTP POST routes.
//
// 本包提供 HTTP RPC 服务端，用于 gorp 框架。
// 本文件实现带 HTTP 传输的 RPCServer 契约，
// 将 RPC handler 注册为 HTTP POST 路由。
package http

import (
	"context"
	"sync"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

// Server implements transportcontract.RPCServer using HTTP.
// It registers RPC handlers as HTTP POST routes on the existing HTTP server.
// Does not create its own HTTP server, delegates to HTTPKey binding.
//
// Server 使用 HTTP 实现 transportcontract.RPCServer。
// 在现有 HTTP 服务器上将 RPC handler 注册为 HTTP POST 路由。
// 不创建自己的 HTTP 服务器，委托给 HTTPKey binding。
type Server struct {
	cfg    *transportcontract.RPCConfig
	c      runtimecontract.Container
	addr   string
	routes sync.Map
}

// NewServer creates a new HTTP Server instance with container for HTTP router access.
//
// NewServer 创建新的 HTTP Server 实例，使用容器访问 HTTP router。
func NewServer(cfg *transportcontract.RPCConfig, c runtimecontract.Container) *Server {
	return &Server{
		cfg: cfg,
		c:   c,
	}
}

// Register stores a service handler for later registration as HTTP route.
// Implements transportcontract.RPCServer.Register.
// Handler must implement transportcontract.HTTPHandler.
//
// Register 存储服务 handler，供后续注册为 HTTP 路由。
// 实现 transportcontract.RPCServer.Register。
// Handler 必须实现 transportcontract.HTTPHandler。
func (s *Server) Register(service string, handler any) error {
	s.routes.Store(service, handler)
	return nil
}

// Start registers all stored RPC handlers as HTTP POST routes on the HTTP server.
// Implements transportcontract.RPCServer.Start.
// Retrieves HTTP server from container and registers routes under "/rpc/{service}".
// Does not start the HTTP server itself (delegated to HTTP provider).
//
// Start 将所有存储的 RPC handler 注册为 HTTP POST 路由。
// 实现 transportcontract.RPCServer.Start。
// 从容器获取 HTTP server，在 "/rpc/{service}" 下注册路由。
// 不启动 HTTP server 本身（委托给 HTTP provider）。
func (s *Server) Start(ctx context.Context) error {
	httpSvc, err := s.c.Make(transportcontract.HTTPKey)
	if err != nil {
		return nil
	}
	httpServer, ok := httpSvc.(transportcontract.HTTP)
	if !ok || httpServer == nil {
		return nil
	}
	router := httpServer.Router()
	if router == nil {
		return nil
	}
	s.routes.Range(func(key, value any) bool {
		service := key.(string)
		handler, ok := value.(transportcontract.HTTPHandler)
		if !ok || handler == nil {
			return true
		}
		router.POST("/rpc/"+service, handler)
		return true
	})
	return nil
}

// Stop is a no-op for HTTP RPC server.
// Implements transportcontract.RPCServer.Stop.
// HTTP server lifecycle is managed by HTTP provider.
//
// Stop 对 HTTP RPC server 是空操作。
// 实现 transportcontract.RPCServer.Stop。
// HTTP server 生命周期由 HTTP provider 管理。
func (s *Server) Stop(ctx context.Context) error {
	return nil
}

// Addr returns the server's address from config or default ":8080".
// Implements transportcontract.RPCServer.Addr.
//
// Addr 从配置返回服务器地址或默认 ":8080"。
// 实现 transportcontract.RPCServer.Addr。
func (s *Server) Addr() string {
	if s.addr != "" {
		return s.addr
	}
	if s.c.IsBind(datacontract.ConfigKey) {
		cfgAny, _ := s.c.Make(datacontract.ConfigKey)
		if cfg, ok := cfgAny.(datacontract.Config); ok {
			return cfg.GetString("app.address")
		}
	}
	return ":8080"
}