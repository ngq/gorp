// Package validate provides a unified validation service using go-playground/validator.
// The package supports locale-aware error messages (zh/en) and custom validation rules.
// Configuration via config.yaml:
//
// 验证服务包，使用 go-playground/validator 提供统一的参数校验能力。
// 支持本地化的错误消息（中文/英文）和自定义校验规则。
// 通过 config.yaml 配置：
//
//	validation:
//	  enabled: true
//	  locale: zh            # 错误消息语言（zh/en）
//	  translate_errors: true # 是否翻译错误（true=中文，false=英文原始错误，性能更好）
//
// Eg:
//
//	// 注册 Provider
//	app.Register(validate.NewProvider())
//
//	// 使用验证服务
//	vSvc := c.MustMake(datacontract.ValidatorKey).(datacontract.Validator)
//	err := vSvc.Validate(ctx, &UserRequest{Name: "test", Email: "test@example.com"})
package validate

import (
	"errors"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

// Provider registers the validation service contract.
//
// Provider 注册验证服务契约。
type Provider struct{}

// NewProvider creates a new validate provider instance.
//
// NewProvider 创建新的验证 Provider 实例。
func NewProvider() *Provider { return &Provider{} }

// Name returns the provider name "validate".
//
// Name 返回 Provider 名称 "validate"。
func (p *Provider) Name() string { return "validate" }

// IsDefer returns true, validation can be deferred until first use.
//
// IsDefer 返回 true，验证服务可延迟初始化。
func (p *Provider) IsDefer() bool { return true }

// Provides returns the validator contract key.
//
// Provides 返回验证器契约键。
func (p *Provider) Provides() []string { return []string{datacontract.ValidatorKey} }

// Register binds the validator service factory to the container.
//
// Register 将验证服务工厂绑定到容器。
func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(datacontract.ValidatorKey, func(c runtimecontract.Container) (any, error) {
		cfg, err := getValidateConfig(c)
		if err != nil {
			return nil, err
		}
		return NewValidatorService(cfg)
	}, true)

	return nil
}

// Boot is a no-op for validate provider.
//
// Boot 验证 Provider 无启动逻辑。
func (p *Provider) Boot(runtimecontract.Container) error { return nil }

// getValidateConfig retrieves validation config from config service.
// Default: enabled=true, locale=zh, translate_errors=true if config not available.
//
// getValidateConfig 从配置服务获取验证配置。
// 默认值：enabled=true, locale=zh, translate_errors=true（配置不可用时）。
func getValidateConfig(c runtimecontract.Container) (*datacontract.ValidatorConfig, error) {
	cfgAny, err := c.Make(datacontract.ConfigKey)
	if err != nil {
		return &datacontract.ValidatorConfig{
			Enabled:         true,
			Locale:          "zh",
			TranslateErrors: true,
		}, nil
	}

	cfg, ok := cfgAny.(datacontract.Config)
	if !ok {
		return nil, errors.New("validate: invalid config service")
	}

	vCfg := &datacontract.ValidatorConfig{
		Enabled:         true,
		Locale:          "zh",
		TranslateErrors: true,
	}

	if v := cfg.Get("validation.enabled"); v != nil {
		vCfg.Enabled = cfg.GetBool("validation.enabled")
	}
	if v := cfg.Get("validation.locale"); v != nil {
		vCfg.Locale = cfg.GetString("validation.locale")
	}
	if v := cfg.Get("validation.translate_errors"); v != nil {
		vCfg.TranslateErrors = cfg.GetBool("validation.translate_errors")
	}

	return vCfg, nil
}