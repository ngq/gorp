// Package host provides service adapters for wrapping HTTP, Cron, and GRPC servers.
// These adapters implement runtimecontract.Hostable interface for lifecycle management.
//
// 本文件提供服务适配器，用于封装 HTTP、Cron 和 GRPC 服务器。
// 这些适配器实现 runtimecontract.Hostable 接口，用于生命周期管理。
package host

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"

	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	"google.golang.org/grpc"
)

// HTTPService wraps transportcontract.HTTP as a Hostable service.
//
// HTTPService 封装 transportcontract.HTTP 为可托管服务。
type HTTPService struct {
	name   string                 // name is the service name.
	                              //
	                               // name 服务名称。
	http   transportcontract.HTTP // http is the HTTP transport.
	                              //
	                               // http HTTP 传输层。
	server *http.Server           // server is the underlying HTTP server.
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

// Start starts the HTTP server in background.
// 非 ErrServerClosed 的错误会通过 slog.Error 记录。
//
// Start 在后台启动 HTTP 服务器。
// 非 ErrServerClosed 的错误会通过 slog.Error 记录。
func (s *HTTPService) Start(ctx context.Context) error {
	go func() {
		if err := s.http.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("http server error", "error", err)
		}
	}()
	return nil
}

// Stop shuts down the HTTP server gracefully.
//
// Stop 优雅关闭 HTTP 服务器。
func (s *HTTPService) Stop(ctx context.Context) error {
	return s.http.Shutdown(ctx)
}

// CronService wraps runtimecontract.Cron as a Hostable service.
//
// CronService 封装 runtimecontract.Cron 为可托管服务。
type CronService struct {
	name string            // name is the service name.
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

// GRPCService wraps grpc.Server as a Hostable service.
//
// GRPCService 封装 grpc.Server 为可托管服务。
type GRPCService struct {
	name   string       // name is the service name.
	                    //
	                     // name 服务名称。
	server *grpc.Server // server is the GRPC server.
	                    //
	                     // server GRPC 服务器。
	lis    net.Listener // lis is the network listener.
	                    //
	                     // lis 网络监听器。
}

// NewGRPCService creates a new GRPC service adapter.
//
// NewGRPCService 创建新的 GRPC 服务适配器。
func NewGRPCService(name string, server *grpc.Server, lis net.Listener) *GRPCService {
	return &GRPCService{name: name, server: server, lis: lis}
}

// Name returns the service name.
//
// Name 返回服务名称。
func (s *GRPCService) Name() string { return s.name }

// Start starts the GRPC server in background.
// 非 grpc.ErrServerStopped 的错误会通过 slog.Error 记录。
//
// Start 在后台启动 GRPC 服务器。
// 非 grpc.ErrServerStopped 的错误会通过 slog.Error 记录。
func (s *GRPCService) Start(ctx context.Context) error {
	go func() {
		if err := s.server.Serve(s.lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			slog.Error("grpc server error", "error", err)
		}
	}()
	return nil
}

// Stop gracefully stops the GRPC server, with fallback to force stop on timeout.
// Core logic: Call GracefulStop in goroutine, wait for done or context timeout.
//
// Stop 优雅关闭 GRPC 服务器，超时时强制关闭。
// 核心逻辑：在 goroutine 中调用 GracefulStop，等待完成或 context 超时。
func (s *GRPCService) Stop(ctx context.Context) error {
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