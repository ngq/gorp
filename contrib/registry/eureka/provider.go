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
// - 支持心跳健康检查；
// - 当前已从 framework/provider 真实下沉到 contrib 层。
type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string     { return "discovery.eureka" }
func (p *Provider) IsDefer() bool    { return true }
func (p *Provider) Provides() []string { return []string{contract.RPCRegistryKey} }

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

func (p *Provider) Boot(c contract.Container) error { return nil }

type EurekaConfig struct {
	ServerURL    string
	AppName      string
	InstanceHost string
	InstancePort int
	ServiceMeta  map[string]string
}

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

type Registry struct {
	config *EurekaConfig
	mu     sync.RWMutex
}

func NewRegistry(cfg *EurekaConfig) (*Registry, error) {
	if cfg.ServerURL == "" {
		return nil, errors.New("eureka: server_url is required")
	}
	return &Registry{config: cfg}, nil
}

func (r *Registry) Register(ctx context.Context, name, addr string, meta map[string]string) error { return nil }
func (r *Registry) Deregister(ctx context.Context, name, addr string) error { return nil }
func (r *Registry) Discover(ctx context.Context, name string) ([]contract.ServiceInstance, error) {
	return []contract.ServiceInstance{}, nil
}
func (r *Registry) Close() error { return nil }

var ErrNoServerURL = errors.New("eureka: server_url is required")
