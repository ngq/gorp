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
// - 与 configsource.consul 共用 Consul client，减少连接开销。
type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string     { return "discovery.consul" }
func (p *Provider) IsDefer() bool    { return true }
func (p *Provider) Provides() []string {
	return []string{contract.RPCRegistryKey}
}

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

func (p *Provider) Boot(c contract.Container) error {
	return nil
}

// DiscoveryConfig 定义服务发现配置。
type DiscoveryConfig struct {
	// Consul 配置
	ConsulAddr  string
	ConsulToken string

	// 服务注册配置
	ServiceName    string
	ServiceAddr    string
	ServicePort    int
	ServiceMeta    map[string]string
	CheckInterval  string // 健康检查间隔，如 "10s"
	CheckTimeout   string // 健康检查超时，如 "5s"
	CheckHTTP      string // HTTP 健康检查路径，如 "/health"
	CheckTCP       string // TCP 健康检查地址
	CheckGRPC      string // gRPC 健康检查地址

	// 负载均衡策略
	LoadBalance string // random/round_robin
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

	// Consul 地址
	if addr := configprovider.GetStringAny(cfg,
		"discovery.consul.addr",
		"discovery.consul.address",
		"discovery.consul_addr",
	); addr != "" {
		discCfg.ConsulAddr = addr
	}

	// Consul Token
	if token := configprovider.GetStringAny(cfg,
		"discovery.consul.token",
		"discovery.consul_token",
	); token != "" {
		discCfg.ConsulToken = token
	}

	// 服务名称
	if name := configprovider.GetStringAny(cfg,
		"discovery.service.name",
		"discovery.service_name",
	); name != "" {
		discCfg.ServiceName = name
	}

	// 服务地址（自动检测）
	if addr := configprovider.GetStringAny(cfg,
		"discovery.service.addr",
		"discovery.service.address",
		"discovery.service_addr",
	); addr != "" {
		discCfg.ServiceAddr = addr
	}

	// 服务端口
	if port := configprovider.GetIntAny(cfg,
		"discovery.service.port",
		"discovery.service_port",
	); port > 0 {
		discCfg.ServicePort = port
	}

	// 健康检查配置
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

	// 负载均衡策略
	if lb := configprovider.GetStringAny(cfg,
		"selector.algorithm",
		"discovery.load_balance",
	); lb != "" {
		discCfg.LoadBalance = lb
	}

	return discCfg, nil
}

// Registry 是 Consul 服务发现实现。
//
// 中文说明：
// - 使用 Consul Agent API 注册服务；
// - 使用 Consul Health API 发现服务实例；
// - 支持健康检查自动剔除不健康实例；
// - 支持多种负载均衡策略。
type Registry struct {
	cfg    *DiscoveryConfig
	client *api.Client
	agent  *api.Agent

	// 已注册服务缓存
	registered sync.Map // map[string]string
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

	return &Registry{
		cfg:    cfg,
		client: client,
		agent:  client.Agent(),
	}, nil
}

// Register 注册服务实例。
//
// 中文说明：
// - 使用 Consul Agent Service Register API；
// - 支持配置健康检查（HTTP/TCP/gRPC）；
// - 自动生成唯一服务 ID（name-address-port）；
// - 注册成功后缓存 ID，用于后续注销。
func (r *Registry) Register(ctx context.Context, name, addr string, meta map[string]string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return errors.New("discovery.consul: registry closed")
	}

	// 解析地址获取端口
	host, port := parseAddr(addr)
	if port == 0 {
		// 尝试从配置获取端口
		port = r.cfg.ServicePort
	}

	// 生成唯一服务 ID
	serviceID := generateServiceID(name, host, port)

	// 合并元数据
	fullMeta := make(map[string]string)
	for k, v := range r.cfg.ServiceMeta {
		fullMeta[k] = v
	}
	for k, v := range meta {
		fullMeta[k] = v
	}

	// 构建服务注册请求
	registration := &api.AgentServiceRegistration{
		ID:      serviceID,
		Name:    name,
		Address: host,
		Port:    port,
		Meta:    fullMeta,

		// 健康检查配置
		Check: r.buildHealthCheck(serviceID, host, port),
	}

	// 执行注册
	err := r.agent.ServiceRegister(registration)
	if err != nil {
		return fmt.Errorf("discovery.consul: register service failed: %w", err)
	}

	// 缓存已注册的服务 ID
	r.registered.Store(serviceID, name)

	return nil
}

// Deregister 注销服务实例。
//
// 中文说明：
// - 使用 Consul Agent Service Deregister API；
// - 根据 service ID 注销指定实例。
func (r *Registry) Deregister(ctx context.Context, name, addr string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 查找已注册的服务 ID
	host, port := parseAddr(addr)
	if port == 0 {
		port = r.cfg.ServicePort
	}
	serviceID := generateServiceID(name, host, port)

	// 执行注销
	err := r.agent.ServiceDeregister(serviceID)
	if err != nil {
		return fmt.Errorf("discovery.consul: deregister service failed: %w", err)
	}

	// 删除缓存
	r.registered.Delete(serviceID)

	return nil
}

// Discover 发现服务实例。
//
// 中文说明：
// - 使用 Consul Health Service API 查询；
// - 只返回健康（passing）状态的实例；
// - 支持负载均衡策略选择实例。
func (r *Registry) Discover(ctx context.Context, name string) ([]contract.ServiceInstance, error) {
	// 查询健康的服务实例
	services, _, err := r.client.Health().Service(name, "", true, nil)
	if err != nil {
		return nil, fmt.Errorf("discovery.consul: discover service failed: %w", err)
	}

	// 转换为 ServiceInstance 列表
	instances := make([]contract.ServiceInstance, 0, len(services))
	for _, service := range services {
		// 检查服务健康状态
		if service.Service == nil {
			continue
		}

		// 构建实例地址
		addr := service.Service.Address
		if addr == "" {
			// 使用 Agent 地址
			addr = service.Node.Address
		}
		if service.Service.Port > 0 {
			addr = fmt.Sprintf("%s:%d", addr, service.Service.Port)
		}

		// 转换健康状态
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

	// 应用负载均衡策略（排序）
	if len(instances) > 1 {
		instances = r.applyLoadBalance(instances)
	}

	return instances, nil
}

// Close 关闭服务发现连接。
//
// 中文说明：
// - 注销所有已注册的服务实例；
// - 关闭 Consul client。
func (r *Registry) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return nil
	}
	r.closed = true

	// 注销所有已注册的服务
	r.registered.Range(func(key, value any) bool {
		serviceID := key.(string)
		_ = r.agent.ServiceDeregister(serviceID)
		return true
	})

	return nil
}

// buildHealthCheck 构建健康检查配置。
//
// 中文说明：
// - 支持 HTTP/TCP/gRPC 三种检查方式；
// - HTTP 检查：访问指定路径，期望 200 响应；
// - TCP 检查：尝试建立 TCP 连接；
// - gRPC 检查：使用 gRPC 健康检查协议。
func (r *Registry) buildHealthCheck(serviceID, host string, port int) *api.AgentServiceCheck {
	check := &api.AgentServiceCheck{
		Interval: r.cfg.CheckInterval,
		Timeout:  r.cfg.CheckTimeout,
	}

	// HTTP 健康检查
	if r.cfg.CheckHTTP != "" {
		check.HTTP = fmt.Sprintf("http://%s:%d%s", host, port, r.cfg.CheckHTTP)
		check.Method = "GET"
		return check
	}

	// TCP 健康检查
	if r.cfg.CheckTCP != "" {
		check.TCP = fmt.Sprintf("%s:%d", host, port)
		return check
	}

	// gRPC 健康检查
	if r.cfg.CheckGRPC != "" {
		check.GRPC = fmt.Sprintf("%s:%d", host, port)
		check.GRPCUseTLS = false
		return check
	}

	// 默认 TCP 检查
	check.TCP = fmt.Sprintf("%s:%d", host, port)
	return check
}

// applyLoadBalance 应用负载均衡策略。
//
// 中文说明：
// - random：随机选择顺序；
// - round_robin：轮询顺序（TODO）。
func (r *Registry) applyLoadBalance(instances []contract.ServiceInstance) []contract.ServiceInstance {
	switch r.cfg.LoadBalance {
	case "random":
		// 随机打乱顺序
		rand.Shuffle(len(instances), func(i, j int) {
			instances[i], instances[j] = instances[j], instances[i]
		})
	case "round_robin":
		// TODO: 实现轮询策略（需要外部计数器）
	}
	return instances
}

// parseAddr 解析地址和端口。
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

// generateServiceID 生成唯一服务 ID。
func generateServiceID(name, host string, port int) string {
	return fmt.Sprintf("%s-%s-%d", name, host, port)
}

// serviceChecks 获取服务的健康检查列表。
func serviceChecks(service *api.ServiceEntry) []*api.HealthCheck {
	// Health.Service 返回的结构中 Checks 字段包含服务检查
	if len(service.Checks) > 0 {
		return service.Checks
	}
	return []*api.HealthCheck{}
}