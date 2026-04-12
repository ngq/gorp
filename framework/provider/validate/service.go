package validate

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	enTranslations "github.com/go-playground/validator/v10/translations/en"
	zhTranslations "github.com/go-playground/validator/v10/translations/zh"

	"github.com/ngq/gorp/framework/contract"
)

// ValidatorService 实现 Validator 接口。
//
// 中文说明：
// - 基于 go-playground/validator/v10；
// - 支持中文/英文错误消息；
// - 验证错误转换为统一 AppError 格式。
type ValidatorService struct {
	validate *validator.Validate
	trans    ut.Translator
	cfg      *contract.ValidatorConfig
}

// NewValidatorService 创建 ValidatorService。
//
// 中文说明：
// - 初始化 validator.Validate；
// - 注册翻译器（中文/英文）；
// - 注册自定义验证规则。
func NewValidatorService(cfg *contract.ValidatorConfig) (*ValidatorService, error) {
	v := validator.New()

	// 使用结构体字段名作为错误字段名
	// 不修改 tag 名称，保持默认的 "validate" tag 用于验证规则

	// 设置翻译器
	uni := ut.New(en.New(), zh.New())
	trans, ok := uni.GetTranslator(cfg.Locale)
	if !ok {
		trans = uni.GetFallback()
	}

	// 注册默认翻译
	switch cfg.Locale {
	case "zh":
		if err := zhTranslations.RegisterDefaultTranslations(v, trans); err != nil {
			// 翻译注册失败时继续，不影响验证功能
		}
	default:
		if err := enTranslations.RegisterDefaultTranslations(v, trans); err != nil {
			// 翻译注册失败时继续，不影响验证功能
		}
	}

	// 注册自定义规则
	for name, ruleCfg := range cfg.CustomRules {
		if ruleCfg.Fn != nil {
			_ = v.RegisterValidationCtx(name, func(ctx context.Context, fl validator.FieldLevel) bool {
				return ruleCfg.Fn(ctx, fl.Field().Interface())
			})
		}
	}

	return &ValidatorService{
		validate: v,
		trans:    trans,
		cfg:      cfg,
	}, nil
}

// Validate 验证结构体。
//
// 中文说明：
// - 使用 validator/v10 进行结构体验证；
// - 验证规则通过 binding tag 定义；
// - 验证失败返回 AppError。
func (s *ValidatorService) Validate(ctx context.Context, obj any) error {
	err := s.validate.StructCtx(ctx, obj)
	if err == nil {
		return nil
	}
	return s.TranslateError(err)
}

// ValidateVar 验证单个变量。
//
// 中文说明：
// - 用于验证单个字段或变量；
// - tag 参数指定验证规则（如 "required,email"）。
func (s *ValidatorService) ValidateVar(ctx context.Context, field any, tag string) error {
	err := s.validate.VarCtx(ctx, field, tag)
	if err == nil {
		return nil
	}
	return s.TranslateError(err)
}

// RegisterCustom 注册自定义验证规则。
//
// 中文说明：
// - name 为规则名称（如 "mobile"、"id_card"）；
// - 注册后可在 binding tag 中使用。
func (s *ValidatorService) RegisterCustom(name string, fn contract.CustomValidateFunc) error {
	return s.validate.RegisterValidationCtx(name, func(ctx context.Context, fl validator.FieldLevel) bool {
		return fn(ctx, fl.Field().Interface())
	})
}

// SetLocale 设置错误消息语言。
//
// 中文说明：
// - 支持 "zh"（中文）和 "en"（英文）；
// - 动态切换翻译器。
func (s *ValidatorService) SetLocale(locale string) error {
	uni := ut.New(en.New(), zh.New())
	trans, ok := uni.GetTranslator(locale)
	if !ok {
		return fmt.Errorf("validate: locale %s not supported", locale)
	}

	// 重新注册翻译
	switch locale {
	case "zh":
		if err := zhTranslations.RegisterDefaultTranslations(s.validate, trans); err != nil {
			return err
		}
	default:
		if err := enTranslations.RegisterDefaultTranslations(s.validate, trans); err != nil {
			return err
		}
	}

	s.trans = trans
	s.cfg.Locale = locale
	return nil
}

// TranslateError 翻译验证错误为 AppError。
//
// 中文说明：
// - 将 validator.ValidationErrors 转换为 AppError；
// - 包含字段级错误详情在 Metadata 中；
// - 使用配置的语言翻译错误消息。
func (s *ValidatorService) TranslateError(err error) contract.AppError {
	if err == nil {
		return nil
	}

	// 处理 validator.ValidationErrors
	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		// 非验证错误，直接包装为 BadRequest
		return contract.BadRequest(contract.ErrorReasonBadRequest, err.Error())
	}

	// 构建错误详情
	details := make([]contract.ValidationError, 0, len(validationErrors))
	for _, fe := range validationErrors {
		details = append(details, contract.ValidationError{
			Field:   fe.Field(), // 使用 JSON tag 名
			Tag:     fe.Tag(),
			Message: fe.Translate(s.trans),
			Value:   fe.Value(),
		})
	}

	// 构建错误消息（取第一个或全部）
	msgs := make([]string, len(details))
	for i, d := range details {
		msgs[i] = d.Message
	}

	// 将详情序列化为 JSON
	detailsJSON, _ := json.Marshal(details)

	return contract.BadRequest(contract.ErrorReasonBadRequest, strings.Join(msgs, "; ")).
		WithMetadata(map[string]string{
			"validation_errors": string(detailsJSON),
			"error_count":       fmt.Sprintf("%d", len(details)),
		})
}

// GetValidator 获取底层的 validator.Validate 实例。
//
// 中文说明：
// - 用于高级用法（如直接访问 validator 方法）。
func (s *ValidatorService) GetValidator() *validator.Validate {
	return s.validate
}

// GetTranslator 获取当前的翻译器。
//
// 中文说明：
// - 用于自定义错误消息翻译。
func (s *ValidatorService) GetTranslator() ut.Translator {
	return s.trans
}