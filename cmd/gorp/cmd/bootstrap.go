package cmd

import (
	"os"
	"sync"

	"github.com/ngq/gorp/app/provider/runtime_provider"
	frameworkbootstrap "github.com/ngq/gorp/framework/bootstrap"
	"github.com/ngq/gorp/framework/contract"
	"github.com/ngq/gorp/framework/provider/orm/ent"

	"github.com/ngq/gorp/framework"
)

type bootstrapOption func(*bootstrapConfig)

type bootstrapConfig struct {
	extraProviders []contract.ServiceProvider
	runtimeProvider contract.ServiceProvider
}

var bootstrapHooks struct {
	mu sync.RWMutex
	extraProviders []contract.ServiceProvider
	runtimeProvider contract.ServiceProvider
}

// WithAppEnv 在 bootstrap 前设置 APP_ENV。
func WithAppEnv(env string) bootstrapOption {
	return func(cfg *bootstrapConfig) {
		if env == "" {
			return
		}
		_ = os.Setenv("APP_ENV", env)
	}
}

// WithExtraProviders 为当前进程追加项目侧 provider。
//
// 中文说明：
// - 用于生成项目在调用 cmd.Execute() 前，把自己的 service/runtime provider 注入 CLI bootstrap；
// - 不修改母仓默认 provider 组，只做顺序明确的附加注册；
// - 传入 nil 会被忽略。
func WithExtraProviders(providers ...contract.ServiceProvider) bootstrapOption {
	return func(cfg *bootstrapConfig) {
		for _, p := range providers {
			if p == nil {
				continue
			}
			cfg.extraProviders = append(cfg.extraProviders, p)
		}
	}
}

// WithRuntimeProvider 指定当前进程使用的 runtime provider。
//
// 中文说明：
// - 生成项目可通过此入口覆盖母仓默认 runtime provider；
// - 它主要服务于共享 CLI 下的 legacy/runtime 命令组装配，不是 starter 默认公开启动入口；
// - 若未指定，则继续回退到母仓 runtime_provider.NewProvider()；
// - 这样母仓与模板项目可以共享同一套 CLI，但各自拥有自己的 runtime 装配。
func WithRuntimeProvider(p contract.ServiceProvider) bootstrapOption {
	return func(cfg *bootstrapConfig) {
		if p != nil {
			cfg.runtimeProvider = p
		}
	}
}

// RegisterBootstrapProviders 注册当前进程的全局 bootstrap hook。
//
// 中文说明：
// - 这是给模板项目高级扩展位使用的稳定入口；
// - 主要用于把项目自己的 runtime/provider 装配注入共享 CLI bootstrap；
// - 普通 starter 用户默认仍应通过项目自己的 `cmd/*/main.go` 启动；
// - 在调用 cmd.Execute() 前执行一次即可；
// - 会覆盖当前进程之前注册的 runtime provider，并替换 extra providers 列表。
func RegisterBootstrapProviders(runtimeProvider contract.ServiceProvider, extraProviders ...contract.ServiceProvider) {
	bootstrapHooks.mu.Lock()
	defer bootstrapHooks.mu.Unlock()

	bootstrapHooks.runtimeProvider = runtimeProvider
	bootstrapHooks.extraProviders = bootstrapHooks.extraProviders[:0]
	for _, p := range extraProviders {
		if p == nil {
			continue
		}
		bootstrapHooks.extraProviders = append(bootstrapHooks.extraProviders, p)
	}
}

func readBootstrapHooks() bootstrapConfig {
	bootstrapHooks.mu.RLock()
	defer bootstrapHooks.mu.RUnlock()

	cfg := bootstrapConfig{}
	if bootstrapHooks.runtimeProvider != nil {
		cfg.runtimeProvider = bootstrapHooks.runtimeProvider
	}
	if len(bootstrapHooks.extraProviders) > 0 {
		cfg.extraProviders = append(cfg.extraProviders, bootstrapHooks.extraProviders...)
	}
	return cfg
}

func bootstrap(opts ...bootstrapOption) (*framework.Application, contract.Container, error) {
	cfg := readBootstrapHooks()
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}

	appRuntime := framework.NewApplication()
	c := appRuntime.Container()

	if err := c.RegisterProviders(frameworkbootstrap.DefaultProviders()...); err != nil {
		return nil, nil, err
	}
	if err := frameworkbootstrap.RegisterSelectedMicroserviceProviders(c); err != nil {
		return nil, nil, err
	}
	if err := c.RegisterProvider(ent.NewProvider()); err != nil {
		return nil, nil, err
	}

	runtimeProvider := cfg.runtimeProvider
	if runtimeProvider == nil {
		runtimeProvider = runtime_provider.NewProvider()
	}
	if err := c.RegisterProvider(runtimeProvider); err != nil {
		return nil, nil, err
	}
	if len(cfg.extraProviders) > 0 {
		if err := c.RegisterProviders(cfg.extraProviders...); err != nil {
			return nil, nil, err
		}
	}

	return appRuntime, c, nil
}
