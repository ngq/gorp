package servicecomb

import (
	"context"
	"errors"
	"sync"

	"github.com/ngq/gorp/framework/contract"
)

// Provider 提供 ServiceComb 服务发现实现。
//
// 中文说明：
// - 使用华为 ServiceComb ServiceCenter 实现服务注册与发现；
// - 兼容 Spring Cloud Huawei 生态；
// - 支持多环境（开发/测试/生产）；
// - 支持服务元数据和标签。
type Provider struct{}

// NewProvider 创建 ServiceComb Provider。
func NewProvider() *Provider { return &Provider{} }

// Name 返回 Provider 名称。
func (p *Provider) Name() string { return "discovery.servicecomb" }

// IsDefer 返回是否延迟加载。
func (p *Provider) IsDefer() bool { return true }

// Provides 返回提供的服务 key。
func (p *Provider) Provides() []string {
	return []string{contract.RPCRegistryKey}
}

// Register 注册 ServiceComb Registry 服务。
func (p *Provider) Register(c contract.Container) error {
	c.Bind(contract.RPCRegistryKey, func(c contract.Container) (any, error) {
		cfg, err := getServiceCombConfig(c)
		if err != nil {
			return nil, err
		}
		return NewRegistry(cfg)
	}, true)

	return nil
}

// Boot 启动 Provider。
func (p *Provider) Boot(c contract.Container) error {
	return nil
}

// ServiceCombConfig 定义 ServiceComb 配置。
type ServiceCombConfig struct {
	// ServerURI ServiceCenter 地址（如 http://servicecenter:30100）
	ServerURI string

	// AppID 应用 ID
	AppID string

	// ServiceName 服务名称
	ServiceName string

	// Version 服务版本
	Version string

	// Environment 环境（development/testing/production）
	Environment string

	// InstanceHost 实例主机
	InstanceHost string

	// InstancePort 实例端口
	InstancePort int

	// ServiceMeta 服务元数据
	ServiceMeta map[string]string

	// Tags 服务标签
	Tags []string
}

// getServiceCombConfig 从容器获取配置。
func getServiceCombConfig(c contract.Container) (*ServiceCombConfig, error) {
	cfgAny, err := c.Make(contract.ConfigKey)
	if err != nil {
		return nil, err
	}

	cfg, ok := cfgAny.(contract.Config)
	if !ok {
		return nil, errors.New("servicecomb: invalid config service")
	}

	servicecombCfg := &ServiceCombConfig{
		Version:     "1.0.0",
		Environment: "production",
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

	return servicecombCfg, nil
}

// Registry ServiceComb 服务注册中心实现。
type Registry struct {
	config *ServiceCombConfig
	mu     sync.RWMutex

	// registeredInstances 已注册的实例
	registeredInstances map[string]bool
}

// NewRegistry 创建 ServiceComb Registry。
func NewRegistry(cfg *ServiceCombConfig) (*Registry, error) {
	if cfg.ServerURI == "" {
		return nil, errors.New("servicecomb: server_uri is required")
	}
	if cfg.AppID == "" {
		return nil, errors.New("servicecomb: app_id is required")
	}

	return &Registry{
		config:              cfg,
		registeredInstances: make(map[string]bool),
	}, nil
}

// Register 注册服务实例。
//
// 中文说明：
// - 向 ServiceCenter 注册服务实例；
// - 支持 AppID/ServiceName/Version 三级命名；
// - 支持元数据和标签。
func (r *Registry) Register(ctx context.Context, name, addr string, meta map[string]string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// TODO: 实现真实的 ServiceComb 注册
	// 需要引入 github.com/go-chassis/go-chassis
	//
	// 示例流程：
	// 1. 构建 microservice 结构
	// 2. 调用 ServiceCenter API 注册
	// 3. 启动心跳维持

	// 合并元数据
	fullMeta := make(map[string]string)
	for k, v := range r.config.ServiceMeta {
		fullMeta[k] = v
	}
	for k, v := range meta {
		fullMeta[k] = v
	}

	// 记录已注册
	r.registeredInstances[name+":"+addr] = true

	return nil
}

// Deregister 注销服务实例。
//
// 中文说明：
// - 从 ServiceCenter 注销服务实例。
func (r *Registry) Deregister(ctx context.Context, name, addr string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// TODO: 实现真实的 ServiceComb 注销

	delete(r.registeredInstances, name+":"+addr)
	return nil
}

// Discover 发现服务实例。
//
// 中文说明：
// - 从 ServiceCenter 查询服务实例列表；
// - 支持 AppID/ServiceName/Version 查询；
// - 支持环境过滤。
func (r *Registry) Discover(ctx context.Context, name string) ([]contract.ServiceInstance, error) {
	// TODO: 实现真实的 ServiceComb 发现
	// 需要引入 github.com/go-chassis/go-chassis
	//
	// 示例流程：
	// 1. 构建 FindServiceInstancesRequest
	// 2. 调用 ServiceCenter API 查询
	// 3. 解析返回的实例列表

	return []contract.ServiceInstance{}, nil
}

// Close 关闭连接。
func (r *Registry) Close() error {
	return nil
}

// ErrServerURIRequired 表示未配置 ServerURI。
var ErrServerURIRequired = errors.New("servicecomb: server_uri is required")

// ErrAppIDRequired 表示未配置 AppID。
var ErrAppIDRequired = errors.New("servicecomb: app_id is required")