// Package validate provides the validation service implementation.
// ValidatorService wraps validator.Validate with locale-aware error translation.
//
// 本文件提供验证服务实现，封装 go-playground/validator。
// ValidatorService 封装 validator.Validate，支持本地化错误翻译。
package validate

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	enTranslations "github.com/go-playground/validator/v10/translations/en"
	zhTranslations "github.com/go-playground/validator/v10/translations/zh"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
)

// ValidatorService implements datacontract.Validator with go-playground/validator.
//
// ValidatorService 使用 go-playground/validator 实现 datacontract.Validator 接口。
type ValidatorService struct {
	validate *validator.Validate         // validate is the underlying validator.
	                                   //
	                                    // validate 底层验证器。
	trans    ut.Translator                // trans is the error translator.
	                                   //
	                                    // trans 错误翻译器。
	cfg      *datacontract.ValidatorConfig // cfg is the validator configuration.
	                                   //
	                                    // cfg 验证配置。
}

// NewValidatorService creates a new validator service with given config.
// Core logic: Initialize validator, setup locale translator, register custom rules.
//
// NewValidatorService 根据配置创建新的验证服务。
// 核心逻辑：初始化验证器、设置本地化翻译器、注册自定义规则。
func NewValidatorService(cfg *datacontract.ValidatorConfig) (*ValidatorService, error) {
	v := validator.New()

	// 注册 JSON tag 名函数，使验证错误返回 JSON 字段名而非 Go 结构体字段名。
	// 例如 UserReq 的 Username 字段带 `json:"username"` tag，错误消息中会显示 username 而非 Username。
	// 如果字段没有 json tag 或 json tag 为 "-"，则回退到 Go 字段名。
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "" || name == "-" {
			return fld.Name
		}
		return name
	})

	// 根据目标 locale 创建对应的 translator。
	// 使用 GetFallback() 获取第一个参数（即目标 locale）的 translator，确保可靠性。
	// 已支持 locale：zh（中文）、en（英文），其他 locale 将回退到英文。
	var trans ut.Translator
	switch cfg.Locale {
	case "zh":
		uni := ut.New(zh.New(), en.New())
		trans = uni.GetFallback()
		_ = zhTranslations.RegisterDefaultTranslations(v, trans)
	default:
		uni := ut.New(en.New(), zh.New())
		trans = uni.GetFallback()
		_ = enTranslations.RegisterDefaultTranslations(v, trans)
	}

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

// Validate validates a struct and returns translated errors.
// Eg:
//
// Validate 校验结构体并返回翻译后的错误。
// Eg:
//
//	type UserReq struct {
//	    Name  string `validate:"required"`
//	    Email string `validate:"required,email"`
//	}
//	err := vSvc.Validate(ctx, &UserReq{Name: "", Email: "invalid"})
func (s *ValidatorService) Validate(ctx context.Context, obj any) error {
	err := s.validate.StructCtx(ctx, obj)
	if err == nil {
		return nil
	}
	return s.TranslateError(err)
}

// ValidateVar validates a single variable against a tag.
// Eg:
//
// ValidateVar 校验单个变量是否符合指定标签规则。
// Eg:
//
//	err := vSvc.ValidateVar(ctx, "test@example.com", "required,email")
func (s *ValidatorService) ValidateVar(ctx context.Context, field any, tag string) error {
	err := s.validate.VarCtx(ctx, field, tag)
	if err == nil {
		return nil
	}
	return s.TranslateError(err)
}

// RegisterCustom registers a custom validation rule.
// Eg:
//
// RegisterCustom 注册自定义校验规则。
// Eg:
//
//	err := vSvc.RegisterCustom("mobile", func(ctx context.Context, field interface{}) bool {
//	    return regexp.MustCompile(`^1[3-9]\d{9}$`).MatchString(fmt.Sprint(field))
//	})
func (s *ValidatorService) RegisterCustom(name string, fn datacontract.CustomValidateFunc) error {
	return s.validate.RegisterValidationCtx(name, func(ctx context.Context, fl validator.FieldLevel) bool {
		return fn(ctx, fl.Field().Interface())
	})
}

// SetLocale changes the error translation locale (zh/en).
//
// SetLocale 更改错误翻译的本地化语言（zh/en）。
func (s *ValidatorService) SetLocale(locale string) error {
	var trans ut.Translator
	switch locale {
	case "zh":
		uni := ut.New(zh.New(), en.New())
		trans = uni.GetFallback()
		if err := zhTranslations.RegisterDefaultTranslations(s.validate, trans); err != nil {
			return err
		}
	case "en":
		uni := ut.New(en.New(), zh.New())
		trans = uni.GetFallback()
		if err := enTranslations.RegisterDefaultTranslations(s.validate, trans); err != nil {
			return err
		}
	default:
		return fmt.Errorf("validate: locale %s not supported", locale)
	}

	s.trans = trans
	s.cfg.Locale = locale
	return nil
}

// TranslateError translates validation errors into localized AppError.
// Core logic: Check if error is ValidationErrors, then translate each field error.
// If TranslateErrors is false, returns raw English errors for better performance.
// Returns error interface; caller can cast to resiliencecontract.AppError if needed.
//
// TranslateError 将验证错误翻译为本地化的 AppError。
// 核心逻辑：检查是否为 ValidationErrors 类型，然后翻译每个字段错误。
// 如果 TranslateErrors 为 false，返回原始英文错误以获得更好性能。
// 返回 error 接口；调用方可在需要时断言为 resiliencecontract.AppError。
func (s *ValidatorService) TranslateError(err error) error {
	if err == nil {
		return nil
	}

	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		return resiliencecontract.BadRequest(resiliencecontract.ErrorReasonBadRequest, err.Error())
	}

	// If TranslateErrors is false, return raw English errors without translation.
	// This saves ~1.6 µs per validation failure (translation + JSON overhead).
	//
	// 如果 TranslateErrors 为 false，返回原始英文错误不做翻译。
	// 这样每次验证失败可节省约 1.6 µs（翻译 + JSON 开销）。
	if !s.cfg.TranslateErrors {
		msgs := make([]string, len(validationErrors))
		for i, fe := range validationErrors {
			msgs[i] = fe.Error() // Raw English error: "Key: 'User.Email' Error:Field validation for 'Email' failed on the 'email' tag"
		}
		return resiliencecontract.BadRequest(
			resiliencecontract.ErrorReasonBadRequest,
			strings.Join(msgs, "; "),
		)
	}

	// Translate errors to configured locale (zh/en).
	// 翻译错误到配置的语言（中文/英文）。
	details := make([]datacontract.ValidationError, 0, len(validationErrors))
	for _, fe := range validationErrors {
		details = append(details, datacontract.ValidationError{
			Field:   fe.Field(),
			Tag:     fe.Tag(),
			Message: fe.Translate(s.trans),
			Value:   fe.Value(),
		})
	}

	msgs := make([]string, len(details))
	for i, d := range details {
		msgs[i] = d.Message
	}

	detailsJSON, _ := json.Marshal(details)

	return resiliencecontract.BadRequest(
		resiliencecontract.ErrorReasonBadRequest,
		strings.Join(msgs, "; "),
	).WithMetadata(map[string]string{
		"validation_errors": string(detailsJSON),
		"error_count":       fmt.Sprintf("%d", len(details)),
	})
}

// GetValidator returns the underlying validator.Validate instance.
//
// GetValidator 返回底层的 validator.Validate 实例，用于高级自定义场景。
func (s *ValidatorService) GetValidator() *validator.Validate {
	return s.validate
}

// GetTranslator returns the error translator for custom error formatting.
//
// GetTranslator 返回错误翻译器，可用于自定义错误格式化。
func (s *ValidatorService) GetTranslator() ut.Translator {
	return s.trans
}