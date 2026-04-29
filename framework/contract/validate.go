package contract

import (
	"context"
)

// ValidatorKey 是 Validator 服务在容器中的绑定 key。
const ValidatorKey = "framework.validator"

// Validator 数据验证接口。
//
// 中文说明：
// - 提供结构体和变量级别的验证能力；
// - 支持自定义验证规则注册；
// - 支持国际化错误消息（中文/英文）；
// - 验证错误统一转换为 AppError 格式。
type Validator interface {
	// Validate 验证结构体。
	//
	// 中文说明：
	// - 使用 validator/v10 进行结构体验证；
	// - 验证失败返回 AppError，包含字段级错误详情；
	// - 验证规则默认通过 `validate` tag 定义。
	Validate(ctx context.Context, obj any) error

	// ValidateVar 验证单个变量。
	//
	// 中文说明：
	// - 用于验证单个字段或变量；
	// - tag 参数指定验证规则（如 "required,email"）。
	ValidateVar(ctx context.Context, field any, tag string) error

	// RegisterCustom 注册自定义验证规则。
	//
	// 中文说明：
	// - name 为规则名称（如 "mobile"、"id_card"）；
	// - 注册后可在 `validate` tag 中使用（如 `validate:"required,mobile"`）。
	RegisterCustom(name string, fn CustomValidateFunc) error

	// SetLocale 设置错误消息语言。
	//
	// 中文说明：
	// - 支持 "zh"（中文）和 "en"（英文）；
	// - 默认使用 "zh"。
	SetLocale(locale string) error

	// TranslateError 翻译验证错误为 AppError。
	//
	// 中文说明：
	// - 将 validator.ValidationErrors 转换为 AppError；
	// - 包含字段级错误详情在 Metadata 中。
	TranslateError(err error) AppError
}

// CustomValidateFunc 自定义验证函数。
//
// 中文说明：
// - ctx 用于传递上下文信息；
// - field 为待验证的字段值；
// - 返回 true 表示验证通过，false 表示验证失败。
type CustomValidateFunc func(ctx context.Context, field any) bool

// ValidationError 验证错误详情。
//
// 中文说明：
// - 描述单个字段的验证错误；
// - 用于构建详细的错误响应。
type ValidationError struct {
	// Field 错误字段名（当前默认使用结构体字段名）
	Field string

	// Tag 验证规则标签（如 "required"、"email"）
	Tag string

	// Message 翻译后的错误消息
	Message string

	// Value 实际值
	Value any
}

// ValidationErrors 多个验证错误。
//
// 中文说明：
// - 包含所有验证失败的字段错误；
// - 实现 error 接口。
type ValidationErrors []ValidationError

// Error 实现 error 接口。
func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return ""
	}

	msgs := make([]string, len(e))
	for i, v := range e {
		msgs[i] = v.Message
	}

	return msgs[0] // 返回第一个错误消息
}

// Errors 返回所有错误消息。
func (e ValidationErrors) Errors() []string {
	if len(e) == 0 {
		return nil
	}

	msgs := make([]string, len(e))
	for i, v := range e {
		msgs[i] = v.Message
	}
	return msgs
}

// ValidatorConfig 验证器配置。
//
// 中文说明：
// - 定义验证器的启用状态、语言、自定义规则等。
type ValidatorConfig struct {
	// Enabled 是否启用验证
	Enabled bool

	// Locale 错误消息语言（默认 "zh"）
	Locale string

	// CustomRules 自定义规则配置
	CustomRules map[string]CustomRuleConfig
}

// CustomRuleConfig 自定义规则配置。
type CustomRuleConfig struct {
	// Name 规则名称
	Name string

	// Message 自定义错误消息（覆盖默认翻译）
	Message string

	// Fn 验证函数（可选）
	Fn CustomValidateFunc
}