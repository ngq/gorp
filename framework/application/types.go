// Package application provides application startup entrypoints for gorp framework.
// This file holds shared types and option contracts for startup helpers.
// Provides runtime aliases, callback contracts, and internal startup config.
//
// 应用启动包提供 gorp 框架的应用启动入口。
// 本文件承载启动辅助所需的共享类型与选项契约。
// 提供 runtime 别名、回调契约和内部启动配置。
package application

import (
	"github.com/ngq/gorp/framework/bootstrap"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

var (
	bootHTTPService    = bootstrap.BootHTTPService
	newHTTPRuntimeFunc = bootstrap.NewHTTPServiceRuntime
)

// HTTPRuntime is the startup runtime exposed to application callbacks.
//
// HTTPRuntime 是 application 回调使用的启动上下文。
type HTTPRuntime = bootstrap.HTTPServiceRuntime

// GRPCRuntime is the startup runtime exposed to application gRPC callbacks.
//
// GRPCRuntime 是 application 回调使用的 gRPC 启动上下文。
type GRPCRuntime = bootstrap.GRPCServiceRuntime

// ServiceProvider reuses the provider declaration from the runtime contract.
//
// ServiceProvider 复用 runtime contract 中的 provider 声明。
type ServiceProvider = runtimecontract.ServiceProvider

// MigrateFunc defines the migration callback contract.
//
// MigrateFunc 定义迁移回调契约。
type MigrateFunc func(*HTTPRuntime) error

// SetupFunc defines the startup setup callback contract.
//
// SetupFunc 定义启动装配回调契约。
type SetupFunc func(*HTTPRuntime) error

// HTTPRouteRegistrar defines the default HTTP route registration callback contract.
//
// HTTPRouteRegistrar 定义默认 HTTP 路由注册回调契约。
type HTTPRouteRegistrar func(router transportcontract.HTTPRouter, container runtimecontract.Container) error

// HTTPServiceOptions is the minimal HTTP options view exposed by the application package.
//
// HTTPServiceOptions 是 application 包暴露的最小 HTTP 选项视图。
type HTTPServiceOptions struct {
	DisableRedis   bool
	DisableGorm    bool
	DisableMetrics bool
	GovernanceMode resiliencecontract.GovernanceMode
	HTTPMode       resiliencecontract.HTTPMode // HTTP 模式维度：contract 或 gin
	GovernanceDisable []string
	GovernanceEnable  []string
	GovernanceProviders map[string]string
}

type runConfig struct {
	httpEnabled bool
	grpcEnabled bool
	httpOpts    bootstrap.HTTPServiceOptions
	migrate     MigrateFunc
	setup       SetupFunc
}

// Option describes an application startup option.
//
// Option 用于声明 application 启动配置。
type Option interface {
	apply(*runConfig)
}

type optionFunc func(*runConfig)

func (f optionFunc) apply(cfg *runConfig) {
	if f != nil {
		f(cfg)
	}
}
