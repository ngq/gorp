// Package application provides application startup entrypoints for gorp framework.
// This file exposes startup option builders for HTTP mode, providers, hooks.
// Lets business code declare startup configuration succinctly.
//
// 应用启动包提供 gorp 框架的应用启动入口。
// 本文件暴露启动选项构造入口，包括 HTTP 模式、provider、钩子。
// 让业务代码可以简洁声明启动配置。
//
// Eg:
//
//	application.Run("my-service",
//	    application.HTTP(),
//	    application.WithProviders(configprovider.NewProvider()),
//	    application.WithHTTPRoutes(registerRoutes),
//	)
package application

import (
	"errors"

	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
)

// HTTP declares that the default HTTP mainline should be used.
//
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
		if h.GovernanceMode != "" {
			cfg.httpOpts.GovernanceMode = string(h.GovernanceMode)
		}
		if len(h.GovernanceDisable) > 0 {
			cfg.httpOpts.GovernanceDisable = append([]string(nil), h.GovernanceDisable...)
		}
		if len(h.GovernanceProviders) > 0 {
			cfg.httpOpts.GovernanceProviders = cloneGovernanceProviders(h.GovernanceProviders)
		}
	})
}

// WithoutHTTP explicitly disables the default HTTP declaration.
//
// WithoutHTTP 显式关闭默认 HTTP 声明。
func WithoutHTTP() Option {
	return optionFunc(func(cfg *runConfig) {
		cfg.httpEnabled = false
	})
}

// Module declares providers for a single module.
//
// Module 声明单个模块的 providers。
func Module(providers ...ServiceProvider) Option {
	return WithProviders(providers...)
}

// Modules declares providers for a group of modules.
//
// Modules 声明一组模块 providers。
func Modules(groups ...[]ServiceProvider) Option {
	return optionFunc(func(cfg *runConfig) {
		for _, providers := range groups {
			WithProviders(providers...).apply(cfg)
		}
	})
}

// WithModule is the explicit named alias of Module.
//
// WithModule 是 Module 的显式命名入口。
func WithModule(providers ...ServiceProvider) Option {
	return Module(providers...)
}

// WithProviders appends provider declarations without changing the startup semantics.
//
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
//
// WithMigrate 声明迁移回调。
//
// Example:
//
//	application.WithMigrate(func(rt *application.HTTPRuntime) error {
//	    return migrateSchema(rt.Container)
//	})
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
//
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
//
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

// WithGovernanceMode declares the startup governance mode explicitly.
//
// WithGovernanceMode 显式声明启动治理模式。
func WithGovernanceMode(mode resiliencecontract.GovernanceMode) Option {
	return optionFunc(func(cfg *runConfig) {
		if mode == "" {
			return
		}
		cfg.httpOpts.GovernanceMode = string(mode)
	})
}

// WithGovernanceDisabled explicitly disables one or more default governance capabilities.
//
// WithGovernanceDisabled 显式关闭一个或多个默认治理能力。
func WithGovernanceDisabled(names ...string) Option {
	return optionFunc(func(cfg *runConfig) {
		if len(names) == 0 {
			return
		}
		cfg.httpOpts.GovernanceDisable = append(cfg.httpOpts.GovernanceDisable, names...)
	})
}

// WithGovernanceEnabled explicitly enables one or more governance capabilities that are off by default.
// This is the symmetric counterpart of WithGovernanceDisabled.
// When the same feature appears in both enable and disable, disable takes precedence.
//
// WithGovernanceEnabled 显式开启一个或多个默认关闭的治理能力。
// 这是 WithGovernanceDisabled 的对称入口。
// 当同一 feature 同时出现在 enable 和 disable 中时，disable 生效。
func WithGovernanceEnabled(names ...string) Option {
	return optionFunc(func(cfg *runConfig) {
		if len(names) == 0 {
			return
		}
		cfg.httpOpts.GovernanceEnable = append(cfg.httpOpts.GovernanceEnable, names...)
	})
}

// WithGovernanceProvider explicitly overrides one governance provider backend.
//
// WithGovernanceProvider 显式覆盖一个治理 provider backend。
func WithGovernanceProvider(name, backend string) Option {
	return optionFunc(func(cfg *runConfig) {
		if name == "" || backend == "" {
			return
		}
		if cfg.httpOpts.GovernanceProviders == nil {
			cfg.httpOpts.GovernanceProviders = make(map[string]string)
		}
		cfg.httpOpts.GovernanceProviders[name] = backend
	})
}

// WithMicroserviceMode selects the default microservice governance mainline.
//
// WithMicroserviceMode 选择默认微服务治理主线。
func WithMicroserviceMode() Option {
	return WithGovernanceMode(resiliencecontract.GovernanceModeMicroservice)
}

// WithMonolithMode selects the lightweight monolith governance mode.
//
// WithMonolithMode 选择轻量单体治理模式。
func WithMonolithMode() Option {
	return WithGovernanceMode(resiliencecontract.GovernanceModeMonolith)
}

// WithGinFirstMode selects the Gin-first governance mode.
//
// WithGinFirstMode 选择 Gin-first 治理模式。
func WithGinFirstMode() Option {
	return WithGovernanceMode(resiliencecontract.GovernanceModeGinFirst)
}

func cloneGovernanceProviders(src map[string]string) map[string]string {
	if len(src) == 0 {
		return nil
	}
	cloned := make(map[string]string, len(src))
	for key, value := range src {
		cloned[key] = value
	}
	return cloned
}
