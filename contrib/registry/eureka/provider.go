// Package eureka provides Eureka service registry provider for the gorp framework.
// This provider implements ServiceRegistry contract with Eureka REST API integration.
//
// 本包提供 gorp 框架 Eureka 服务注册中心 provider。
// 本 provider 实现 ServiceRegistry 契约，集成 Eureka REST API。
package eureka

import (
	"errors"
	"time"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

// Provider 提供 Eureka 服务发现实现。
//
// 中文说明：
//   - 使用 Netflix Eureka 实现服务注册与发现；
//   - 兼容 Spring Cloud 生态；
//   - 支持心跳健康检查。
//   - 当前状态：部分可用
//   - 说明：已完成 P2 第一版最小注册/发现闭环，具备 Register / Deregister / Discover 与 fake client 行为测试；
//     但当前仍未覆盖完整 Eureka 心跳与续租产品化语义。
type Provider struct{}

// NewProvider creates a new Eureka provider instance.
//
// NewProvider 创建新的 Eureka provider 实例。
func NewProvider() *Provider { return &Provider{} }

// Name returns the provider identifier "registry.eureka".
//
// Name 返回 provider 标识符 "registry.eureka"。
func (p *Provider) Name() string { return "registry.eureka" }

// IsDefer returns true for lazy initialization.
//
// IsDefer 返回 true，延迟初始化。
func (p *Provider) IsDefer() bool { return true }

// Provides returns the contract keys this provider satisfies.
//
// Provides 返回此 provider 满足的契约键。
func (p *Provider) Provides() []string {
	return []string{transportcontract.RPCRegistryKey}
}

// Register binds the Eureka registry to the container.
//
// Register 将 Eureka 注册中心绑定到容器。
func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(transportcontract.RPCRegistryKey, func(c runtimecontract.Container) (any, error) {
		cfg, err := getEurekaConfig(c)
		if err != nil {
			return nil, err
		}
		return NewRegistry(cfg)
	}, true)

	return nil
}

// Boot does nothing for lazy providers.
//
// Boot 延迟 provider 不需要 boot 操作。
func (p *Provider) Boot(c runtimecontract.Container) error { return nil }

// EurekaConfig 定义 Eureka 配置。
//
// 配置项说明：
//   - ServerURL: Eureka 服务端地址（必需）
//   - AppName: 应用名称
//   - InstanceHost: 实例主机名
//   - InstancePort: 实例端口
//   - ServiceMeta: 服务级 metadata
//   - HeartbeatInterval: 心跳间隔
//   - HeartbeatRetryBackoff: 心跳重试退避时间
//   - WatchInterval: 监听轮询间隔
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

// getEurekaConfig extracts Eureka configuration from the container's config binding.
//
// getEurekaConfig 从容器的 config binding 中提取 Eureka 配置。
// 配置路径格式：discovery.eureka.xxx
func getEurekaConfig(c runtimecontract.Container) (*EurekaConfig, error) {
	cfgAny, err := c.Make(datacontract.ConfigKey)
	if err != nil {
		return nil, err
	}

	cfg, ok := cfgAny.(datacontract.Config)
	if !ok {
		return nil, errors.New("eureka: invalid config service")
	}

	// 初始化默认配置
	eurekaCfg := &EurekaConfig{
		HeartbeatRetryBackoff: time.Second,
		WatchInterval:         5 * time.Second,
	}

	// 从配置文件读取各项配置
	if v := cfg.Get("discovery.eureka.server_url"); v != nil {
		eurekaCfg.ServerURL = cfg.GetString("discovery.eureka.server_url")
	}
	if v := cfg.Get("discovery.eureka.app_name"); v != nil {
		eurekaCfg.AppName = cfg.GetString("discovery.eureka.app_name")
	}
	if v := cfg.Get("discovery.eureka.instance_host"); v != nil {
		eurekaCfg.InstanceHost = cfg.GetString("discovery.eureka.instance_host")
	}
	if v := cfg.Get("discovery.eureka.instance_port"); v != nil {
		eurekaCfg.InstancePort = cfg.GetInt("discovery.eureka.instance_port")
	}
	// 心跳间隔（秒）
	if v := cfg.Get("discovery.eureka.heartbeat_interval_seconds"); v != nil {
		if seconds := cfg.GetInt("discovery.eureka.heartbeat_interval_seconds"); seconds > 0 {
			eurekaCfg.HeartbeatInterval = time.Duration(seconds) * time.Second
		}
	}
	// 心跳重试退避（毫秒）
	if v := cfg.Get("discovery.eureka.heartbeat_retry_backoff_ms"); v != nil {
		if ms := cfg.GetInt("discovery.eureka.heartbeat_retry_backoff_ms"); ms > 0 {
			eurekaCfg.HeartbeatRetryBackoff = time.Duration(ms) * time.Millisecond
		}
	}
	// 监听轮询间隔（毫秒）
	if v := cfg.Get("discovery.eureka.watch_interval_ms"); v != nil {
		if ms := cfg.GetInt("discovery.eureka.watch_interval_ms"); ms > 0 {
			eurekaCfg.WatchInterval = time.Duration(ms) * time.Millisecond
		}
	}

	return eurekaCfg, nil
}