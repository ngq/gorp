package servicecomb

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	internalnative "github.com/ngq/gorp/contrib/internal/native"
	"github.com/ngq/gorp/contrib/registry/internal/lifecycle"
	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

var (
	ErrServerURIRequired = errors.New("servicecomb: server_uri is required")
	ErrAppIDRequired     = errors.New("servicecomb: app_id is required")
	ErrServiceNotFound   = errors.New("servicecomb: service not found")
	ErrRegistryClosed    = errors.New("servicecomb: registry closed")
	ErrAlreadyRegistered = errors.New("servicecomb: instance already registered")
)

// Provider 提供 ServiceComb 服务发现实现。
//
// 中文说明：
//   - 使用华为 ServiceComb ServiceCenter 实现服务注册与发现；
//   - 兼容 Spring Cloud Huawei 生态；
//   - 支持多环境（开发/测试/生产）；
//   - 支持服务元数据和标签。
//   - 当前状态：部分可用
//   - 说明：已完成 P2 第一版最小注册/发现闭环，具备 Register / Deregister / Discover 与 fake client 行为测试；
//     但当前仍未覆盖完整心跳、实例治理与 SDK 产品化语义。
type Provider struct{}

func NewProvider() *Provider           { return &Provider{} }
func (p *Provider) Name() string       { return "registry.servicecomb" }
func (p *Provider) IsDefer() bool      { return true }
func (p *Provider) Provides() []string { return []string{transportcontract.RPCRegistryKey} }

func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(transportcontract.RPCRegistryKey, func(c runtimecontract.Container) (any, error) {
		cfg, err := getServiceCombConfig(c)
		if err != nil {
			return nil, err
		}
		return NewRegistry(cfg)
	}, true)
	return nil
}

func (p *Provider) Boot(c runtimecontract.Container) error { return nil }

type ServiceCombConfig struct {
	ServerURI             string
	AppID                 string
	ServiceName           string
	Version               string
	Environment           string
	InstanceHost          string
	InstancePort          int
	ServiceMeta           map[string]string
	Tags                  []string
	HeartbeatInterval     time.Duration
	HeartbeatRetryBackoff time.Duration
	WatchInterval         time.Duration
}

func getServiceCombConfig(c runtimecontract.Container) (*ServiceCombConfig, error) {
	cfgAny, err := c.Make(datacontract.ConfigKey)
	if err != nil {
		return nil, err
	}
	cfg, ok := cfgAny.(datacontract.Config)
	if !ok {
		return nil, errors.New("servicecomb: invalid config service")
	}

	servicecombCfg := &ServiceCombConfig{
		Version:               "1.0.0",
		Environment:           "production",
		HeartbeatRetryBackoff: time.Second,
		WatchInterval:         time.Second,
	}
	if v := cfg.Get("discovery.servicecomb.server_uri"); v != nil {
		servicecombCfg.ServerURI = cfg.GetString("discovery.servicecomb.server_uri")
	}
	if v := cfg.Get("discovery.servicecomb.app_id"); v != nil {
		servicecombCfg.AppID = cfg.GetString("discovery.servicecomb.app_id")
	}
	if v := cfg.Get("discovery.servicecomb.service_name"); v != nil {
		servicecombCfg.ServiceName = cfg.GetString("discovery.servicecomb.service_name")
	}
	if v := cfg.Get("discovery.servicecomb.version"); v != nil {
		servicecombCfg.Version = cfg.GetString("discovery.servicecomb.version")
	}
	if v := cfg.Get("discovery.servicecomb.environment"); v != nil {
		servicecombCfg.Environment = cfg.GetString("discovery.servicecomb.environment")
	}
	if v := cfg.Get("discovery.servicecomb.heartbeat_interval_seconds"); v != nil {
		if seconds := cfg.GetInt("discovery.servicecomb.heartbeat_interval_seconds"); seconds > 0 {
			servicecombCfg.HeartbeatInterval = time.Duration(seconds) * time.Second
		}
	}
	if v := cfg.Get("discovery.servicecomb.heartbeat_retry_backoff_ms"); v != nil {
		if ms := cfg.GetInt("discovery.servicecomb.heartbeat_retry_backoff_ms"); ms > 0 {
			servicecombCfg.HeartbeatRetryBackoff = time.Duration(ms) * time.Millisecond
		}
	}
	if v := cfg.Get("discovery.servicecomb.watch_interval_ms"); v != nil {
		if ms := cfg.GetInt("discovery.servicecomb.watch_interval_ms"); ms > 0 {
			servicecombCfg.WatchInterval = time.Duration(ms) * time.Millisecond
		}
	}
	return servicecombCfg, nil
}

type serviceCombClient interface {
	Register(ctx context.Context, cfg *ServiceCombConfig, name, addr string, meta map[string]string) error
	Deregister(ctx context.Context, cfg *ServiceCombConfig, name, addr string) error
	Heartbeat(ctx context.Context, cfg *ServiceCombConfig, name, addr string) error
	Discover(ctx context.Context, cfg *ServiceCombConfig, name string) ([]transportcontract.ServiceInstance, error)
}

type nativeClientProvider interface {
	Underlying() any
}

type Registry struct {
	config *ServiceCombConfig
	client serviceCombClient

	mu            sync.RWMutex
	registered    map[string]map[string]string
	renewals      map[string]context.CancelFunc
	endpointCache map[string][]transportcontract.ServiceInstance
	watchCache    map[string]string
	closed        bool
	closeMu       sync.Mutex
	watchCancels  []context.CancelFunc
}

func NewRegistry(cfg *ServiceCombConfig) (*Registry, error) {
	return NewRegistryWithClient(cfg, &inMemoryServiceCombClient{})
}

func NewRegistryWithClient(cfg *ServiceCombConfig, client serviceCombClient) (*Registry, error) {
	if cfg.ServerURI == "" {
		return nil, ErrServerURIRequired
	}
	if cfg.AppID == "" {
		return nil, ErrAppIDRequired
	}
	if client == nil {
		return nil, errors.New("servicecomb: client is required")
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
	key := name + "|" + addr
	if _, exists := r.registered[key]; exists {
		return ErrAlreadyRegistered
	}
	if err := r.client.Register(ctx, r.config, name, addr, meta); err != nil {
		return err
	}
	r.registered[key] = cloneStringMap(meta)
	delete(r.endpointCache, name)
	delete(r.watchCache, name)
	r.startHeartbeatLocked(name, addr, meta)
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
	key := name + "|" + addr
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
	sortServiceInstances(instances)
	r.mu.Lock()
	r.endpointCache[name] = append([]transportcontract.ServiceInstance(nil), instances...)
	r.mu.Unlock()
	return instances, nil
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

func (r *Registry) Underlying() any {
	if provider, ok := r.client.(nativeClientProvider); ok {
		if native := provider.Underlying(); native != nil {
			return native
		}
	}
	return r.client
}

func (r *Registry) As(target any) bool {
	return internalnative.As(r.Underlying(), target)
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
	emit := func(instances []transportcontract.ServiceInstance) bool {
		current := snapshotKey(instances)

		r.mu.Lock()
		last := r.watchCache[name]
		if current == last {
			r.mu.Unlock()
			return true
		}
		r.watchCache[name] = current
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

	go func() {
		defer close(ch)

		instances, err := r.Discover(watchCtx, name)
		if err != nil {
			if errors.Is(err, ErrServiceNotFound) {
				emit(nil)
			}
			return
		}
		emit(instances)

		ticker := time.NewTicker(r.watchInterval())
		defer ticker.Stop()

		for {
			select {
			case <-watchCtx.Done():
				return
			case <-ticker.C:
				latest, err := r.client.Discover(watchCtx, r.config, name)
				if err != nil {
					if errors.Is(err, ErrServiceNotFound) {
						r.mu.Lock()
						delete(r.endpointCache, name)
						r.mu.Unlock()
						emit(nil)
					}
					continue
				}
				sortServiceInstances(latest)
				emit(latest)
			}
		}
	}()

	return ch, nil
}

func (r *Registry) startHeartbeatLocked(name, addr string, meta map[string]string) {
	if r.config.HeartbeatInterval <= 0 {
		return
	}
	key := name + "|" + addr
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
				return r.recoverRegistration(name, addr)
			},
			ShouldRecover: func(err error) bool {
				return errors.Is(err, ErrServiceNotFound)
			},
		})
	}()
}

func (r *Registry) recoverRegistration(name, addr string) error {
	r.mu.RLock()
	if r.closed {
		r.mu.RUnlock()
		return ErrRegistryClosed
	}
	meta := cloneStringMap(r.registered[name+"|"+addr])
	r.mu.RUnlock()
	return r.client.Register(context.Background(), r.config, name, addr, meta)
}

type inMemoryServiceCombClient struct {
	mu       sync.RWMutex
	services map[string][]transportcontract.ServiceInstance
}

func (c *inMemoryServiceCombClient) Underlying() any {
	return c
}

func (c *inMemoryServiceCombClient) Register(ctx context.Context, cfg *ServiceCombConfig, name, addr string, meta map[string]string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.services == nil {
		c.services = make(map[string][]transportcontract.ServiceInstance)
	}
	fullMeta := make(map[string]string)
	for k, v := range cfg.ServiceMeta {
		fullMeta[k] = v
	}
	for k, v := range meta {
		fullMeta[k] = v
	}
	fullMeta["version"] = cfg.Version
	fullMeta["environment"] = cfg.Environment
	instance := transportcontract.ServiceInstance{
		ID:       generateInstanceID(name, addr),
		Name:     name,
		Address:  addr,
		Metadata: fullMeta,
		Healthy:  true,
	}
	c.services[name] = append(c.services[name], instance)
	return nil
}

func (c *inMemoryServiceCombClient) Deregister(ctx context.Context, cfg *ServiceCombConfig, name, addr string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	instances := c.services[name]
	for i, inst := range instances {
		if inst.Address == addr {
			c.services[name] = append(instances[:i], instances[i+1:]...)
			return nil
		}
	}
	return ErrServiceNotFound
}

func (c *inMemoryServiceCombClient) Heartbeat(ctx context.Context, cfg *ServiceCombConfig, name, addr string) error {
	return nil
}

func (c *inMemoryServiceCombClient) Discover(ctx context.Context, cfg *ServiceCombConfig, name string) ([]transportcontract.ServiceInstance, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	instances := c.services[name]
	if len(instances) == 0 {
		return nil, ErrServiceNotFound
	}
	result := make([]transportcontract.ServiceInstance, len(instances))
	copy(result, instances)
	sortServiceInstances(result)
	return result, nil
}

func generateInstanceID(name, addr string) string {
	return fmt.Sprintf("%s-%s", name, addr)
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

func (r *Registry) watchInterval() time.Duration {
	if r.config.WatchInterval > 0 {
		return r.config.WatchInterval
	}
	if r.config.HeartbeatRetryBackoff > 0 {
		return r.config.HeartbeatRetryBackoff
	}
	if r.config.HeartbeatInterval > 0 {
		return r.config.HeartbeatInterval
	}
	return time.Second
}

func snapshotKey(instances []transportcontract.ServiceInstance) string {
	if len(instances) == 0 {
		return "<empty>"
	}
	parts := make([]string, 0, len(instances))
	for _, inst := range instances {
		parts = append(parts, inst.ID+"|"+inst.Address+"|"+fmt.Sprintf("%t", inst.Healthy))
	}
	sort.Strings(parts)
	result := ""
	for _, part := range parts {
		result += part + ";"
	}
	return result
}

func sortServiceInstances(instances []transportcontract.ServiceInstance) {
	sort.Slice(instances, func(i, j int) bool {
		if instances[i].ID != instances[j].ID {
			return instances[i].ID < instances[j].ID
		}
		return instances[i].Address < instances[j].Address
	})
}
