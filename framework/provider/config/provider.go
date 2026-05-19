// Package config provides configuration service for gorp framework.
// Supports multiple config sources: local files, environment variables, remote config.
// Implements layered loading: base files + env overlay + env directory + env vars.
//
// 配置服务包，提供 gorp 框架的配置能力。
// 支持多种配置源：本地文件、环境变量、远程配置。
// 实现分层加载：基础文件 + 环境覆盖文件 + 环境目录 + 环境变量。
package config

import (
	"os"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

// Provider registers the configuration service contract.
// Reads config from "config.yaml" by default, with layered overlay support.
// Core logic: Bind Service instance, load config in Boot phase.
//
// Provider 注册配置服务契约。
// 默认从 "config.yaml" 读取配置，支持分层覆盖。
// 核心逻辑：绑定 Service 实例、在 Boot 阶段加载配置。
type Provider struct {
	cfg *Service
}

// NewProvider creates a new config provider with initialized service instance.
//
// NewProvider 创建新的配置 provider，初始化服务实例。
func NewProvider() *Provider {
	return &Provider{cfg: NewService()}
}

// Name returns provider name for identification.
//
// Name 返回 provider 名称，用于标识。
func (p *Provider) Name() string { return "config" }

// IsDefer indicates config provider should not defer loading.
// Config must be loaded before other providers that depend on it.
//
// IsDefer 表示配置 provider 不应延迟加载。
// 配置必须在其他依赖它的 provider 之前加载。
func (p *Provider) IsDefer() bool { return false }

// Provides returns the capability keys this provider exposes.
// Exposes ConfigKey for application-wide configuration access.
//
// Provides 返回 provider 暴露的能力键。
// 暴露 ConfigKey 用于应用级配置访问。
func (p *Provider) Provides() []string { return []string{datacontract.ConfigKey} }

// DependsOn returns the keys this provider depends on.
// Config provider has no dependencies - it's a root provider.
//
// DependsOn 返回该 provider 依赖的 key。
// Config provider 无依赖，是根 provider。
func (p *Provider) DependsOn() []string { return nil }

// Register binds the config factory to the container.
// Core logic: Create config instance, bind to container with singleton flag.
//
// Register 将配置工厂绑定到容器。
// 核心逻辑：创建配置实例、绑定到容器并标记为单例。
func (p *Provider) Register(c runtimecontract.Container) error {
	// Bind a stable pointer so Boot() can load into it.
	cfg := p.cfg
	c.Bind(datacontract.ConfigKey, func(runtimecontract.Container) (any, error) {
		return cfg, nil
	}, true)
	return nil
}

// Boot loads configuration content based on APP_ENV environment variable.
// Core logic: Normalize env, load layered config files, apply env vars.
//
// Boot 根据 APP_ENV 环境变量装载配置内容。
// 核心逻辑：规范化环境名、加载分层配置文件、应用环境变量。
func (p *Provider) Boot(runtimecontract.Container) error {
	env := NormalizeEnv(os.Getenv("APP_ENV"))
	// 中文说明：
	// - APP_ENV 是整个配置装载流程的入口变量；
	// - framework 级统一约定使用 dev / test / prod；
	return p.cfg.Load(env)
}
