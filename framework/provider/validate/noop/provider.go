// Package noop provides a no-op validator implementation for monolith scenarios.
// This validator always passes validation and does nothing.
// Use in monolith applications where validation is handled elsewhere.
//
// 空验证器实现包，用于单体应用场景。
// 此验证器始终通过验证，不执行任何操作。
// 用于单体应用，验证逻辑在其他地方处理。
package noop

import (
	"context"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

// Provider registers a no-op validator contract.
//
// Provider 注册空验证器契约。
type Provider struct{}

// NewProvider creates a new no-op validator provider instance.
//
// NewProvider 创建新的空验证器 Provider 实例。
func NewProvider() *Provider { return &Provider{} }

// Name returns the provider name "validate.noop".
//
// Name 返回 Provider 名称 "validate.noop"。
func (p *Provider) Name() string { return "validate.noop" }

// IsDefer returns true, validator can be deferred until first use.
//
// IsDefer 返回 true，验证器可延迟初始化直到首次使用。
func (p *Provider) IsDefer() bool { return true }

// Provides returns the validator contract key.
//
// Provides 返回验证器契约键。
func (p *Provider) Provides() []string { return []string{datacontract.ValidatorKey} }

// DependsOn returns the keys this provider depends on.
// Noop validator has no dependencies.
//
// DependsOn 返回该 provider 依赖的 key。
// Noop validator 无依赖。
func (p *Provider) DependsOn() []string { return nil }

// Register binds the no-op validator to the container.
//
// Register 将空验证器绑定到容器。
func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(datacontract.ValidatorKey, func(c runtimecontract.Container) (any, error) {
		return &noopValidator{}, nil
	}, true)
	return nil
}

// Boot is a no-op for this provider.
//
// Boot 此 Provider 无启动逻辑。
func (p *Provider) Boot(runtimecontract.Container) error { return nil }

// noopValidator implements datacontract.Validator with no-op behavior.
//
// noopValidator 使用空行为实现 datacontract.Validator 接口。
type noopValidator struct{}

// Validate always returns nil (no validation performed).
//
// Validate 始终返回 nil（不执行验证）。
func (v *noopValidator) Validate(ctx context.Context, obj any) error {
	return nil
}

// ValidateVar always returns nil (no validation performed).
//
// ValidateVar 始终返回 nil（不执行验证）。
func (v *noopValidator) ValidateVar(ctx context.Context, field any, tag string) error {
	return nil
}

// RegisterCustom does nothing and returns nil.
//
// RegisterCustom 不执行任何操作并返回 nil。
func (v *noopValidator) RegisterCustom(name string, fn datacontract.CustomValidateFunc) error {
	return nil
}

// SetLocale does nothing and returns nil.
//
// SetLocale 不执行任何操作并返回 nil。
func (v *noopValidator) SetLocale(locale string) error {
	return nil
}

// TranslateError wraps the error as BadRequest AppError.
// Returns error interface; caller can cast to resiliencecontract.AppError if needed.
//
// TranslateError 将错误包装为 BadRequest AppError。
// 返回 error 接口；调用方可在需要时断言为 resiliencecontract.AppError。
func (v *noopValidator) TranslateError(err error) error {
	if err == nil {
		return nil
	}
	return resiliencecontract.BadRequest(resiliencecontract.ErrorReasonBadRequest, err.Error())
}
