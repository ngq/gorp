// Application scenarios:
// - Expose the root-package application startup surface and top-level aliases for runtime/app assembly.
// - Keep business startup code on a short public path while delegating concrete behavior to framework/application.
// - Re-export common startup options, callbacks, errors, and runtime access types.
//
// 适用场景：
// - 暴露根包层的应用启动入口，以及 runtime/app 装配所需的顶层别名。
// - 让业务启动代码走更短的公共路径，同时把具体实现委托给 `framework/application`。
// - 重导出常用启动选项、回调、错误和 runtime 访问类型。
package gorp

import (
	"context"

	"github.com/ngq/gorp/framework/application"
	"github.com/ngq/gorp/framework/contract/data"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	"github.com/ngq/gorp/framework/contract/runtime"
)

// DBRuntimeKey is the container binding key of the DB runtime capability.
//
// DBRuntimeKey 是 DB runtime 能力的容器绑定键。
const DBRuntimeKey = data.DBRuntimeKey

var (
	// ErrServiceNameRequired indicates that the service name is missing.
	//
	// ErrServiceNameRequired 表示缺少服务名。
	ErrServiceNameRequired = application.ErrServiceNameRequired
	// ErrNoServiceDeclared indicates that no runnable service has been declared.
	//
	// ErrNoServiceDeclared 表示未声明可运行服务。
	ErrNoServiceDeclared = application.ErrNoServiceDeclared
	// ErrHTTPRouteRegistrationFailed indicates that HTTP route registration failed.
	//
	// ErrHTTPRouteRegistrationFailed 表示 HTTP 路由注册失败。
	ErrHTTPRouteRegistrationFailed = application.ErrHTTPRouteRegistrationFailed
	// ErrHTTPRuntimeUnavailable indicates that the HTTP runtime is unavailable during route setup.
	//
	// ErrHTTPRuntimeUnavailable 表示在路由注册阶段缺少可用 HTTP runtime。
	ErrHTTPRuntimeUnavailable = application.ErrHTTPRuntimeUnavailable
	// ErrSetupFailed indicates that the setup callback failed.
	//
	// ErrSetupFailed 表示 setup 回调执行失败。
	ErrSetupFailed = application.ErrSetupFailed
	// ErrMigrateFailed indicates that the migrate callback failed.
	//
	// ErrMigrateFailed 表示 migrate 回调执行失败。
	ErrMigrateFailed = application.ErrMigrateFailed
	// ErrStartupCanceled indicates that startup was canceled before boot completed.
	//
	// ErrStartupCanceled 表示启动在完成前已被取消。
	ErrStartupCanceled = application.ErrStartupCanceled
	// ErrHTTPServiceRunFailed indicates that booting the default HTTP service failed.
	//
	// ErrHTTPServiceRunFailed 表示默认 HTTP 服务启动失败。
	ErrHTTPServiceRunFailed = application.ErrHTTPServiceRunFailed
	// ErrHTTPRuntimeBuildFailed indicates that building the HTTP runtime failed.
	//
	// ErrHTTPRuntimeBuildFailed 表示 HTTP runtime 构建失败。
	ErrHTTPRuntimeBuildFailed = application.ErrHTTPRuntimeBuildFailed
)

// HTTPRuntime is the top-level alias of the application HTTP runtime.
//
// HTTPRuntime 是 application HTTP runtime 的顶层别名。
type HTTPRuntime = application.HTTPRuntime

// HTTPServiceOptions is the top-level alias of application HTTP service options.
//
// HTTPServiceOptions 是 application HTTP 服务选项的顶层别名。
type HTTPServiceOptions = application.HTTPServiceOptions

// GovernanceMode is the top-level alias of the runtime governance mode.
//
// GovernanceMode 是运行时治理模式的顶层别名。
type GovernanceMode = resiliencecontract.GovernanceMode

// ServiceProvider is the top-level alias of the runtime service provider contract.
//
// ServiceProvider 是 runtime service provider 契约的顶层别名。
type ServiceProvider = runtime.ServiceProvider

// MigrateFunc is the top-level alias of the application migrate callback contract.
//
// MigrateFunc 是 application migrate 回调契约的顶层别名。
type MigrateFunc = application.MigrateFunc

// SetupFunc is the top-level alias of the application setup callback contract.
//
// SetupFunc 是 application setup 回调契约的顶层别名。
type SetupFunc = application.SetupFunc

// HTTPRouteRegistrar is the top-level alias of the HTTP route registration callback contract.
//
// HTTPRouteRegistrar 是 HTTP 路由注册回调契约的顶层别名。
type HTTPRouteRegistrar = application.HTTPRouteRegistrar

// Option is the top-level alias of the application startup option contract.
//
// Option 是 application 启动选项契约的顶层别名。
type Option = application.Option

// Run boots the default HTTP mainline with top-level gorp options.
//
// Run 使用顶层 gorp 选项启动默认 HTTP 主线。
//
// Example:
//
//	err := gorp.Run(
//	    "user-service",
//	    gorp.HTTP(),
//	    gorp.WithHTTPRoutes(func(router gorp.HTTPRouter, c gorp.Container) error {
//	        registerRoutes(router)
//	        return nil
//	    }),
//	)
func Run(serviceName string, options ...Option) error {
	return application.Run(serviceName, options...)
}

// Start is an alias of Run.
//
// Start 是 Run 的同义入口。
func Start(serviceName string, options ...Option) error {
	return application.Start(serviceName, options...)
}

// RunContext boots the default HTTP mainline with an explicit context.
//
// RunContext 使用显式 context 启动默认 HTTP 主线。
func RunContext(ctx context.Context, serviceName string, options ...Option) error {
	return application.RunContext(ctx, serviceName, options...)
}

// BuildHTTPRuntime builds the HTTP runtime without starting listeners.
//
// BuildHTTPRuntime 构建 HTTP runtime，但不启动监听。
func BuildHTTPRuntime(serviceName string, options ...Option) (*HTTPRuntime, error) {
	return application.BuildHTTPRuntime(serviceName, options...)
}

// Build is an alias of BuildHTTPRuntime.
//
// Build 是 BuildHTTPRuntime 的同义入口。
func Build(serviceName string, options ...Option) (*HTTPRuntime, error) {
	return application.Build(serviceName, options...)
}

// HTTP declares that the default HTTP mainline should be used.
//
// HTTP 声明使用默认 HTTP 主线。
func HTTP(opts ...HTTPServiceOptions) Option {
	return application.HTTP(opts...)
}

// WithoutHTTP explicitly disables the default HTTP declaration.
//
// WithoutHTTP 显式关闭默认 HTTP 声明。
func WithoutHTTP() Option {
	return application.WithoutHTTP()
}

// Module declares providers for a single module.
//
// Module 声明单个模块的 providers。
func Module(providers ...ServiceProvider) Option {
	return application.Module(providers...)
}

// Modules declares providers for a group of modules.
//
// Modules 声明一组模块的 providers。
func Modules(groups ...[]ServiceProvider) Option {
	return application.Modules(groups...)
}

// WithModule is the explicit named alias of Module.
//
// WithModule 是 Module 的显式命名入口。
func WithModule(providers ...ServiceProvider) Option {
	return application.WithModule(providers...)
}

// WithProviders appends provider declarations to the startup options.
//
// WithProviders 向启动选项追加 provider 声明。
func WithProviders(providers ...ServiceProvider) Option {
	return application.WithProviders(providers...)
}

// WithMigrate declares a migrate callback.
//
// WithMigrate 声明 migrate 回调。
func WithMigrate(fn MigrateFunc) Option {
	return application.WithMigrate(fn)
}

// WithSetup declares a setup callback.
//
// WithSetup 声明 setup 回调。
func WithSetup(fn SetupFunc) Option {
	return application.WithSetup(fn)
}

// WithHTTPRoutes declares the default HTTP route registration callback.
//
// WithHTTPRoutes 声明默认 HTTP 路由注册回调。
func WithHTTPRoutes(register HTTPRouteRegistrar) Option {
	return application.WithHTTPRoutes(register)
}

// WithGovernanceMode declares the runtime governance mode explicitly.
//
// WithGovernanceMode 显式声明运行时治理模式。
func WithGovernanceMode(mode GovernanceMode) Option {
	return application.WithGovernanceMode(mode)
}

// WithMicroserviceMode selects the default microservice governance mainline.
//
// WithMicroserviceMode 选择默认微服务治理主线。
func WithMicroserviceMode() Option {
	return application.WithMicroserviceMode()
}

// WithMonolithMode selects the default monolith governance mainline.
//
// WithMonolithMode 选择默认单体治理主线。
func WithMonolithMode() Option {
	return application.WithMonolithMode()
}
