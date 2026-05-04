package eureka

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	internalnative "github.com/ngq/gorp/contrib/internal/native"
	"github.com/ngq/gorp/contrib/registry/internal/lifecycle"
	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

var (
	ErrNoServerURL       = errors.New("eureka: server_url is required")
	ErrServiceNotFound   = errors.New("eureka: service not found")
	ErrRegistryClosed    = errors.New("eureka: registry closed")
	ErrAlreadyRegistered = errors.New("eureka: instance already registered")
)

// Provider 提供 Eureka 服务发现实现。
//
// 中文说明：
//   - 使用 Netflix Eureka 实现服务注册与发现；
//   - 兼容 Spring Cloud 生态；
//   - 支持心跳健康检查。
//   - 当前状态：部分可用
//   - 说明：已完成 P2 第一版最小注册/发现闭环，具备 Register / Deregister / Discover 与 fake client 行为测试；
//     但当前仍未覆盖完整 Eureka 心跳与续租产品化语义。
type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string  { return "registry.eureka" }
func (p *Provider) IsDefer() bool { return true }
func (p *Provider) Provides() []string {
	return []string{transportcontract.RPCRegistryKey}
}

func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(transportcontract.RPCRegistryKey, func(c runtimecontract.Container) (any, error) {
		cfg, err := getEurekaConfig(c)
		if err != nil {
			return nil, err
		}
		return NewRegistry(cfg)
	}, true)

	return nil
}

func (p *Provider) Boot(c runtimecontract.Container) error { return nil }

type EurekaConfig struct {
	ServerURL             string
	AppName               string
	InstanceHost          string
	InstancePort          int
	ServiceMeta           map[string]string
	HeartbeatInterval     time.Duration
	HeartbeatRetryBackoff time.Duration
	WatchInterval         time.Duration
}

func getEurekaConfig(c runtimecontract.Container) (*EurekaConfig, error) {
	cfgAny, err := c.Make(datacontract.ConfigKey)
	if err != nil {
		return nil, err
	}

	cfg, ok := cfgAny.(datacontract.Config)
	if !ok {
		return nil, errors.New("eureka: invalid config service")
	}

	eurekaCfg := &EurekaConfig{
		HeartbeatRetryBackoff: time.Second,
		WatchInterval:         5 * time.Second,
	}

	if v := cfg.Get("discovery.eureka.server_url"); v != nil {
		eurekaCfg.ServerURL = cfg.GetString("discovery.eureka.server_url")
	}
	if v := cfg.Get("discovery.eureka.app_name"); v != nil {
		eurekaCfg.AppName = cfg.GetString("discovery.eureka.app_name")
	}
	if v := cfg.Get("discovery.eureka.instance_host"); v != nil {
		eurekaCfg.InstanceHost = cfg.GetString("discovery.eureka.instance_host")
	}
	if v := cfg.Get("discovery.eureka.instance_port"); v != nil {
		eurekaCfg.InstancePort = cfg.GetInt("discovery.eureka.instance_port")
	}
	if v := cfg.Get("discovery.eureka.heartbeat_interval_seconds"); v != nil {
		if seconds := cfg.GetInt("discovery.eureka.heartbeat_interval_seconds"); seconds > 0 {
			eurekaCfg.HeartbeatInterval = time.Duration(seconds) * time.Second
		}
	}
	if v := cfg.Get("discovery.eureka.heartbeat_retry_backoff_ms"); v != nil {
		if ms := cfg.GetInt("discovery.eureka.heartbeat_retry_backoff_ms"); ms > 0 {
			eurekaCfg.HeartbeatRetryBackoff = time.Duration(ms) * time.Millisecond
		}
	}
	if v := cfg.Get("discovery.eureka.watch_interval_ms"); v != nil {
		if ms := cfg.GetInt("discovery.eureka.watch_interval_ms"); ms > 0 {
			eurekaCfg.WatchInterval = time.Duration(ms) * time.Millisecond
		}
	}

	return eurekaCfg, nil
}

type eurekaClient interface {
	Register(ctx context.Context, cfg *EurekaConfig, name, addr string, meta map[string]string) error
	Deregister(ctx context.Context, cfg *EurekaConfig, name, addr string) error
	Heartbeat(ctx context.Context, cfg *EurekaConfig, name, addr string) error
	Discover(ctx context.Context, cfg *EurekaConfig, name string) ([]transportcontract.ServiceInstance, error)
	Watch(ctx context.Context, cfg *EurekaConfig, name string, onUpdate func([]transportcontract.ServiceInstance)) error
}

// HTTPClientProvider exposes the current HTTP transport object for native down-dive.
type HTTPClientProvider interface {
	HTTPClient() *http.Client
}

type Registry struct {
	config *EurekaConfig
	client eurekaClient

	mu            sync.RWMutex
	registered    map[string]map[string]string
	renewals      map[string]context.CancelFunc
	endpointCache map[string][]transportcontract.ServiceInstance
	watchCache    map[string]string
	closeMu       sync.Mutex
	closed        bool
	watchCancels  []context.CancelFunc
}

func NewRegistry(cfg *EurekaConfig) (*Registry, error) {
	return NewRegistryWithClient(cfg, newHTTPEurekaClient())
}

func NewRegistryWithClient(cfg *EurekaConfig, client eurekaClient) (*Registry, error) {
	if cfg.ServerURL == "" {
		return nil, ErrNoServerURL
	}
	if client == nil {
		return nil, errors.New("eureka: client is required")
	}

	return &Registry{
		config:        cfg,
		client:        client,
		registered:    make(map[string]map[string]string),
		renewals:      make(map[string]context.CancelFunc),
		endpointCache: make(map[string][]transportcontract.ServiceInstance),
		watchCache:    make(map[string]string),
	}, nil
}

func (r *Registry) Register(ctx context.Context, name, addr string, meta map[string]string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return ErrRegistryClosed
	}
	key := instanceKey(name, addr)
	if _, exists := r.registered[key]; exists {
		return ErrAlreadyRegistered
	}
	if err := r.client.Register(ctx, r.config, name, addr, meta); err != nil {
		return err
	}
	r.registered[key] = cloneStringMap(meta)
	delete(r.endpointCache, name)
	delete(r.watchCache, name)
	r.startHeartbeatLocked(name, addr)
	return nil
}

func (r *Registry) Deregister(ctx context.Context, name, addr string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return ErrRegistryClosed
	}
	if err := r.client.Deregister(ctx, r.config, name, addr); err != nil {
		return err
	}
	key := instanceKey(name, addr)
	if cancel, ok := r.renewals[key]; ok {
		cancel()
		delete(r.renewals, key)
	}
	delete(r.registered, key)
	delete(r.endpointCache, name)
	delete(r.watchCache, name)
	return nil
}

func (r *Registry) Discover(ctx context.Context, name string) ([]transportcontract.ServiceInstance, error) {
	r.mu.RLock()
	if instances, ok := r.endpointCache[name]; ok && len(instances) > 0 {
		cached := append([]transportcontract.ServiceInstance(nil), instances...)
		r.mu.RUnlock()
		return cached, nil
	}
	closed := r.closed
	r.mu.RUnlock()
	if closed {
		return nil, ErrRegistryClosed
	}

	instances, err := r.client.Discover(ctx, r.config, name)
	if err != nil {
		return nil, err
	}

	r.mu.Lock()
	r.endpointCache[name] = append([]transportcontract.ServiceInstance(nil), instances...)
	r.mu.Unlock()
	return instances, nil
}

// Underlying returns the current native client object used by this registry.
func (r *Registry) Underlying() any {
	return r.client
}

// As projects the current native client into the requested target when possible.
func (r *Registry) As(target any) bool {
	return internalnative.As(r.client, target)
}

func (r *Registry) Watch(ctx context.Context, name string) (<-chan []transportcontract.ServiceInstance, error) {
	r.closeMu.Lock()
	defer r.closeMu.Unlock()

	if r.closed {
		return nil, ErrRegistryClosed
	}

	watchCtx, cancel := context.WithCancel(ctx)
	r.watchCancels = append(r.watchCancels, cancel)

	ch := make(chan []transportcontract.ServiceInstance, 10)
	var workers sync.WaitGroup
	emit := func(instances []transportcontract.ServiceInstance) bool {
		key := snapshotKey(instances)

		r.mu.Lock()
		last := r.watchCache[name]
		if last == key {
			r.mu.Unlock()
			return true
		}
		r.watchCache[name] = key
		if len(instances) == 0 {
			delete(r.endpointCache, name)
		} else {
			r.endpointCache[name] = append([]transportcontract.ServiceInstance(nil), instances...)
		}
		r.mu.Unlock()

		select {
		case ch <- append([]transportcontract.ServiceInstance(nil), instances...):
			return true
		case <-watchCtx.Done():
			return false
		default:
			return true
		}
	}

	workers.Add(1)
	go func() {
		defer workers.Done()
		for {
			err := r.client.Watch(watchCtx, r.config, name, func(instances []transportcontract.ServiceInstance) {
				emit(instances)
			})
			if err == nil || watchCtx.Err() != nil {
				return
			}
			if !isRetryableWatchError(err) {
				return
			}
			select {
			case <-watchCtx.Done():
				return
			case <-time.After(watchRetryInterval(r.config)):
			}
		}
	}()

	workers.Add(1)
	go func() {
		defer workers.Done()
		instances, err := r.Discover(watchCtx, name)
		if err != nil {
			if errors.Is(err, ErrServiceNotFound) {
				emit(nil)
			}
			return
		}
		emit(instances)
	}()

	go func() {
		workers.Wait()
		close(ch)
	}()

	return ch, nil
}

func (r *Registry) Close() error {
	r.closeMu.Lock()
	defer r.closeMu.Unlock()

	if r.closed {
		return nil
	}
	r.closed = true
	for key, cancel := range r.renewals {
		cancel()
		delete(r.renewals, key)
	}
	for _, cancel := range r.watchCancels {
		cancel()
	}
	r.watchCancels = nil
	return nil
}

func (r *Registry) startHeartbeatLocked(name, addr string) {
	if r.config.HeartbeatInterval <= 0 {
		return
	}
	key := instanceKey(name, addr)
	if cancel, ok := r.renewals[key]; ok {
		cancel()
	}
	renewCtx, cancel := context.WithCancel(context.Background())
	r.renewals[key] = cancel

	go func() {
		lifecycle.RunHeartbeatLoop(renewCtx, lifecycle.HeartbeatLoopConfig{
			Interval:     r.config.HeartbeatInterval,
			RetryBackoff: r.config.HeartbeatRetryBackoff,
			Heartbeat: func(ctx context.Context) error {
				return r.client.Heartbeat(ctx, r.config, name, addr)
			},
			Recover: func(ctx context.Context) error {
				return r.recoverRegistration(ctx, name, addr)
			},
			ShouldRecover: func(err error) bool {
				return errors.Is(err, ErrServiceNotFound)
			},
		})
	}()
}

func (r *Registry) recoverRegistration(ctx context.Context, name, addr string) error {
	r.mu.RLock()
	if r.closed {
		r.mu.RUnlock()
		return ErrRegistryClosed
	}
	meta := cloneStringMap(r.registered[instanceKey(name, addr)])
	r.mu.RUnlock()
	return r.client.Register(ctx, r.config, name, addr, meta)
}

type httpEurekaClient struct {
	httpClient *http.Client
}

func newHTTPEurekaClient() eurekaClient {
	return &httpEurekaClient{
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *httpEurekaClient) HTTPClient() *http.Client {
	return c.httpClient
}

func (c *httpEurekaClient) Register(ctx context.Context, cfg *EurekaConfig, name, addr string, meta map[string]string) error {
	payload := map[string]any{
		"instance": map[string]any{
			"app":      name,
			"hostName": cfg.InstanceHost,
			"ipAddr":   hostFromAddr(addr),
			"port": map[string]any{
				"$":        portFromAddr(addr, cfg.InstancePort),
				"@enabled": true,
			},
			"vipAddress": name,
			"metadata":   mergeMeta(cfg.ServiceMeta, meta),
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(cfg.ServerURL, "/")+"/eureka/apps/"+url.PathEscape(strings.ToUpper(name)), bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("eureka: register failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("eureka: register failed with status %d", resp.StatusCode)
	}
	return nil
}

func (c *httpEurekaClient) Deregister(ctx context.Context, cfg *EurekaConfig, name, addr string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, strings.TrimRight(cfg.ServerURL, "/")+"/eureka/apps/"+url.PathEscape(strings.ToUpper(name))+"/"+url.PathEscape(instanceID(name, addr)), nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("eureka: deregister failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return ErrServiceNotFound
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("eureka: deregister failed with status %d", resp.StatusCode)
	}
	return nil
}

func (c *httpEurekaClient) Heartbeat(ctx context.Context, cfg *EurekaConfig, name, addr string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, strings.TrimRight(cfg.ServerURL, "/")+"/eureka/apps/"+url.PathEscape(strings.ToUpper(name))+"/"+url.PathEscape(instanceID(name, addr)), nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("eureka: heartbeat failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return ErrServiceNotFound
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("eureka: heartbeat failed with status %d", resp.StatusCode)
	}
	return nil
}

func (c *httpEurekaClient) Discover(ctx context.Context, cfg *EurekaConfig, name string) ([]transportcontract.ServiceInstance, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(cfg.ServerURL, "/")+"/eureka/apps/"+url.PathEscape(strings.ToUpper(name)), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("eureka: discover failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrServiceNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("eureka: discover failed with status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("eureka: read discover response failed: %w", err)
	}

	var payload eurekaDiscoverResponse
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("eureka: decode discover response failed: %w", err)
	}

	instances := payload.Application.Instance
	if len(instances) == 0 {
		return nil, ErrServiceNotFound
	}

	result := make([]transportcontract.ServiceInstance, 0, len(instances))
	for _, instance := range instances {
		port := instance.Port.Value
		if port == 0 {
			port = cfg.InstancePort
		}
		result = append(result, transportcontract.ServiceInstance{
			ID:       instance.InstanceID,
			Name:     strings.ToLower(instance.App),
			Address:  fmt.Sprintf("%s:%d", instance.IPAddr, port),
			Metadata: instance.Metadata,
			Healthy:  strings.EqualFold(instance.Status, "UP"),
		})
	}
	sortServiceInstances(result)
	return result, nil
}

func (c *httpEurekaClient) Watch(ctx context.Context, cfg *EurekaConfig, name string, onUpdate func([]transportcontract.ServiceInstance)) error {
	interval := cfg.WatchInterval
	if interval <= 0 {
		interval = 5 * time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	var last string
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			instances, err := c.Discover(ctx, cfg, name)
			if err != nil {
				if errors.Is(err, ErrServiceNotFound) {
					if last != "[]" {
						last = "[]"
						onUpdate([]transportcontract.ServiceInstance{})
					}
					continue
				}
				continue
			}

			payload, marshalErr := json.Marshal(instances)
			if marshalErr != nil {
				continue
			}
			current := string(payload)
			if current == last {
				continue
			}
			last = current
			onUpdate(instances)
		}
	}
}

type eurekaDiscoverResponse struct {
	Application struct {
		Instance []struct {
			InstanceID string            `json:"instanceId"`
			App        string            `json:"app"`
			IPAddr     string            `json:"ipAddr"`
			Status     string            `json:"status"`
			Metadata   map[string]string `json:"metadata"`
			Port       struct {
				Value int `json:"$"`
			} `json:"port"`
		} `json:"instance"`
	} `json:"application"`
}

func instanceKey(name, addr string) string {
	return name + "|" + addr
}

func instanceID(name, addr string) string {
	return strings.ToUpper(name) + ":" + addr
}

func hostFromAddr(addr string) string {
	parts := strings.Split(addr, ":")
	if len(parts) == 0 {
		return addr
	}
	return parts[0]
}

func portFromAddr(addr string, fallback int) int {
	parts := strings.Split(addr, ":")
	if len(parts) < 2 {
		return fallback
	}
	var port int
	_, _ = fmt.Sscanf(parts[len(parts)-1], "%d", &port)
	if port == 0 {
		return fallback
	}
	return port
}

func mergeMeta(base map[string]string, extra map[string]string) map[string]string {
	result := make(map[string]string, len(base)+len(extra))
	for k, v := range base {
		result[k] = v
	}
	for k, v := range extra {
		result[k] = v
	}
	return result
}

func cloneStringMap(input map[string]string) map[string]string {
	if input == nil {
		return nil
	}
	out := make(map[string]string, len(input))
	for k, v := range input {
		out[k] = v
	}
	return out
}

func sortServiceInstances(instances []transportcontract.ServiceInstance) {
	sort.Slice(instances, func(i, j int) bool {
		if instances[i].ID != instances[j].ID {
			return instances[i].ID < instances[j].ID
		}
		return instances[i].Address < instances[j].Address
	})
}

func snapshotKey(instances []transportcontract.ServiceInstance) string {
	if len(instances) == 0 {
		return "<empty>"
	}
	parts := make([]string, 0, len(instances))
	for _, instance := range instances {
		parts = append(parts, instance.ID+"|"+instance.Address+"|"+fmt.Sprintf("%t", instance.Healthy))
	}
	sort.Strings(parts)
	return strings.Join(parts, ";")
}

func watchRetryInterval(cfg *EurekaConfig) time.Duration {
	if cfg.WatchInterval > 0 {
		return cfg.WatchInterval
	}
	if cfg.HeartbeatRetryBackoff > 0 {
		return cfg.HeartbeatRetryBackoff
	}
	return time.Second
}

func isRetryableWatchError(err error) bool {
	if err == nil {
		return false
	}
	return !errors.Is(err, ErrRegistryClosed) && !errors.Is(err, context.Canceled)
}
