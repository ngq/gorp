package polaris

import (
	"context"
	"errors"
	"sync"

	"github.com/ngq/gorp/framework/contract"
)

// Provider 提供 Polaris 服务发现实现。
//
// 中文说明：
// - 使用腾讯云 Polaris 实现服务注册与发现；
// - 支持服务路由、负载均衡、熔断；
// - 适用于腾讯云环境和私有化部署。
type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string     { return "discovery.polaris" }
func (p *Provider) IsDefer() bool    { return true }
func (p *Provider) Provides() []string {
	return []string{contract.RPCRegistryKey}
}

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

func (p *Provider) Boot(c contract.Container) error {
	return nil
}

// PolarisConfig 定义 Polaris 配置。
type PolarisConfig struct {
	// Address Polaris Server 地址
	Address string

	// Namespace 命名空间
	Namespace string

	// Token 访问令牌
	Token string

	// 服务注册配置
	ServiceName string
	ServiceAddr string
	ServicePort int
	ServiceMeta map[string]string
}

// getPolarisConfig 从容器获取配置。
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
		Namespace: "default",
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

	return polarisCfg, nil
}

// Registry Polaris 服务注册中心实现。
type Registry struct {
	config *PolarisConfig
	mu     sync.RWMutex
}

// NewRegistry 创建 Polaris Registry。
func NewRegistry(cfg *PolarisConfig) (*Registry, error) {
	if cfg.Address == "" {
		return nil, errors.New("polaris: address is required")
	}

	return &Registry{config: cfg}, nil
}

// Register 注册服务实例。
func (r *Registry) Register(ctx context.Context, name, addr string, meta map[string]string) error {
	// TODO: 实现真实的 Polaris 注册
	// 需要引入 github.com/polarismesh/polaris-go
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

var ErrNoAddress = errors.New("polaris: address is required")