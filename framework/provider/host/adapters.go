// Package host provides service adapters for wrapping HTTP, Cron, and GRPC servers.
// These adapters implement runtimecontract.Hostable interface for lifecycle management.
//
// 本文件提供服务适配器，用于封装 HTTP、Cron 和 GRPC 服务器。
// 这些适配器实现 runtimecontract.Hostable 接口，用于生命周期管理。
package host

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"sync/atomic"

	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

// grpcServable 是 gRPC server 的最小运行接口。*grpc.Server 满足该接口，
// 但本包不 import grpc，避免纯 HTTP 应用被强制链接 grpc-go。
//
// grpcServable 是 gRPC server 的最小运行接口。
type grpcServable interface {
	Serve(lis net.Listener) error
	GracefulStop()
	Stop()
}

// HTTPService wraps transportcontract.HTTP as a Hostable service.
//
// HTTPService 封装 transportcontract.HTTP 为可托管服务。
type HTTPService struct {
	name string // name is the service name.
	//
	// name 服务名称。
	http transportcontract.HTTP // http is the HTTP transport.
	//
	// http HTTP 传输层。
	server *http.Server // server is the underlying HTTP server.
	//
	// server 底层 HTTP 服务器。
}

// NewHTTPService creates a new HTTP service adapter.
//
// NewHTTPService 创建新的 HTTP 服务适配器。
func NewHTTPService(name string, h transportcontract.HTTP) *HTTPService {
	return &HTTPService{
		name:   name,
		http:   h,
		server: h.Server(),
	}
}

// Name returns the service name.
//
// Name 返回服务名称。
func (s *HTTPService) Name() string { return s.name }

// Start binds the HTTP listener synchronously and serves in the background.
//
// Start 同步绑定 HTTP 监听地址，然后在后台提供服务。
func (s *HTTPService) Start(ctx context.Context) error {
	return s.http.Start(ctx)
}

// Stop shuts down the HTTP server gracefully.
//
// Stop 优雅关闭 HTTP 服务器。
func (s *HTTPService) Stop(ctx context.Context) error {
	return s.http.Stop(ctx)
}

// CronService wraps runtimecontract.Cron as a Hostable service.
//
// CronService 封装 runtimecontract.Cron 为可托管服务。
type CronService struct {
	name string // name is the service name.
	//
	// name 服务名称。
	cron runtimecontract.Cron // cron is the cron scheduler.
	//
	// cron Cron 调度器。
}

// NewCronService creates a new Cron service adapter.
//
// NewCronService 创建新的 Cron 服务适配器。
func NewCronService(name string, c runtimecontract.Cron) *CronService {
	return &CronService{name: name, cron: c}
}

// Name returns the service name.
//
// Name 返回服务名称。
func (s *CronService) Name() string { return s.name }

// Start starts the cron scheduler.
//
// Start 启动 Cron 调度器。
func (s *CronService) Start(ctx context.Context) error {
	s.cron.Start()
	return nil
}

// Stop stops the cron scheduler and waits for completion.
// Core logic: Call cron.Stop, wait for Done channel or context timeout.
//
// Stop 停止 Cron 调度器并等待完成。
// 核心逻辑：调用 cron.Stop，等待 Done channel 或 context 超时。
func (s *CronService) Stop(ctx context.Context) error {
	stopped := s.cron.Stop()
	select {
	case <-stopped.Done():
		return nil
	case <-ctx.Done():
		return fmt.Errorf("cron stop context done: %w", ctx.Err())
	}
}

// GRPCService wraps a gRPC server (via grpcServable interface) as a Hostable service.
//
// GRPCService 封装 gRPC server（经 grpcServable 接口）为可托管服务。
type GRPCService struct {
	name    string // name 服务名称。
	server  grpcServable
	lis     net.Listener
	stopped atomic.Bool // stopped 记录是否被主动停止，用于区分正常停止与异常
}

// NewGRPCService creates a new GRPC service adapter.
//
// NewGRPCService 创建新的 GRPC 服务适配器。
func NewGRPCService(name string, server grpcServable, lis net.Listener) *GRPCService {
	return &GRPCService{name: name, server: server, lis: lis}
}

// Name returns the service name.
//
// Name 返回服务名称。
func (s *GRPCService) Name() string { return s.name }

// Start starts the GRPC server in background.
// 主动停止（Stop）后 Serve 返回的错误会被忽略，避免依赖 grpc.ErrServerStopped。
//
// Start 在后台启动 GRPC 服务器。
func (s *GRPCService) Start(ctx context.Context) error {
	go func() {
		if err := s.server.Serve(s.lis); err != nil && !s.stopped.Load() {
			slog.Error("grpc server error", "error", err)
		}
	}()
	return nil
}

// Stop gracefully stops the GRPC server, with fallback to force stop on timeout.
//
// Stop 优雅关闭 GRPC 服务器，超时时强制关闭。
func (s *GRPCService) Stop(ctx context.Context) error {
	s.stopped.Store(true)
	done := make(chan struct{})
	go func() {
		s.server.GracefulStop()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		s.server.Stop()
		return fmt.Errorf("grpc stop context done: %w", ctx.Err())
	}
}
