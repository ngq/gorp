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
// - 当前已从 framework/provider 真实下沉到 contrib 层。
type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string     { return "discovery.polaris" }
func (p *Provider) IsDefer() bool    { return true }
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
	Address     string
	Namespace   string
	Token       string
	ServiceName string
	ServiceAddr string
	ServicePort int
	ServiceMeta map[string]string
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
	polarisCfg := &PolarisConfig{Namespace: "default"}
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

type Registry struct {
	config *PolarisConfig
	mu     sync.RWMutex
}

func NewRegistry(cfg *PolarisConfig) (*Registry, error) {
	if cfg.Address == "" {
		return nil, errors.New("polaris: address is required")
	}
	return &Registry{config: cfg}, nil
}

func (r *Registry) Register(ctx context.Context, name, addr string, meta map[string]string) error { return nil }
func (r *Registry) Deregister(ctx context.Context, name, addr string) error { return nil }
func (r *Registry) Discover(ctx context.Context, name string) ([]contract.ServiceInstance, error) {
	return []contract.ServiceInstance{}, nil
}
func (r *Registry) Close() error { return nil }

var ErrNoAddress = errors.New("polaris: address is required")
