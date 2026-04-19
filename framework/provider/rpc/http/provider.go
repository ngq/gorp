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

	"github.com/ngq/gorp/framework/contract"
	configprovider "github.com/ngq/gorp/framework/provider/config"
)

// Provider 提供 HTTP RPC 实现。
//
// 中文说明：
// - 基于 HTTP REST API 实现服务间调用；
// - service 映射为 baseURL，method 映射为 path；
// - 支持服务发现集成（通过 Registry 获取目标地址）；
// - 无需 gRPC/protobuf，适合简单服务间通信。
type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string  { return "rpc.http" }
func (p *Provider) IsDefer() bool { return true }
func (p *Provider) Provides() []string {
	return []string{contract.RPCClientKey, contract.RPCServerKey}
}

func (p *Provider) Register(c contract.Container) error {
	// 注册 HTTP RPCClient
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

		// 获取 CircuitBreaker（如果可用）
		var circuitBreaker contract.CircuitBreaker
		if c.IsBind(contract.CircuitBreakerKey) {
			cbAny, _ := c.Make(contract.CircuitBreakerKey)
			circuitBreaker, _ = cbAny.(contract.CircuitBreaker)
		}

		return NewClient(cfg, registry, selector, metadataPropagator, serviceAuth, tracer, circuitBreaker), nil
	}, true)

	// 注册 HTTP RPCServer（复用 Gin Engine）
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
		return &contract.RPCConfig{Mode: "http"}, nil
	}

	cfg, ok := cfgAny.(contract.Config)
	if !ok {
		return &contract.RPCConfig{Mode: "http"}, nil
	}

	rpcCfg := &contract.RPCConfig{
		Mode:      "http",
		TimeoutMS: 30000, // 默认 30 秒
	}

	if mode := configprovider.GetStringAny(cfg,
		"rpc.mode",
	); mode != "" {
		rpcCfg.Mode = mode
	}
	if baseURL := configprovider.GetStringAny(cfg,
		"rpc.http.base_url",
		"rpc.base_url",
	); baseURL != "" {
		rpcCfg.BaseURL = baseURL
	}
	if timeout := configprovider.GetIntAny(cfg,
		"rpc.timeout_ms",
		"rpc.timeout",
	); timeout > 0 {
		rpcCfg.TimeoutMS = timeout
	}

	return rpcCfg, nil
}

// Client 是 HTTP RPC 客户端实现。
//
// 中文说明：
// - 使用标准 net/http 发起请求；
// - 支持服务发现：优先从 Registry 获取地址，否则使用 BaseURL；
// - 请求/响应使用 JSON 序列化；
// - 当启用 circuit_breaker 时，会在真正发起下游调用前经过统一保护。
type Client struct {
	cfg                *contract.RPCConfig
	registry           contract.ServiceRegistry
	selector           contract.Selector
	metadataPropagator contract.MetadataPropagator
	serviceAuth        contract.ServiceAuthenticator
	tracer             contract.Tracer
	circuitBreaker     contract.CircuitBreaker
	httpCli            *http.Client

	// 服务地址缓存（当前仅用于非 discovery fallback 场景）
	serviceCache sync.Map // map[string]*cachedAddr
}

type cachedAddr struct {
	addr     string
	expireAt time.Time
}

// NewClient 创建 HTTP RPC 客户端。
func NewClient(
	cfg *contract.RPCConfig,
	registry contract.ServiceRegistry,
	selector contract.Selector,
	metadataPropagator contract.MetadataPropagator,
	serviceAuth contract.ServiceAuthenticator,
	tracer contract.Tracer,
	circuitBreaker contract.CircuitBreaker,
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
		httpCli: &http.Client{
			Timeout: timeout,
		},
	}
}

// Call 执行 RPC 调用。
//
// 中文说明：
// - service: 目标服务名称（如 "user-service"）；
// - method: 方法路径（如 "/api/user/get"）；
// - 自动拼接完整 URL 并发送 POST 请求；
// - 支持服务发现优先选择地址；
// - selector 的 done 回调会在调用结束后统一回传；
// - 若启用了 circuit_breaker，则同一资源名会被统一纳入熔断保护。
func (c *Client) Call(ctx context.Context, service, method string, req, resp any) error {
	// 序列化请求
	reqBody, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("rpc: marshal request failed: %w", err)
	}

	// 获取目标地址（若走 selector，会返回 done callback）
	addr, done, err := c.resolveTarget(ctx, service)
	if err != nil {
		return fmt.Errorf("rpc: resolve address failed: %w", err)
	}
	if done != nil {
		defer func() {
			done(ctx, contract.DoneInfo{Err: err, BytesSent: true, BytesReceived: err == nil})
		}()
	}

	err = c.doWithCircuitBreaker(ctx, service, method, func() error {
		// 拼接完整 URL
		fullURL := c.buildURL(addr, method)

		// 发送请求
		httpReq, reqErr := http.NewRequestWithContext(ctx, http.MethodPost, fullURL, bytes.NewReader(reqBody))
		if reqErr != nil {
			return fmt.Errorf("rpc: create request failed: %w", reqErr)
		}
		httpReq.Header.Set("Content-Type", "application/json")

		// 注入 metadata（如果存在）
		if c.metadataPropagator != nil {
			carrier := &headerCarrier{header: httpReq.Header}
			c.metadataPropagator.Inject(ctx, carrier)
		}

		// 注入 tracing 上下文（如果存在）
		if c.tracer != nil {
			carrier := &headerCarrier{header: httpReq.Header}
			_ = c.tracer.Inject(ctx, carrier)
		}

		// 注入服务间认证令牌（如果启用）
		if c.serviceAuth != nil {
			if token, tokenErr := c.serviceAuth.GenerateToken(ctx, service); tokenErr == nil && strings.TrimSpace(token) != "" {
				httpReq.Header.Set("X-Service-Token", token)
			}
		}

		// 透传 TraceID（如果存在）
		if traceID := ctx.Value("trace_id"); traceID != nil {
			httpReq.Header.Set("X-Trace-ID", fmt.Sprintf("%v", traceID))
		}

		httpResp, callErr := c.httpCli.Do(httpReq)
		if callErr != nil {
			return fmt.Errorf("rpc: request failed: %w", callErr)
		}
		defer httpResp.Body.Close()

		// 检查响应状态
		if httpResp.StatusCode >= 400 {
			body, _ := io.ReadAll(httpResp.Body)
			return fmt.Errorf("rpc: server error %d: %s", httpResp.StatusCode, string(body))
		}

		// 反序列化响应
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
	})
	if err != nil {
		return err
	}

	return nil
}

// CallRaw 执行原始数据 RPC 调用。
func (c *Client) CallRaw(ctx context.Context, service, method string, data []byte) ([]byte, error) {
	addr, done, err := c.resolveTarget(ctx, service)
	if err != nil {
		return nil, err
	}
	if done != nil {
		defer func() {
			done(ctx, contract.DoneInfo{Err: err, BytesSent: true, BytesReceived: err == nil})
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

// Close 关闭客户端。
func (c *Client) Close() error {
	c.httpCli.CloseIdleConnections()
	return nil
}

// resolveTarget 解析服务目标地址。
//
// 中文说明：
// - 优先从 Registry 发现服务实例；
// - 如存在 Selector，则统一通过 Selector 选择实例；
// - 返回 done callback 给调用方在请求完成后回传结果；
// - 如果 Registry 不可用，则回退到 BaseURL 或默认服务地址。
func (c *Client) resolveTarget(ctx context.Context, service string) (string, contract.DoneFunc, error) {
	// discovery + selector 主链路
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

	// 回退到 BaseURL
	if c.cfg.BaseURL != "" {
		return c.cfg.BaseURL, nil, nil
	}

	// 默认服务地址
	return fmt.Sprintf("http://%s", service), nil, nil
}

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

func (c *Client) doWithCircuitBreaker(ctx context.Context, service, method string, fn func() error) error {
	if c.circuitBreaker == nil {
		return fn()
	}
	return c.circuitBreaker.Do(ctx, c.circuitBreakerResource(service, method), fn)
}

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

func sanitizeCircuitBreakerSegment(segment string) string {
	segment = strings.TrimSpace(segment)
	segment = strings.Trim(segment, "/")
	if segment == "" {
		return "unknown"
	}
	replacer := strings.NewReplacer("/", ".", " ", "_", ":", ".")
	return replacer.Replace(segment)
}

// buildURL 拼接完整 URL。
func (c *Client) buildURL(addr, method string) string {
	// 确保 method 以 / 开头
	if !strings.HasPrefix(method, "/") {
		method = "/" + method
	}

	// 确保 addr 有协议前缀
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

// Server 是 HTTP RPC 服务端实现。
//
// 中文说明：
// - 复用 Gin Engine 暴露 HTTP API；
// - 无需单独监听端口，与 HTTP 服务共享；
// - Register 将 handler 注册到 Gin 路由。
type Server struct {
	cfg    *contract.RPCConfig
	c      contract.Container
	addr   string
	routes sync.Map // map[string]any
}

// NewServer 创建 HTTP RPC 服务端。
func NewServer(cfg *contract.RPCConfig, c contract.Container) *Server {
	return &Server{
		cfg: cfg,
		c:   c,
	}
}

// Register 注册服务处理器。
//
// 中文说明：
// - handler 类型应为 gin.HandlerFunc；
// - 实际注册在 Gin Engine 上，路径为 /rpc/{service}/{method}。
func (s *Server) Register(service string, handler any) error {
	s.routes.Store(service, handler)
	return nil
}

// Start 启动 RPC 服务。
//
// 中文说明：
// - HTTP RPC 不需要单独启动，复用 HTTP 服务；
// - 返回 nil 表示成功。
func (s *Server) Start(ctx context.Context) error {
	// HTTP RPC 复用 Gin Engine，无需单独启动
	// 注册路由到 Gin
	if s.c.IsBind(contract.HTTPEngineKey) {
		engineAny, _ := s.c.Make(contract.HTTPEngineKey)
		// 使用类型断言处理 gin.Engine
		if engine, ok := engineAny.(interface{ POST(string, any) }); ok {
			s.routes.Range(func(key, value any) bool {
				service := key.(string)
				handler := value
				// 注册到 /rpc/{service}/ 路径
				engine.POST("/rpc/"+service, handler)
				return true
			})
		}
	}
	return nil
}

// Stop 停止 RPC 服务。
func (s *Server) Stop(ctx context.Context) error {
	// HTTP RPC 随 HTTP 服务关闭，无需单独处理
	return nil
}

// Addr 返回服务地址。
func (s *Server) Addr() string {
	if s.addr != "" {
		return s.addr
	}
	// 从配置获取 HTTP 服务地址
	if s.c.IsBind(contract.ConfigKey) {
		cfgAny, _ := s.c.Make(contract.ConfigKey)
		if cfg, ok := cfgAny.(contract.Config); ok {
			return cfg.GetString("app.address")
		}
	}
	return ":8080"
}
