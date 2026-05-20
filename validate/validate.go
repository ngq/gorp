// Application scenarios:
// - Expose the unified validation capability through a top-level convenience package.
// - Re-export validator contracts so business code can depend on short package paths.
// - Provide container-based helper functions for object and field validation.
//
// 适用场景：
// - 通过顶层便捷包暴露统一校验能力。
// - 重新导出校验器契约，让业务代码可以依赖更短的包路径。
// - 提供基于容器的对象校验和字段校验 helper。
package validate

import (
	"context"

	"github.com/ngq/gorp/framework/container"
	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

// Validator is the top-level alias of the unified validator contract.
//
// Validator 是统一校验器契约的顶层别名。
type Validator = datacontract.Validator

// ValidatorConfig is the top-level alias of the validator config contract.
//
// ValidatorConfig 是校验器配置契约的顶层别名。
type ValidatorConfig = datacontract.ValidatorConfig

// ValidationError is the top-level alias of the validation error contract.
//
// ValidationError 是校验错误契约的顶层别名。
type ValidationError = datacontract.ValidationError

// ValidationErrors is the top-level alias of the validation errors contract.
//
// ValidationErrors 是校验错误集合契约的顶层别名。
//
//nolint:errname // 类型别名，原始定义在 datacontract 包
type ValidationErrors = datacontract.ValidationErrors

// CustomValidateFunc is the top-level alias of the custom validate callback contract.
//
// CustomValidateFunc 是自定义校验回调契约的顶层别名。
type CustomValidateFunc = datacontract.CustomValidateFunc

// CustomRuleConfig is the top-level alias of the custom rule config contract.
//
// CustomRuleConfig 是自定义规则配置契约的顶层别名。
type CustomRuleConfig = datacontract.CustomRuleConfig

// Get returns the unified validator from the container.
//
// Get 从容器获取统一校验器。
func Get(c runtimecontract.Container) (datacontract.Validator, error) {
	return container.MakeValidator(c)
}

// GetOrPanic returns the unified validator from the container and panics on failure.
//
// GetOrPanic 从容器获取统一校验器，失败时 panic。
func GetOrPanic(c runtimecontract.Container) datacontract.Validator {
	return container.MustMakeValidator(c)
}

// Validate validates one object using the validator resolved from the container.
//
// Validate 使用容器中的校验器校验对象。
//
// Example:
//
//	err := validate.Validate(ctx, c, &CreateUserRequest{Name: "alice"})
func Validate(ctx context.Context, c runtimecontract.Container, obj any) error {
	validatorSvc, err := Get(c)
	if err != nil {
		return err
	}
	return validatorSvc.Validate(ctx, obj)
}

// ValidateVar validates one field value using the validator resolved from the container.
//
// ValidateVar 使用容器中的校验器校验单个字段值。
func ValidateVar(ctx context.Context, c runtimecontract.Container, field any, tag string) error {
	validatorSvc, err := Get(c)
	if err != nil {
		return err
	}
	return validatorSvc.ValidateVar(ctx, field, tag)
}
