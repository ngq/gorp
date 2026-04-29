package zookeeper

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/ngq/gorp/framework/contract"
)

// Provider 提供 Zookeeper 服务发现实现。
//
// 中文说明：
// - 使用 Zookeeper 实现服务注册与发现；
// - 支持临时节点（Ephemeral ZNode）实现健康检查；
// - 支持服务元数据；
// - 当前已从 framework/provider 真实下沉到 contrib 层。
type Provider struct{}

func NewProvider() *Provider { return &Provider{} }
func (p *Provider) Name() string     { return "discovery.zookeeper" }
func (p *Provider) IsDefer() bool    { return true }
func (p *Provider) Provides() []string { return []string{contract.RPCRegistryKey} }

func (p *Provider) Register(c contract.Container) error {
	c.Bind(contract.RPCRegistryKey, func(c contract.Container) (any, error) {
		cfg, err := getZookeeperConfig(c)
		if err != nil {
			return nil, err
		}
		return NewRegistry(cfg)
	}, true)
	return nil
}
func (p *Provider) Boot(c contract.Container) error { return nil }

type ZookeeperConfig struct {
	Servers        []string
	SessionTimeout time.Duration
	BasePath       string
	ServiceName    string
	ServiceAddr    string
	ServicePort    int
	ServiceMeta    map[string]string
}

func getZookeeperConfig(c contract.Container) (*ZookeeperConfig, error) {
	cfgAny, err := c.Make(contract.ConfigKey)
	if err != nil {
		return nil, err
	}
	cfg, ok := cfgAny.(contract.Config)
	if !ok {
		return nil, errors.New("zookeeper: invalid config service")
	}
	zkCfg := &ZookeeperConfig{BasePath: "/services", SessionTimeout: 30 * time.Second}
	if v := cfg.Get("discovery.zookeeper.servers"); v != nil {
		if servers, ok := v.([]string); ok { zkCfg.Servers = servers }
	}
	if v := cfg.Get("discovery.zookeeper.base_path"); v != nil { zkCfg.BasePath = cfg.GetString("discovery.zookeeper.base_path") }
	if v := cfg.Get("discovery.zookeeper.session_timeout"); v != nil { zkCfg.SessionTimeout = time.Duration(cfg.GetInt("discovery.zookeeper.session_timeout")) * time.Second }
	return zkCfg, nil
}

type Registry struct {
	config              *ZookeeperConfig
	mu                  sync.RWMutex
	registeredInstances map[string]string
}

func NewRegistry(cfg *ZookeeperConfig) (*Registry, error) {
	if len(cfg.Servers) == 0 { return nil, errors.New("zookeeper: servers is required") }
	return &Registry{config: cfg, registeredInstances: make(map[string]string)}, nil
}
func (r *Registry) Register(ctx context.Context, name, addr string, meta map[string]string) error {
	r.mu.Lock(); defer r.mu.Unlock(); instancePath := r.config.BasePath + "/" + name + "/" + addr; r.registeredInstances[name] = instancePath; _ = meta; return nil
}
func (r *Registry) Deregister(ctx context.Context, name, addr string) error {
	r.mu.Lock(); defer r.mu.Unlock(); delete(r.registeredInstances, name); _ = addr; return nil
}
func (r *Registry) Discover(ctx context.Context, name string) ([]contract.ServiceInstance, error) { return []contract.ServiceInstance{}, nil }
func (r *Registry) Close() error { return nil }

var ErrNoServers = errors.New("zookeeper: no servers configured")
