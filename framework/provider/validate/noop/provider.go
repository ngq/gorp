package noop

import (
	"context"

	"github.com/ngq/gorp/framework/contract"
)

// Provider 提供 noop Validator 实现。
//
// 中文说明：
// - 单体模式下使用，零依赖；
// - 所有验证操作直接通过，不执行实际验证；
// - 错误消息不做翻译处理。
type Provider struct{}

// NewProvider 创建 noop Validator Provider。
func NewProvider() *Provider { return &Provider{} }

// Name 返回 Provider 名称。
func (p *Provider) Name() string { return "validate.noop" }

// IsDefer 返回是否延迟加载。
func (p *Provider) IsDefer() bool { return true }

// Provides 返回提供的服务 key。
func (p *Provider) Provides() []string {
	return []string{contract.ValidatorKey}
}

// Register 注册 noop Validator 服务。
func (p *Provider) Register(c contract.Container) error {
	c.Bind(contract.ValidatorKey, func(c contract.Container) (any, error) {
		return &noopValidator{}, nil
	}, true)

	return nil
}

// Boot 启动 Provider。
func (p *Provider) Boot(c contract.Container) error {
	return nil
}

// noopValidator noop Validator 实现。
//
// 中文说明：
// - 所有验证操作直接返回 nil（通过）；
// - 不执行任何验证逻辑。
type noopValidator struct{}

// Validate 总是返回 nil（不验证）。
func (v *noopValidator) Validate(ctx context.Context, obj any) error {
	return nil
}

// ValidateVar 总是返回 nil（不验证）。
func (v *noopValidator) ValidateVar(ctx context.Context, field any, tag string) error {
	return nil
}

// RegisterCustom 空操作（noop 模式不支持自定义规则）。
func (v *noopValidator) RegisterCustom(name string, fn contract.CustomValidateFunc) error {
	return nil
}

// SetLocale 空操作（noop 模式不支持语言切换）。
func (v *noopValidator) SetLocale(locale string) error {
	return nil
}

// TranslateError 将原始错误包装为 AppError。
//
// 中文说明：
// - noop 模式下不做翻译处理；
// - 直接包装为 BadRequest 错误。
func (v *noopValidator) TranslateError(err error) contract.AppError {
	if err == nil {
		return nil
	}
	return contract.BadRequest(contract.ErrorReasonBadRequest, err.Error())
}