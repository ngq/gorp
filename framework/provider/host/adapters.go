package host

import (
	"context"
	"net"
	"net/http"

	"github.com/ngq/gorp/framework/contract"
	"google.golang.org/grpc"
)

// HTTPService 是 HTTP 服务的 Hostable 适配器。
//
// 中文说明：
// - 把 contract.HTTP 包装成 Host 可编排的服务对象；
// - 这样 HTTP 服务就能统一纳入 Host 的启动/停止顺序管理。
type HTTPService struct {
	name   string
	http   contract.HTTP
	server *http.Server
}

// NewHTTPService 创建 HTTP Hostable 适配器。
//
// 中文说明：
// - 把 contract.HTTP 的运行与关闭能力桥接到 Hostable 接口上；
// - name 用于 Host 生命周期列表与日志定位。
func NewHTTPService(name string, h contract.HTTP) *HTTPService {
	return &HTTPService{
		name:   name,
		http:   h,
		server: h.Server(),
	}
}

// Name 返回服务名。
func (s *HTTPService) Name() string { return s.name }

// Start 启动 HTTP 服务。
//
// 中文说明：
// - 真实监听过程放到 goroutine 中执行，避免阻塞 Host 启动编排。
func (s *HTTPService) Start(ctx context.Context) error {
	go func() {
		if err := s.http.Run(); err != nil && err != http.ErrServerClosed {
			// ignore here; command layer still owns logging/reporting strategy
		}
	}()
	return nil
}

// Stop 触发 HTTP 服务优雅关闭。
//
// 中文说明：
// - 直接复用 contract.HTTP 的 Shutdown 语义；
// - 由 Host 统一传入关闭超时上下文。
func (s *HTTPService) Stop(ctx context.Context) error {
	return s.http.Shutdown(ctx)
}

// CronService 是 Cron 服务的 Hostable 适配器。
//
// 中文说明：
// - 把 contract.Cron 包装成 Host 可编排的服务对象；
// - 让定时任务调度也能纳入统一生命周期管理。
type CronService struct {
	name string
	cron contract.Cron
}

// NewCronService 创建 Cron Hostable 适配器。
//
// 中文说明：
// - 把 contract.Cron 的 Start/Stop 能力桥接到 Hostable 接口上；
// - name 用于 Host 生命周期列表与日志定位。
func NewCronService(name string, c contract.Cron) *CronService {
	return &CronService{name: name, cron: c}
}

// Name 返回服务名。
func (s *CronService) Name() string { return s.name }

// Start 启动 Cron 调度器。
//
// 中文说明：
// - Cron 本身启动很快，因此这里直接同步调用 Start；
// - 真正的任务执行仍由底层调度器异步触发。
func (s *CronService) Start(ctx context.Context) error {
	s.cron.Start()
	return nil
}

// Stop 停止 Cron 调度器并等待收尾。
//
// 中文说明：
// - 先调用底层 Stop，再等待已在执行中的任务退出；
// - 如果外部 ctx 先超时，则返回 ctx.Err()。
func (s *CronService) Stop(ctx context.Context) error {
	stopped := s.cron.Stop()
	select {
	case <-stopped.Done():
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// GRPCService 是 gRPC 服务的 Hostable 适配器。
//
// 中文说明：
// - 把标准 grpc.Server + Listener 包装成 Host 可编排服务；
// - 统一接入 Host 的启动与优雅停止流程。
type GRPCService struct {
	name   string
	server *grpc.Server
	lis    net.Listener
}

// NewGRPCService 创建 gRPC Hostable 适配器。
//
// 中文说明：
// - 把 grpc.Server 的 Serve/GracefulStop 生命周期桥接到 Hostable；
// - name 用于 Host 生命周期列表与日志定位。
func NewGRPCService(name string, server *grpc.Server, lis net.Listener) *GRPCService {
	return &GRPCService{name: name, server: server, lis: lis}
}

// Name 返回服务名。
func (s *GRPCService) Name() string { return s.name }

// Start 启动 gRPC 服务。
//
// 中文说明：
// - 与 HTTP 适配器一致，监听过程放到 goroutine 中执行，避免阻塞 Host。
func (s *GRPCService) Start(ctx context.Context) error {
	go func() {
		_ = s.server.Serve(s.lis)
	}()
	return nil
}

// Stop 优先尝试优雅关闭，超时后强制停止。
//
// 中文说明：
// - 先走 GracefulStop，让已建立连接有机会自然收尾；
// - 如果外部 ctx 已超时，则退化到 Stop 强制退出。
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
		return ctx.Err()
	}
}
