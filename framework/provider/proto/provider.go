package proto

import (
	"errors"

	"github.com/ngq/gorp/framework/contract"
)

// Provider 提供 ProtoGenerator 实现。
//
// 中文说明：
// - 支持三种工作流：Proto-first / Service-first / Route-first；
// - Proto-first：标准 protoc 生成；
// - Service-first：Go Service 接口 → Proto；
// - Route-first：Gin 路由 → Proto。
type Provider struct{}

// NewProvider 创建 Proto Generator Provider。
func NewProvider() *Provider { return &Provider{} }

// Name 返回 Provider 名称。
func (p *Provider) Name() string { return "proto.generator" }

// IsDefer 返回是否延迟加载。
func (p *Provider) IsDefer() bool { return true }

// Provides 返回提供的服务 key。
func (p *Provider) Provides() []string {
	return []string{contract.ProtoGeneratorKey}
}

// Register 注册 ProtoGenerator 服务。
//
// 中文说明：
// - 从容器获取配置，创建 Generator；
// - 配置格式见 ProtoGeneratorConfig 结构体。
func (p *Provider) Register(c contract.Container) error {
	c.Bind(contract.ProtoGeneratorKey, func(c contract.Container) (any, error) {
		cfg, err := getProtoConfig(c)
		if err != nil {
			return nil, err
		}
		return NewGenerator(cfg)
	}, true)

	return nil
}

// Boot 启动 Provider。
func (p *Provider) Boot(c contract.Container) error {
	return nil
}

// getProtoConfig 从容器获取 Proto 生成器配置。
//
// 中文说明：
// - 配置路径：proto.enabled、proto.strategy 等；
// - 未配置时使用默认值。
func getProtoConfig(c contract.Container) (*contract.ProtoGeneratorConfig, error) {
	cfgAny, err := c.Make(contract.ConfigKey)
	if err != nil {
		// 配置服务不可用时使用默认配置
		return &contract.ProtoGeneratorConfig{
			Enabled:               true,
			Strategy:              "protoc",
			DefaultProtoDir:       "api/proto",
			IncludeHTTPAnnotation: false,
		}, nil
	}

	cfg, ok := cfgAny.(contract.Config)
	if !ok {
		return nil, errors.New("proto: invalid config service")
	}

	protoCfg := &contract.ProtoGeneratorConfig{
		Enabled:               true,
		Strategy:              "protoc",
		DefaultProtoDir:       "api/proto",
		IncludeHTTPAnnotation: false,
	}

	// 读取配置项
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
