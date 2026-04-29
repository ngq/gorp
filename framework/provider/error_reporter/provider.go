package error_reporter

import (
	"github.com/ngq/gorp/framework/contract"
)

// Provider 错误上报服务提供者。
type Provider struct {
	config contract.ErrorReporterConfig
}

// NewProvider 创建错误上报服务提供者。
func NewProvider(config contract.ErrorReporterConfig) *Provider {
	return &Provider{config: config}
}

// Name 返回 provider 名称。
func (p *Provider) Name() string { return "error_reporter" }

// IsDefer 表示 error_reporter provider 不走延迟加载。
func (p *Provider) IsDefer() bool { return false }

// Provides 返回当前 provider 暴露的能力 key。
func (p *Provider) Provides() []string { return []string{contract.ErrorReporterKey} }

// Register 绑定错误上报服务。
func (p *Provider) Register(c contract.Container) error {
	c.Bind(contract.ErrorReporterKey, func(c contract.Container) (interface{}, error) {
		// 如果配置了 Sentry DSN，使用 Sentry adapter
		if p.config.Enabled && p.config.DSN != "" {
			return NewSentryAdapter(p.config), nil
		}
		// 否则使用日志 fallback
		logAny, err := c.Make(contract.LogKey)
		if err != nil {
			return nil, err
		}
		logger := logAny.(contract.Logger)
		return NewLogReporter(logger), nil
	}, true)
	return nil
}

// Boot error_reporter provider 无额外启动逻辑。
func (p *Provider) Boot(contract.Container) error { return nil }