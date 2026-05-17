// Package grpc provides the gRPC-based RPC client for the gorp framework.
// This file implements the RPCClient contract with connection pooling,
// service discovery integration, and governance middleware chain.
//
// 本包提供 gorp 框架基于 gRPC 的 RPC 客户端实现。
// 本文件实现 RPCClient 契约，包含连接池、服务发现集成和治理中间件链。
package grpc

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	discoverycontract "github.com/ngq/gorp/framework/contract/discovery"
	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	securitycontract "github.com/ngq/gorp/framework/contract/security"
	supportcontract "github.com/ngq/gorp/framework/contract/support"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	appgrpc "github.com/ngq/gorp/framework/provider/grpc"
	metadatamw "github.com/ngq/gorp/framework/provider/metadata/middleware"
	tracingmw "github.com/ngq/gorp/framework/provider/tracing/middleware"
	rpcgovernance "github.com/ngq/gorp/framework/rpc/governance"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// Client implements transportcontract.RPCClient using gRPC.
// It manages connection pooling, service discovery, metadata propagation,
// tracing, service authentication, circuit breaker and retry.
//
// Connection Pool Behavior:
// - Connections are cached by address in a sync.Map with no upper limit.
// - Each unique service address creates one gRPC connection that is reused.
// - Connections are closed when Client.Close() is called.
// - For scenarios with many services, consider implementing a connection pool
//   with LRU eviction or using a connection pool library.
//
// Client 使用 gRPC 实现 transportcontract.RPCClient。
// 管理连接池、服务发现、metadata 传播、tracing、服务认证、熔断和重试。
//
// 连接池行为：
// - 连接按地址缓存在 sync.Map 中，无上限。
// - 每个唯一服务地址创建一个 gRPC 连接并复用。
// - 调用 Client.Close() 时关闭所有连接。
// - 对于服务数量多的场景，建议实现带 LRU 淘汰的连接池或使用连接池库。
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

// NewClient creates a new gRPC Client instance with full governance capabilities.
//
// NewClient 创建具备完整治理能力的新 gRPC Client 实例。
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

// Call invokes a gRPC method on the specified service with governance middleware chain.
// Implements transportcontract.RPCClient.Call.
//
// Call 在指定服务上调用 gRPC 方法，应用治理中间件链。
// 实现 transportcontract.RPCClient.Call。
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

		// 注入 metadata propagation
		if c.metadataPropagator != nil {
			carrier := newGRPCMetadataCarrier(md)
			c.metadataPropagator.Inject(callCtx, carrier)
		}
		// 注入 tracing context
		if c.tracer != nil {
			carrier := newGRPCMetadataCarrier(md)
			_ = c.tracer.Inject(callCtx, carrier)
		}
		// 注入 service auth token
		if c.serviceAuth != nil {
			if token, tokenErr := c.serviceAuth.GenerateToken(callCtx, service); tokenErr == nil && token != "" {
				md.Set("x-service-token", token)
			}
		}
		// 注入 trace-id
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

// CallRaw is not supported for gRPC as it requires protobuf messages.
// Implements transportcontract.RPCClient.CallRaw.
//
// CallRaw gRPC 不支持原始字节调用，需要 protobuf 消息。
// 实现 transportcontract.RPCClient.CallRaw。
func (c *Client) CallRaw(ctx context.Context, service, method string, data []byte) ([]byte, error) {
	return nil, errors.New("rpc: gRPC does not support raw bytes, use protobuf")
}

// Conn returns the underlying gRPC connection for advanced usage.
// Allows direct access to gRPC client connection for streaming or custom invocations.
//
// Conn 返回底层 gRPC 连接供高级使用。
// 允许直接访问 gRPC 客户端连接进行流式调用或自定义调用。
func (c *Client) Conn(ctx context.Context, service string) (*grpc.ClientConn, error) {
	conn, _, err := c.getConn(ctx, service)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// Close closes all pooled connections and marks the client as closed.
// Implements transportcontract.RPCClient.Close.
//
// Close 关闭所有池化连接并标记客户端为已关闭。
// 实现 transportcontract.RPCClient.Close。
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

// getConn returns a pooled gRPC connection for the given service.
// Uses service discovery when available, falls back to direct target.
//
// getConn 返回给定服务的池化 gRPC 连接。
// 可用时使用服务发现，回退到直接目标地址。
func (c *Client) getConn(ctx context.Context, service string) (*grpc.ClientConn, discoverycontract.DoneFunc, error) {
	addr, done, err := c.resolveTarget(ctx, service)
	if err != nil {
		return nil, nil, err
	}

	// 检查连接池
	if cached, ok := c.connPool.Load(addr); ok {
		conn := cached.(*grpc.ClientConn)
		if conn.GetState().String() != "SHUTDOWN" {
			return conn, done, nil
		}
		c.connPool.Delete(addr)
	}

	// 构建 dial options
	opts := []grpc.DialOption{}
	if c.cfg.Insecure {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	serviceName := service
	if serviceName == "" {
		serviceName = addr
	}

	// 构建拦截器链
	var unaryInterceptors []grpc.UnaryClientInterceptor
	var streamInterceptors []grpc.StreamClientInterceptor

	// metadata propagation interceptor
	if c.metadataPropagator != nil {
		unaryInterceptors = append(unaryInterceptors, metadatamw.UnaryClientInterceptor(c.metadataPropagator))
		streamInterceptors = append(streamInterceptors, metadatamw.StreamClientInterceptor(c.metadataPropagator))
	}
	// tracing interceptor
	if c.tracer != nil {
		unaryInterceptors = append(unaryInterceptors, tracingmw.UnaryClientInterceptor(c.tracer, serviceName))
	}
	// service auth interceptor
	if c.serviceAuth != nil {
		unaryInterceptors = append(unaryInterceptors, serviceAuthUnaryClientInterceptor(c.serviceAuth, serviceName))
		streamInterceptors = append(streamInterceptors, serviceAuthStreamClientInterceptor(c.serviceAuth, serviceName))
	}
	// circuit breaker interceptor
	if c.circuitBreaker != nil {
		unaryInterceptors = append(unaryInterceptors, c.circuitBreakerUnaryInterceptor(serviceName))
		streamInterceptors = append(streamInterceptors, c.circuitBreakerStreamInterceptor(serviceName))
	}
	// default grpc interceptors
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

// resolveTarget resolves the target address using service discovery or direct config.
//
// resolveTarget 使用服务发现或直接配置解析目标地址。
func (c *Client) resolveTarget(ctx context.Context, service string) (string, discoverycontract.DoneFunc, error) {
	// 尝试服务发现
	if c.registry != nil {
		instances, err := c.registry.Discover(ctx, service)
		if err == nil && len(instances) > 0 {
			// 使用 selector 选择实例
			if c.selector != nil {
				selected, done, err := c.selector.Select(ctx, instances)
				if err == nil {
					return selected.Address, done, nil
				}
			}
			// 回退到第一个健康实例
			for _, inst := range instances {
				if inst.Healthy {
					return inst.Address, nil, nil
				}
			}
		}
	}

	// 回退到配置的目标地址
	if c.cfg.Target != "" {
		return c.cfg.Target, nil, nil
	}
	return service, nil, nil
}

// circuitBreakerUnaryInterceptor creates a circuit breaker unary client interceptor.
//
// circuitBreakerUnaryInterceptor 创建熔断器一元客户端拦截器。
func (c *Client) circuitBreakerUnaryInterceptor(service string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		return c.circuitBreaker.Do(ctx, c.circuitBreakerResource(service, method), func() error {
			return invoker(ctx, method, req, reply, cc, opts...)
		})
	}
}

// circuitBreakerStreamInterceptor creates a circuit breaker stream client interceptor.
//
// circuitBreakerStreamInterceptor 创建熔断器流式客户端拦截器。
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

// circuitBreakerResource generates the resource name for circuit breaker.
//
// circuitBreakerResource 生成熔断器资源名称。
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

// sanitizeCircuitBreakerSegment sanitizes a segment for circuit breaker resource name.
//
// sanitizeCircuitBreakerSegment 清理熔断器资源名称的片段。
func sanitizeCircuitBreakerSegment(segment string) string {
	segment = strings.TrimSpace(segment)
	segment = strings.Trim(segment, "/")
	if segment == "" {
		return "unknown"
	}
	replacer := strings.NewReplacer("/", ".", " ", "_", ":", ".")
	return replacer.Replace(segment)
}