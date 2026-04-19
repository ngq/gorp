package grpc

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	appgrpc "github.com/ngq/gorp/app/grpc"
	"github.com/ngq/gorp/framework/contract"
	configprovider "github.com/ngq/gorp/framework/provider/config"
	metadatamw "github.com/ngq/gorp/framework/provider/metadata/middleware"
	serviceauthtoken "github.com/ngq/gorp/framework/provider/serviceauth/token"
	tracingmw "github.com/ngq/gorp/framework/provider/tracing/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	grpc_health_v1 "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
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
	return []string{
		contract.RPCClientKey,
		contract.RPCServerKey,
		contract.GRPCConnFactoryKey,
		contract.GRPCServerRegistrarKey,
	}
}

func (p *Provider) Register(c contract.Container) error {
	// 注册 Proto-first gRPC 连接工厂。
	c.Bind(contract.GRPCConnFactoryKey, func(c contract.Container) (any, error) {
		cfg, _ := getConfig(c)
		return newClientFromContainer(c, cfg), nil
	}, true)

	// 注册旧统一 RPCClient 抽象，底层继续复用 Proto-first 连接工厂实现。
	c.Bind(contract.RPCClientKey, func(c contract.Container) (any, error) {
		return c.Make(contract.GRPCConnFactoryKey)
	}, true)

	// 注册 Proto-first gRPC 服务端注册器。
	c.Bind(contract.GRPCServerRegistrarKey, func(c contract.Container) (any, error) {
		cfg, _ := getConfig(c)
		return NewServer(cfg, c), nil
	}, true)

	// 注册旧统一 RPCServer 抽象，底层继续复用 Proto-first 服务端注册器实现。
	c.Bind(contract.RPCServerKey, func(c contract.Container) (any, error) {
		return c.Make(contract.GRPCServerRegistrarKey)
	}, true)

	return nil
}

func (p *Provider) Boot(c contract.Container) error {
	return nil
}

func newClientFromContainer(c contract.Container, cfg *contract.RPCConfig) *Client {
	var registry contract.ServiceRegistry
	if c.IsBind(contract.RPCRegistryKey) {
		regAny, _ := c.Make(contract.RPCRegistryKey)
		registry, _ = regAny.(contract.ServiceRegistry)
	}

	var selector contract.Selector
	if c.IsBind(contract.SelectorKey) {
		selAny, _ := c.Make(contract.SelectorKey)
		selector, _ = selAny.(contract.Selector)
	}

	var metadataPropagator contract.MetadataPropagator
	if c.IsBind(contract.MetadataPropagatorKey) {
		mdAny, _ := c.Make(contract.MetadataPropagatorKey)
		metadataPropagator, _ = mdAny.(contract.MetadataPropagator)
	}

	var serviceAuth contract.ServiceAuthenticator
	if c.IsBind(contract.ServiceAuthKey) {
		authAny, _ := c.Make(contract.ServiceAuthKey)
		serviceAuth, _ = authAny.(contract.ServiceAuthenticator)
	}

	var tracer contract.Tracer
	if c.IsBind(contract.TracerKey) {
		tracerAny, _ := c.Make(contract.TracerKey)
		tracer, _ = tracerAny.(contract.Tracer)
	}

	var circuitBreaker contract.CircuitBreaker
	if c.IsBind(contract.CircuitBreakerKey) {
		cbAny, _ := c.Make(contract.CircuitBreakerKey)
		circuitBreaker, _ = cbAny.(contract.CircuitBreaker)
	}

	return NewClient(cfg, registry, selector, metadataPropagator, serviceAuth, tracer, circuitBreaker)
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
		Insecure:  true,
		TimeoutMS: 30000,
	}

	if mode := configprovider.GetStringAny(cfg, "rpc.mode"); mode != "" {
		rpcCfg.Mode = mode
	}
	if target := configprovider.GetStringAny(cfg, "rpc.grpc.target", "rpc.target"); target != "" {
		rpcCfg.Target = target
	}
	if insecure, ok := configprovider.GetBoolAny(cfg, "rpc.grpc.insecure"); ok {
		rpcCfg.Insecure = insecure
	}
	if timeout := configprovider.GetIntAny(cfg, "rpc.timeout_ms", "rpc.timeout"); timeout > 0 {
		rpcCfg.TimeoutMS = timeout
	}
	if addr := configprovider.GetStringAny(cfg, "rpc.grpc.address", "rpc.address"); addr != "" {
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
	circuitBreaker     contract.CircuitBreaker

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
	circuitBreaker contract.CircuitBreaker,
) *Client {
	return &Client{
		cfg:                cfg,
		registry:           registry,
		selector:           selector,
		metadataPropagator: metadataPropagator,
		serviceAuth:        serviceAuth,
		tracer:             tracer,
		circuitBreaker:     circuitBreaker,
	}
}

// Call 执行 RPC 调用。
//
// 中文说明：
// - service: 目标服务名称；
// - method: gRPC 方法全名（如 "/user.UserService/GetUser"）；
// - req/resp: protobuf 请求/响应对象；
// - 使用 Invoke 发起 gRPC 调用；
// - 这是旧统一 RPC 抽象留下的兼容入口，Proto-first 主线应优先使用 Conn + pb.NewXxxClient。
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

	timeout := time.Duration(c.cfg.TimeoutMS) * time.Millisecond
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	}

	if c.metadataPropagator != nil {
		carrier := newGRPCMetadataCarrier(md)
		c.metadataPropagator.Inject(ctx, carrier)
	}
	if c.tracer != nil {
		carrier := newGRPCMetadataCarrier(md)
		_ = c.tracer.Inject(ctx, carrier)
	}
	if c.serviceAuth != nil {
		if token, tokenErr := c.serviceAuth.GenerateToken(ctx, service); tokenErr == nil && token != "" {
			md.Set("x-service-token", token)
		}
	}
	if traceID := ctx.Value("trace_id"); traceID != nil {
		md.Set("x-trace-id", fmt.Sprintf("%v", traceID))
	}
	ctx = metadata.NewOutgoingContext(ctx, md)

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

// Conn 按服务名返回可复用的 gRPC 连接。
//
// 中文说明：
// - 这是 Proto-first 客户端主线使用的正式入口；
// - 业务侧拿到连接后应继续使用 `pb.NewXxxClient(conn)` 发起调用；
// - discovery / selector / metadata / tracing / serviceauth / 连接复用仍由 framework 负责。
func (c *Client) Conn(ctx context.Context, service string) (*grpc.ClientConn, error) {
	conn, _, err := c.getConn(ctx, service)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// Close 关闭所有连接。
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}
	c.closed = true

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

	serviceName := service
	if serviceName == "" {
		serviceName = addr
	}

	var unaryInterceptors []grpc.UnaryClientInterceptor
	var streamInterceptors []grpc.StreamClientInterceptor
	if c.metadataPropagator != nil {
		unaryInterceptors = append(unaryInterceptors, metadatamw.UnaryClientInterceptor(c.metadataPropagator))
		streamInterceptors = append(streamInterceptors, metadatamw.StreamClientInterceptor(c.metadataPropagator))
	}
	if c.tracer != nil {
		unaryInterceptors = append(unaryInterceptors, tracingmw.UnaryClientInterceptor(c.tracer, serviceName))
	}
	if c.serviceAuth != nil {
		unaryInterceptors = append(unaryInterceptors, serviceauthtoken.UnaryClientInterceptor(c.serviceAuth, serviceName))
		streamInterceptors = append(streamInterceptors, serviceauthtoken.StreamClientInterceptor(c.serviceAuth, serviceName))
	}
	if c.circuitBreaker != nil {
		unaryInterceptors = append(unaryInterceptors, c.circuitBreakerUnaryInterceptor(serviceName))
		streamInterceptors = append(streamInterceptors, c.circuitBreakerStreamInterceptor(serviceName))
	}
	unaryInterceptors = append(unaryInterceptors, appgrpc.UnaryClientInterceptor())
	streamInterceptors = append(streamInterceptors, appgrpc.StreamClientInterceptor())
	if len(unaryInterceptors) > 0 {
		opts = append(opts, grpc.WithChainUnaryInterceptor(unaryInterceptors...))
	}
	if len(streamInterceptors) > 0 {
		opts = append(opts, grpc.WithChainStreamInterceptor(streamInterceptors...))
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

func (c *Client) circuitBreakerUnaryInterceptor(service string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		return c.circuitBreaker.Do(ctx, c.circuitBreakerResource(service, method), func() error {
			return invoker(ctx, method, req, reply, cc, opts...)
		})
	}
}

func (c *Client) circuitBreakerStreamInterceptor(service string) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		var stream grpc.ClientStream
		err := c.circuitBreaker.Do(ctx, c.circuitBreakerResource(service, method), func() error {
			var callErr error
			stream, callErr = streamer(ctx, desc, cc, method, opts...)
			return callErr
		})
		return stream, err
	}
}

func (c *Client) circuitBreakerResource(service, method string) string {
	parts := []string{"rpc", "grpc"}
	if service != "" {
		parts = append(parts, sanitizeCircuitBreakerSegment(service))
	}
	if method != "" {
		parts = append(parts, sanitizeCircuitBreakerSegment(method))
	}
	return strings.Join(parts, ".")
}

func sanitizeCircuitBreakerSegment(segment string) string {
	segment = strings.TrimSpace(segment)
	segment = strings.Trim(segment, "/")
	if segment == "" {
		return "unknown"
	}
	replacer := strings.NewReplacer("/", ".", " ", "_", ":", ".")
	return replacer.Replace(segment)
}

// Server 是 gRPC RPC 服务端实现。
//
// 中文说明：
// - 使用 grpc.Server 暴露服务；
// - 支持独立监听端口（与 HTTP 服务分离）；
// - Register 注册 protobuf service implementation。
type Server struct {
	cfg *contract.RPCConfig
	c   contract.Container

	server *grpc.Server
	addr   string

	// 服务注册
	services sync.Map // map[string]any

	// 运行状态
	mu       sync.Mutex
	running  bool
	listener net.Listener
}

// NewServer 创建 gRPC RPC 服务端。
func NewServer(cfg *contract.RPCConfig, c contract.Container) *Server {
	return &Server{cfg: cfg, c: c}
}

func (s *Server) newGRPCServer() *grpc.Server {
	opts := []grpc.ServerOption{}
	if !s.cfg.Insecure {
		// TODO: 支持 TLS
	}

	var unaryInterceptors []grpc.UnaryServerInterceptor
	var streamInterceptors []grpc.StreamServerInterceptor
	unaryInterceptors = append(unaryInterceptors, appgrpc.UnaryServerInterceptor())
	streamInterceptors = append(streamInterceptors, appgrpc.StreamServerInterceptor())
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
				streamInterceptors = append(streamInterceptors, metadatamw.StreamServerInterceptor(propagator))
			}
		}
	}
	if s.c.IsBind(contract.ServiceAuthKey) {
		if authAny, err := s.c.Make(contract.ServiceAuthKey); err == nil {
			if authenticator, ok := authAny.(contract.ServiceAuthenticator); ok {
				unaryInterceptors = append(unaryInterceptors, serviceauthtoken.UnaryServerInterceptor(authenticator))
				streamInterceptors = append(streamInterceptors, serviceauthtoken.StreamServerInterceptor(authenticator))
			}
		}
	}
	if len(unaryInterceptors) > 0 {
		opts = append(opts, grpc.ChainUnaryInterceptor(unaryInterceptors...))
	}
	if len(streamInterceptors) > 0 {
		opts = append(opts, grpc.ChainStreamInterceptor(streamInterceptors...))
	}

	srv := grpc.NewServer(opts...)
	hs := health.NewServer()
	hs.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(srv, hs)
	reflection.Register(srv)
	return srv
}

// Register 注册服务处理器。
//
// 中文说明：
// - 这是旧统一 RPC 抽象留下的弱类型注册入口；
// - Proto-first 正式主线应优先使用下方的 `RegisterProto`；
// - 当前保留该方法仅用于兼容历史抽象，不再作为公开 gRPC 主路径。
func (s *Server) Register(service string, handler any) error {
	s.services.Store(service, handler)
	return nil
}

// RegisterProto 注册标准 protobuf service。
//
// 中文说明：
// - 业务侧应通过 `RegisterProto(func(server *grpc.Server) error { ... })` 挂接 `pb.RegisterXxxServer(...)`；
// - 这样业务主线可以保持标准 gRPC register 心智；
// - 若底层 server 尚未初始化，会先按当前配置创建一个可注册的 gRPC server。
func (s *Server) RegisterProto(register func(server *grpc.Server) error) error {
	if register == nil {
		return nil
	}
	return register(s.Server())
}

// Start 启动 gRPC 服务。
//
// 中文说明：
// - 创建 grpc.Server 并注册所有服务；
// - 开始监听指定端口；
// - 与 HTTP 服务端口分离；
// - 若业务已通过 Proto-first 注册入口预先创建 server，则直接复用，不再覆盖。
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

	s.services.Range(func(key, value any) bool {
		return true
	})

	s.running = true
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

	s.server.GracefulStop()
	s.running = false
	return nil
}

// Addr 返回服务监听地址。
func (s *Server) Addr() string {
	return s.addr
}

// Server 返回底层 grpc.Server。
//
// 中文说明：
// - 这是 Proto-first 注册器接口的一部分；
// - 常规业务注册优先走 `RegisterProto`；
// - 暴露该方法主要用于保留最小逃生舱与底层扩展能力。
func (s *Server) Server() *grpc.Server {
	if s.server == nil {
		s.server = s.newGRPCServer()
	}
	return s.server
}

// GRPCServer 返回底层 grpc.Server，供业务代码注册 protobuf service。
//
// 中文说明：
// - 这是旧名称保留，方便现有内部代码逐步迁移；
// - Proto-first 主线应优先使用 `Server()` / `RegisterProto()`。
func (s *Server) GRPCServer() *grpc.Server {
	return s.Server()
}
