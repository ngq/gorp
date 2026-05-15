// Package metadata provides metadata propagation service for gorp framework.
// Supports automatic metadata injection and extraction across HTTP/gRPC boundaries.
// Configurable propagation prefix and constant metadata.
//
// 元数据包提供元数据传播服务，用于 gorp 框架。
// 支持跨 HTTP/gRPC 边界的自动元数据注入和提取。
// 可配置传播前缀和常量元数据。
package metadata

import (
	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	configprovider "github.com/ngq/gorp/framework/provider/config"
	"github.com/ngq/gorp/framework/provider/metadata/propagator"
)

// Provider registers metadata propagation service.
// Core logic: Create Metadata and Propagator, bind to container.
//
// Provider 注册元数据传播服务。
// 核心逻辑：创建 Metadata 和 Propagator、绑定到容器。
type Provider struct{}

// NewProvider creates a new metadata provider.
//
// NewProvider 创建新的元数据 provider。
func NewProvider() *Provider { return &Provider{} }

// Name returns provider name for identification.
//
// Name 返回 provider 名称，用于标识。
func (p *Provider) Name() string { return "metadata.default" }

// IsDefer indicates metadata provider should defer loading.
// Can be loaded after transport providers.
//
// IsDefer 表示元数据 provider 应延迟加载。
// 可以在传输 provider 之后加载。
func (p *Provider) IsDefer() bool { return true }

// Provides returns the capability keys this provider exposes.
// Exposes MetadataKey and MetadataPropagatorKey.
//
// Provides 返回 provider 暴露的能力键。
// 暴露 MetadataKey 和 MetadataPropagatorKey。
func (p *Provider) Provides() []string {
	return []string{transportcontract.MetadataKey, transportcontract.MetadataPropagatorKey}
}

// DependsOn returns the keys this provider depends on.
// Metadata provider depends on Config for propagation configuration.
//
// DependsOn 返回该 provider 依赖的 key。
// Metadata provider 依赖 Config 获取传播配置。
func (p *Provider) DependsOn() []string { return []string{datacontract.ConfigKey} }

// Register binds metadata services to the container.
// Core logic: Create Metadata instance, create Propagator with config, bind both.
//
// Register 将元数据服务绑定到容器。
// 核心逻辑：创建 Metadata 实例、创建带配置的 Propagator、绑定两者。
func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(transportcontract.MetadataKey, func(c runtimecontract.Container) (any, error) {
		return transportcontract.NewMetadata(), nil
	}, true)
	c.Bind(transportcontract.MetadataPropagatorKey, func(c runtimecontract.Container) (any, error) {
		cfg := readMetadataConfig(c)
		return propagator.NewDefaultPropagator(cfg.PropagatePrefix, cfg.ConstantMetadata), nil
	}, true)
	return nil
}

// Boot initializes the metadata provider.
// No additional startup logic required.
//
// Boot 初始化元数据 provider。
// 无需额外启动逻辑。
func (p *Provider) Boot(runtimecontract.Container) error { return nil }

func readMetadataConfig(c runtimecontract.Container) transportcontract.MetadataConfig {
	cfgAny, err := c.Make(datacontract.ConfigKey)
	if err != nil {
		return transportcontract.MetadataConfig{
			PropagatePrefix: []string{"x-md-"},
		}
	}
	cfg, ok := cfgAny.(datacontract.Config)
	if !ok {
		return transportcontract.MetadataConfig{
			PropagatePrefix: []string{"x-md-"},
		}
	}

	metaCfg := transportcontract.MetadataConfig{
		PropagatePrefix: []string{"x-md-"},
	}
	if prefixes := configprovider.GetStringSliceAny(cfg, "metadata.propagate_prefix"); len(prefixes) > 0 {
		metaCfg.PropagatePrefix = prefixes
	}
	if constant := configprovider.GetStringMapAny(cfg, "metadata.constant_metadata"); len(constant) > 0 {
		metaCfg.ConstantMetadata = constant
	}
	if maxSize := configprovider.GetIntAny(cfg, "metadata.max_size"); maxSize > 0 {
		metaCfg.MaxSize = maxSize
	}
	return metaCfg
}
