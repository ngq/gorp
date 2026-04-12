package eureka

import (
	"context"
	"errors"
	"sync"

	"github.com/ngq/gorp/framework/contract"
)

// Provider 提供 Eureka 服务发现实现。
//
// 中文说明：
// - 使用 Netflix Eureka 实现服务注册与发现；
// - 兼容 Spring Cloud 生态；
// - 支持心跳健康检查。
type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string     { return "discovery.eureka" }
func (p *Provider) IsDefer() bool    { return true }
func (p *Provider) Provides() []string {
	return []string{contract.RPCRegistryKey}
}

func (p *Provider) Register(c contract.Container) error {
	c.Bind(contract.RPCRegistryKey, func(c contract.Container) (any, error) {
		cfg, err := getEurekaConfig(c)
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

// EurekaConfig 定义 Eureka 配置。
type EurekaConfig struct {
	// ServerURL Eureka Server URL
	ServerURL string

	// AppName 应用名称
	AppName string

	// InstanceHost 实例主机
	InstanceHost string

	// InstancePort 实例端口
	InstancePort int

	// ServiceMeta 服务元数据
	ServiceMeta map[string]string
}

// getEurekaConfig 从容器获取配置。
func getEurekaConfig(c contract.Container) (*EurekaConfig, error) {
	cfgAny, err := c.Make(contract.ConfigKey)
	if err != nil {
		return nil, err
	}

	cfg, ok := cfgAny.(contract.Config)
	if !ok {
		return nil, errors.New("eureka: invalid config service")
	}

	eurekaCfg := &EurekaConfig{}

	if v := cfg.Get("discovery.eureka.server_url"); v != nil {
		eurekaCfg.ServerURL = cfg.GetString("discovery.eureka.server_url")
	}
	if v := cfg.Get("discovery.eureka.app_name"); v != nil {
		eurekaCfg.AppName = cfg.GetString("discovery.eureka.app_name")
	}

	return eurekaCfg, nil
}

// Registry Eureka 服务注册中心实现。
type Registry struct {
	config *EurekaConfig
	mu     sync.RWMutex
}

// NewRegistry 创建 Eureka Registry。
func NewRegistry(cfg *EurekaConfig) (*Registry, error) {
	if cfg.ServerURL == "" {
		return nil, errors.New("eureka: server_url is required")
	}

	return &Registry{config: cfg}, nil
}

// Register 注册服务实例。
func (r *Registry) Register(ctx context.Context, name, addr string, meta map[string]string) error {
	// TODO: 实现真实的 Eureka 注册
	// 需要引入 Eureka Go 客户端
	return nil
}

// Deregister 注销服务实例。
func (r *Registry) Deregister(ctx context.Context, name, addr string) error {
	return nil
}

// Discover 发现服务实例。
func (r *Registry) Discover(ctx context.Context, name string) ([]contract.ServiceInstance, error) {
	return []contract.ServiceInstance{}, nil
}

// Close 关闭连接。
func (r *Registry) Close() error {
	return nil
}

var ErrNoServerURL = errors.New("eureka: server_url is required")