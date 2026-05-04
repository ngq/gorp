package data

import (
	"context"

	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
)

const ValidatorKey = "framework.validator"

type Validator interface {
	Validate(ctx context.Context, obj any) error
	ValidateVar(ctx context.Context, field any, tag string) error
	RegisterCustom(name string, fn CustomValidateFunc) error
	SetLocale(locale string) error
	TranslateError(err error) resiliencecontract.AppError
}

type CustomValidateFunc func(ctx context.Context, field any) bool

type ValidationError struct {
	Field   string
	Tag     string
	Message string
	Value   any
}

type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return ""
	}

	msgs := make([]string, len(e))
	for i, v := range e {
		msgs[i] = v.Message
	}

	return msgs[0]
}

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

type ValidatorConfig struct {
	Enabled bool
	Locale  string

	CustomRules map[string]CustomRuleConfig
}

type CustomRuleConfig struct {
	Name    string
	Message string
	Fn      CustomValidateFunc
}
