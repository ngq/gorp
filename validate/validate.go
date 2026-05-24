// Application scenarios:
// - Expose the unified validation capability through a top-level convenience package.
// - Re-export validator contracts so business code can depend on short package paths.
// - Provide context-based helper functions for object and field validation.
//
// 适用场景：
// - 通过顶层便捷包暴露统一校验能力。
// - 重新导出校验器契约，让业务代码可以依赖更短的包路径。
// - 提供基于 context 的对象校验和字段校验 helper。
package validate

import (
	"context"

	"github.com/ngq/gorp/framework/container"
	frameworkcontainer "github.com/ngq/gorp/framework/container"
	datacontract "github.com/ngq/gorp/framework/contract/data"
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

// resolveValidator 从 context 解析容器并获取校验器实例。
// 内部 helper：先通过 frameworkcontainer.Resolve(ctx) 拿到容器，
// 再用 container.MakeWith 按契约键解析出 Validator。
func resolveValidator(ctx context.Context) (datacontract.Validator, error) {
	cont := frameworkcontainer.Resolve(ctx)
	if cont == nil {
		return nil, context.DeadlineExceeded
	}
	return container.MakeWith[datacontract.Validator](cont, datacontract.ValidatorKey)
}

// GetService 从容器获取统一校验器。
// 通过 context 解析容器后按 ValidatorKey 查找校验器实例，
// 找不到或容器不可用时返回 error。
//
// Example:
//
//	validatorSvc, err := validate.GetService(ctx)
func GetService(ctx context.Context) (datacontract.Validator, error) {
	return resolveValidator(ctx)
}

// MustGetService 从容器获取统一校验器，失败时 panic。
// 适用于启动阶段或明确期望校验器一定已注册的场景。
//
// Example:
//
//	validatorSvc := validate.MustGetService(ctx)
func MustGetService(ctx context.Context) datacontract.Validator {
	svc, err := resolveValidator(ctx)
	if err != nil {
		panic(err)
	}
	return svc
}

// Validate 使用容器中的校验器校验对象。
// 通过 context 解析容器获取校验器，然后调用其 Validate 方法完成结构体校验。
// 如果容器不可用或校验器未注册，返回 error。
//
// Example:
//
//	err := validate.Validate(ctx, &CreateUserRequest{Name: "alice"})
func Validate(ctx context.Context, obj any) error {
	validatorSvc, err := resolveValidator(ctx)
	if err != nil {
		return err
	}
	return validatorSvc.Validate(ctx, obj)
}

// ValidateVar 使用容器中的校验器校验单个字段值。
// 通过 context 解析容器获取校验器，然后调用其 ValidateVar 方法完成字段校验。
// tag 参数为校验规则标签，如 "required,email"。
//
// Example:
//
//	err := validate.ValidateVar(ctx, "alice@example.com", "required,email")
func ValidateVar(ctx context.Context, field any, tag string) error {
	validatorSvc, err := resolveValidator(ctx)
	if err != nil {
		return err
	}
	return validatorSvc.ValidateVar(ctx, field, tag)
}
