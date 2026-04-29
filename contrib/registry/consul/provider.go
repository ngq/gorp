package consul

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"

	"github.com/ngq/gorp/framework/contract"
	configprovider "github.com/ngq/gorp/framework/provider/config"
	"github.com/hashicorp/consul/api"
)

// Provider 提供 Consul 服务发现实现。
//
// 中文说明：
// - 使用 Consul Agent API 实现服务注册与发现；
// - 支持健康检查（HTTP/TCP/gRPC）；
// - 支持服务元数据（版本、权重等）；
// - 当前已从 framework/provider 真实下沉到 contrib 层。
type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string     { return "discovery.consul" }
func (p *Provider) IsDefer() bool    { return true }
func (p *Provider) Provides() []string { return []string{contract.RPCRegistryKey} }

func (p *Provider) Register(c contract.Container) error {
	c.Bind(contract.RPCRegistryKey, func(c contract.Container) (any, error) {
		cfg, err := getDiscoveryConfig(c)
		if err != nil {
			return nil, err
		}
		return NewRegistry(cfg)
	}, true)
	return nil
}

func (p *Provider) Boot(c contract.Container) error { return nil }

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

// getDiscoveryConfig 从容器获取服务发现配置。
func getDiscoveryConfig(c contract.Container) (*DiscoveryConfig, error) {
	cfgAny, err := c.Make(contract.ConfigKey)
	if err != nil {
		return nil, err
	}
	cfg, ok := cfgAny.(contract.Config)
	if !ok {
		return nil, errors.New("discovery: invalid config service")
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
	if name := configprovider.GetStringAny(cfg,
		"discovery.service.name",
		"discovery.service_name",
	); name != "" {
		discCfg.ServiceName = name
	}
	if addr := configprovider.GetStringAny(cfg,
		"discovery.service.addr",
		"discovery.service.address",
		"discovery.service_addr",
	); addr != "" {
		discCfg.ServiceAddr = addr
	}
	if port := configprovider.GetIntAny(cfg,
		"discovery.service.port",
		"discovery.service_port",
	); port > 0 {
		discCfg.ServicePort = port
	}
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
	if lb := configprovider.GetStringAny(cfg,
		"selector.algorithm",
		"discovery.load_balance",
	); lb != "" {
		discCfg.LoadBalance = lb
	}
	return discCfg, nil
}

// Registry 是 Consul 服务发现实现。
type Registry struct {
	cfg        *DiscoveryConfig
	client     *api.Client
	agent      *api.Agent
	registered sync.Map
	mu         sync.Mutex
	closed     bool
}

// NewRegistry 创建 Consul 服务发现实例。
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
		return nil, fmt.Errorf("discovery.consul: create client failed: %w", err)
	}
	return &Registry{cfg: cfg, client: client, agent: client.Agent()}, nil
}

// Register 注册服务实例。
func (r *Registry) Register(ctx context.Context, name, addr string, meta map[string]string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return errors.New("discovery.consul: registry closed")
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
		return fmt.Errorf("discovery.consul: register service failed: %w", err)
	}
	r.registered.Store(serviceID, name)
	return nil
}

// Deregister 注销服务实例。
func (r *Registry) Deregister(ctx context.Context, name, addr string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	host, port := parseAddr(addr)
	if port == 0 {
		port = r.cfg.ServicePort
	}
	serviceID := generateServiceID(name, host, port)
	if err := r.agent.ServiceDeregister(serviceID); err != nil {
		return fmt.Errorf("discovery.consul: deregister service failed: %w", err)
	}
	r.registered.Delete(serviceID)
	return nil
}

// Discover 发现服务实例。
func (r *Registry) Discover(ctx context.Context, name string) ([]contract.ServiceInstance, error) {
	services, _, err := r.client.Health().Service(name, "", true, nil)
	if err != nil {
		return nil, fmt.Errorf("discovery.consul: discover service failed: %w", err)
	}
	instances := make([]contract.ServiceInstance, 0, len(services))
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
		instances = append(instances, contract.ServiceInstance{
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

// Close 关闭服务发现连接。
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

func (r *Registry) applyLoadBalance(instances []contract.ServiceInstance) []contract.ServiceInstance {
	switch r.cfg.LoadBalance {
	case "random":
		rand.Shuffle(len(instances), func(i, j int) {
			instances[i], instances[j] = instances[j], instances[i]
		})
	case "round_robin":
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
