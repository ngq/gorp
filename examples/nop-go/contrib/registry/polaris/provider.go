// Package polaris provides Polaris service registry provider for the gorp framework.
// This provider implements ServiceRegistry contract with Polaris SDK integration.
//
// 本包提供 gorp 框架 Polaris 服务注册中心 provider。
// 本 provider 实现 ServiceRegistry 契约，集成 Polaris SDK。
package polaris

import (
	"errors"
	"time"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	configprovider "github.com/ngq/gorp/framework/provider/config"
)

// Provider 提供 Polaris 服务发现实现。
type Provider struct {
	BaseRegistryProvider
}

// NewProvider creates a new Polaris provider instance.
func NewProvider() *Provider {
	p := &Provider{}
	p.NameStr = "registry.polaris"
	p.GetConfig = func(c runtimecontract.Container) (any, error) {
		return getPolarisConfig(c)
	}
	p.NewRegistry = func(cfg any) (transportcontract.ServiceRegistry, error) {
		return NewRegistry(cfg.(*PolarisConfig))
	}
	return p
}

// PolarisConfig 定义 Polaris 配置。
type PolarisConfig struct {
	Address            string
	Namespace          string
	Token              string
	ServiceName        string
	ServiceAddr        string
	ServicePort        int
	ServiceMeta        map[string]string
	WatchRetryInterval time.Duration
}

func getPolarisConfig(c runtimecontract.Container) (*PolarisConfig, error) {
	cfgAny, err := c.Make(datacontract.ConfigKey)
	if err != nil {
		return nil, err
	}
	cfg, ok := cfgAny.(datacontract.Config)
	if !ok {
		return nil, errors.New("registry.polaris: invalid config service")
	}

	polarisCfg := &PolarisConfig{
		Namespace:          "default",
		WatchRetryInterval: 200 * time.Millisecond,
	}
	polarisCfg.Address = configprovider.GetStringAny(cfg, "discovery.polaris.address")
	polarisCfg.Namespace = configprovider.GetStringAny(cfg, "discovery.polaris.namespace")
	if polarisCfg.Namespace == "" {
		polarisCfg.Namespace = "default"
	}
	polarisCfg.Token = configprovider.GetStringAny(cfg, "discovery.polaris.token")
	if ms := configprovider.GetIntAny(cfg, "discovery.polaris.watch_retry_interval_ms"); ms > 0 {
		polarisCfg.WatchRetryInterval = time.Duration(ms) * time.Millisecond
	}
	return polarisCfg, nil
}
