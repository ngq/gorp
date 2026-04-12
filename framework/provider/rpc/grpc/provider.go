package grpc

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/ngq/gorp/framework/contract"
	configprovider "github.com/ngq/gorp/framework/provider/config"
	metadatamw "github.com/ngq/gorp/framework/provider/metadata/middleware"
	serviceauthtoken "github.com/ngq/gorp/framework/provider/serviceauth/token"
	tracingmw "github.com/ngq/gorp/framework/provider/tracing/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// Provider 提供 gRPC RPC 实现。
//
// 中文说明：
// - 基于 gRPC/protobuf 实现服务间调用；
// - 性能更高，适合高频服务间通信；
// - 支持服务发现集成；
// - 需要项目引入 google.golang.org/grpc 依赖。
type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string  { return "rpc.grpc" }
func (p *Provider) IsDefer() bool { return true }
func (p *Provider) Provides() []string {
	return []string{contract.RPCClientKey, contract.RPCServerKey}
}

func (p *Provider) Register(c contract.Container) error {
	// 注册 gRPC RPCClient
	c.Bind(contract.RPCClientKey, func(c contract.Container) (any, error) {
		cfg, _ := getConfig(c)

		// 获取 Registry（如果可用）
		var registry contract.ServiceRegistry
		if c.IsBind(contract.RPCRegistryKey) {
			regAny, _ := c.Make(contract.RPCRegistryKey)
			registry, _ = regAny.(contract.ServiceRegistry)
		}

		// 获取 Selector（如果可用）
		var selector contract.Selector
		if c.IsBind(contract.SelectorKey) {
			selAny, _ := c.Make(contract.SelectorKey)
			selector, _ = selAny.(contract.Selector)
		}

		// 获取 MetadataPropagator（如果可用）
		var metadataPropagator contract.MetadataPropagator
		if c.IsBind(contract.MetadataPropagatorKey) {
			mdAny, _ := c.Make(contract.MetadataPropagatorKey)
			metadataPropagator, _ = mdAny.(contract.MetadataPropagator)
		}

		// 获取 ServiceAuthenticator（如果可用）
		var serviceAuth contract.ServiceAuthenticator
		if c.IsBind(contract.ServiceAuthKey) {
			authAny, _ := c.Make(contract.ServiceAuthKey)
			serviceAuth, _ = authAny.(contract.ServiceAuthenticator)
		}

		// 获取 Tracer（如果可用）
		var tracer contract.Tracer
		if c.IsBind(contract.TracerKey) {
			tracerAny, _ := c.Make(contract.TracerKey)
			tracer, _ = tracerAny.(contract.Tracer)
		}

		return NewClient(cfg, registry, selector, metadataPropagator, serviceAuth, tracer), nil
	}, true)

	// 注册 gRPC RPCServer
	c.Bind(contract.RPCServerKey, func(c contract.Container) (any, error) {
		cfg, _ := getConfig(c)
		return NewServer(cfg, c), nil
	}, true)

	return nil
}

func (p *Provider) Boot(c contract.Container) error {
	return nil
}

// getConfig 从容器获取 RPC 配置。
func getConfig(c contract.Container) (*contract.RPCConfig, error) {
	cfgAny, err := c.Make(contract.ConfigKey)
	if err != nil {
		return &contract.RPCConfig{Mode: "grpc"}, nil
	}

	cfg, ok := cfgAny.(contract.Config)
	if !ok {
		return &contract.RPCConfig{Mode: "grpc"}, nil
	}

	rpcCfg := &contract.RPCConfig{
		Mode:      "grpc",
		Insecure:  true, // 默认不使用 TLS
		TimeoutMS: 30000,
	}

	if mode := configprovider.GetStringAny(cfg,
		"rpc.mode",
	); mode != "" {
		rpcCfg.Mode = mode
	}
	if target := configprovider.GetStringAny(cfg,
		"rpc.grpc.target",
		"rpc.target",
	); target != "" {
		rpcCfg.Target = target
	}
	if insecure, ok := configprovider.GetBoolAny(cfg,
		"rpc.grpc.insecure",
	); ok {
		rpcCfg.Insecure = insecure
	}
	if timeout := configprovider.GetIntAny(cfg,
		"rpc.timeout_ms",
		"rpc.timeout",
	); timeout > 0 {
		rpcCfg.TimeoutMS = timeout
	}
	if addr := configprovider.GetStringAny(cfg,
		"rpc.grpc.address",
		"rpc.address",
	); addr != "" {
		rpcCfg.Address = addr
	}

	return rpcCfg, nil
}

// Client 是 gRPC RPC 客户端实现。
//
// 中文说明：
// - 使用 grpc.ClientConn 发起调用；
// - 支持服务发现：优先从 Registry 获取地址；
// - Call 方法需要传入 gRPC 请求/响应对象。
type Client struct {
	cfg                *contract.RPCConfig
	registry           contract.ServiceRegistry
	selector           contract.Selector
	metadataPropagator contract.MetadataPropagator
	serviceAuth        contract.ServiceAuthenticator
	tracer             contract.Tracer

	// 连接池：按地址缓存 ClientConn
	connPool sync.Map // map[string]*grpc.ClientConn

	// 连接管理
	mu     sync.Mutex
	closed bool
}

// NewClient 创建 gRPC RPC 客户端。
func NewClient(
	cfg *contract.RPCConfig,
	registry contract.ServiceRegistry,
	selector contract.Selector,
	metadataPropagator contract.MetadataPropagator,
	serviceAuth contract.ServiceAuthenticator,
	tracer contract.Tracer,
) *Client {
	return &Client{
		cfg:                cfg,
		registry:           registry,
		selector:           selector,
		metadataPropagator: metadataPropagator,
		serviceAuth:        serviceAuth,
		tracer:             tracer,
	}
}

// Call 执行 RPC 调用。
//
// 中文说明：
// - service: 目标服务名称；
// - method: gRPC 方法全名（如 "/user.UserService/GetUser"）；
// - req/resp: protobuf 请求/响应对象；
// - 使用 Invoke 发起 gRPC 调用。
func (c *Client) Call(ctx context.Context, service, method string, req, resp any) error {
	conn, done, err := c.getConn(ctx, service)
	if err != nil {
		return fmt.Errorf("rpc: get connection failed: %w", err)
	}
	if done != nil {
		defer func() {
			done(ctx, contract.DoneInfo{Err: err, BytesSent: true, BytesReceived: err == nil})
		}()
	}

	// 设置超时
	timeout := time.Duration(c.cfg.TimeoutMS) * time.Millisecond
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	// 准备 outgoing metadata
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	}

	// 注入 metadata（如果存在）
	if c.metadataPropagator != nil {
		carrier := newGRPCMetadataCarrier(md)
		c.metadataPropagator.Inject(ctx, carrier)
	}

	// 注入 tracing 上下文（如果存在）
	if c.tracer != nil {
		carrier := newGRPCMetadataCarrier(md)
		_ = c.tracer.Inject(ctx, carrier)
	}

	// 注入服务间认证令牌（如果启用）
	if c.serviceAuth != nil {
		if token, tokenErr := c.serviceAuth.GenerateToken(ctx, service); tokenErr == nil && token != "" {
			md.Set("x-service-token", token)
		}
	}

	// 透传 TraceID
	if traceID := ctx.Value("trace_id"); traceID != nil {
		md.Set("x-trace-id", fmt.Sprintf("%v", traceID))
	}
	ctx = metadata.NewOutgoingContext(ctx, md)

	// 发起调用
	err = conn.Invoke(ctx, method, req, resp)
	if err != nil {
		return fmt.Errorf("rpc: invoke failed: %w", err)
	}

	return nil
}

// CallRaw 执行原始数据 RPC 调用。
//
// 中文说明：
// - gRPC 不推荐直接传原始字节；
// - 此方法返回错误，建议使用 protobuf 定义。
func (c *Client) CallRaw(ctx context.Context, service, method string, data []byte) ([]byte, error) {
	return nil, errors.New("rpc: gRPC does not support raw bytes, use protobuf")
}

// Close 关闭所有连接。
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}
	c.closed = true

	// 关闭所有缓存的连接
	c.connPool.Range(func(key, value any) bool {
		if conn, ok := value.(*grpc.ClientConn); ok {
			conn.Close()
		}
		return true
	})

	return nil
}

// getConn 获取或创建服务连接。
//
// 中文说明：
// - 从连接池获取，不存在则创建；
// - 支持服务发现动态选择地址。
func (c *Client) getConn(ctx context.Context, service string) (*grpc.ClientConn, contract.DoneFunc, error) {
	addr, done, err := c.resolveTarget(ctx, service)
	if err != nil {
		return nil, nil, err
	}

	// 尝试从缓存获取（按地址缓存，避免把 selector 绕过去）
	if cached, ok := c.connPool.Load(addr); ok {
		conn := cached.(*grpc.ClientConn)
		if conn.GetState().String() != "SHUTDOWN" {
			return conn, done, nil
		}
		c.connPool.Delete(addr)
	}

	opts := []grpc.DialOption{}
	if c.cfg.Insecure {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	conn, err := grpc.NewClient(addr, opts...)
	if err != nil {
		return nil, nil, fmt.Errorf("rpc: dial failed: %w", err)
	}

	c.connPool.Store(addr, conn)
	return conn, done, nil
}

// resolveTarget 解析服务目标地址。
func (c *Client) resolveTarget(ctx context.Context, service string) (string, contract.DoneFunc, error) {
	if c.registry != nil {
		instances, err := c.registry.Discover(ctx, service)
		if err == nil && len(instances) > 0 {
			if c.selector != nil {
				selected, done, err := c.selector.Select(ctx, instances)
				if err == nil {
					return selected.Address, done, nil
				}
			}
			for _, inst := range instances {
				if inst.Healthy {
					return inst.Address, nil, nil
				}
			}
		}
	}

	if c.cfg.Target != "" {
		return c.cfg.Target, nil, nil
	}
	return service, nil, nil
}

func getConfigService(c contract.Container) contract.Config {
	cfgAny, err := c.Make(contract.ConfigKey)
	if err != nil {
		return nil
	}
	cfg, _ := cfgAny.(contract.Config)
	return cfg
}

type grpcMetadataCarrier struct {
	md metadata.MD
}

func newGRPCMetadataCarrier(md metadata.MD) *grpcMetadataCarrier {
	if md == nil {
		md = metadata.New(nil)
	}
	return &grpcMetadataCarrier{md: md}
}

func (c *grpcMetadataCarrier) Get(key string) string {
	values := c.md.Get(key)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func (c *grpcMetadataCarrier) Set(key, value string) {
	c.md.Set(key, value)
}

func (c *grpcMetadataCarrier) Add(key, value string) {
	c.md.Append(key, value)
}

func (c *grpcMetadataCarrier) Keys() []string {
	keys := make([]string, 0, len(c.md))
	for k := range c.md {
		keys = append(keys, k)
	}
	return keys
}

func (c *grpcMetadataCarrier) Values(key string) []string {
	return c.md.Get(key)
}

// Server 是 gRPC RPC 服务端实现。
//
// 中文说明：
// - 使用 grpc.Server 暴露服务；
// - 支持独立监听端口（与 HTTP 服务分离）；
// - Register 注册 protobuf service implementation。
type Server struct {
	cfg   *contract.RPCConfig
	c     contract.Container

	server *grpc.Server
	addr   string

	// 服务注册
	services sync.Map // map[string]any

	// 运行状态
	mu     sync.Mutex
	running bool
	listener net.Listener
}

// NewServer 创建 gRPC RPC 服务端。
func NewServer(cfg *contract.RPCConfig, c contract.Container) *Server {
	return &Server{
		cfg: cfg,
		c:   c,
	}
}

// Register 注册服务处理器。
//
// 中文说明：
// - handler 类型应为 protobuf service implementation；
// - 使用 grpc.Server.RegisterService 注册。
func (s *Server) Register(service string, handler any) error {
	s.services.Store(service, handler)
	return nil
}

// Start 启动 gRPC 服务。
//
// 中文说明：
// - 创建 grpc.Server 并注册所有服务；
// - 开始监听指定端口；
// - 与 HTTP 服务端口分离。
func (s *Server) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return errors.New("rpc: server already running")
	}

	// 确定监听地址
	addr := s.cfg.Address
	if addr == "" {
		addr = ":9090" // 默认 gRPC 端口
	}

	// 创建监听器
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("rpc: listen failed: %w", err)
	}
	s.listener = lis
	s.addr = lis.Addr().String()

	// 创建 gRPC Server
	opts := []grpc.ServerOption{}
	if !s.cfg.Insecure {
		// TODO: 支持 TLS
	}

	var unaryInterceptors []grpc.UnaryServerInterceptor
	if s.c.IsBind(contract.TracerKey) {
		if tracerAny, err := s.c.Make(contract.TracerKey); err == nil {
			if tracer, ok := tracerAny.(contract.Tracer); ok {
				serviceName := configprovider.GetStringAny(getConfigService(s.c), "service.name", "tracing.service_name")
				if serviceName == "" {
					serviceName = "grpc-service"
				}
				unaryInterceptors = append(unaryInterceptors, tracingmw.UnaryServerInterceptor(tracer, serviceName))
			}
		}
	}
	if s.c.IsBind(contract.MetadataPropagatorKey) {
		if propagatorAny, err := s.c.Make(contract.MetadataPropagatorKey); err == nil {
			if propagator, ok := propagatorAny.(contract.MetadataPropagator); ok {
				unaryInterceptors = append(unaryInterceptors, metadatamw.UnaryServerInterceptor(propagator))
			}
		}
	}
	if s.c.IsBind(contract.ServiceAuthKey) {
		if authAny, err := s.c.Make(contract.ServiceAuthKey); err == nil {
			if authenticator, ok := authAny.(contract.ServiceAuthenticator); ok {
				unaryInterceptors = append(unaryInterceptors, serviceauthtoken.UnaryServerInterceptor(authenticator))
			}
		}
	}
	if len(unaryInterceptors) > 0 {
		opts = append(opts, grpc.ChainUnaryInterceptor(unaryInterceptors...))
	}
	s.server = grpc.NewServer(opts...)

	// 注册服务（这里简化处理，实际需要 protobuf 生成的 desc）
	// 用户需要手动调用 gRPC 标准注册方法，这里只做占位
	s.services.Range(func(key, value any) bool {
		// 实际注册由业务代码处理
		return true
	})

	// 标记运行状态
	s.running = true

	// 启动服务（非阻塞）
	go func() {
		s.server.Serve(lis)
	}()

	return nil
}

// Stop 停止 gRPC 服务。
func (s *Server) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	// 优雅关闭
	s.server.GracefulStop()
	s.running = false

	return nil
}

// Addr 返回服务监听地址。
func (s *Server) Addr() string {
	return s.addr
}

// GRPCServer 返回底层 grpc.Server，供业务代码注册 protobuf service。
//
// 中文说明：
// - 业务代码需要用此方法获取 grpc.Server；
// - 然后调用 pb.RegisterXXXServer(s.GRPCServer(), handler)。
func (s *Server) GRPCServer() *grpc.Server {
	return s.server
}