package metadata

import (
	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	configprovider "github.com/ngq/gorp/framework/provider/config"
	"github.com/ngq/gorp/framework/provider/metadata/propagator"
)

type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string { return "metadata.default" }

func (p *Provider) IsDefer() bool { return true }

func (p *Provider) Provides() []string {
	return []string{transportcontract.MetadataKey, transportcontract.MetadataPropagatorKey}
}

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
