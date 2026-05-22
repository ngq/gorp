// Package error_reporter provides error reporting provider for gorp framework.
// Registers error reporting service with Sentry or Log backend based on configuration.
// Supports synchronous and asynchronous error reporting with stack trace.
//
// 错误上报包提供错误上报 provider，用于 gorp 框架。
// 根据配置注册 Sentry 或 Log 后端的错误上报服务。
// 支持带堆栈的同步和异步错误上报。
package error_reporter

import (
	"github.com/ngq/gorp/framework/container"
	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

// Provider provides error reporting service for gorp framework.
// Core logic: Bind reporter factory, choose Sentry or Log based on config.
//
// Provider 提供错误上报服务，用于 gorp 框架。
// 核心逻辑：绑定上报器工厂、根据配置选择 Sentry 或日志。
type Provider struct {
	config resiliencecontract.ErrorReporterConfig
}

// NewProvider creates a new error reporter provider with configuration.
//
// NewProvider 创建新的错误上报 provider，携带配置。
func NewProvider(config resiliencecontract.ErrorReporterConfig) *Provider {
	return &Provider{config: config}
}

// Name returns provider name for identification.
//
// Name 返回 provider 名称，用于标识。
func (p *Provider) Name() string { return "error_reporter" }

// IsDefer indicates error reporter should not defer loading.
// Must be available for early error capture.
//
// IsDefer 表示错误上报器不应延迟加载。
// 必须尽早可用以捕获错误。
func (p *Provider) IsDefer() bool { return false }

// Provides returns the capability keys this provider exposes.
// Exposes ErrorReporterKey for error reporting service.
//
// Provides 返回 provider 暴露的能力键。
// 暴露 ErrorReporterKey 用于错误上报服务。
func (p *Provider) Provides() []string { return []string{resiliencecontract.ErrorReporterKey} }

// Register binds the error reporter factory to the container.
// Core logic: Create Sentry adapter if DSN configured, otherwise create Log reporter.
//
// Register 将错误上报器工厂绑定到容器。
// 核心逻辑：若配置了 DSN 则创建 Sentry 适配器，否则创建日志上报器。
func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(resiliencecontract.ErrorReporterKey, func(c runtimecontract.Container) (interface{}, error) {
		if p.config.Enabled && p.config.DSN != "" {
			return NewSentryAdapter(p.config), nil
		}

		logger, err := container.MakeWith[observabilitycontract.Logger](c, observabilitycontract.LogKey)
		if err != nil {
			return nil, err
		}
		return NewLogReporter(logger), nil
	}, true)
	return nil
}

// Boot initializes the error reporter provider.
// No additional startup logic required.
//
// Boot 初始化错误上报 provider。
// 无需额外启动逻辑。
func (p *Provider) Boot(runtimecontract.Container) error { return nil }
