package facade

import (
	"context"
	"errors"
	"strings"

	"github.com/ngq/gorp/framework/bootstrap"
	frameworkcontainer "github.com/ngq/gorp/framework/container"
	"github.com/ngq/gorp/framework/contract"
	frameworkgrpc "github.com/ngq/gorp/framework/provider/grpc"
)

var (
	// ErrServiceNameRequired 表示未提供服务名。
	ErrServiceNameRequired = errors.New("facade: serviceName is required")
	// ErrNoServiceDeclared 表示未声明可运行服务。
	ErrNoServiceDeclared = errors.New("facade: no service declared")
	// ErrHTTPRouteRegistrationFailed 表示 HTTP 路由注册失败。
	ErrHTTPRouteRegistrationFailed = errors.New("facade: http route registration failed")
	// ErrHTTPRuntimeUnavailable 表示 HTTP 路由注册阶段缺少可用 runtime。
	ErrHTTPRuntimeUnavailable = errors.New("facade: http runtime unavailable")
	// ErrSetupFailed 表示 setup 回调执行失败。
	ErrSetupFailed = errors.New("facade: setup failed")
	// ErrMigrateFailed 表示 migrate 回调执行失败。
	ErrMigrateFailed = errors.New("facade: migrate failed")
	// ErrStartupCanceled 表示启动前 context 已取消。
	ErrStartupCanceled = errors.New("facade: startup canceled")
	// ErrHTTPServiceRunFailed 表示 HTTP 服务启动失败。
	ErrHTTPServiceRunFailed = errors.New("facade: http service run failed")
	// ErrHTTPRuntimeBuildFailed 表示 HTTP runtime 构建失败。
	ErrHTTPRuntimeBuildFailed = errors.New("facade: http runtime build failed")
)

var (
	bootHTTPService    = bootstrap.BootHTTPService
	newHTTPRuntimeFunc = bootstrap.NewHTTPServiceRuntime
)

// HTTPRuntime 是 facade 回调使用的启动上下文。
type HTTPRuntime = bootstrap.HTTPServiceRuntime

// ServiceProvider 复用 contract 层 provider 声明。
type ServiceProvider = contract.ServiceProvider

// MigrateFunc 定义迁移回调契约。
type MigrateFunc func(*HTTPRuntime) error

// SetupFunc 定义启动装配回调契约。
type SetupFunc func(*HTTPRuntime) error

// HTTPRouteRegistrar 定义默认 HTTP 路由注册契约。
type HTTPRouteRegistrar func(router contract.HTTPRouter, container contract.Container) error

// HTTPServiceOptions 是 facade 暴露的最小 HTTP 选项视图。
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

// Option 用于声明 facade 启动配置。
type Option interface {
	apply(*runConfig)
}

type optionFunc func(*runConfig)

func (f optionFunc) apply(cfg *runConfig) {
	if f != nil {
		f(cfg)
	}
}

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

// Run 启动默认 HTTP 主线。
// 说明：业务运行入口仍是项目自己的 main；facade 只提供启动装配 helper。
func Run(serviceName string, options ...Option) error {
	return RunContext(context.Background(), serviceName, options...)
}

// Start 是 Run 的同义入口。
func Start(serviceName string, options ...Option) error {
	return Run(serviceName, options...)
}

// RunContext 使用显式 context 启动默认主线。
// 当前语义：仅在启动前检查取消，运行中关闭流程仍由 bootstrap 处理。
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

// ensureStartupContext 在启动前校验 context 状态。
// 语义：启动前已取消优先返回取消错误，不进入后续参数与装配流程。
func ensureStartupContext(ctx context.Context) error {
	if ctx == nil {
		return nil
	}
	return ctx.Err()
}

// BuildHTTPRuntime 仅构建启动上下文，不启动监听。
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

// Build 是 BuildHTTPRuntime 的同义入口。
func Build(serviceName string, options ...Option) (*HTTPRuntime, error) {
	return BuildHTTPRuntime(serviceName, options...)
}

// HTTP 声明使用默认 HTTP 主线。
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

// WithoutHTTP 显式关闭默认 HTTP 声明。
func WithoutHTTP() Option {
	return optionFunc(func(cfg *runConfig) {
		cfg.httpEnabled = false
	})
}

// Module 声明单个模块的 providers。
func Module(providers ...ServiceProvider) Option {
	return WithProviders(providers...)
}

// Modules 声明一组模块 providers。
func Modules(groups ...[]ServiceProvider) Option {
	return optionFunc(func(cfg *runConfig) {
		for _, providers := range groups {
			WithProviders(providers...).apply(cfg)
		}
	})
}

// WithModule 是 Module 的显式命名入口。
func WithModule(providers ...ServiceProvider) Option {
	return Module(providers...)
}

// WithProviders 追加 providers 声明，不改变底层选型语义。
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

// WithSetup 声明装配回调。
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

// WithHTTPRoutes 声明默认 HTTP 路由注册回调。
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

// MakeGRPCConnFactory 获取 Proto-first gRPC 连接工厂。
func MakeGRPCConnFactory(c contract.Container) (contract.GRPCConnFactory, error) {
	return frameworkcontainer.MakeGRPCConnFactory(c)
}

// MakeGRPCServerRegistrar 获取 Proto-first gRPC 服务端注册器。
func MakeGRPCServerRegistrar(c contract.Container) (contract.GRPCServerRegistrar, error) {
	return frameworkcontainer.MakeGRPCServerRegistrar(c)
}

// MakeDistributedLock 获取分布式锁能力。
func MakeDistributedLock(c contract.Container) (contract.DistributedLock, error) {
	return frameworkcontainer.MakeDistributedLock(c)
}

// MakeMessagePublisher 获取消息发布能力。
func MakeMessagePublisher(c contract.Container) (contract.MessagePublisher, error) {
	return frameworkcontainer.MakeMessagePublisher(c)
}

// MakeMessageSubscriber 获取消息订阅能力。
func MakeMessageSubscriber(c contract.Container) (contract.MessageSubscriber, error) {
	return frameworkcontainer.MakeMessageSubscriber(c)
}

// WithServiceIdentity 把服务身份写入上下文。
func WithServiceIdentity(ctx context.Context, identity *contract.ServiceIdentity) context.Context {
	return contract.NewServiceIdentityContext(ctx, identity)
}

// FromServiceIdentity 读取上下文中的服务身份。
func FromServiceIdentity(ctx context.Context) (*contract.ServiceIdentity, bool) {
	return contract.FromServiceIdentityContext(ctx)
}

// GetGRPCTraceID 从 gRPC context 读取 trace id。
func GetGRPCTraceID(ctx context.Context) string {
	return frameworkgrpc.GetTraceID(ctx)
}

// GetGRPCRequestID 从 gRPC context 读取 request id。
func GetGRPCRequestID(ctx context.Context) string {
	return frameworkgrpc.GetRequestID(ctx)
}
