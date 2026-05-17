// Package gorp provides the root-package application startup surface for gorp framework.
// Exposes top-level aliases for runtime/app assembly and common startup options.
// Re-export runtime access types, callbacks, errors from framework/application.
//
// Gorp 包提供 gorp 框架的根包层应用启动入口。
// 暴露 runtime/app 装配所需的顶层别名和常用启动选项。
// 从 framework/application 重导出 runtime 访问类型、回调、错误。
//
// Eg:
//
//	gorp.Run(
//	    gorp.WithProviders(configprovider.NewProvider()),
//	    gorp.WithHTTPRoutes(registerRoutes),
//	)
//
// HTTP is enabled by default. Use gorp.HTTP() only when you need HTTPServiceOptions.
// Use gorp.WithoutHTTP() to explicitly disable the HTTP mainline.
// Service name is read from config file (app.name) automatically.
//
// HTTP 默认已启用。仅当需要传入 HTTPServiceOptions 时才调用 gorp.HTTP()。
// 使用 gorp.WithoutHTTP() 可显式关闭 HTTP 主线。
// 服务名自动从配置文件 app.name 读取。
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
	// ErrGRPCServiceRunFailed indicates that booting the gRPC service failed.
	//
	// ErrGRPCServiceRunFailed 表示 gRPC 服务启动失败。
	ErrGRPCServiceRunFailed = application.ErrGRPCServiceRunFailed
	// ErrGRPCRuntimeBuildFailed indicates that building the gRPC runtime failed.
	//
	// ErrGRPCRuntimeBuildFailed 表示 gRPC runtime 构建失败。
	ErrGRPCRuntimeBuildFailed = application.ErrGRPCRuntimeBuildFailed
)

// HTTPRuntime is the top-level alias of the application HTTP runtime.
//
// HTTPRuntime 是 application HTTP runtime 的顶层别名。
type HTTPRuntime = application.HTTPRuntime

// GRPCRuntime is the top-level alias of the application gRPC runtime.
//
// GRPCRuntime 是 application gRPC runtime 的顶层别名。
type GRPCRuntime = application.GRPCRuntime

// HTTPServiceOptions is the top-level alias of application HTTP service options.
//
// HTTPServiceOptions 是 application HTTP 服务选项的顶层别名。
type HTTPServiceOptions = application.HTTPServiceOptions

// GovernanceMode is the top-level alias of the runtime governance mode.
//
// GovernanceMode 是运行时治理模式的顶层别名。
type GovernanceMode = resiliencecontract.GovernanceMode

const (
	// GovernanceModeMono keeps the runtime on the lightweight mono governance mainline.
	//
	// GovernanceModeMono 表示继续使用轻量、单体优先的默认治理主线。
	GovernanceModeMono = resiliencecontract.GovernanceModeMono
	// GovernanceModeMicro enables the default microservice governance mainline.
	//
	// GovernanceModeMicro 表示启用默认微服务治理主线。
	GovernanceModeMicro = resiliencecontract.GovernanceModeMicro
)

// HTTPMode is the top-level alias of the HTTP handling abstraction mode.
//
// HTTPMode 是 HTTP 处理抽象模式的顶层别名。
type HTTPMode = resiliencecontract.HTTPMode

const (
	// HTTPModeContract uses gorp.HTTPContext abstraction.
	//
	// HTTPModeContract 使用 gorp.HTTPContext 契约抽象。
	HTTPModeContract = resiliencecontract.HTTPModeContract
	// HTTPModeGin uses native gin.Context directly.
	//
	// HTTPModeGin 使用原生 gin.Context。
	HTTPModeGin = resiliencecontract.HTTPModeGin
)

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
// Service name is read from config file (app.name) automatically.
//
// Run 使用顶层 gorp 选项启动默认 HTTP 主线。
// 服务名自动从配置文件 app.name 读取。
//
// Example:
//
//	err := gorp.Run(
//	    gorp.HTTP(),
//	    gorp.WithHTTPRoutes(func(router gorp.HTTPRouter, c gorp.Container) error {
//	        registerRoutes(router)
//	        return nil
//	    }),
//	)
func Run(options ...Option) error {
	return application.Run(options...)
}

// Start is an alias of Run.
//
// Deprecated: Use Run instead. Start is kept for backward compatibility.
//
// Start 是 Run 的同义入口。
//
// Deprecated: 请使用 Run。Start 仅为向后兼容保留。
func Start(options ...Option) error {
	return application.Start(options...)
}

// RunContext boots the default HTTP mainline with an explicit context.
// Service name is read from config file (app.name) automatically.
//
// RunContext 使用显式 context 启动默认 HTTP 主线。
// 服务名自动从配置文件 app.name 读取。
func RunContext(ctx context.Context, options ...Option) error {
	return application.RunContext(ctx, options...)
}

// BuildHTTPRuntime builds the HTTP runtime without starting listeners.
// Service name is read from config file (app.name) automatically.
//
// BuildHTTPRuntime 构建 HTTP runtime，但不启动监听。
// 服务名自动从配置文件 app.name 读取。
func BuildHTTPRuntime(options ...Option) (*HTTPRuntime, error) {
	return application.BuildHTTPRuntime(options...)
}

// Build is an alias of BuildHTTPRuntime.
//
// Deprecated: Use BuildHTTPRuntime instead. Build is kept for backward compatibility.
//
// Build 是 BuildHTTPRuntime 的同义入口。
//
// Deprecated: 请使用 BuildHTTPRuntime。Build 仅为向后兼容保留。
func Build(options ...Option) (*HTTPRuntime, error) {
	return application.Build(options...)
}

// HTTP declares that the default HTTP mainline should be used.
// HTTP declares HTTP service options. The HTTP mainline is enabled by default,
// so calling HTTP() without arguments is unnecessary. Use HTTP() only when you
// need to pass HTTPServiceOptions. Use WithoutHTTP() to explicitly disable HTTP.
//
// HTTP 声明 HTTP 服务选项。HTTP 主线默认已启用，无参调用 HTTP() 是冗余的。
// 仅当需要传入 HTTPServiceOptions 时才调用。使用 WithoutHTTP() 可显式关闭 HTTP 主线。
func HTTP(opts ...HTTPServiceOptions) Option {
	return application.HTTP(opts...)
}

// GRPC declares that the gRPC service should be started alongside the HTTP mainline.
// This enables the gRPC server in the HTTPServiceRuntime, allowing users to register
// proto services via rt.GRPCServer in the setup callback.
//
// GRPC 声明在 HTTP 主线之外同时启动 gRPC 服务。
// 这会在 HTTPServiceRuntime 中启用 gRPC 服务器，允许用户在 setup 回调中
// 通过 rt.GRPCServer 注册 proto 服务。
//
// Example:
//
//	gorp.Run("user-service",
//	    gorp.HTTP(),
//	    gorp.GRPC(),
//	    gorp.WithMicroMode(),
//	    gorp.WithSetup(func(rt *gorp.HTTPRuntime) error {
//	        pb.RegisterUserServiceServer(rt.GRPCServer, userService)
//	        return nil
//	    }),
//	)
func GRPC() Option {
	return application.GRPC()
}

// WithoutHTTP explicitly disables the default HTTP declaration.
//
// WithoutHTTP 显式关闭默认 HTTP 声明。
func WithoutHTTP() Option {
	return application.WithoutHTTP()
}

// Module declares providers for a single module.
//
// Deprecated: Use WithProviders instead. Module is kept for backward compatibility.
//
// Module 声明单个模块的 providers。
//
// Deprecated: 请使用 WithProviders。Module 仅为向后兼容保留。
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
// Deprecated: Use WithProviders instead. WithModule is kept for backward compatibility.
//
// WithModule 是 Module 的显式命名入口。
//
// Deprecated: 请使用 WithProviders。WithModule 仅为向后兼容保留。
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

// WithGovernanceDisabled explicitly disables one or more default governance capabilities.
//
// WithGovernanceDisabled 显式关闭一个或多个默认治理能力。
func WithGovernanceDisabled(names ...string) Option {
	return application.WithGovernanceDisabled(names...)
}

// WithGovernanceEnabled explicitly enables one or more governance capabilities that are off by default.
// This is the symmetric counterpart of WithGovernanceDisabled.
// When the same feature appears in both enable and disable, disable takes precedence.
//
// WithGovernanceEnabled 显式开启一个或多个默认关闭的治理能力。
// 这是 WithGovernanceDisabled 的对称入口。
// 当同一 feature 同时出现在 enable 和 disable 中时，disable 生效。
func WithGovernanceEnabled(names ...string) Option {
	return application.WithGovernanceEnabled(names...)
}

// WithGovernanceProvider explicitly overrides one governance provider backend.
//
// WithGovernanceProvider 显式覆盖一个治理 provider backend。
func WithGovernanceProvider(name, backend string) Option {
	return application.WithGovernanceProvider(name, backend)
}

// WithMicroMode selects the microservice governance mainline.
//
// WithMicroMode 选择微服务治理主线。
func WithMicroMode() Option {
	return application.WithMicroMode()
}

// WithMonoMode selects the mono governance mainline.
//
// WithMonoMode 选择单体治理主线。
func WithMonoMode() Option {
	return application.WithMonoMode()
}

// WithMicroGovernance selects microservice governance and HTTP contract mode.
//
// WithMicroGovernance 选择微服务治理 + HTTP 契约模式。
func WithMicroGovernance() Option {
	return application.WithMicroGovernance()
}

// WithMonoGovernance selects mono governance (HTTP mode left to default or explicit).
//
// WithMonoGovernance 选择单体治理（HTTP 模式由默认值或显式参数决定）。
func WithMonoGovernance() Option {
	return application.WithMonoGovernance()
}

// WithHTTPMode declares the HTTP handling mode explicitly.
//
// WithHTTPMode 显式声明 HTTP 处理抽象模式。
func WithHTTPMode(mode HTTPMode) Option {
	return application.WithHTTPMode(mode)
}

// WithContractHTTPMode selects the gorp.HTTPContext contract abstraction.
//
// WithContractHTTPMode 选择 gorp.HTTPContext 契约抽象。
func WithContractHTTPMode() Option {
	return application.WithContractHTTPMode()
}

// WithGinHTTPMode selects the native gin.Context mode.
//
// WithGinHTTPMode 选择原生 gin.Context 模式。
func WithGinHTTPMode() Option {
	return application.WithGinHTTPMode()
}
