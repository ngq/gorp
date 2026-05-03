package polaris

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	internalnative "github.com/ngq/gorp/contrib/internal/native"
	"github.com/ngq/gorp/framework/contract"
	polarissdk "github.com/polarismesh/polaris-go"
	polarismodel "github.com/polarismesh/polaris-go/pkg/model"
)

var (
	ErrNoAddress         = errors.New("polaris: address is required")
	ErrServiceNotFound   = errors.New("polaris: service not found")
	ErrRegistryClosed    = errors.New("polaris: registry closed")
	ErrAlreadyRegistered = errors.New("polaris: instance already registered")
)

// Provider 提供 Polaris 服务发现实现。
//
// 中文说明：
//   - 使用腾讯云 Polaris 实现服务注册与发现；
//   - 支持服务路由、负载均衡、熔断；
//   - 适用于腾讯云环境和私有化部署。
//   - 当前状态：部分可用
//   - 说明：已完成 P2 第一版最小注册/发现闭环，具备 Register / Deregister / Discover 与 fake client 行为测试；
//     但当前仍未覆盖完整路由规则、健康检查与 SDK 产品化语义。
type Provider struct{}

func NewProvider() *Provider           { return &Provider{} }
func (p *Provider) Name() string       { return "registry.polaris" }
func (p *Provider) IsDefer() bool      { return true }
func (p *Provider) Provides() []string { return []string{contract.RPCRegistryKey} }

func (p *Provider) Register(c contract.Container) error {
	c.Bind(contract.RPCRegistryKey, func(c contract.Container) (any, error) {
		cfg, err := getPolarisConfig(c)
		if err != nil {
			return nil, err
		}
		return NewRegistry(cfg)
	}, true)
	return nil
}

func (p *Provider) Boot(c contract.Container) error { return nil }

type PolarisConfig struct {
	Address            string
	Namespace          string
	Token              string
	ServiceName        string
	ServiceAddr        string
	ServicePort        int
	ServiceMeta        map[string]string
	WatchRetryInterval time.Duration
}

func getPolarisConfig(c contract.Container) (*PolarisConfig, error) {
	cfgAny, err := c.Make(contract.ConfigKey)
	if err != nil {
		return nil, err
	}
	cfg, ok := cfgAny.(contract.Config)
	if !ok {
		return nil, errors.New("polaris: invalid config service")
	}

	polarisCfg := &PolarisConfig{
		Namespace:          "default",
		WatchRetryInterval: 200 * time.Millisecond,
	}
	if v := cfg.Get("discovery.polaris.address"); v != nil {
		polarisCfg.Address = cfg.GetString("discovery.polaris.address")
	}
	if v := cfg.Get("discovery.polaris.namespace"); v != nil {
		polarisCfg.Namespace = cfg.GetString("discovery.polaris.namespace")
	}
	if v := cfg.Get("discovery.polaris.token"); v != nil {
		polarisCfg.Token = cfg.GetString("discovery.polaris.token")
	}
	if v := cfg.Get("discovery.polaris.watch_retry_interval_ms"); v != nil {
		if ms := cfg.GetInt("discovery.polaris.watch_retry_interval_ms"); ms > 0 {
			polarisCfg.WatchRetryInterval = time.Duration(ms) * time.Millisecond
		}
	}
	return polarisCfg, nil
}

type polarisRegistryClient interface {
	Register(ctx context.Context, cfg *PolarisConfig, name, addr string, meta map[string]string) error
	Deregister(ctx context.Context, cfg *PolarisConfig, name, addr string) error
	Discover(ctx context.Context, cfg *PolarisConfig, name string) ([]contract.ServiceInstance, error)
	Watch(ctx context.Context, cfg *PolarisConfig, name string, onUpdate func([]contract.ServiceInstance)) error
}

type polarisRegistryNativeClient interface {
	Underlying() any
}

type Registry struct {
	config *PolarisConfig
	client polarisRegistryClient

	mu            sync.RWMutex
	endpointCache map[string][]contract.ServiceInstance
	registered    map[string]struct{}
	closeMu       sync.Mutex
	closed        bool
	watchCancels  []context.CancelFunc
}

func NewRegistry(cfg *PolarisConfig) (*Registry, error) {
	return NewRegistryWithClient(cfg, newOfficialPolarisRegistryClient())
}

func NewRegistryWithClient(cfg *PolarisConfig, client polarisRegistryClient) (*Registry, error) {
	if cfg.Address == "" {
		return nil, ErrNoAddress
	}
	if client == nil {
		return nil, errors.New("polaris: registry client is required")
	}
	return &Registry{
		config:        cfg,
		client:        client,
		endpointCache: make(map[string][]contract.ServiceInstance),
		registered:    make(map[string]struct{}),
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
	r.registered[key] = struct{}{}
	delete(r.endpointCache, name)
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
	delete(r.registered, name+"|"+addr)
	delete(r.endpointCache, name)
	return nil
}

func (r *Registry) Discover(ctx context.Context, name string) ([]contract.ServiceInstance, error) {
	r.mu.RLock()
	if instances, ok := r.endpointCache[name]; ok && len(instances) > 0 {
		cached := append([]contract.ServiceInstance(nil), instances...)
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
	r.endpointCache[name] = append([]contract.ServiceInstance(nil), instances...)
	r.mu.Unlock()
	return instances, nil
}

func (r *Registry) Watch(ctx context.Context, name string) (<-chan []contract.ServiceInstance, error) {
	r.closeMu.Lock()
	defer r.closeMu.Unlock()

	if r.closed {
		return nil, ErrRegistryClosed
	}

	watchCtx, cancel := context.WithCancel(ctx)
	r.watchCancels = append(r.watchCancels, cancel)

	ch := make(chan []contract.ServiceInstance, 10)
	var emitMu sync.Mutex
	lastSnapshot := ""
	emit := func(instances []contract.ServiceInstance) {
		sorted := append([]contract.ServiceInstance(nil), instances...)
		sortServiceInstances(sorted)
		current := snapshotKey(sorted)

		emitMu.Lock()
		if current == lastSnapshot {
			emitMu.Unlock()
			return
		}
		lastSnapshot = current
		emitMu.Unlock()

		r.mu.Lock()
		if len(sorted) == 0 {
			delete(r.endpointCache, name)
		} else {
			r.endpointCache[name] = append([]contract.ServiceInstance(nil), sorted...)
		}
		r.mu.Unlock()

		select {
		case ch <- append([]contract.ServiceInstance(nil), sorted...):
		case <-watchCtx.Done():
		default:
		}
	}

	go func() {
		defer close(ch)
		for {
			err := r.client.Watch(watchCtx, r.config, name, func(instances []contract.ServiceInstance) {
				emit(instances)
			})
			if err == nil || watchCtx.Err() != nil {
				return
			}
			select {
			case <-watchCtx.Done():
				return
			case <-time.After(r.config.WatchRetryInterval):
			}
		}
	}()

	go func() {
		instances, err := r.Discover(watchCtx, name)
		if err != nil {
			if errors.Is(err, ErrServiceNotFound) {
				emit([]contract.ServiceInstance{})
			}
			return
		}
		emit(instances)
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
	for _, cancel := range r.watchCancels {
		cancel()
	}
	r.watchCancels = nil
	if closer, ok := r.client.(interface{ Close() error }); ok {
		return closer.Close()
	}
	return nil
}

func (r *Registry) Underlying() any {
	if provider, ok := r.client.(polarisRegistryNativeClient); ok {
		if native := provider.Underlying(); native != nil {
			return native
		}
	}
	return r.client
}

func (r *Registry) As(target any) bool {
	return internalnative.As(r.Underlying(), target)
}

type officialPolarisRegistryClient struct {
	mu       sync.Mutex
	context  any
	provider polarissdk.ProviderAPI
	consumer polarissdk.ConsumerAPI
}

type polarisRegistryNative struct {
	Context  any
	Provider polarissdk.ProviderAPI
	Consumer polarissdk.ConsumerAPI
}

func newOfficialPolarisRegistryClient() polarisRegistryClient {
	return &officialPolarisRegistryClient{}
}

func (c *officialPolarisRegistryClient) Underlying() any {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.provider == nil && c.consumer == nil {
		return nil
	}
	return &polarisRegistryNative{
		Context:  c.context,
		Provider: c.provider,
		Consumer: c.consumer,
	}
}

func (c *officialPolarisRegistryClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.provider != nil {
		c.provider.Destroy()
	}
	if c.consumer != nil {
		c.consumer.Destroy()
	}
	if destroyer, ok := c.context.(interface{ Destroy() }); ok {
		destroyer.Destroy()
	}
	c.context = nil
	c.provider = nil
	c.consumer = nil
	return nil
}

func (c *officialPolarisRegistryClient) Register(ctx context.Context, cfg *PolarisConfig, name, addr string, meta map[string]string) error {
	provider, _, err := c.ensureClients(cfg)
	if err != nil {
		return err
	}
	host, port, err := splitHostPort(addr)
	if err != nil {
		return err
	}

	request := &polarissdk.InstanceRegisterRequest{}
	request.Service = name
	request.ServiceToken = cfg.Token
	request.Namespace = cfg.Namespace
	request.Host = host
	request.Port = port
	request.Metadata = mergePolarisMetadata(cfg.ServiceMeta, meta)
	request.SetHealthy(true)
	_, err = provider.RegisterInstance(request)
	return translatePolarisRegistryError(err)
}

func (c *officialPolarisRegistryClient) Deregister(ctx context.Context, cfg *PolarisConfig, name, addr string) error {
	provider, _, err := c.ensureClients(cfg)
	if err != nil {
		return err
	}
	host, port, err := splitHostPort(addr)
	if err != nil {
		return err
	}

	request := &polarissdk.InstanceDeRegisterRequest{}
	request.Service = name
	request.ServiceToken = cfg.Token
	request.Namespace = cfg.Namespace
	request.Host = host
	request.Port = port
	err = provider.Deregister(request)
	return translatePolarisRegistryError(err)
}

func (c *officialPolarisRegistryClient) Discover(ctx context.Context, cfg *PolarisConfig, name string) ([]contract.ServiceInstance, error) {
	_, consumer, err := c.ensureClients(cfg)
	if err != nil {
		return nil, err
	}
	request := &polarissdk.GetInstancesRequest{}
	request.Namespace = cfg.Namespace
	request.Service = name
	response, err := consumer.GetInstances(request)
	if err != nil {
		return nil, translatePolarisRegistryError(err)
	}
	instances := polarisInstancesToContract(response)
	if len(instances) == 0 {
		return nil, ErrServiceNotFound
	}
	return instances, nil
}

func (c *officialPolarisRegistryClient) Watch(ctx context.Context, cfg *PolarisConfig, name string, onUpdate func([]contract.ServiceInstance)) error {
	_, consumer, err := c.ensureClients(cfg)
	if err != nil {
		return err
	}
	request := &polarissdk.WatchServiceRequest{}
	request.Key = polarismodel.ServiceKey{
		Namespace: cfg.Namespace,
		Service:   name,
	}
	response, err := consumer.WatchService(request)
	if err != nil {
		return translatePolarisRegistryError(err)
	}
	if response != nil && response.GetAllInstancesResp != nil {
		onUpdate(polarisInstancesToContract(response.GetAllInstancesResp))
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case event, ok := <-response.EventChannel:
			if !ok {
				return nil
			}
			if event == nil {
				continue
			}
			instances, convErr := c.Discover(ctx, cfg, name)
			if convErr != nil {
				if errors.Is(convErr, ErrServiceNotFound) {
					onUpdate([]contract.ServiceInstance{})
					continue
				}
				return convErr
			}
			onUpdate(instances)
		}
	}
}

func (c *officialPolarisRegistryClient) ensureClients(cfg *PolarisConfig) (polarissdk.ProviderAPI, polarissdk.ConsumerAPI, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.provider != nil && c.consumer != nil {
		return c.provider, c.consumer, nil
	}

	addresses, err := normalizeRegistryPolarisAddresses(cfg.Address)
	if err != nil {
		return nil, nil, err
	}
	context, err := polarissdk.NewSDKContextByAddress(addresses...)
	if err != nil {
		return nil, nil, translatePolarisRegistryError(err)
	}
	c.context = context
	c.provider = polarissdk.NewProviderAPIByContext(context)
	c.consumer = polarissdk.NewConsumerAPIByContext(context)
	return c.provider, c.consumer, nil
}

func splitHostPort(addr string) (string, int, error) {
	host, portString, err := net.SplitHostPort(addr)
	if err != nil {
		return "", 0, fmt.Errorf("polaris: invalid address %q: %w", addr, err)
	}
	port, err := net.LookupPort("tcp", portString)
	if err != nil {
		return "", 0, fmt.Errorf("polaris: invalid port %q: %w", portString, err)
	}
	return host, port, nil
}

func mergePolarisMetadata(base map[string]string, override map[string]string) map[string]string {
	if len(base) == 0 && len(override) == 0 {
		return nil
	}
	merged := make(map[string]string, len(base)+len(override))
	for k, v := range base {
		merged[k] = v
	}
	for k, v := range override {
		merged[k] = v
	}
	return merged
}

func polarisInstancesToContract(response *polarismodel.InstancesResponse) []contract.ServiceInstance {
	if response == nil {
		return nil
	}
	sourceInstances := response.GetInstances()
	instances := make([]contract.ServiceInstance, 0, len(sourceInstances))
	for _, instance := range sourceInstances {
		if instance == nil {
			continue
		}
		instances = append(instances, contract.ServiceInstance{
			ID:       instance.GetId(),
			Name:     instance.GetService(),
			Address:  fmt.Sprintf("%s:%d", instance.GetHost(), instance.GetPort()),
			Metadata: instance.GetMetadata(),
			Healthy:  instance.IsHealthy(),
		})
	}
	sortServiceInstances(instances)
	return instances
}

func normalizeRegistryPolarisAddresses(raw string) ([]string, error) {
	parts := strings.Split(raw, ",")
	addresses := make([]string, 0, len(parts))
	for _, part := range parts {
		candidate := strings.TrimSpace(part)
		if candidate == "" {
			continue
		}
		if strings.Contains(candidate, "://") {
			parsed, err := url.Parse(candidate)
			if err != nil {
				return nil, fmt.Errorf("polaris: invalid address: %w", err)
			}
			if parsed.Host != "" {
				candidate = parsed.Host
			}
		}
		addresses = append(addresses, candidate)
	}
	if len(addresses) == 0 {
		return nil, ErrNoAddress
	}
	return addresses, nil
}

func translatePolarisRegistryError(err error) error {
	if err == nil {
		return nil
	}
	message := strings.ToLower(err.Error())
	switch {
	case strings.Contains(message, "not found"), strings.Contains(message, "404"):
		return ErrServiceNotFound
	case strings.Contains(message, "connection refused"),
		strings.Contains(message, "dial tcp"),
		strings.Contains(message, "timeout"),
		strings.Contains(message, "no such host"),
		strings.Contains(message, "unavailable"):
		return fmt.Errorf("polaris: source unavailable: %w", err)
	default:
		return err
	}
}

type inMemoryPolarisClient struct {
	mu       sync.RWMutex
	services map[string][]contract.ServiceInstance
	watchers map[string][]chan []contract.ServiceInstance
}

func (c *inMemoryPolarisClient) Register(ctx context.Context, cfg *PolarisConfig, name, addr string, meta map[string]string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.services == nil {
		c.services = make(map[string][]contract.ServiceInstance)
	}
	fullMeta := make(map[string]string)
	for k, v := range cfg.ServiceMeta {
		fullMeta[k] = v
	}
	for k, v := range meta {
		fullMeta[k] = v
	}
	instance := contract.ServiceInstance{
		ID:       generateServiceID(name, addr),
		Name:     name,
		Address:  addr,
		Metadata: fullMeta,
		Healthy:  true,
	}
	c.services[name] = append(c.services[name], instance)
	c.notifyWatchersLocked(name)
	return nil
}

func (c *inMemoryPolarisClient) Deregister(ctx context.Context, cfg *PolarisConfig, name, addr string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	instances := c.services[name]
	for i, inst := range instances {
		if inst.Address == addr {
			c.services[name] = append(instances[:i], instances[i+1:]...)
			c.notifyWatchersLocked(name)
			return nil
		}
	}
	return ErrServiceNotFound
}

func (c *inMemoryPolarisClient) Discover(ctx context.Context, cfg *PolarisConfig, name string) ([]contract.ServiceInstance, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	instances := c.services[name]
	if len(instances) == 0 {
		return nil, ErrServiceNotFound
	}
	result := make([]contract.ServiceInstance, len(instances))
	copy(result, instances)
	return result, nil
}

func (c *inMemoryPolarisClient) Watch(ctx context.Context, cfg *PolarisConfig, name string, onUpdate func([]contract.ServiceInstance)) error {
	ch := make(chan []contract.ServiceInstance, 4)

	c.mu.Lock()
	if c.watchers == nil {
		c.watchers = make(map[string][]chan []contract.ServiceInstance)
	}
	c.watchers[name] = append(c.watchers[name], ch)
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		watchers := c.watchers[name]
		for i, watcher := range watchers {
			if watcher == ch {
				c.watchers[name] = append(watchers[:i], watchers[i+1:]...)
				break
			}
		}
		if len(c.watchers[name]) == 0 {
			delete(c.watchers, name)
		}
		c.mu.Unlock()
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		case instances := <-ch:
			onUpdate(instances)
		}
	}
}

func (c *inMemoryPolarisClient) notifyWatchersLocked(name string) {
	watchers := append([]chan []contract.ServiceInstance(nil), c.watchers[name]...)
	instances := append([]contract.ServiceInstance(nil), c.services[name]...)
	for _, watcher := range watchers {
		select {
		case watcher <- instances:
		default:
		}
	}
}

func generateServiceID(name, addr string) string {
	return fmt.Sprintf("%s-%s", name, addr)
}

func sortServiceInstances(instances []contract.ServiceInstance) {
	sort.Slice(instances, func(i, j int) bool {
		if instances[i].ID != instances[j].ID {
			return instances[i].ID < instances[j].ID
		}
		return instances[i].Address < instances[j].Address
	})
}

func snapshotKey(instances []contract.ServiceInstance) string {
	if len(instances) == 0 {
		return ""
	}
	var builder strings.Builder
	for _, inst := range instances {
		builder.WriteString(inst.ID)
		builder.WriteString("|")
		builder.WriteString(inst.Address)
		builder.WriteString("|")
		if inst.Healthy {
			builder.WriteString("1")
		} else {
			builder.WriteString("0")
		}
		builder.WriteString(";")
	}
	return builder.String()
}
