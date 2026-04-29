package validate

import (
	"context"
	"testing"

	"github.com/ngq/gorp/framework/contract"
	"github.com/stretchr/testify/require"
)

type exportValidatorStub struct{}

func (s *exportValidatorStub) Validate(context.Context, any) error                    { return nil }
func (s *exportValidatorStub) ValidateVar(context.Context, any, string) error         { return nil }
func (s *exportValidatorStub) RegisterCustom(string, contract.CustomValidateFunc) error { return nil }
func (s *exportValidatorStub) SetLocale(string) error                                 { return nil }
func (s *exportValidatorStub) TranslateError(error) contract.AppError                 { return nil }

type exportValidateContainerStub struct {
	validator contract.Validator
}

func (s *exportValidateContainerStub) Bind(string, contract.Factory, bool)                {}
func (s *exportValidateContainerStub) IsBind(string) bool                                 { return true }
func (s *exportValidateContainerStub) MustMake(key string) any                            { v, _ := s.Make(key); return v }
func (s *exportValidateContainerStub) RegisterProvider(contract.ServiceProvider) error     { return nil }
func (s *exportValidateContainerStub) RegisterProviders(...contract.ServiceProvider) error { return nil }
func (s *exportValidateContainerStub) Make(key string) (any, error) {
	if key == contract.ValidatorKey {
		return s.validator, nil
	}
	return nil, context.DeadlineExceeded
}

func TestExportedValidateHelpers(t *testing.T) {
	stub := &exportValidatorStub{}
	containerStub := &exportValidateContainerStub{validator: stub}

	validatorSvc, err := Make(containerStub)
	require.NoError(t, err)
	require.Same(t, stub, validatorSvc)
	require.Same(t, stub, MustMake(containerStub))

	err = Validate(context.Background(), containerStub, struct{ Name string }{Name: "alice"})
	require.NoError(t, err)
	err = ValidateVar(context.Background(), containerStub, "alice@example.com", "required,email")
	require.NoError(t, err)

	var _ Validator = stub
	var _ = ValidatorConfig{Enabled: true, Locale: "zh"}
	var _ = ValidationError{Field: "name"}
	var _ ValidationErrors
	var _ CustomValidateFunc
	var _ = CustomRuleConfig{Name: "mobile"}
}
