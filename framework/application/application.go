package application

import (
	"context"
	"errors"
	"strings"

	"github.com/ngq/gorp/framework/bootstrap"
	frameworkcontainer "github.com/ngq/gorp/framework/container"
	datacontract "github.com/ngq/gorp/framework/contract/data"
	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	securitycontract "github.com/ngq/gorp/framework/contract/security"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	frameworkgrpc "github.com/ngq/gorp/framework/provider/grpc"
)

var (
	// ErrServiceNameRequired indicates that the service name is missing.
	// ErrServiceNameRequired 表示未提供服务名。
	ErrServiceNameRequired = errors.New("application: serviceName is required")
	// ErrNoServiceDeclared indicates that no runnable service has been declared.
	// ErrNoServiceDeclared 表示未声明可运行服务。
	ErrNoServiceDeclared = errors.New("application: no service declared")
	// ErrHTTPRouteRegistrationFailed indicates that HTTP route registration failed.
	// ErrHTTPRouteRegistrationFailed 表示 HTTP 路由注册失败。
	ErrHTTPRouteRegistrationFailed = errors.New("application: http route registration failed")
	// ErrHTTPRuntimeUnavailable indicates that the HTTP runtime is unavailable during route setup.
	// ErrHTTPRuntimeUnavailable 表示 HTTP 路由注册阶段缺少可用 runtime。
	ErrHTTPRuntimeUnavailable = errors.New("application: http runtime unavailable")
	// ErrSetupFailed indicates that the setup callback failed.
	// ErrSetupFailed 表示 setup 回调执行失败。
	ErrSetupFailed = errors.New("application: setup failed")
	// ErrMigrateFailed indicates that the migrate callback failed.
	// ErrMigrateFailed 表示 migrate 回调执行失败。
	ErrMigrateFailed = errors.New("application: migrate failed")
	// ErrStartupCanceled indicates that startup was canceled before boot completed.
	// ErrStartupCanceled 表示启动前 context 已取消。
	ErrStartupCanceled = errors.New("application: startup canceled")
	// ErrHTTPServiceRunFailed indicates that booting the default HTTP service failed.
	// ErrHTTPServiceRunFailed 表示 HTTP 服务启动失败。
	ErrHTTPServiceRunFailed = errors.New("application: http service run failed")
	// ErrHTTPRuntimeBuildFailed indicates that building the HTTP runtime failed.
	// ErrHTTPRuntimeBuildFailed 表示 HTTP runtime 构建失败。
	ErrHTTPRuntimeBuildFailed = errors.New("application: http runtime build failed")
)

var (
	bootHTTPService    = bootstrap.BootHTTPService
	newHTTPRuntimeFunc = bootstrap.NewHTTPServiceRuntime
)

// HTTPRuntime is the startup runtime exposed to application callbacks.
// HTTPRuntime 是 application 回调使用的启动上下文。
type HTTPRuntime = bootstrap.HTTPServiceRuntime

// ServiceProvider reuses the provider declaration from the runtime contract.
// ServiceProvider 复用 runtime contract 中的 provider 声明。
type ServiceProvider = runtimecontract.ServiceProvider

// MigrateFunc defines the migration callback contract.
// MigrateFunc 定义迁移回调契约。
type MigrateFunc func(*HTTPRuntime) error

// SetupFunc defines the startup setup callback contract.
// SetupFunc 定义启动装配回调契约。
type SetupFunc func(*HTTPRuntime) error

// HTTPRouteRegistrar defines the default HTTP route registration callback contract.
// HTTPRouteRegistrar 定义默认 HTTP 路由注册契约。
type HTTPRouteRegistrar func(router transportcontract.HTTPRouter, container runtimecontract.Container) error

// HTTPServiceOptions is the minimal HTTP options view exposed by the application package.
// HTTPServiceOptions 是 application 包暴露的最小 HTTP 选项视图。
type HTTPServiceOptions struct {
	DisableRedis   bool
	DisableGorm    bool
	DisableMetrics bool
}

type runConfig struct {
	httpEnabled bool
	httpOpts    bootstrap.HTTPServiceOptions
	migrate     MigrateFunc
	setup       SetupFunc
}

// Option describes an application startup option.
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

// resolveRunConfig resolves and normalizes startup options.
// resolveRunConfig 解析并归一化启动配置。
func resolveRunConfig(serviceName string, options ...Option) (runConfig, error) {
	serviceName = strings.TrimSpace(serviceName)
	if serviceName == "" {
		return runConfig{}, ErrServiceNameRequired
	}

	cfg := runConfig{httpEnabled: true}
	for _, opt := range options {
		if opt != nil {
			opt.apply(&cfg)
		}
	}
	if !cfg.httpEnabled {
		return runConfig{}, ErrNoServiceDeclared
	}
	return cfg, nil
}

// Run boots the default HTTP mainline with application options.
// 说明：业务运行入口仍是项目自己的 main；application 只提供启动装配 helper。
// Run 启动默认 HTTP 主线。
//
// Example:
//
//	err := application.Run(
//	    "user-service",
//	    application.HTTP(),
//	    application.WithProviders(myProvider),
//	    application.WithHTTPRoutes(func(router transportcontract.HTTPRouter, c runtimecontract.Container) error {
//	        registerRoutes(router)
//	        return nil
//	    }),
//	)
func Run(serviceName string, options ...Option) error {
	return RunContext(context.Background(), serviceName, options...)
}

// Start is an alias of Run.
// Start 是 Run 的同义入口。
func Start(serviceName string, options ...Option) error {
	return Run(serviceName, options...)
}

// RunContext boots the default mainline with an explicit context.
// 当前语义：仅在启动前检查取消，运行中关闭流程仍由 bootstrap 处理。
// RunContext 使用显式 context 启动默认主线。
func RunContext(ctx context.Context, serviceName string, options ...Option) error {
	if err := ensureStartupContext(ctx); err != nil {
		return errors.Join(ErrStartupCanceled, err)
	}
	if ctx == nil {
		ctx = context.Background()
	}

	normalizedServiceName := strings.TrimSpace(serviceName)
	cfg, err := resolveRunConfig(normalizedServiceName, options...)
	if err != nil {
		return err
	}
	if err := bootHTTPService(normalizedServiceName, cfg.httpOpts, cfg.migrate, cfg.setup); err != nil {
		return errors.Join(ErrHTTPServiceRunFailed, err)
	}
	return nil
}

// ensureStartupContext validates the context state before startup.
// 语义：启动前已取消优先返回取消错误，不进入后续参数与装配流程。
// ensureStartupContext 在启动前校验 context 状态。
func ensureStartupContext(ctx context.Context) error {
	if ctx == nil {
		return nil
	}
	return ctx.Err()
}

// BuildHTTPRuntime builds the startup runtime without starting listeners.
// BuildHTTPRuntime 仅构建启动上下文，不启动监听。
//
// Example:
//
//	rt, err := application.BuildHTTPRuntime("user-service", application.HTTP())
//	if err != nil {
//	    return err
//	}
//	defer rt.Container.Close()
func BuildHTTPRuntime(serviceName string, options ...Option) (*HTTPRuntime, error) {
	normalizedServiceName := strings.TrimSpace(serviceName)
	cfg, err := resolveRunConfig(normalizedServiceName, options...)
	if err != nil {
		return nil, err
	}
	rt, err := newHTTPRuntimeFunc(normalizedServiceName, cfg.httpOpts)
	if err != nil {
		return nil, errors.Join(ErrHTTPRuntimeBuildFailed, err)
	}
	return rt, nil
}

// Build is an alias of BuildHTTPRuntime.
// Build 是 BuildHTTPRuntime 的同义入口。
func Build(serviceName string, options ...Option) (*HTTPRuntime, error) {
	return BuildHTTPRuntime(serviceName, options...)
}

// HTTP declares that the default HTTP mainline should be used.
// HTTP 声明使用默认 HTTP 主线。
//
// Example:
//
//	application.HTTP(application.HTTPServiceOptions{
//	    DisableRedis: true,
//	})
func HTTP(opts ...HTTPServiceOptions) Option {
	return optionFunc(func(cfg *runConfig) {
		cfg.httpEnabled = true
		if len(opts) == 0 {
			return
		}
		h := opts[0]
		cfg.httpOpts.DisableRedis = h.DisableRedis
		cfg.httpOpts.DisableGorm = h.DisableGorm
		cfg.httpOpts.DisableMetrics = h.DisableMetrics
	})
}

// WithoutHTTP explicitly disables the default HTTP declaration.
// WithoutHTTP 显式关闭默认 HTTP 声明。
func WithoutHTTP() Option {
	return optionFunc(func(cfg *runConfig) {
		cfg.httpEnabled = false
	})
}

// Module declares providers for a single module.
// Module 声明单个模块的 providers。
func Module(providers ...ServiceProvider) Option {
	return WithProviders(providers...)
}

// Modules declares providers for a group of modules.
// Modules 声明一组模块 providers。
func Modules(groups ...[]ServiceProvider) Option {
	return optionFunc(func(cfg *runConfig) {
		for _, providers := range groups {
			WithProviders(providers...).apply(cfg)
		}
	})
}

// WithModule is the explicit named alias of Module.
// WithModule 是 Module 的显式命名入口。
func WithModule(providers ...ServiceProvider) Option {
	return Module(providers...)
}

// WithProviders appends provider declarations without changing the startup semantics.
// WithProviders 追加 providers 声明，不改变底层启动语义。
//
// Example:
//
//	application.WithProviders(
//	    configprovider.NewProvider(),
//	    cacheprovider.NewProvider(),
//	)
func WithProviders(providers ...ServiceProvider) Option {
	return optionFunc(func(cfg *runConfig) {
		if len(providers) == 0 {
			return
		}
		existing := make(map[string]struct{}, len(cfg.httpOpts.ExtraProviders))
		for _, p := range cfg.httpOpts.ExtraProviders {
			if p == nil {
				continue
			}
			existing[p.Name()] = struct{}{}
		}
		for _, p := range providers {
			if p == nil {
				continue
			}
			name := p.Name()
			if _, ok := existing[name]; ok {
				continue
			}
			existing[name] = struct{}{}
			cfg.httpOpts.ExtraProviders = append(cfg.httpOpts.ExtraProviders, p)
		}
	})
}

// WithMigrate declares a migration callback.
// WithMigrate 声明迁移回调。
func WithMigrate(fn func(*HTTPRuntime) error) Option {
	return optionFunc(func(cfg *runConfig) {
		if fn == nil {
			return
		}
		cfg.migrate = func(rt *HTTPRuntime) error {
			if err := fn(rt); err != nil {
				return errors.Join(ErrMigrateFailed, err)
			}
			return nil
		}
	})
}

// WithSetup declares a setup callback.
// WithSetup 声明装配回调。
//
// Example:
//
//	application.WithSetup(func(rt *application.HTTPRuntime) error {
//	    return registerHTTP(rt.Router)
//	})
func WithSetup(fn func(*HTTPRuntime) error) Option {
	return optionFunc(func(cfg *runConfig) {
		var next SetupFunc
		if fn != nil {
			next = func(rt *HTTPRuntime) error {
				if err := fn(rt); err != nil {
					return errors.Join(ErrSetupFailed, err)
				}
				return nil
			}
		}
		cfg.setup = composeSetup(cfg.setup, next)
	})
}

// WithHTTPRoutes declares the default HTTP route registration callback.
// WithHTTPRoutes 声明默认 HTTP 路由注册回调。
//
// Example:
//
//	application.WithHTTPRoutes(func(router transportcontract.HTTPRouter, c runtimecontract.Container) error {
//	    api := router.Group("/api")
//	    api.GET("/ping", pingHandler)
//	    return nil
//	})
func WithHTTPRoutes(register HTTPRouteRegistrar) Option {
	return WithSetup(func(rt *HTTPRuntime) error {
		if register == nil {
			return nil
		}
		if rt == nil {
			return errors.Join(ErrHTTPRuntimeUnavailable, errors.New("runtime is nil"))
		}
		if rt.Router == nil {
			return errors.Join(ErrHTTPRuntimeUnavailable, errors.New("runtime router is nil"))
		}
		if rt.Container == nil {
			return errors.Join(ErrHTTPRuntimeUnavailable, errors.New("runtime container is nil"))
		}
		if err := register(rt.Router, rt.Container); err != nil {
			return errors.Join(ErrHTTPRouteRegistrationFailed, err)
		}
		return nil
	})
}

func composeSetup(prev, next SetupFunc) SetupFunc {
	switch {
	case prev == nil:
		return next
	case next == nil:
		return prev
	default:
		return func(rt *HTTPRuntime) error {
			if err := prev(rt); err != nil {
				return err
			}
			return next(rt)
		}
	}
}

// MakeGRPCConnFactory returns the proto-first gRPC connection factory from the container.
// MakeGRPCConnFactory 获取 Proto-first gRPC 连接工厂。
func MakeGRPCConnFactory(c runtimecontract.Container) (transportcontract.GRPCConnFactory, error) {
	return frameworkcontainer.MakeGRPCConnFactory(c)
}

// MakeGRPCServerRegistrar returns the proto-first gRPC server registrar from the container.
// MakeGRPCServerRegistrar 获取 Proto-first gRPC 服务端注册器。
func MakeGRPCServerRegistrar(c runtimecontract.Container) (transportcontract.GRPCServerRegistrar, error) {
	return frameworkcontainer.MakeGRPCServerRegistrar(c)
}

// MakeDistributedLock returns the distributed lock capability from the container.
// MakeDistributedLock 获取分布式锁能力。
func MakeDistributedLock(c runtimecontract.Container) (datacontract.DistributedLock, error) {
	return frameworkcontainer.MakeDistributedLock(c)
}

// MakeMessagePublisher returns the message publishing capability from the container.
// MakeMessagePublisher 获取消息发布能力。
func MakeMessagePublisher(c runtimecontract.Container) (integrationcontract.MessagePublisher, error) {
	return frameworkcontainer.MakeMessagePublisher(c)
}

// MakeMessageSubscriber returns the message subscription capability from the container.
// MakeMessageSubscriber 获取消息订阅能力。
func MakeMessageSubscriber(c runtimecontract.Container) (integrationcontract.MessageSubscriber, error) {
	return frameworkcontainer.MakeMessageSubscriber(c)
}

// WithServiceIdentity writes service identity into the context.
// WithServiceIdentity 把服务身份写入上下文。
func WithServiceIdentity(ctx context.Context, identity *securitycontract.ServiceIdentity) context.Context {
	return securitycontract.NewServiceIdentityContext(ctx, identity)
}

// FromServiceIdentity reads service identity from the context.
// FromServiceIdentity 读取上下文中的服务身份。
func FromServiceIdentity(ctx context.Context) (*securitycontract.ServiceIdentity, bool) {
	return securitycontract.FromServiceIdentityContext(ctx)
}

// GetGRPCTraceID reads the trace id from a gRPC context.
// GetGRPCTraceID 从 gRPC context 读取 trace id。
func GetGRPCTraceID(ctx context.Context) string {
	return frameworkgrpc.GetTraceID(ctx)
}

// GetGRPCRequestID reads the request id from a gRPC context.
// GetGRPCRequestID 从 gRPC context 读取 request id。
func GetGRPCRequestID(ctx context.Context) string {
	return frameworkgrpc.GetRequestID(ctx)
}
