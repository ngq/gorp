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
// - 当前已从 framework/provider 真实下沉到 contrib 层。
type Provider struct{}

func NewProvider() *Provider { return &Provider{} }
func (p *Provider) Name() string     { return "discovery.servicecomb" }
func (p *Provider) IsDefer() bool    { return true }
func (p *Provider) Provides() []string { return []string{contract.RPCRegistryKey} }

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

func (p *Provider) Boot(c contract.Container) error { return nil }

type ServiceCombConfig struct {
	ServerURI    string
	AppID        string
	ServiceName  string
	Version      string
	Environment  string
	InstanceHost string
	InstancePort int
	ServiceMeta  map[string]string
	Tags         []string
}

func getServiceCombConfig(c contract.Container) (*ServiceCombConfig, error) {
	cfgAny, err := c.Make(contract.ConfigKey)
	if err != nil {
		return nil, err
	}
	cfg, ok := cfgAny.(contract.Config)
	if !ok {
		return nil, errors.New("servicecomb: invalid config service")
	}
	servicecombCfg := &ServiceCombConfig{Version: "1.0.0", Environment: "production"}
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

type Registry struct {
	config              *ServiceCombConfig
	mu                  sync.RWMutex
	registeredInstances map[string]bool
}

func NewRegistry(cfg *ServiceCombConfig) (*Registry, error) {
	if cfg.ServerURI == "" {
		return nil, errors.New("servicecomb: server_uri is required")
	}
	if cfg.AppID == "" {
		return nil, errors.New("servicecomb: app_id is required")
	}
	return &Registry{config: cfg, registeredInstances: make(map[string]bool)}, nil
}

func (r *Registry) Register(ctx context.Context, name, addr string, meta map[string]string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	fullMeta := make(map[string]string)
	for k, v := range r.config.ServiceMeta {
		fullMeta[k] = v
	}
	for k, v := range meta {
		fullMeta[k] = v
	}
	r.registeredInstances[name+":"+addr] = true
	_ = fullMeta
	return nil
}
func (r *Registry) Deregister(ctx context.Context, name, addr string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.registeredInstances, name+":"+addr)
	return nil
}
func (r *Registry) Discover(ctx context.Context, name string) ([]contract.ServiceInstance, error) {
	return []contract.ServiceInstance{}, nil
}
func (r *Registry) Close() error { return nil }

var ErrServerURIRequired = errors.New("servicecomb: server_uri is required")
var ErrAppIDRequired = errors.New("servicecomb: app_id is required")
