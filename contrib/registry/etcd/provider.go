// Package etcd provides etcd service registry provider for the gorp framework.
// This provider implements ServiceRegistry contract with etcd SDK integration.
//
// 本包提供 gorp 框架 etcd 服务注册中心 provider。
// 本 provider 实现 ServiceRegistry 契约，集成 etcd SDK。
package etcd

import (
	"errors"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	configprovider "github.com/ngq/gorp/framework/provider/config"
)

// Provider 提供 etcd 服务发现实现。
//
// 中文说明：
//   - 使用 etcd KV + Lease API 实现服务注册与发现；
//   - 通过租约 TTL 实现最小健康检查与自动下线；
//   - 当前已补齐 KeepAlive 失效后的最小重新注册闭环。
type Provider struct{}

// NewProvider creates a new etcd provider instance.
//
// NewProvider 创建新的 etcd provider 实例。
func NewProvider() *Provider { return &Provider{} }

// Name returns the provider identifier "registry.etcd".
//
// Name 返回 provider 标识符 "registry.etcd"。
func (p *Provider) Name() string { return "registry.etcd" }

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

// Register binds the etcd registry to the container.
//
// Register 将 etcd 注册中心绑定到容器。
func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(transportcontract.RPCRegistryKey, func(c runtimecontract.Container) (any, error) {
		cfg, err := getDiscoveryConfig(c)
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

// DiscoveryConfig 定义 etcd 服务发现配置。
type DiscoveryConfig struct {
	// EtcdEndpoints etcd 集群地址列表
	EtcdEndpoints []string
	// EtcdUsername etcd 用户名（可选）
	EtcdUsername string
	// EtcdPassword etcd 密码（可选）
	EtcdPassword string

	// ServicePath 服务注册路径前缀
	ServicePath string
	// LeaseTTL 租约 TTL（秒）
	LeaseTTL int64

	// ServiceName 服务名称
	ServiceName string
	// ServiceAddr 服务地址
	ServiceAddr string
	// ServicePort 服务端口
	ServicePort int
	// ServiceMeta 服务元数据
	ServiceMeta map[string]string

	// LoadBalance 负载均衡策略
	LoadBalance string
}

// getDiscoveryConfig extracts etcd configuration from the container's config binding.
//
// getDiscoveryConfig 从容器的 config binding 中提取 etcd 配置。
func getDiscoveryConfig(c runtimecontract.Container) (*DiscoveryConfig, error) {
	cfgAny, err := c.Make(datacontract.ConfigKey)
	if err != nil {
		return nil, err
	}

	cfg, ok := cfgAny.(datacontract.Config)
	if !ok {
		return nil, errors.New("discovery: invalid config service")
	}

	discCfg := &DiscoveryConfig{
		ServicePath: "/services/",
		LeaseTTL:    10,
		LoadBalance: "random",
	}

	// 解析 etcd endpoints
	endpoints := configprovider.GetStringSliceAny(cfg,
		"discovery.etcd.endpoints",
		"discovery.etcd_endpoints",
	)
	if len(endpoints) == 0 {
		endpoints = []string{"localhost:2379"}
	}
	discCfg.EtcdEndpoints = endpoints

	// 解析 etcd 认证信息
	if username := configprovider.GetStringAny(cfg,
		"discovery.etcd.username",
		"discovery.etcd_username",
	); username != "" {
		discCfg.EtcdUsername = username
	}
	if password := configprovider.GetStringAny(cfg,
		"discovery.etcd.password",
		"discovery.etcd_password",
	); password != "" {
		discCfg.EtcdPassword = password
	}

	// 解析服务路径
	if servicePath := configprovider.GetStringAny(cfg,
		"discovery.service.path",
		"discovery.service_path",
	); servicePath != "" {
		discCfg.ServicePath = servicePath
	}

	// 解析租约 TTL
	if ttl := configprovider.GetIntAny(cfg,
		"discovery.etcd.lease_ttl",
		"discovery.lease_ttl",
	); ttl > 0 {
		discCfg.LeaseTTL = int64(ttl)
	}

	// 解析服务名称
	if name := configprovider.GetStringAny(cfg,
		"discovery.service.name",
		"discovery.service_name",
	); name != "" {
		discCfg.ServiceName = name
	}

	// 解析服务地址
	if addr := configprovider.GetStringAny(cfg,
		"discovery.service.addr",
		"discovery.service.address",
		"discovery.service_addr",
	); addr != "" {
		discCfg.ServiceAddr = addr
	}

	// 解析服务端口
	if port := configprovider.GetIntAny(cfg,
		"discovery.service.port",
		"discovery.service_port",
	); port > 0 {
		discCfg.ServicePort = port
	}

	// 解析负载均衡策略
	if lb := configprovider.GetStringAny(cfg,
		"selector.algorithm",
		"discovery.load_balance",
	); lb != "" {
		discCfg.LoadBalance = lb
	}

	return discCfg, nil
}