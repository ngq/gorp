// Package grpc provides the gRPC-based RPC server for the gorp framework.
// This file implements the RPCServer contract with health checking,
// reflection, and middleware chain integration.
//
// 本包提供 gorp 框架基于 gRPC 的 RPC 服务端实现。
// 本文件实现 RPCServer 契约，包含健康检查、反射和中间件链集成。
package grpc

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	securitycontract "github.com/ngq/gorp/framework/contract/security"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	configprovider "github.com/ngq/gorp/framework/provider/config"
	appgrpc "github.com/ngq/gorp/framework/provider/grpc"
	metadatamw "github.com/ngq/gorp/framework/provider/metadata/middleware"
	tracingmw "github.com/ngq/gorp/framework/provider/tracing/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	grpc_health_v1 "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

// Server implements transportcontract.RPCServer using gRPC.
// It manages gRPC server lifecycle, health checking, reflection,
// and middleware chain integration.
//
// Server 使用 gRPC 实现 transportcontract.RPCServer。
// 管理 gRPC 服务器生命周期、健康检查、反射和中间件链集成。
type Server struct {
	cfg      *transportcontract.RPCConfig
	c        runtimecontract.Container
	server   *grpc.Server
	addr     string
	services sync.Map
	mu       sync.Mutex
	running  bool
	listener net.Listener
}

// NewServer creates a new gRPC Server instance with container for middleware resolution.
//
// NewServer 创建新的 gRPC Server 实例，使用容器解析中间件依赖。
func NewServer(cfg *transportcontract.RPCConfig, c runtimecontract.Container) *Server {
	return &Server{cfg: cfg, c: c}
}

// Register stores a service handler for later registration.
// Implements transportcontract.RPCServer.Register.
//
// Register 存储服务 handler 供后续注册。
// 实现 transportcontract.RPCServer.Register。
func (s *Server) Register(service string, handler any) error {
	s.services.Store(service, handler)
	return nil
}

// RegisterProto registers a gRPC service using a registration function.
// Allows direct registration of protobuf-generated service implementations.
//
// RegisterProto 使用注册函数注册 gRPC 服务。
// 允许直接注册 protobuf 生成的服务实现。
func (s *Server) RegisterProto(register func(server *grpc.Server) error) error {
	if register == nil {
		return nil
	}
	return register(s.Server())
}

// Start starts the gRPC server on the configured address.
// Implements transportcontract.RPCServer.Start.
//
// Start 在配置地址启动 gRPC 服务器。
// 实现 transportcontract.RPCServer.Start。
func (s *Server) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return errors.New("rpc: server already running")
	}

	addr := s.cfg.Address
	if addr == "" {
		addr = ":9090"
	}

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("rpc: listen failed: %w", err)
	}
	s.listener = lis
	s.addr = lis.Addr().String()

	if s.server == nil {
		s.server = s.newGRPCServer()
	}

	s.running = true
	go func() {
		s.server.Serve(lis)
	}()

	return nil
}

// Stop gracefully stops the gRPC server.
// Implements transportcontract.RPCServer.Stop.
//
// Stop 优雅停止 gRPC 服务器。
// 实现 transportcontract.RPCServer.Stop。
func (s *Server) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	s.server.GracefulStop()
	s.running = false
	return nil
}

// Addr returns the server's listening address.
// Implements transportcontract.RPCServer.Addr.
//
// Addr 返回服务器的监听地址。
// 实现 transportcontract.RPCServer.Addr。
func (s *Server) Addr() string {
	return s.addr
}

// Server returns the underlying gRPC server instance.
// Lazy initializes the server if not already created.
//
// Server 返回底层 gRPC 服务器实例。
// 如果尚未创建则延迟初始化。
func (s *Server) Server() *grpc.Server {
	if s.server == nil {
		s.server = s.newGRPCServer()
	}
	return s.server
}

// GRPCServer returns the underlying gRPC server (alias for Server).
// Provides explicit gRPC type access for advanced usage.
//
// GRPGServer 返回底层 gRPC 服务器（Server 的别名）。
// 提供显式 gRPC 类型访问供高级使用。
func (s *Server) GRPCServer() *grpc.Server {
	return s.Server()
}

// newGRPCServer constructs a gRPC server with middleware chain from container.
// The interceptor chain follows the HTTP middleware order:
// recovery → logging → timeout → tracing → metadata → serviceauth → metrics
//
// newGRPCServer 从容器解析中间件链构建 gRPC 服务器。
// interceptor 链顺序与 HTTP middleware 对齐：
// recovery → logging → timeout → tracing → metadata → serviceauth → metrics
func (s *Server) newGRPCServer() *grpc.Server {
	opts := []grpc.ServerOption{}

	var unaryInterceptors []grpc.UnaryServerInterceptor
	var streamInterceptors []grpc.StreamServerInterceptor

	// 1. recovery (最外层，panic 恢复)
	// recovery - panic 恢复
	unaryInterceptors = append(unaryInterceptors, recoveryUnaryServerInterceptor())
	streamInterceptors = append(streamInterceptors, recoveryStreamServerInterceptor())

	// 2. logging (default grpc interceptors)
	// logging - 默认 grpc interceptors
	unaryInterceptors = append(unaryInterceptors, appgrpc.UnaryServerInterceptor())
	streamInterceptors = append(streamInterceptors, appgrpc.StreamServerInterceptor())

	// 3. timeout (请求超时控制)
	// timeout - 请求超时控制
	if s.cfg.TimeoutMS > 0 {
		unaryInterceptors = append(unaryInterceptors, timeoutUnaryServerInterceptor(time.Duration(s.cfg.TimeoutMS)*time.Millisecond))
	}

	// 4. tracing (分布式追踪)
	// tracing - 分布式追踪
	if s.c.IsBind(observabilitycontract.TracerKey) {
		if tracerAny, err := s.c.Make(observabilitycontract.TracerKey); err == nil {
			if tracer, ok := tracerAny.(observabilitycontract.Tracer); ok {
				serviceName := configprovider.GetStringAny(getConfigService(s.c), "service.name", "tracing.service_name")
				if serviceName == "" {
					serviceName = "grpc-service"
				}
				unaryInterceptors = append(unaryInterceptors, tracingmw.UnaryServerInterceptor(tracer, serviceName))
			}
		}
	}

	// 5. metadata propagation (metadata 传播)
	// metadata - metadata 传播
	if s.c.IsBind(transportcontract.MetadataPropagatorKey) {
		if propagatorAny, err := s.c.Make(transportcontract.MetadataPropagatorKey); err == nil {
			if propagator, ok := propagatorAny.(transportcontract.MetadataPropagator); ok {
				unaryInterceptors = append(unaryInterceptors, metadatamw.UnaryServerInterceptor(propagator))
				streamInterceptors = append(streamInterceptors, metadatamw.StreamServerInterceptor(propagator))
			}
		}
	}

	// 6. service auth (服务间认证)
	// serviceauth - 服务间认证
	if s.c.IsBind(securitycontract.ServiceAuthKey) {
		if authAny, err := s.c.Make(securitycontract.ServiceAuthKey); err == nil {
			if authenticator, ok := authAny.(securitycontract.ServiceAuthenticator); ok {
				unaryInterceptors = append(unaryInterceptors, serviceAuthUnaryServerInterceptor(authenticator))
				streamInterceptors = append(streamInterceptors, serviceAuthStreamServerInterceptor(authenticator))
			}
		}
	}

	// 7. rate limit (从容器解析)
	// rate_limit - 从容器解析
	if s.c.IsBind(resiliencecontract.RateLimiterKey) {
		if limiterAny, err := s.c.Make(resiliencecontract.RateLimiterKey); err == nil {
			if limiter, ok := limiterAny.(resiliencecontract.RateLimiter); ok {
				unaryInterceptors = append(unaryInterceptors, rateLimitUnaryServerInterceptor(limiter))
			}
		}
	}

	// 8. loadshedding (从容器解析)
	// loadshedding - 从容器解析
	if s.c.IsBind(resiliencecontract.LoadShedderKey) {
		if lsAny, err := s.c.Make(resiliencecontract.LoadShedderKey); err == nil {
			if ls, ok := lsAny.(resiliencecontract.LoadShedder); ok {
				unaryInterceptors = append(unaryInterceptors, loadSheddingUnaryServerInterceptor(ls))
			}
		}
	}

	// 9. metrics (请求指标)
	// metrics - 请求指标
	unaryInterceptors = append(unaryInterceptors, metricsUnaryServerInterceptor())

	if len(unaryInterceptors) > 0 {
		opts = append(opts, grpc.ChainUnaryInterceptor(unaryInterceptors...))
	}
	if len(streamInterceptors) > 0 {
		opts = append(opts, grpc.ChainStreamInterceptor(streamInterceptors...))
	}

	srv := grpc.NewServer(opts...)

	// 注册健康检查服务
	hs := health.NewServer()
	hs.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(srv, hs)

	// 注册反射服务
	reflection.Register(srv)

	return srv
}

// rateLimitUnaryServerInterceptor creates a unary server interceptor for rate limiting.
// Rejects requests when the rate limit is exceeded.
//
// rateLimitUnaryServerInterceptor 创建限流的一元服务端拦截器。
// 当超过限流阈值时拒绝请求。
func rateLimitUnaryServerInterceptor(limiter resiliencecontract.RateLimiter) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if limiter != nil {
			if err := limiter.Allow(ctx, info.FullMethod); err != nil {
				return nil, status.Error(codes.ResourceExhausted, "rate limit exceeded")
			}
		}
		return handler(ctx, req)
	}
}

// loadSheddingUnaryServerInterceptor creates a unary server interceptor for load shedding.
// Rejects requests when the system is overloaded.
//
// loadSheddingUnaryServerInterceptor 创建过载保护的一元服务端拦截器。
// 当系统过载时拒绝请求。
func loadSheddingUnaryServerInterceptor(ls resiliencecontract.LoadShedder) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if ls != nil {
			if err := ls.Allow(ctx, info.FullMethod); err != nil {
				return nil, status.Error(codes.Unavailable, "service overloaded")
			}
			callResp, callErr := handler(ctx, req)
			ls.Done(ctx, info.FullMethod, callErr)
			return callResp, callErr
		}
		return handler(ctx, req)
	}
}

// getConfigService retrieves the config binding from container for server-side use.
//
// getConfigService 从容器检索 config binding，用于服务端。
func getConfigService(c runtimecontract.Container) datacontract.Config {
	cfgAny, err := c.Make(datacontract.ConfigKey)
	if err != nil {
		return nil
	}
	cfg, _ := cfgAny.(datacontract.Config)
	return cfg
}