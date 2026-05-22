// Package zookeeper provides Zookeeper service registry provider for the gorp framework.
// This provider implements ServiceRegistry contract with Zookeeper SDK integration.
//
// 本包提供 gorp 框架 Zookeeper 服务注册中心 provider。
// 本 provider 实现 ServiceRegistry 契约，集成 Zookeeper SDK。
package zookeeper

import (
	"errors"
	"time"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	configprovider "github.com/ngq/gorp/framework/provider/config"
)

// Provider 提供 Zookeeper 服务发现实现。
type Provider struct {
	BaseRegistryProvider
}

// NewProvider creates a new Zookeeper provider instance.
func NewProvider() *Provider {
	p := &Provider{}
	p.NameStr = "registry.zookeeper"
	p.GetConfig = func(c runtimecontract.Container) (any, error) {
		return getZookeeperConfig(c)
	}
	p.NewRegistry = func(cfg any) (transportcontract.ServiceRegistry, error) {
		return NewRegistry(cfg.(*ZookeeperConfig))
	}
	return p
}

// ZookeeperConfig 定义 Zookeeper 配置。
type ZookeeperConfig struct {
	Servers            []string
	SessionTimeout     time.Duration
	WatchRetryInterval time.Duration
	BasePath           string
	ServiceName        string
	ServiceAddr        string
	ServicePort        int
	ServiceMeta        map[string]string
}

func getZookeeperConfig(c runtimecontract.Container) (*ZookeeperConfig, error) {
	cfgAny, err := c.Make(datacontract.ConfigKey)
	if err != nil {
		return nil, err
	}

	cfg, ok := cfgAny.(datacontract.Config)
	if !ok {
		return nil, errors.New("registry.zookeeper: invalid config service")
	}

	zkCfg := &ZookeeperConfig{
		BasePath:           "/services",
		SessionTimeout:     30 * time.Second,
		WatchRetryInterval: 200 * time.Millisecond,
	}

	if v := cfg.Get("discovery.zookeeper.servers"); v != nil {
		if servers, ok := v.([]string); ok {
			zkCfg.Servers = servers
		}
	}

	zkCfg.BasePath = configprovider.GetStringAny(cfg, "discovery.zookeeper.base_path")
	if zkCfg.BasePath == "" {
		zkCfg.BasePath = "/services"
	}
	if seconds := configprovider.GetIntAny(cfg, "discovery.zookeeper.session_timeout"); seconds > 0 {
		zkCfg.SessionTimeout = time.Duration(seconds) * time.Second
	}
	if ms := configprovider.GetIntAny(cfg, "discovery.zookeeper.watch_retry_interval_ms"); ms > 0 {
		zkCfg.WatchRetryInterval = time.Duration(ms) * time.Millisecond
	}

	return zkCfg, nil
}
