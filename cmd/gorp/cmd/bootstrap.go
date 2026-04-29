package cmd

import (
	"os"
	"sync"

	frameworkbootstrap "github.com/ngq/gorp/framework/bootstrap"
	"github.com/ngq/gorp/framework/contract"
	"github.com/ngq/gorp/framework"
)

type bootstrapOption func(*bootstrapConfig)

type bootstrapConfig struct {
	extraProviders  []contract.ServiceProvider
	runtimeProvider contract.ServiceProvider
}

var bootstrapHooks struct {
	mu              sync.RWMutex
	extraProviders  []contract.ServiceProvider
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
// - 这是工具链内部 bootstrap 的附加 provider 扩展位；
// - 主要服务于生成项目在共享 CLI / 兼容命令链路下补充自己的 provider；
// - 不修改 framework 默认 provider 组，只做顺序明确的附加注册；
// - 传入 nil 会被忽略；
// - 业务服务默认公开启动入口仍应优先走项目自己的 `cmd/*/main.go`。
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
// - 这是工具链内部 bootstrap 的兼容 runtime 扩展位；
// - 主要服务于共享 CLI / legacy 辅助命令链路，不是 starter 默认公开启动入口；
// - 若未指定，则继续保持 framework/bootstrap.NewCLIApplication() 当前装配结果；
// - 这样母仓与模板项目可以共享同一套 CLI，但业务服务默认主线仍落在项目自己的启动入口。
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
// - 这是给模板项目高级扩展位使用的兼容入口，不是业务服务默认公开启动入口；
// - 主要用于在共享 CLI / legacy 辅助命令链路下，把项目自己的 runtime/provider 装配注入工具链 bootstrap；
// - 普通 starter 用户默认仍应通过项目自己的 `cmd/*/main.go` 启动；
// - examples 如保留，也只应作为参考与可选验证素材，不构成这层 hook 的默认语义来源；
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

	appRuntime, c, err := frameworkbootstrap.NewCLIApplication()
	if err != nil {
		return nil, nil, err
	}

	runtimeProvider := cfg.runtimeProvider
	if runtimeProvider != nil {
		if err := c.RegisterProvider(runtimeProvider); err != nil {
			return nil, nil, err
		}
	}
	if len(cfg.extraProviders) > 0 {
		if err := c.RegisterProviders(cfg.extraProviders...); err != nil {
			return nil, nil, err
		}
	}

	return appRuntime, c, nil
}
