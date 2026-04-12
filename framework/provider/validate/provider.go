package validate

import (
	"errors"

	"github.com/ngq/gorp/framework/contract"
)

// Provider 提供 Validator 实现。
//
// 中文说明：
// - 基于 go-playground/validator/v10 实现；
// - 支持国际化错误消息（中文/英文）；
// - 支持自定义验证规则；
// - 验证错误转换为统一 AppError 格式。
type Provider struct{}

// NewProvider 创建 Validator Provider。
func NewProvider() *Provider { return &Provider{} }

// Name 返回 Provider 名称。
func (p *Provider) Name() string { return "validate" }

// IsDefer 返回是否延迟加载。
func (p *Provider) IsDefer() bool { return true }

// Provides 返回提供的服务 key。
func (p *Provider) Provides() []string {
	return []string{contract.ValidatorKey}
}

// Register 注册 Validator 服务。
//
// 中文说明：
// - 从容器获取配置，创建 ValidatorService；
// - 配置格式见 ValidatorConfig 结构体。
func (p *Provider) Register(c contract.Container) error {
	c.Bind(contract.ValidatorKey, func(c contract.Container) (any, error) {
		cfg, err := getValidateConfig(c)
		if err != nil {
			return nil, err
		}
		return NewValidatorService(cfg)
	}, true)

	return nil
}

// Boot 启动 Provider。
func (p *Provider) Boot(c contract.Container) error {
	return nil
}

// getValidateConfig 从容器获取验证器配置。
//
// 中文说明：
// - 配置路径：validation.enabled、validation.locale 等；
// - 未配置时使用默认值。
func getValidateConfig(c contract.Container) (*contract.ValidatorConfig, error) {
	cfgAny, err := c.Make(contract.ConfigKey)
	if err != nil {
		// 配置服务不可用时使用默认配置
		return &contract.ValidatorConfig{
			Enabled: true,
			Locale:  "zh",
		}, nil
	}

	cfg, ok := cfgAny.(contract.Config)
	if !ok {
		return nil, errors.New("validate: invalid config service")
	}

	vCfg := &contract.ValidatorConfig{
		Enabled: true,
		Locale:  "zh",
	}

	// 读取配置项
	if v := cfg.Get("validation.enabled"); v != nil {
		vCfg.Enabled = cfg.GetBool("validation.enabled")
	}
	if v := cfg.Get("validation.locale"); v != nil {
		vCfg.Locale = cfg.GetString("validation.locale")
	}

	return vCfg, nil
}