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
)

// Provider 提供 Zookeeper 服务发现实现。
//
// 中文说明：
//   - 使用 Zookeeper 实现服务注册与发现；
//   - 支持临时节点（Ephemeral ZNode）实现健康检查；
//   - 支持服务元数据；
//   - 适用于已有 Zookeeper 集群的环境。
//   - 当前状态：部分可用
//   - 说明：已完成 P2 第一版最小注册/发现闭环，具备真实 Zookeeper 后端抽象与 fake backend 行为测试；
//     但当前仍未覆盖 watcher、重连与更完整 session 产品化语义。
type Provider struct{}

// NewProvider creates a new Zookeeper provider instance.
//
// NewProvider 创建新的 Zookeeper provider 实例。
func NewProvider() *Provider { return &Provider{} }

// Name returns the provider identifier "registry.zookeeper".
//
// Name 返回 provider 标识符 "registry.zookeeper"。
func (p *Provider) Name() string       { return "registry.zookeeper" }

// IsDefer returns true for lazy initialization.
//
// IsDefer 返回 true，延迟初始化。
func (p *Provider) IsDefer() bool      { return true }

// Provides returns the contract keys this provider satisfies.
//
// Provides 返回此 provider 满足的契约键。
func (p *Provider) Provides() []string { return []string{transportcontract.RPCRegistryKey} }

// Register binds the Zookeeper registry to the container.
//
// Register 将 Zookeeper 注册中心绑定到容器。
func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(transportcontract.RPCRegistryKey, func(c runtimecontract.Container) (any, error) {
		cfg, err := getZookeeperConfig(c)
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

// ZookeeperConfig 定义 Zookeeper 配置。
//
// ZookeeperConfig 定义 Zookeeper 注册中心所需配置项。
type ZookeeperConfig struct {
	Servers            []string        // Zookeeper 集群服务器地址列表
	SessionTimeout     time.Duration   // 会话超时时间
	WatchRetryInterval time.Duration   // Watch 重试间隔
	BasePath           string          // 服务注册根路径（如 /services）
	ServiceName        string          // 服务名称（可选，用于静态注册）
	ServiceAddr        string          // 服务地址（可选，用于静态注册）
	ServicePort        int             // 服务端口（可选，用于静态注册）
	ServiceMeta        map[string]string // 服务元数据（可选）
}

// getZookeeperConfig extracts Zookeeper configuration from the container's config binding.
//
// getZookeeperConfig 从容器的 config binding 中提取 Zookeeper 配置。
func getZookeeperConfig(c runtimecontract.Container) (*ZookeeperConfig, error) {
	cfgAny, err := c.Make(datacontract.ConfigKey)
	if err != nil {
		return nil, err
	}

	cfg, ok := cfgAny.(datacontract.Config)
	if !ok {
		return nil, errors.New("zookeeper: invalid config service")
	}

	zkCfg := &ZookeeperConfig{
		BasePath:           "/services",
		SessionTimeout:     30 * time.Second,
		WatchRetryInterval: 200 * time.Millisecond,
	}

	// 从配置中提取服务器列表
	if v := cfg.Get("discovery.zookeeper.servers"); v != nil {
		if servers, ok := v.([]string); ok {
			zkCfg.Servers = servers
		}
	}

	// 从配置中提取基础路径
	if v := cfg.Get("discovery.zookeeper.base_path"); v != nil {
		zkCfg.BasePath = cfg.GetString("discovery.zookeeper.base_path")
	}

	// 从配置中提取会话超时时间（秒）
	if v := cfg.Get("discovery.zookeeper.session_timeout"); v != nil {
		if seconds := cfg.GetInt("discovery.zookeeper.session_timeout"); seconds > 0 {
			zkCfg.SessionTimeout = time.Duration(seconds) * time.Second
		}
	}

	// 从配置中提取 Watch 重试间隔（毫秒）
	if v := cfg.Get("discovery.zookeeper.watch_retry_interval_ms"); v != nil {
		if ms := cfg.GetInt("discovery.zookeeper.watch_retry_interval_ms"); ms > 0 {
			zkCfg.WatchRetryInterval = time.Duration(ms) * time.Millisecond
		}
	}

	return zkCfg, nil
}