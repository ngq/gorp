package proto

import (
	"errors"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string  { return "proto.generator" }
func (p *Provider) IsDefer() bool { return true }
func (p *Provider) Provides() []string {
	return []string{integrationcontract.ProtoGeneratorKey}
}

func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(integrationcontract.ProtoGeneratorKey, func(c runtimecontract.Container) (any, error) {
		cfg, err := getProtoConfig(c)
		if err != nil {
			return nil, err
		}
		return NewGenerator(cfg)
	}, true)

	return nil
}

func (p *Provider) Boot(c runtimecontract.Container) error {
	return nil
}

func getProtoConfig(c runtimecontract.Container) (*integrationcontract.ProtoGeneratorConfig, error) {
	cfgAny, err := c.Make(datacontract.ConfigKey)
	if err != nil {
		return &integrationcontract.ProtoGeneratorConfig{
			Enabled:               true,
			Strategy:              "protoc",
			DefaultProtoDir:       "api/proto",
			IncludeHTTPAnnotation: false,
		}, nil
	}

	cfg, ok := cfgAny.(datacontract.Config)
	if !ok {
		return nil, errors.New("proto: invalid config service")
	}

	protoCfg := &integrationcontract.ProtoGeneratorConfig{
		Enabled:               true,
		Strategy:              "protoc",
		DefaultProtoDir:       "api/proto",
		IncludeHTTPAnnotation: false,
	}

	if v := cfg.Get("proto.enabled"); v != nil {
		protoCfg.Enabled = cfg.GetBool("proto.enabled")
	}
	if v := cfg.Get("proto.strategy"); v != nil {
		protoCfg.Strategy = cfg.GetString("proto.strategy")
	}
	if v := cfg.Get("proto.default_proto_dir"); v != nil {
		protoCfg.DefaultProtoDir = cfg.GetString("proto.default_proto_dir")
	}
	if v := cfg.Get("proto.include_http_annotation"); v != nil {
		protoCfg.IncludeHTTPAnnotation = cfg.GetBool("proto.include_http_annotation")
	}

	return protoCfg, nil
}
