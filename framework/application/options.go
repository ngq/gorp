// Application scenarios:
// - Expose startup option builders for the application package.
// - Let business code declare HTTP mode, provider sets, migration hooks, and route registration succinctly.
// - Keep option composition stable while allowing setup and route hooks to grow incrementally.
//
// 适用场景：
// - 暴露 application 包的启动选项构造入口。
// - 让业务代码可以简洁声明 HTTP 模式、provider 集合、迁移钩子和路由注册逻辑。
// - 在保持选项组合语义稳定的前提下，支持 setup 与路由钩子逐步扩展。
package application

import "errors"

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
