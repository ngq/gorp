// Package proto provides proto generator service for gorp framework.
// Supports three workflows: Proto-first, Service-first, Route-first.
// Generates Go code from proto, or proto from Go service/Gin routes.
//
// Proto 包提供 proto 生成器服务，用于 gorp 框架。
// 支持三种工作流：Proto-first、Service-first、Route-first。
// 从 proto 生成 Go 代码，或从 Go service/Gin 路由生成 proto。
package proto

import (
	"errors"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

// Provider registers proto generator service.
// Core logic: Read proto config, create Generator, bind to container.
//
// Provider 注册 proto 生成器服务。
// 核心逻辑：读取 proto 配置、创建 Generator、绑定到容器。
type Provider struct{}

// NewProvider creates a new proto generator provider.
//
// NewProvider 创建新的 proto 生成器 provider。
func NewProvider() *Provider { return &Provider{} }

// Name returns provider name for identification.
//
// Name 返回 provider 名称，用于标识。
func (p *Provider) Name() string  { return "proto.generator" }

// IsDefer indicates proto generator should defer loading.
// Generator is typically used during build/deployment phase.
//
// IsDefer 表示 proto 生成器应延迟加载。
// 生成器通常在构建/部署阶段使用。
func (p *Provider) IsDefer() bool { return true }

// Provides returns the capability keys this provider exposes.
// Exposes ProtoGeneratorKey for proto generation service.
//
// Provides 返回 provider 暴露的能力键。
// 暴露 ProtoGeneratorKey 用于 proto 生成服务。
func (p *Provider) Provides() []string {
	return []string{integrationcontract.ProtoGeneratorKey}
}

// Register binds the proto generator factory to the container.
// Core logic: Read config, create Generator instance, bind to container.
//
// Register 将 proto 生成器工厂绑定到容器。
// 核心逻辑：读取配置、创建 Generator 实例、绑定到容器。
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

// Boot initializes the proto generator provider.
// No additional startup logic required.
//
// Boot 初始化 proto 生成器 provider。
// 无需额外启动逻辑。
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
