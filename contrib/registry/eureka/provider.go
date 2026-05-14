// Package eureka provides Eureka service registry provider for the gorp framework.
// This provider implements ServiceRegistry contract with Eureka REST API integration.
//
// 本包提供 gorp 框架 Eureka 服务注册中心 provider。
// 本 provider 实现 ServiceRegistry 契约，集成 Eureka REST API。
package eureka

import (
	"errors"
	"time"

	"github.com/ngq/gorp/contrib/internal/baseregistry"
	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	configprovider "github.com/ngq/gorp/framework/provider/config"
)

// Provider 提供 Eureka 服务发现实现。
type Provider struct {
	baseregistry.BaseRegistryProvider
}

// NewProvider creates a new Eureka provider instance.
func NewProvider() *Provider {
	p := &Provider{}
	p.NameStr = "registry.eureka"
	p.GetConfig = func(c runtimecontract.Container) (any, error) {
		return getEurekaConfig(c)
	}
	p.NewRegistry = func(cfg any) (transportcontract.ServiceRegistry, error) {
		return NewRegistry(cfg.(*EurekaConfig))
	}
	return p
}

// EurekaConfig 定义 Eureka 配置。
type EurekaConfig struct {
	ServerURL             string
	AppName               string
	InstanceHost          string
	InstancePort          int
	ServiceMeta           map[string]string
	HeartbeatInterval     time.Duration
	HeartbeatRetryBackoff time.Duration
	WatchInterval         time.Duration
}

func getEurekaConfig(c runtimecontract.Container) (*EurekaConfig, error) {
	cfgAny, err := c.Make(datacontract.ConfigKey)
	if err != nil {
		return nil, err
	}

	cfg, ok := cfgAny.(datacontract.Config)
	if !ok {
		return nil, errors.New("registry.eureka: invalid config service")
	}

	eurekaCfg := &EurekaConfig{
		HeartbeatRetryBackoff: time.Second,
		WatchInterval:         5 * time.Second,
	}

	eurekaCfg.ServerURL = configprovider.GetStringAny(cfg, "discovery.eureka.server_url")
	eurekaCfg.AppName = configprovider.GetStringAny(cfg, "discovery.eureka.app_name")
	eurekaCfg.InstanceHost = configprovider.GetStringAny(cfg, "discovery.eureka.instance_host")
	if port := configprovider.GetIntAny(cfg, "discovery.eureka.instance_port"); port > 0 {
		eurekaCfg.InstancePort = port
	}
	if seconds := configprovider.GetIntAny(cfg, "discovery.eureka.heartbeat_interval_seconds"); seconds > 0 {
		eurekaCfg.HeartbeatInterval = time.Duration(seconds) * time.Second
	}
	if ms := configprovider.GetIntAny(cfg, "discovery.eureka.heartbeat_retry_backoff_ms"); ms > 0 {
		eurekaCfg.HeartbeatRetryBackoff = time.Duration(ms) * time.Millisecond
	}
	if ms := configprovider.GetIntAny(cfg, "discovery.eureka.watch_interval_ms"); ms > 0 {
		eurekaCfg.WatchInterval = time.Duration(ms) * time.Millisecond
	}

	return eurekaCfg, nil
}
