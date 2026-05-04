package host

import (
	"context"
	"net"
	"net/http"

	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	"google.golang.org/grpc"
)

type HTTPService struct {
	name   string
	http   transportcontract.HTTP
	server *http.Server
}

func NewHTTPService(name string, h transportcontract.HTTP) *HTTPService {
	return &HTTPService{
		name:   name,
		http:   h,
		server: h.Server(),
	}
}

func (s *HTTPService) Name() string { return s.name }

func (s *HTTPService) Start(ctx context.Context) error {
	go func() {
		if err := s.http.Run(); err != nil && err != http.ErrServerClosed {
		}
	}()
	return nil
}

func (s *HTTPService) Stop(ctx context.Context) error {
	return s.http.Shutdown(ctx)
}

type CronService struct {
	name string
	cron runtimecontract.Cron
}

func NewCronService(name string, c runtimecontract.Cron) *CronService {
	return &CronService{name: name, cron: c}
}

func (s *CronService) Name() string { return s.name }

func (s *CronService) Start(ctx context.Context) error {
	s.cron.Start()
	return nil
}

func (s *CronService) Stop(ctx context.Context) error {
	stopped := s.cron.Stop()
	select {
	case <-stopped.Done():
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

type GRPCService struct {
	name   string
	server *grpc.Server
	lis    net.Listener
}

func NewGRPCService(name string, server *grpc.Server, lis net.Listener) *GRPCService {
	return &GRPCService{name: name, server: server, lis: lis}
}

func (s *GRPCService) Name() string { return s.name }

func (s *GRPCService) Start(ctx context.Context) error {
	go func() {
		_ = s.server.Serve(s.lis)
	}()
	return nil
}

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
