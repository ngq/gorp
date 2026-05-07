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

	datacontract "github.com/ngq/gorp/framework/contract/data"
	discoverycontract "github.com/ngq/gorp/framework/contract/discovery"
	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	securitycontract "github.com/ngq/gorp/framework/contract/security"
	supportcontract "github.com/ngq/gorp/framework/contract/support"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	configprovider "github.com/ngq/gorp/framework/provider/config"
	rpcgovernance "github.com/ngq/gorp/framework/rpc/governance"
)

type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string  { return "rpc.http" }
func (p *Provider) IsDefer() bool { return true }
func (p *Provider) Provides() []string {
	return []string{transportcontract.RPCClientKey, transportcontract.RPCServerKey}
}

func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(transportcontract.RPCClientKey, func(c runtimecontract.Container) (any, error) {
		cfg, _ := getConfig(c)

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

		return NewClient(cfg, registry, selector, metadataPropagator, serviceAuth, tracer, circuitBreaker, retry), nil
	}, true)

	c.Bind(transportcontract.RPCServerKey, func(c runtimecontract.Container) (any, error) {
		cfg, _ := getConfig(c)
		return NewServer(cfg, c), nil
	}, true)

	return nil
}

func (p *Provider) Boot(c runtimecontract.Container) error {
	return nil
}

func getConfig(c runtimecontract.Container) (*transportcontract.RPCConfig, error) {
	cfgAny, err := c.Make(datacontract.ConfigKey)
	if err != nil {
		return &transportcontract.RPCConfig{Mode: "http"}, nil
	}

	cfg, ok := cfgAny.(datacontract.Config)
	if !ok {
		return &transportcontract.RPCConfig{Mode: "http"}, nil
	}

	rpcCfg := &transportcontract.RPCConfig{
		Mode:      "http",
		TimeoutMS: 30000,
	}

	if mode := configprovider.GetStringAny(cfg, "rpc.mode"); mode != "" {
		rpcCfg.Mode = mode
	}
	if baseURL := configprovider.GetStringAny(cfg, "rpc.http.base_url", "rpc.base_url"); baseURL != "" {
		rpcCfg.BaseURL = baseURL
	}
	if timeout := configprovider.GetIntAny(cfg, "rpc.timeout_ms", "rpc.timeout"); timeout > 0 {
		rpcCfg.TimeoutMS = timeout
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
	httpCli            *http.Client
	serviceCache       sync.Map
}

type cachedAddr struct {
	addr     string
	expireAt time.Time
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

func (c *Client) Close() error {
	c.httpCli.CloseIdleConnections()
	return nil
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

	if c.cfg.BaseURL != "" {
		return c.cfg.BaseURL, nil, nil
	}

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

type Server struct {
	cfg    *transportcontract.RPCConfig
	c      runtimecontract.Container
	addr   string
	routes sync.Map
}

func NewServer(cfg *transportcontract.RPCConfig, c runtimecontract.Container) *Server {
	return &Server{
		cfg: cfg,
		c:   c,
	}
}

func (s *Server) Register(service string, handler any) error {
	s.routes.Store(service, handler)
	return nil
}

func (s *Server) Start(ctx context.Context) error {
	httpSvc, err := s.c.Make(transportcontract.HTTPKey)
	if err != nil {
		return nil
	}
	httpServer, ok := httpSvc.(transportcontract.HTTP)
	if !ok || httpServer == nil {
		return nil
	}
	router := httpServer.Router()
	if router == nil {
		return nil
	}
	s.routes.Range(func(key, value any) bool {
		service := key.(string)
		handler, ok := value.(transportcontract.HTTPHandler)
		if !ok || handler == nil {
			return true
		}
		router.POST("/rpc/"+service, handler)
		return true
	})
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	return nil
}

func (s *Server) Addr() string {
	if s.addr != "" {
		return s.addr
	}
	if s.c.IsBind(datacontract.ConfigKey) {
		cfgAny, _ := s.c.Make(datacontract.ConfigKey)
		if cfg, ok := cfgAny.(datacontract.Config); ok {
			return cfg.GetString("app.address")
		}
	}
	return ":8080"
}
