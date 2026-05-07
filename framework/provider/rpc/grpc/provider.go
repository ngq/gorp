package grpc

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	discoverycontract "github.com/ngq/gorp/framework/contract/discovery"
	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	securitycontract "github.com/ngq/gorp/framework/contract/security"
	supportcontract "github.com/ngq/gorp/framework/contract/support"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	configprovider "github.com/ngq/gorp/framework/provider/config"
	appgrpc "github.com/ngq/gorp/framework/provider/grpc"
	metadatamw "github.com/ngq/gorp/framework/provider/metadata/middleware"
	tracingmw "github.com/ngq/gorp/framework/provider/tracing/middleware"
	rpcgovernance "github.com/ngq/gorp/framework/rpc/governance"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	grpc_health_v1 "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
)

type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string  { return "rpc.grpc" }
func (p *Provider) IsDefer() bool { return true }
func (p *Provider) Provides() []string {
	return []string{
		transportcontract.RPCClientKey,
		transportcontract.RPCServerKey,
		transportcontract.GRPCConnFactoryKey,
		transportcontract.GRPCServerRegistrarKey,
	}
}

func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(transportcontract.GRPCConnFactoryKey, func(c runtimecontract.Container) (any, error) {
		cfg, _ := getGRPCConfig(c)
		return newClientFromContainer(c, cfg), nil
	}, true)

	c.Bind(transportcontract.RPCClientKey, func(c runtimecontract.Container) (any, error) {
		return c.Make(transportcontract.GRPCConnFactoryKey)
	}, true)

	c.Bind(transportcontract.GRPCServerRegistrarKey, func(c runtimecontract.Container) (any, error) {
		cfg, _ := getGRPCConfig(c)
		return NewServer(cfg, c), nil
	}, true)

	c.Bind(transportcontract.RPCServerKey, func(c runtimecontract.Container) (any, error) {
		return c.Make(transportcontract.GRPCServerRegistrarKey)
	}, true)

	return nil
}

func (p *Provider) Boot(c runtimecontract.Container) error {
	return nil
}

func newClientFromContainer(c runtimecontract.Container, cfg *transportcontract.RPCConfig) *Client {
	var registry transportcontract.ServiceRegistry
	if c.IsBind(transportcontract.RPCRegistryKey) {
		regAny, _ := c.Make(transportcontract.RPCRegistryKey)
		registry, _ = regAny.(transportcontract.ServiceRegistry)
	}

	var selector discoverycontract.Selector
	if c.IsBind(discoverycontract.SelectorKey) {
		selAny, _ := c.Make(discoverycontract.SelectorKey)
		selector, _ = selAny.(discoverycontract.Selector)
	}

	var metadataPropagator transportcontract.MetadataPropagator
	if c.IsBind(transportcontract.MetadataPropagatorKey) {
		mdAny, _ := c.Make(transportcontract.MetadataPropagatorKey)
		metadataPropagator, _ = mdAny.(transportcontract.MetadataPropagator)
	}

	var serviceAuth securitycontract.ServiceTokenIssuer
	if c.IsBind(securitycontract.ServiceAuthKey) {
		authAny, _ := c.Make(securitycontract.ServiceAuthKey)
		serviceAuth, _ = authAny.(securitycontract.ServiceTokenIssuer)
	}

	var tracer observabilitycontract.Tracer
	if c.IsBind(observabilitycontract.TracerKey) {
		tracerAny, _ := c.Make(observabilitycontract.TracerKey)
		tracer, _ = tracerAny.(observabilitycontract.Tracer)
	}

	var circuitBreaker resiliencecontract.CircuitBreaker
	if c.IsBind(resiliencecontract.CircuitBreakerKey) {
		cbAny, _ := c.Make(resiliencecontract.CircuitBreakerKey)
		circuitBreaker, _ = cbAny.(resiliencecontract.CircuitBreaker)
	}
	var retry resiliencecontract.Retry
	if c.IsBind(resiliencecontract.RetryKey) {
		retryAny, _ := c.Make(resiliencecontract.RetryKey)
		retry, _ = retryAny.(resiliencecontract.Retry)
	}

	return NewClient(cfg, registry, selector, metadataPropagator, serviceAuth, tracer, circuitBreaker, retry)
}

func getGRPCConfig(c runtimecontract.Container) (*transportcontract.RPCConfig, error) {
	cfgAny, err := c.Make(datacontract.ConfigKey)
	if err != nil {
		return &transportcontract.RPCConfig{Mode: "grpc"}, nil
	}

	cfg, ok := cfgAny.(datacontract.Config)
	if !ok {
		return &transportcontract.RPCConfig{Mode: "grpc"}, nil
	}

	rpcCfg := &transportcontract.RPCConfig{
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
	if insecureFlag, ok := configprovider.GetBoolAny(cfg, "rpc.grpc.insecure"); ok {
		rpcCfg.Insecure = insecureFlag
	}
	if timeout := configprovider.GetIntAny(cfg, "rpc.timeout_ms", "rpc.timeout"); timeout > 0 {
		rpcCfg.TimeoutMS = timeout
	}
	if addr := configprovider.GetStringAny(cfg, "rpc.grpc.address", "rpc.address"); addr != "" {
		rpcCfg.Address = addr
	}

	return rpcCfg, nil
}

type Client struct {
	cfg                *transportcontract.RPCConfig
	registry           transportcontract.ServiceRegistry
	selector           discoverycontract.Selector
	metadataPropagator transportcontract.MetadataPropagator
	serviceAuth        securitycontract.ServiceTokenIssuer
	tracer             observabilitycontract.Tracer
	circuitBreaker     resiliencecontract.CircuitBreaker
	retry              resiliencecontract.Retry
	connPool           sync.Map
	mu                 sync.Mutex
	closed             bool
}

func NewClient(
	cfg *transportcontract.RPCConfig,
	registry transportcontract.ServiceRegistry,
	selector discoverycontract.Selector,
	metadataPropagator transportcontract.MetadataPropagator,
	serviceAuth securitycontract.ServiceTokenIssuer,
	tracer observabilitycontract.Tracer,
	circuitBreaker resiliencecontract.CircuitBreaker,
	retry resiliencecontract.Retry,
) *Client {
	return &Client{
		cfg:                cfg,
		registry:           registry,
		selector:           selector,
		metadataPropagator: metadataPropagator,
		serviceAuth:        serviceAuth,
		tracer:             tracer,
		circuitBreaker:     circuitBreaker,
		retry:              retry,
	}
}

func (c *Client) Call(ctx context.Context, service, method string, req, resp any) error {
	conn, done, err := c.getConn(ctx, service)
	if err != nil {
		return fmt.Errorf("rpc: get connection failed: %w", err)
	}
	startedAt := time.Now()
	if done != nil {
		defer func() {
			latency := time.Since(startedAt)
			if latency <= 0 {
				latency = time.Nanosecond
			}
			done(ctx, discoverycontract.DoneInfo{
				Err:           err,
				BytesSent:     true,
				BytesReceived: err == nil,
				Latency:       latency,
			})
		}()
	}

	invoker := rpcgovernance.Apply(func(callCtx context.Context, service, method string, req, resp any) error {
		md, ok := metadata.FromOutgoingContext(callCtx)
		if !ok {
			md = metadata.New(nil)
		}

		if c.metadataPropagator != nil {
			carrier := newGRPCMetadataCarrier(md)
			c.metadataPropagator.Inject(callCtx, carrier)
		}
		if c.tracer != nil {
			carrier := newGRPCMetadataCarrier(md)
			_ = c.tracer.Inject(callCtx, carrier)
		}
		if c.serviceAuth != nil {
			if token, tokenErr := c.serviceAuth.GenerateToken(callCtx, service); tokenErr == nil && token != "" {
				md.Set("x-service-token", token)
			}
		}
		if traceID, ok := supportcontract.FromTraceIDContext(callCtx); ok {
			md.Set("x-trace-id", traceID)
		}
		callCtx = metadata.NewOutgoingContext(callCtx, md)

		return conn.Invoke(callCtx, method, req, resp)
	},
		rpcgovernance.TimeoutMiddleware(time.Duration(c.cfg.TimeoutMS)*time.Millisecond),
		rpcgovernance.RetryMiddlewareWithResource(c.retry, c.circuitBreakerResource),
	)

	err = invoker(ctx, service, method, req, resp)
	if err != nil {
		return fmt.Errorf("rpc: invoke failed: %w", err)
	}

	return nil
}

func (c *Client) CallRaw(ctx context.Context, service, method string, data []byte) ([]byte, error) {
	return nil, errors.New("rpc: gRPC does not support raw bytes, use protobuf")
}

func (c *Client) Conn(ctx context.Context, service string) (*grpc.ClientConn, error) {
	conn, _, err := c.getConn(ctx, service)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

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

func (c *Client) getConn(ctx context.Context, service string) (*grpc.ClientConn, discoverycontract.DoneFunc, error) {
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
		unaryInterceptors = append(unaryInterceptors, serviceAuthUnaryClientInterceptor(c.serviceAuth, serviceName))
		streamInterceptors = append(streamInterceptors, serviceAuthStreamClientInterceptor(c.serviceAuth, serviceName))
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

func (c *Client) resolveTarget(ctx context.Context, service string) (string, discoverycontract.DoneFunc, error) {
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

func getConfigService(c runtimecontract.Container) datacontract.Config {
	cfgAny, err := c.Make(datacontract.ConfigKey)
	if err != nil {
		return nil
	}
	cfg, _ := cfgAny.(datacontract.Config)
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

func NewServer(cfg *transportcontract.RPCConfig, c runtimecontract.Container) *Server {
	return &Server{cfg: cfg, c: c}
}

func (s *Server) newGRPCServer() *grpc.Server {
	opts := []grpc.ServerOption{}
	if !s.cfg.Insecure {
	}

	var unaryInterceptors []grpc.UnaryServerInterceptor
	var streamInterceptors []grpc.StreamServerInterceptor
	unaryInterceptors = append(unaryInterceptors, appgrpc.UnaryServerInterceptor())
	streamInterceptors = append(streamInterceptors, appgrpc.StreamServerInterceptor())
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
	if s.c.IsBind(transportcontract.MetadataPropagatorKey) {
		if propagatorAny, err := s.c.Make(transportcontract.MetadataPropagatorKey); err == nil {
			if propagator, ok := propagatorAny.(transportcontract.MetadataPropagator); ok {
				unaryInterceptors = append(unaryInterceptors, metadatamw.UnaryServerInterceptor(propagator))
				streamInterceptors = append(streamInterceptors, metadatamw.StreamServerInterceptor(propagator))
			}
		}
	}
	if s.c.IsBind(securitycontract.ServiceAuthKey) {
		if authAny, err := s.c.Make(securitycontract.ServiceAuthKey); err == nil {
			if authenticator, ok := authAny.(securitycontract.ServiceAuthenticator); ok {
				unaryInterceptors = append(unaryInterceptors, serviceAuthUnaryServerInterceptor(authenticator))
				streamInterceptors = append(streamInterceptors, serviceAuthStreamServerInterceptor(authenticator))
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

func (s *Server) Register(service string, handler any) error {
	s.services.Store(service, handler)
	return nil
}

func (s *Server) RegisterProto(register func(server *grpc.Server) error) error {
	if register == nil {
		return nil
	}
	return register(s.Server())
}

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

func (s *Server) Addr() string {
	return s.addr
}

func (s *Server) Server() *grpc.Server {
	if s.server == nil {
		s.server = s.newGRPCServer()
	}
	return s.server
}

func (s *Server) GRPCServer() *grpc.Server {
	return s.Server()
}

type serviceAuthWrappedStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *serviceAuthWrappedStream) Context() context.Context { return w.ctx }

func serviceAuthUnaryClientInterceptor(auth securitycontract.ServiceTokenIssuer, targetService string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if token, err := auth.GenerateToken(ctx, targetService); err == nil && strings.TrimSpace(token) != "" {
			md, ok := metadata.FromOutgoingContext(ctx)
			if !ok {
				md = metadata.New(nil)
			}
			md.Set("x-service-token", token)
			ctx = metadata.NewOutgoingContext(ctx, md)
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func serviceAuthStreamClientInterceptor(auth securitycontract.ServiceTokenIssuer, targetService string) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		if token, err := auth.GenerateToken(ctx, targetService); err == nil && strings.TrimSpace(token) != "" {
			md, ok := metadata.FromOutgoingContext(ctx)
			if !ok {
				md = metadata.New(nil)
			}
			md.Set("x-service-token", token)
			ctx = metadata.NewOutgoingContext(ctx, md)
		}
		return streamer(ctx, desc, cc, method, opts...)
	}
}

func serviceAuthUnaryServerInterceptor(auth securitycontract.ServiceAuthenticator) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if values := md.Get("x-service-token"); len(values) > 0 && strings.TrimSpace(values[0]) != "" {
				ctx = context.WithValue(ctx, "x-service-token", values[0])
			}
		}
		identity, err := auth.Authenticate(ctx)
		if err != nil {
			return nil, fmt.Errorf("rpc: service authentication failed: %w", err)
		}
		if identity != nil {
			ctx = securitycontract.NewServiceIdentityContext(ctx, identity)
		}
		return handler(ctx, req)
	}
}

func serviceAuthStreamServerInterceptor(auth securitycontract.ServiceAuthenticator) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := ss.Context()
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if values := md.Get("x-service-token"); len(values) > 0 && strings.TrimSpace(values[0]) != "" {
				ctx = context.WithValue(ctx, "x-service-token", values[0])
			}
		}
		identity, err := auth.Authenticate(ctx)
		if err != nil {
			return fmt.Errorf("rpc: service authentication failed: %w", err)
		}
		if identity != nil {
			ctx = securitycontract.NewServiceIdentityContext(ctx, identity)
		}
		return handler(srv, &serviceAuthWrappedStream{ServerStream: ss, ctx: ctx})
	}
}
