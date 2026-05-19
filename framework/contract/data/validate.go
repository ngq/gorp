// Application scenarios:
// - Define the validation contract shared by validation providers and HTTP middleware.
// - Standardize object validation, locale switching, custom rule registration, and error translation.
// - Provide reusable models for validation errors and validation-specific config.
//
// 适用场景：
// - 定义校验 provider 和 HTTP 中间件共同依赖的校验契约。
// - 统一对象校验、语言切换、自定义规则注册和错误翻译语义。
// - 为校验错误和校验专用配置提供可复用的数据模型。
package data

import (
	"context"
)

// ValidatorKey is the container key for the validator capability.
//
// ValidatorKey 是校验器能力的容器键。
const ValidatorKey = "framework.validator"

// Validator defines the object validation capability exposed by the framework.
//
// Validator 定义框架对外暴露的对象校验能力。
type Validator interface {
	// Validate validates the target object.
	//
	// Validate 校验目标对象。
	Validate(ctx context.Context, obj any) error

	// ValidateVar validates a single value against a validation tag.
	//
	// ValidateVar 使用指定 tag 校验单个值。
	ValidateVar(ctx context.Context, field any, tag string) error

	// RegisterCustom registers a named custom validation rule.
	//
	// RegisterCustom 注册具名自定义校验规则。
	RegisterCustom(name string, fn CustomValidateFunc) error

	// SetLocale switches the active validation locale.
	//
	// SetLocale 切换当前校验语言环境。
	SetLocale(locale string) error

	// TranslateError converts a raw validation error into an app error.
	// Returns an error that can be cast to resiliencecontract.AppError if needed.
	//
	// TranslateError 将原始校验错误转换为应用错误。
	// 返回的 error 可在需要时断言为 resiliencecontract.AppError。
	TranslateError(err error) error
}

// CustomValidateFunc defines a custom validation rule function.
//
// CustomValidateFunc 定义自定义校验规则函数。
type CustomValidateFunc func(ctx context.Context, field any) bool

// ValidationError describes one field-level validation failure.
//
// ValidationError 描述一次字段级校验失败。
type ValidationError struct {
	Field   string
	Tag     string
	Message string
	Value   any
}

// ValidationErrors is a collection of validation failures.
//
// ValidationErrors 是校验失败集合。
//
//nolint:errname // 命名表示"多个校验错误的集合"，符合语义
type ValidationErrors []ValidationError

// Error returns the first validation message as the aggregate error text.
//
// Error 返回首个校验消息作为聚合错误文本。
func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return ""
	}

	// Build the message list once so callers can reuse the same aggregation semantics.
	// 先统一构建消息列表，保证调用方拿到一致的聚合语义。
	msgs := make([]string, len(e))
	for i, v := range e {
		msgs[i] = v.Message
	}

	// Keep the Error() contract concise by exposing the first validation message.
	// 保持 Error() 结果简洁，仅暴露首个校验消息。
	return msgs[0]
}

// Errors returns all validation messages as plain strings.
//
// Errors 以纯字符串形式返回全部校验消息。
func (e ValidationErrors) Errors() []string {
	if len(e) == 0 {
		return nil
	}

	// Preserve the full validation message set for callers that need all failures.
	// 保留完整校验消息集合，供需要全部失败信息的调用方使用。
	msgs := make([]string, len(e))
	for i, v := range e {
		msgs[i] = v.Message
	}
	return msgs
}

// ValidatorConfig describes validator-level configuration.
//
// ValidatorConfig 描述校验器级配置。
type ValidatorConfig struct {
	Enabled         bool
	Locale          string
	TranslateErrors bool // TranslateErrors controls whether to translate validation errors.
	// When false, returns raw English errors for better performance.
	//
	// TranslateErrors 控制是否翻译校验错误。
	// 为 false 时返回原始英文错误，性能更好。

	CustomRules map[string]CustomRuleConfig
}

// CustomRuleConfig describes one custom validation rule declaration.
//
// CustomRuleConfig 描述单个自定义校验规则声明。
type CustomRuleConfig struct {
	Name    string
	Message string
	Fn      CustomValidateFunc
}
