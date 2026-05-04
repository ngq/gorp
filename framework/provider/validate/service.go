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

	datacontract "github.com/ngq/gorp/framework/contract/data"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
)

type ValidatorService struct {
	validate *validator.Validate
	trans    ut.Translator
	cfg      *datacontract.ValidatorConfig
}

func NewValidatorService(cfg *datacontract.ValidatorConfig) (*ValidatorService, error) {
	v := validator.New()

	uni := ut.New(en.New(), zh.New())
	trans, ok := uni.GetTranslator(cfg.Locale)
	if !ok {
		trans = uni.GetFallback()
	}

	switch cfg.Locale {
	case "zh":
		_ = zhTranslations.RegisterDefaultTranslations(v, trans)
	default:
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

func (s *ValidatorService) Validate(ctx context.Context, obj any) error {
	err := s.validate.StructCtx(ctx, obj)
	if err == nil {
		return nil
	}
	return s.TranslateError(err)
}

func (s *ValidatorService) ValidateVar(ctx context.Context, field any, tag string) error {
	err := s.validate.VarCtx(ctx, field, tag)
	if err == nil {
		return nil
	}
	return s.TranslateError(err)
}

func (s *ValidatorService) RegisterCustom(name string, fn datacontract.CustomValidateFunc) error {
	return s.validate.RegisterValidationCtx(name, func(ctx context.Context, fl validator.FieldLevel) bool {
		return fn(ctx, fl.Field().Interface())
	})
}

func (s *ValidatorService) SetLocale(locale string) error {
	uni := ut.New(en.New(), zh.New())
	trans, ok := uni.GetTranslator(locale)
	if !ok {
		return fmt.Errorf("validate: locale %s not supported", locale)
	}

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

func (s *ValidatorService) TranslateError(err error) resiliencecontract.AppError {
	if err == nil {
		return nil
	}

	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return resiliencecontract.BadRequest(resiliencecontract.ErrorReasonBadRequest, err.Error())
	}

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

func (s *ValidatorService) GetValidator() *validator.Validate {
	return s.validate
}

func (s *ValidatorService) GetTranslator() ut.Translator {
	return s.trans
}
