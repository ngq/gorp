// Package consul provides Consul service registry implementation for gorp.
//
// Consul 注册中心 Provider，实现 transportcontract.ServiceRegistry 契约。
// 支持服务注册、发现、注销、健康检查。
//
// 使用示例：
//
//  cfg := &DiscoveryConfig{
//      ConsulAddr:    "localhost:8500",
//      CheckInterval: "10s",
//      CheckTimeout:  "5s",
//  }
//  registry, err := NewRegistry(cfg)
//  if err != nil {
//      panic(err)
//  }
//  defer registry.Close()
//
//  err = registry.Register(ctx, "my-service", "192.168.1.100:8080", nil)
//
// 配置路径：discovery.consul.*
package consul

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/hashicorp/consul/api"
	"github.com/ngq/gorp/contrib/internal/baseregistry"
	internalnative "github.com/ngq/gorp/contrib/internal/native"
	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	configprovider "github.com/ngq/gorp/framework/provider/config"
)

// Provider 提供 Consul 服务发现实现。
type Provider struct {
	baseregistry.BaseRegistryProvider
}

// NewProvider creates a new Consul registry provider.
func NewProvider() *Provider {
	p := &Provider{}
	p.NameStr = "registry.consul"
	p.GetConfig = func(c runtimecontract.Container) (any, error) {
		return getDiscoveryConfig(c)
	}
	p.NewRegistry = func(cfg any) (transportcontract.ServiceRegistry, error) {
		return NewRegistry(cfg.(*DiscoveryConfig))
	}
	return p
}

// DiscoveryConfig 定义服务发现配置。
type DiscoveryConfig struct {
	ConsulAddr    string
	ConsulToken   string
	ServiceName   string
	ServiceAddr   string
	ServicePort   int
	ServiceMeta   map[string]string
	CheckInterval string
	CheckTimeout  string
	CheckHTTP     string
	CheckTCP      string
	CheckGRPC     string
	LoadBalance   string
}

func getDiscoveryConfig(c runtimecontract.Container) (*DiscoveryConfig, error) {
	cfg, err := readConfig(c)
	if err != nil {
		return nil, err
	}

	discCfg := &DiscoveryConfig{
		ConsulAddr:    "localhost:8500",
		CheckInterval: "10s",
		CheckTimeout:  "5s",
		LoadBalance:   "random",
	}

	if addr := configprovider.GetStringAny(cfg,
		"discovery.consul.addr",
		"discovery.consul.address",
		"discovery.consul_addr",
	); addr != "" {
		discCfg.ConsulAddr = addr
	}
	if token := configprovider.GetStringAny(cfg,
		"discovery.consul.token",
		"discovery.consul_token",
	); token != "" {
		discCfg.ConsulToken = token
	}

	sc := baseregistry.ReadServiceConfig(cfg)
	discCfg.ServiceName = sc.ServiceName
	discCfg.ServiceAddr = sc.ServiceAddr
	discCfg.ServicePort = sc.ServicePort
	discCfg.LoadBalance = sc.LoadBalance

	if interval := configprovider.GetStringAny(cfg,
		"discovery.check.interval",
		"discovery.check_interval",
	); interval != "" {
		discCfg.CheckInterval = interval
	}
	if timeout := configprovider.GetStringAny(cfg,
		"discovery.check.timeout",
		"discovery.check_timeout",
	); timeout != "" {
		discCfg.CheckTimeout = timeout
	}
	if http := configprovider.GetStringAny(cfg,
		"discovery.check.http",
		"discovery.check_http",
	); http != "" {
		discCfg.CheckHTTP = http
	}
	if tcp := configprovider.GetStringAny(cfg,
		"discovery.check.tcp",
		"discovery.check_tcp",
	); tcp != "" {
		discCfg.CheckTCP = tcp
	}
	if grpc := configprovider.GetStringAny(cfg,
		"discovery.check.grpc",
		"discovery.check_grpc",
	); grpc != "" {
		discCfg.CheckGRPC = grpc
	}
	return discCfg, nil
}

func readConfig(c runtimecontract.Container) (datacontract.Config, error) {
	cfgAny, err := c.Make(datacontract.ConfigKey)
	if err != nil {
		return nil, err
	}
	cfg, ok := cfgAny.(datacontract.Config)
	if !ok {
		return nil, errors.New("discovery: invalid config service")
	}
	return cfg, nil
}

type Registry struct {
	cfg        *DiscoveryConfig
	client     *api.Client
	agent      *api.Agent
	registered sync.Map
	mu         sync.Mutex
	lbCounter  atomic.Uint64
	closed     bool
}

func NewRegistry(cfg *DiscoveryConfig) (*Registry, error) {
	consulCfg := api.DefaultConfig()
	if cfg.ConsulAddr != "" {
		consulCfg.Address = cfg.ConsulAddr
	}
	if cfg.ConsulToken != "" {
		consulCfg.Token = cfg.ConsulToken
	}
	client, err := api.NewClient(consulCfg)
	if err != nil {
		return nil, fmt.Errorf("registry.consul: create client failed: %w", err)
	}
	return &Registry{cfg: cfg, client: client, agent: client.Agent()}, nil
}

func (r *Registry) Register(ctx context.Context, name, addr string, meta map[string]string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return errors.New("registry.consul: registry closed")
	}

	host, port := parseAddr(addr)
	if port == 0 {
		port = r.cfg.ServicePort
	}
	serviceID := generateServiceID(name, host, port)

	fullMeta := make(map[string]string)
	for k, v := range r.cfg.ServiceMeta {
		fullMeta[k] = v
	}
	for k, v := range meta {
		fullMeta[k] = v
	}

	registration := &api.AgentServiceRegistration{
		ID:      serviceID,
		Name:    name,
		Address: host,
		Port:    port,
		Meta:    fullMeta,
		Check:   r.buildHealthCheck(serviceID, host, port),
	}
	if err := r.agent.ServiceRegister(registration); err != nil {
		return fmt.Errorf("registry.consul: register service failed: %w", err)
	}
	r.registered.Store(serviceID, name)
	return nil
}

func (r *Registry) Deregister(ctx context.Context, name, addr string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	host, port := parseAddr(addr)
	if port == 0 {
		port = r.cfg.ServicePort
	}
	serviceID := generateServiceID(name, host, port)
	if err := r.agent.ServiceDeregister(serviceID); err != nil {
		return fmt.Errorf("registry.consul: deregister service failed: %w", err)
	}
	r.registered.Delete(serviceID)
	return nil
}

func (r *Registry) Discover(ctx context.Context, name string) ([]transportcontract.ServiceInstance, error) {
	services, _, err := r.client.Health().Service(name, "", true, nil)
	if err != nil {
		return nil, fmt.Errorf("registry.consul: discover service failed: %w", err)
	}
	instances := make([]transportcontract.ServiceInstance, 0, len(services))
	for _, service := range services {
		if service.Service == nil {
			continue
		}
		addr := service.Service.Address
		if addr == "" {
			addr = service.Node.Address
		}
		if service.Service.Port > 0 {
			addr = fmt.Sprintf("%s:%d", addr, service.Service.Port)
		}
		healthy := false
		for _, check := range serviceChecks(service) {
			if check.Status == "passing" {
				healthy = true
				break
			}
		}
		instances = append(instances, transportcontract.ServiceInstance{
			ID:       service.Service.ID,
			Name:     service.Service.Service,
			Address:  addr,
			Metadata: service.Service.Meta,
			Healthy:  healthy,
		})
	}
	if len(instances) > 1 {
		instances = r.applyLoadBalance(instances)
	}
	return instances, nil
}

func (r *Registry) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return nil
	}
	r.closed = true
	r.registered.Range(func(key, value any) bool {
		serviceID := key.(string)
		_ = r.agent.ServiceDeregister(serviceID)
		return true
	})
	return nil
}

func (r *Registry) Underlying() any {
	return r.client
}

func (r *Registry) As(target any) bool {
	return internalnative.As(r.client, target)
}

func (r *Registry) buildHealthCheck(serviceID, host string, port int) *api.AgentServiceCheck {
	check := &api.AgentServiceCheck{Interval: r.cfg.CheckInterval, Timeout: r.cfg.CheckTimeout}
	if r.cfg.CheckHTTP != "" {
		check.HTTP = fmt.Sprintf("http://%s:%d%s", host, port, r.cfg.CheckHTTP)
		check.Method = "GET"
		return check
	}
	if r.cfg.CheckTCP != "" {
		check.TCP = fmt.Sprintf("%s:%d", host, port)
		return check
	}
	if r.cfg.CheckGRPC != "" {
		check.GRPC = fmt.Sprintf("%s:%d", host, port)
		check.GRPCUseTLS = false
		return check
	}
	check.TCP = fmt.Sprintf("%s:%d", host, port)
	return check
}

// applyLoadBalance applies load balance strategy to instances.
// Note: The registry layer only reorders instances for initial selection preference.
// The actual per-request weighted/round-robin selection is performed by the
// Selector provider (WRR/P2C) at the RPC layer.
//
// applyLoadBalance 对实例应用负载均衡策略。
// 注意：注册中心层仅对实例重新排序以表达初始选择偏好。
// 实际的逐请求加权/轮询选择由 Selector provider（WRR/P2C）在 RPC 层执行。
func (r *Registry) applyLoadBalance(instances []transportcontract.ServiceInstance) []transportcontract.ServiceInstance {
	switch r.cfg.LoadBalance {
	case "random":
		rand.Shuffle(len(instances), func(i, j int) {
			instances[i], instances[j] = instances[j], instances[i]
		})
	case "round_robin":
		// Rotate instances by current counter to achieve round-robin ordering.
		// The actual per-request selection is done by the Selector provider (WRR/P2C).
		if len(instances) > 1 {
			r.lbCounter.Add(1)
			offset := int(r.lbCounter.Load()) % len(instances)
			instances = append(instances[offset:], instances[:offset]...)
		}
	}
	return instances
}

func parseAddr(addr string) (host string, port int) {
	parts := strings.Split(addr, ":")
	if len(parts) == 2 {
		host = parts[0]
		port, _ = strconv.Atoi(parts[1])
	} else {
		host = addr
	}
	return host, port
}

func generateServiceID(name, host string, port int) string {
	return fmt.Sprintf("%s-%s-%d", name, host, port)
}

func serviceChecks(service *api.ServiceEntry) []*api.HealthCheck {
	if len(service.Checks) > 0 {
		return service.Checks
	}
	return []*api.HealthCheck{}
}
