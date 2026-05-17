// Application scenarios:
// - Verify the top-level validate package exports and helper behavior.
// - Protect validator aliasing and container-based validation helpers from regressions.
// - Document expected usage through focused export tests.
//
// 适用场景：
// - 验证顶层 validate 包的导出能力和 helper 行为。
// - 防止校验器别名和基于容器的校验 helper 回归。
// - 通过聚焦型导出测试固化预期用法。
package validate

import (
	"context"
	"io"
	"testing"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	"github.com/stretchr/testify/require"
)

type exportValidatorStub struct{}

func (s *exportValidatorStub) Validate(context.Context, any) error            { return nil }
func (s *exportValidatorStub) ValidateVar(context.Context, any, string) error { return nil }
func (s *exportValidatorStub) RegisterCustom(string, datacontract.CustomValidateFunc) error {
	return nil
}
func (s *exportValidatorStub) SetLocale(string) error                    { return nil }
func (s *exportValidatorStub) TranslateError(error) error               { return nil }

type exportValidateContainerStub struct {
	validator datacontract.Validator
}

func (s *exportValidateContainerStub) Bind(string, runtimecontract.Factory, bool)                      {}
func (s *exportValidateContainerStub) NamedBind(string, string, runtimecontract.Factory, bool)          {}
func (s *exportValidateContainerStub) IsBind(string) bool                                               { return true }
func (s *exportValidateContainerStub) IsBindNamed(string, string) bool                                  { return false }
func (s *exportValidateContainerStub) MustMake(key string) any                                          { v, _ := s.Make(key); return v }
func (s *exportValidateContainerStub) MustMakeNamed(string, string) any                                 { return nil }
func (s *exportValidateContainerStub) RegisterCloser(string, io.Closer)                                 {}
func (s *exportValidateContainerStub) Destroy() error                                                   { return nil }
func (s *exportValidateContainerStub) RegisteredProviders() []runtimecontract.ProviderInfo              { return nil }
func (s *exportValidateContainerStub) DebugPrint() string                                               { return "" }
func (s *exportValidateContainerStub) ProviderDAG() runtimecontract.ProviderDAG                          { return runtimecontract.ProviderDAG{} }
func (s *exportValidateContainerStub) MakeNamed(string, string) (any, error)                            { return nil, nil }
func (s *exportValidateContainerStub) RegisterProvider(runtimecontract.ServiceProvider) error {
	return nil
}
func (s *exportValidateContainerStub) RegisterProviders(...runtimecontract.ServiceProvider) error {
	return nil
}
func (s *exportValidateContainerStub) Make(key string) (any, error) {
	if key == datacontract.ValidatorKey {
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
