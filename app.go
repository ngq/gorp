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
//	gorp.Run("my-service",
//	    gorp.HTTP(),
//	    gorp.WithProviders(configprovider.NewProvider()),
//	    gorp.WithHTTPRoutes(registerRoutes),
//	)
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
	// GovernanceModeMonolith keeps the runtime on the lightweight monolith governance mainline.
	//
	// GovernanceModeMonolith 表示继续使用轻量、单体优先的默认治理主线。
	GovernanceModeMonolith = resiliencecontract.GovernanceModeMonolith
	// GovernanceModeGinFirst keeps Gin-native development ergonomics on the shared governance mainline.
	//
	// GovernanceModeGinFirst 表示在共享治理主线下优先保留 Gin 原生开发体验。
	GovernanceModeGinFirst = resiliencecontract.GovernanceModeGinFirst
	// GovernanceModeMicroservice enables the default microservice governance mainline.
	//
	// GovernanceModeMicroservice 表示启用默认微服务治理主线。
	GovernanceModeMicroservice = resiliencecontract.GovernanceModeMicroservice
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
//	    gorp.WithMicroserviceMode(),
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

// WithGinFirstMode selects the Gin-first governance mainline.
//
// WithGinFirstMode 选择 Gin-first 治理主线。
func WithGinFirstMode() Option {
	return application.WithGinFirstMode()
}
