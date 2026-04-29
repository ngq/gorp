package validate

import (
	"context"

	"github.com/ngq/gorp/framework/container"
	"github.com/ngq/gorp/framework/contract"
)

type Validator = contract.Validator
type ValidatorConfig = contract.ValidatorConfig
type ValidationError = contract.ValidationError
type ValidationErrors = contract.ValidationErrors
type CustomValidateFunc = contract.CustomValidateFunc

type CustomRuleConfig = contract.CustomRuleConfig

// Make 从容器获取统一验证器。
func Make(c contract.Container) (contract.Validator, error) {
	return container.MakeValidator(c)
}

// MustMake 从容器获取统一验证器，失败 panic。
func MustMake(c contract.Container) contract.Validator {
	return container.MustMakeValidator(c)
}

// Validate 使用容器中的验证器校验结构体。
func Validate(ctx context.Context, c contract.Container, obj any) error {
	validatorSvc, err := Make(c)
	if err != nil {
		return err
	}
	return validatorSvc.Validate(ctx, obj)
}

// ValidateVar 使用容器中的验证器校验单个字段。
func ValidateVar(ctx context.Context, c contract.Container, field any, tag string) error {
	validatorSvc, err := Make(c)
	if err != nil {
		return err
	}
	return validatorSvc.ValidateVar(ctx, field, tag)
}
