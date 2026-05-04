package noop

import (
	"context"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string { return "validate.noop" }

func (p *Provider) IsDefer() bool { return true }

func (p *Provider) Provides() []string { return []string{datacontract.ValidatorKey} }

func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(datacontract.ValidatorKey, func(c runtimecontract.Container) (any, error) {
		return &noopValidator{}, nil
	}, true)
	return nil
}

func (p *Provider) Boot(runtimecontract.Container) error { return nil }

type noopValidator struct{}

func (v *noopValidator) Validate(ctx context.Context, obj any) error {
	return nil
}

func (v *noopValidator) ValidateVar(ctx context.Context, field any, tag string) error {
	return nil
}

func (v *noopValidator) RegisterCustom(name string, fn datacontract.CustomValidateFunc) error {
	return nil
}

func (v *noopValidator) SetLocale(locale string) error {
	return nil
}

func (v *noopValidator) TranslateError(err error) resiliencecontract.AppError {
	if err == nil {
		return nil
	}
	return resiliencecontract.BadRequest(resiliencecontract.ErrorReasonBadRequest, err.Error())
}
