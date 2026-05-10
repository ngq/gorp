// Package http provides HTTP RPC client for gorp framework.
// This file implements the RPCClient contract with HTTP transport,
// service discovery, metadata propagation, tracing, circuit breaker and retry.
//
// 本包提供 HTTP RPC 客户端，用于 gorp 框架。
// 本文件实现带 HTTP 传输的 RPCClient 契约，
// 包含服务发现、元数据传播、追踪、熔断器和重试。
package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	discoverycontract "github.com/ngq/gorp/framework/contract/discovery"
	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	securitycontract "github.com/ngq/gorp/framework/contract/security"
	supportcontract "github.com/ngq/gorp/framework/contract/support"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	rpcgovernance "github.com/ngq/gorp/framework/rpc/governance"
)

// Client implements transportcontract.RPCClient using HTTP.
// It manages HTTP connections, service discovery, metadata propagation,
// tracing, service authentication, circuit breaker and retry.
//
// Client 使用 HTTP 实现 transportcontract.RPCClient。
// 管理 HTTP 连接、服务发现、元数据传播、追踪、服务认证、熔断器和重试。
type Client struct {
	cfg                *transportcontract.RPCConfig
	registry           transportcontract.ServiceRegistry
	selector           discoverycontract.Selector
	metadataPropagator transportcontract.MetadataPropagator
	serviceAuth        securitycontract.ServiceTokenIssuer
	tracer             observabilitycontract.Tracer
	circuitBreaker     resiliencecontract.CircuitBreaker
	retry              resiliencecontract.Retry
	httpCli            *http.Client
	serviceCache       sync.Map
}

// cachedAddr stores cached service address with expiration.
//
// cachedAddr 存储带过期时间的缓存服务地址。
type cachedAddr struct {
	addr     string
	expireAt time.Time
}

// NewClient creates a new HTTP Client instance with full governance capabilities.
// Timeout is configured from RPCConfig.TimeoutMS with 30s default.
//
// NewClient 创建具备完整治理能力的新 HTTP Client 实例。
// 超时时间从 RPCConfig.TimeoutMS 配置，默认 30 秒。
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
	timeout := time.Duration(cfg.TimeoutMS) * time.Millisecond
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	return &Client{
		cfg:                cfg,
		registry:           registry,
		selector:           selector,
		metadataPropagator: metadataPropagator,
		serviceAuth:        serviceAuth,
		tracer:             tracer,
		circuitBreaker:     circuitBreaker,
		retry:              retry,
		httpCli: &http.Client{
			Timeout: timeout,
		},
	}
}

// Call invokes an HTTP RPC method on the specified service with governance middleware chain.
// Implements transportcontract.RPCClient.Call.
// Marshals request to JSON, resolves target address, applies metadata/tracing/auth headers,
// executes request with timeout/retry/circuit-breaker middleware, and unmarshals response.
//
// Call 在指定服务上调用 HTTP RPC 方法，应用治理中间件链。
// 实现 transportcontract.RPCClient.Call。
// 将请求序列化为 JSON、解析目标地址、应用 metadata/tracing/auth 头、
// 执行带 timeout/retry/circuit-breaker 中间件的请求、反序列化响应。
func (c *Client) Call(ctx context.Context, service, method string, req, resp any) error {
	reqBody, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("rpc: marshal request failed: %w", err)
	}

	addr, done, err := c.resolveTarget(ctx, service)
	if err != nil {
		return fmt.Errorf("rpc: resolve address failed: %w", err)
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
		fullURL := c.buildURL(addr, method)

		httpReq, reqErr := http.NewRequestWithContext(callCtx, http.MethodPost, fullURL, bytes.NewReader(reqBody))
		if reqErr != nil {
			return fmt.Errorf("rpc: create request failed: %w", reqErr)
		}
		httpReq.Header.Set("Content-Type", "application/json")

		if c.metadataPropagator != nil {
			carrier := &headerCarrier{header: httpReq.Header}
			c.metadataPropagator.Inject(callCtx, carrier)
		}

		if c.tracer != nil {
			carrier := &headerCarrier{header: httpReq.Header}
			_ = c.tracer.Inject(callCtx, carrier)
		}

		if c.serviceAuth != nil {
			if token, tokenErr := c.serviceAuth.GenerateToken(callCtx, service); tokenErr == nil && strings.TrimSpace(token) != "" {
				httpReq.Header.Set("X-Service-Token", token)
			}
		}

		if traceID, ok := supportcontract.FromTraceIDContext(callCtx); ok {
			httpReq.Header.Set("X-Trace-ID", traceID)
		}

		httpResp, callErr := c.httpCli.Do(httpReq)
		if callErr != nil {
			return fmt.Errorf("rpc: request failed: %w", callErr)
		}
		defer httpResp.Body.Close()

		if httpResp.StatusCode >= 400 {
			body, _ := io.ReadAll(httpResp.Body)
			return fmt.Errorf("rpc: server error %d: %s", httpResp.StatusCode, string(body))
		}

		if resp != nil {
			respBody, readErr := io.ReadAll(httpResp.Body)
			if readErr != nil {
				return fmt.Errorf("rpc: read response failed: %w", readErr)
			}
			if unmarshalErr := json.Unmarshal(respBody, resp); unmarshalErr != nil {
				return fmt.Errorf("rpc: unmarshal response failed: %w", unmarshalErr)
			}
		}

		return nil
	},
		rpcgovernance.TimeoutMiddleware(time.Duration(c.cfg.TimeoutMS)*time.Millisecond),
		rpcgovernance.RetryMiddlewareWithResource(c.retry, c.circuitBreakerResource),
	)
	err = c.doWithCircuitBreaker(ctx, service, method, func() error {
		return invoker(ctx, service, method, req, resp)
	})
	if err != nil {
		return err
	}

	return nil
}

// CallRaw invokes an HTTP RPC method with raw bytes payload.
// Implements transportcontract.RPCClient.CallRaw.
// Uses application/octet-stream content type and returns raw response bytes.
//
// CallRaw 使用原始字节负载调用 HTTP RPC 方法。
// 实现 transportcontract.RPCClient.CallRaw。
// 使用 application/octet-stream 内容类型并返回原始响应字节。
func (c *Client) CallRaw(ctx context.Context, service, method string, data []byte) ([]byte, error) {
	addr, done, err := c.resolveTarget(ctx, service)
	if err != nil {
		return nil, err
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

	var respBody []byte
	err = c.doWithCircuitBreaker(ctx, service, method, func() error {
		fullURL := c.buildURL(addr, method)

		httpReq, reqErr := http.NewRequestWithContext(ctx, http.MethodPost, fullURL, bytes.NewReader(data))
		if reqErr != nil {
			return reqErr
		}
		httpReq.Header.Set("Content-Type", "application/octet-stream")

		httpResp, callErr := c.httpCli.Do(httpReq)
		if callErr != nil {
			return callErr
		}
		defer httpResp.Body.Close()

		respBody, callErr = io.ReadAll(httpResp.Body)
		return callErr
	})
	if err != nil {
		return nil, err
	}

	return respBody, nil
}

// Close closes idle HTTP connections.
// Implements transportcontract.RPCClient.Close.
//
// Close 关闭空闲 HTTP 连接。
// 实现 transportcontract.RPCClient.Close。
func (c *Client) Close() error {
	c.httpCli.CloseIdleConnections()
	return nil
}

// resolveTarget resolves the target address using service discovery or fallback config.
// Returns address and optional DoneFunc for selector callback.
//
// resolveTarget 使用服务发现或回退配置解析目标地址。
// 返回地址和可选的 DoneFunc 用于选择器回调。
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

	if c.cfg.BaseURL != "" {
		return c.cfg.BaseURL, nil, nil
	}

	return fmt.Sprintf("http://%s", service), nil, nil
}

// doWithCircuitBreaker executes fn with circuit breaker protection if available.
//
// doWithCircuitBreaker 在熔断器保护下执行 fn（如果可用）。
func (c *Client) doWithCircuitBreaker(ctx context.Context, service, method string, fn func() error) error {
	if c.circuitBreaker == nil {
		return fn()
	}
	return c.circuitBreaker.Do(ctx, c.circuitBreakerResource(service, method), fn)
}

// circuitBreakerResource generates the resource name for circuit breaker.
// Format: "rpc.http.{service}.{method}" with sanitized segments.
//
// circuitBreakerResource 生成熔断器资源名称。
// 格式："rpc.http.{service}.{method}"，片段经过清理。
func (c *Client) circuitBreakerResource(service, method string) string {
	parts := []string{"rpc", "http"}
	if service != "" {
		parts = append(parts, sanitizeCircuitBreakerSegment(service))
	}
	if method != "" {
		parts = append(parts, sanitizeCircuitBreakerSegment(method))
	}
	return strings.Join(parts, ".")
}

// sanitizeCircuitBreakerSegment sanitizes a segment for circuit breaker resource name.
// Trims whitespace, removes leading/trailing slashes, replaces /, spaces, : with . or _.
//
// sanitizeCircuitBreakerSegment 清理熔断器资源名称的片段。
// 去除空白、移除前后斜杠、替换 /、空格、: 为 . 或 _。
func sanitizeCircuitBreakerSegment(segment string) string {
	segment = strings.TrimSpace(segment)
	segment = strings.Trim(segment, "/")
	if segment == "" {
		return "unknown"
	}
	replacer := strings.NewReplacer("/", ".", " ", "_", ":", ".")
	return replacer.Replace(segment)
}

// buildURL constructs the full URL from address and method path.
// Ensures method starts with "/" and address has http/https scheme.
//
// buildURL 从地址和方法路径构造完整 URL。
// 确保方法以 "/" 开头，地址有 http/https 协议。
func (c *Client) buildURL(addr, method string) string {
	if !strings.HasPrefix(method, "/") {
		method = "/" + method
	}

	if !strings.HasPrefix(addr, "http://") && !strings.HasPrefix(addr, "https://") {
		addr = "http://" + addr
	}

	u, err := url.Parse(addr)
	if err != nil {
		return addr + method
	}
	u.Path = method
	return u.String()
}

// headerCarrier implements transportcontract.TextMapCarrier for HTTP headers.
// Used for metadata propagation and tracing context injection/extraction.
//
// headerCarrier 实现 transportcontract.TextMapCarrier 用于 HTTP 头。
// 用于 metadata 传播和 tracing 上下文注入/提取。
type headerCarrier struct {
	header http.Header
}

func (c *headerCarrier) Get(key string) string {
	return c.header.Get(key)
}

func (c *headerCarrier) Set(key, value string) {
	c.header.Set(key, value)
}

func (c *headerCarrier) Add(key, value string) {
	c.header.Add(key, value)
}

func (c *headerCarrier) Keys() []string {
	keys := make([]string, 0, len(c.header))
	for k := range c.header {
		keys = append(keys, k)
	}
	return keys
}

func (c *headerCarrier) Values(key string) []string {
	return c.header.Values(key)
}