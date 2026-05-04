package validate

import (
	"context"

	"github.com/ngq/gorp/framework/container"
	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

type Validator = datacontract.Validator
type ValidatorConfig = datacontract.ValidatorConfig
type ValidationError = datacontract.ValidationError
type ValidationErrors = datacontract.ValidationErrors
type CustomValidateFunc = datacontract.CustomValidateFunc

type CustomRuleConfig = datacontract.CustomRuleConfig

// Make 从容器获取统一验证器。
func Make(c runtimecontract.Container) (datacontract.Validator, error) {
	return container.MakeValidator(c)
}

// MustMake 从容器获取统一验证器，失败 panic。
func MustMake(c runtimecontract.Container) datacontract.Validator {
	return container.MustMakeValidator(c)
}

// Validate 使用容器中的验证器校验结构体。
func Validate(ctx context.Context, c runtimecontract.Container, obj any) error {
	validatorSvc, err := Make(c)
	if err != nil {
		return err
	}
	return validatorSvc.Validate(ctx, obj)
}

// ValidateVar 使用容器中的验证器校验单个字段。
func ValidateVar(ctx context.Context, c runtimecontract.Container, field any, tag string) error {
	validatorSvc, err := Make(c)
	if err != nil {
		return err
	}
	return validatorSvc.ValidateVar(ctx, field, tag)
}
