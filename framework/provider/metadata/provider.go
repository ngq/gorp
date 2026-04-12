package metadata

import (
	"github.com/ngq/gorp/framework/contract"
	configprovider "github.com/ngq/gorp/framework/provider/config"
	"github.com/ngq/gorp/framework/provider/metadata/propagator"
)

// Provider 提供默认 metadata 传播能力。
//
// 中文说明：
// - 统一绑定 Metadata 与 MetadataPropagator；
// - Metadata 本体使用内存实现，供上下文传递和测试场景复用；
// - MetadataPropagator 默认使用前缀传播策略，供 HTTP / gRPC middleware 统一使用。
type Provider struct{}

// NewProvider 创建 metadata provider。
func NewProvider() *Provider { return &Provider{} }

// Name 返回 provider 名称。
func (p *Provider) Name() string { return "metadata.default" }

// IsDefer 延迟加载。
func (p *Provider) IsDefer() bool { return true }

// Provides 返回对外能力 key。
func (p *Provider) Provides() []string {
	return []string{contract.MetadataKey, contract.MetadataPropagatorKey}
}

// Register 绑定 metadata 与传播器。
func (p *Provider) Register(c contract.Container) error {
	c.Bind(contract.MetadataKey, func(c contract.Container) (any, error) {
		return contract.NewMetadata(), nil
	}, true)
	c.Bind(contract.MetadataPropagatorKey, func(c contract.Container) (any, error) {
		cfg := readMetadataConfig(c)
		return propagator.NewDefaultPropagator(cfg.PropagatePrefix, cfg.ConstantMetadata), nil
	}, true)
	return nil
}

// Boot 无额外启动逻辑。
func (p *Provider) Boot(contract.Container) error { return nil }

func readMetadataConfig(c contract.Container) contract.MetadataConfig {
	cfgAny, err := c.Make(contract.ConfigKey)
	if err != nil {
		return contract.MetadataConfig{
			PropagatePrefix: []string{"x-md-"},
		}
	}
	cfg, ok := cfgAny.(contract.Config)
	if !ok {
		return contract.MetadataConfig{
			PropagatePrefix: []string{"x-md-"},
		}
	}

	metaCfg := contract.MetadataConfig{
		PropagatePrefix: []string{"x-md-"},
	}
	if prefixes := configprovider.GetStringSliceAny(cfg,
		"metadata.propagate_prefix",
	); len(prefixes) > 0 {
		metaCfg.PropagatePrefix = prefixes
	}
	if constant := configprovider.GetStringMapAny(cfg,
		"metadata.constant_metadata",
	); len(constant) > 0 {
		metaCfg.ConstantMetadata = constant
	}
	if maxSize := configprovider.GetIntAny(cfg,
		"metadata.max_size",
	); maxSize > 0 {
		metaCfg.MaxSize = maxSize
	}
	return metaCfg
}

func toString(v any) string {
	switch s := v.(type) {
	case string:
		return s
	default:
		return ""
	}
}
