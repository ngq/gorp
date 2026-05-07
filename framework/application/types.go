// Application scenarios:
// - Hold the shared types and option contracts used by application startup helpers.
// - Provide one place for runtime aliases, callback contracts, and internal startup config.
// - Keep run/options/accessor files focused on behavior instead of repeating shared declarations.
//
// 适用场景：
// - 承载 application 启动辅助所需的共享类型与选项契约。
// - 为 runtime 别名、回调契约和内部启动配置提供统一定义位置。
// - 让 run/options/accessor 等文件专注于行为实现，而不是重复声明公共类型。
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
}

type runConfig struct {
	httpEnabled bool
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
