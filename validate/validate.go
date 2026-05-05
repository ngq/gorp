package validate

import (
	"context"

	"github.com/ngq/gorp/framework/container"
	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

// Validator is the top-level alias of the unified validator contract.
// Validator 是统一验证器契约的顶层别名。
type Validator = datacontract.Validator

// ValidatorConfig is the top-level alias of the validator config contract.
// ValidatorConfig 是验证器配置契约的顶层别名。
type ValidatorConfig = datacontract.ValidatorConfig

// ValidationError is the top-level alias of the validation error contract.
// ValidationError 是校验错误契约的顶层别名。
type ValidationError = datacontract.ValidationError

// ValidationErrors is the top-level alias of the validation errors contract.
// ValidationErrors 是校验错误集合契约的顶层别名。
type ValidationErrors = datacontract.ValidationErrors

// CustomValidateFunc is the top-level alias of the custom validate callback contract.
// CustomValidateFunc 是自定义校验回调契约的顶层别名。
type CustomValidateFunc = datacontract.CustomValidateFunc

// CustomRuleConfig is the top-level alias of the custom rule config contract.
// CustomRuleConfig 是自定义规则配置契约的顶层别名。
type CustomRuleConfig = datacontract.CustomRuleConfig

// Make returns the unified validator from the container.
// Make 从容器获取统一验证器。
func Make(c runtimecontract.Container) (datacontract.Validator, error) {
	return container.MakeValidator(c)
}

// MustMake returns the unified validator from the container and panics on failure.
// MustMake 从容器获取统一验证器，失败 panic。
func MustMake(c runtimecontract.Container) datacontract.Validator {
	return container.MustMakeValidator(c)
}

// Validate validates a struct using the validator from the container.
// Validate 使用容器中的验证器校验结构体。
//
// Example:
//
//	err := validate.Validate(ctx, c, &CreateUserRequest{Name: "alice"})
func Validate(ctx context.Context, c runtimecontract.Container, obj any) error {
	validatorSvc, err := Make(c)
	if err != nil {
		return err
	}
	return validatorSvc.Validate(ctx, obj)
}

// ValidateVar validates a single field using the validator from the container.
// ValidateVar 使用容器中的验证器校验单个字段。
func ValidateVar(ctx context.Context, c runtimecontract.Container, field any, tag string) error {
	validatorSvc, err := Make(c)
	if err != nil {
		return err
	}
	return validatorSvc.ValidateVar(ctx, field, tag)
}
