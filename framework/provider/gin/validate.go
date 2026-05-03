package gin

import (
	"reflect"

	"github.com/gin-gonic/gin"

	"github.com/ngq/gorp/framework/contract"
)

const (
	validatedBodyGinKey = "validated_body"
	validatorGinKey     = "validator"
)

func storeValidatedBody(c *gin.Context, obj interface{}) {
	if c == nil {
		return
	}
	c.Set(validatedBodyGinKey, obj)
	if c.Request != nil {
		c.Request = c.Request.WithContext(contract.NewValidatedBodyContext(c.Request.Context(), obj))
	}
}

func storeValidator(c *gin.Context, validator contract.Validator) {
	if c == nil {
		return
	}
	c.Set(validatorGinKey, validator)
}

func clonePointerValue(objType interface{}) (interface{}, bool) {
	if objType == nil {
		return nil, false
	}
	t := reflect.TypeOf(objType)
	if t.Kind() != reflect.Ptr {
		return nil, false
	}
	return reflect.New(t.Elem()).Interface(), true
}

// ValidateBodyMiddleware 创建请求体验证中间件。
//
// 中文说明：
// - 这是 provider 侧的验证入口，负责承接当前 HTTP provider 下的请求绑定与验证；
// - 默认业务主线如果只需要在 handler 内部完成绑定与验证，优先直接使用 `HTTPContext.BindJSON(...) + Validator.Validate(...)`；
// - 当项目确实希望把“绑定 + 验证 + 统一错误响应”前移到中间件层时，再显式接入这里的 middleware；
// - 验证成功后的对象应继续通过 request context / helper 暴露给后续链路，而不是要求业务直接依赖 Gin-only 存储。
func ValidateBodyMiddleware(validator contract.Validator, objType interface{}) contract.HTTPMiddleware {
	return func(c contract.HTTPContext, next contract.HTTPNext) {
		if validator == nil || objType == nil {
			if next != nil {
				next()
			}
			return
		}

		binder, ok := clonePointerValue(objType)
		if !ok {
			appErr := contract.BadRequest(contract.ErrorReasonBadRequest, "validate middleware requires pointer objType")
			respondWithError(c, appErr)
			return
		}
		if err := c.BindJSON(binder); err != nil {
			appErr := contract.BadRequest(contract.ErrorReasonBadRequest, "invalid request body: "+err.Error())
			respondWithError(c, appErr)
			return
		}
		if err := validator.Validate(c.Context(), binder); err != nil {
			appErr, ok := err.(contract.AppError)
			if !ok {
				appErr = contract.BadRequest(contract.ErrorReasonBadRequest, err.Error())
			}
			respondWithError(c, appErr)
			return
		}
		if gc, ok := unwrapGinContext(c); ok {
			storeValidatedBody(gc, binder)
		}
		if next != nil {
			next()
		}
	}
}

// ValidateMiddleware 创建通用验证中间件。
//
// 中文说明：
// - 它只负责把 Validator 暴露给后续 Gin helper / 兼容路径；
// - 默认 framework 主线不要求业务必须依赖它；
// - 如果项目仍保留 provider-specific 验证辅助写法，可显式接入该 middleware。
func ValidateMiddleware(validator contract.Validator) contract.HTTPMiddleware {
	return func(c contract.HTTPContext, next contract.HTTPNext) {
		if gc, ok := unwrapGinContext(c); ok {
			storeValidator(gc, validator)
		}
		if next != nil {
			next()
		}
	}
}

// ValidateBody 辅助函数：验证请求体。
//
// 中文说明：
// - 在 handler 中使用；
// - 自动绑定和验证；
// - 验证失败自动响应错误。
//
// 返回：
// - nil: 验证成功
// - 非 nil: 验证失败（错误已响应，handler 应 return）
//
// 使用示例：
//
//	func loginHandler(c *gin.Context) {
//	    var req LoginRequest
//	    if err := gin.ValidateBody(c, validator, &req); err != nil {
//	        return // 错误已处理
//	    }
//	    // 使用 req ...
//	}
func ValidateBody(c *gin.Context, validator contract.Validator, obj interface{}) error {
	// 绑定 JSON
	if err := c.ShouldBindJSON(obj); err != nil {
		appErr := contract.BadRequest(contract.ErrorReasonBadRequest, "invalid request body: "+err.Error())
		respondWithError(newHTTPContext(c), appErr)
		return appErr
	}

	// 执行验证
	if err := validator.Validate(c.Request.Context(), obj); err != nil {
		appErr, ok := err.(contract.AppError)
		if !ok {
			appErr = contract.BadRequest(contract.ErrorReasonBadRequest, err.Error())
		}
		respondWithError(newHTTPContext(c), appErr)
		return appErr
	}

	storeValidatedBody(c, obj)
	return nil
}

// ValidateQuery 辅助函数：验证 Query 参数。
//
// 中文说明：
// - 验证 URL 查询参数；
// - 使用 form tag 绑定。
func ValidateQuery(c *gin.Context, validator contract.Validator, obj interface{}) error {
	// 绑定 Query
	if err := c.ShouldBindQuery(obj); err != nil {
		appErr := contract.BadRequest(contract.ErrorReasonBadRequest, "invalid query parameters: "+err.Error())
		respondWithError(newHTTPContext(c), appErr)
		return appErr
	}

	// 执行验证
	if err := validator.Validate(c.Request.Context(), obj); err != nil {
		appErr, ok := err.(contract.AppError)
		if !ok {
			appErr = contract.BadRequest(contract.ErrorReasonBadRequest, err.Error())
		}
		respondWithError(newHTTPContext(c), appErr)
		return appErr
	}

	return nil
}

// ValidateForm 辅助函数：验证 Form 数据。
//
// 中文说明：
// - 验证表单数据；
// - 使用 form tag 绑定。
func ValidateForm(c *gin.Context, validator contract.Validator, obj interface{}) error {
	// 绑定 Form
	if err := c.ShouldBind(obj); err != nil {
		appErr := contract.BadRequest(contract.ErrorReasonBadRequest, "invalid form data: "+err.Error())
		respondWithError(newHTTPContext(c), appErr)
		return appErr
	}

	// 执行验证
	if err := validator.Validate(c.Request.Context(), obj); err != nil {
		appErr, ok := err.(contract.AppError)
		if !ok {
			appErr = contract.BadRequest(contract.ErrorReasonBadRequest, err.Error())
		}
		respondWithError(newHTTPContext(c), appErr)
		return appErr
	}

	return nil
}

// respondWithError 统一错误响应。
//
// 中文说明：
// - 使用统一的响应格式；
// - 包含 code、reason、message 字段。
func respondWithError(c contract.HTTPContext, err contract.AppError) {
	status := err.GetStatus()

	response := ValidateErrorResponse{
		Code:    int(status.Code),
		Reason:  string(status.Reason),
		Message: status.Message,
	}

	if status.Metadata != nil {
		if errorsJSON, ok := status.Metadata["validation_errors"]; ok {
			response.Details = errorsJSON
		}
	}

	c.JSON(response.Code, response)
}

// ValidateErrorResponse 验证错误响应格式。
//
// 中文说明：
// - 统一的验证错误响应结构；
// - 包含字段级错误详情。
type ValidateErrorResponse struct {
	// Code HTTP 状态码
	Code int `json:"code"`

	// Reason 错误原因
	Reason string `json:"reason"`

	// Message 错误消息
	Message string `json:"message"`

	// Details 验证错误详情（JSON）
	Details string `json:"details,omitempty"`
}

// GetValidator 从 context 获取 Validator。
//
// 中文说明：
// - 配合 ValidateMiddleware 使用；
// - 在 handler 中获取验证器实例。
func GetValidator(c *gin.Context) contract.Validator {
	if v, exists := c.Get(validatorGinKey); exists {
		if validator, ok := v.(contract.Validator); ok {
			return validator
		}
	}
	return nil
}

// GetValidatedBody 从 context 获取已验证的请求体。
//
// 中文说明：
// - 配合 ValidateBodyMiddleware 使用；
// - 在 handler 中获取已验证的对象。
func GetValidatedBody(c *gin.Context) interface{} {
	if v, exists := c.Get(validatedBodyGinKey); exists {
		return v
	}
	if c != nil && c.Request != nil {
		if v, ok := contract.FromValidatedBodyContext(c.Request.Context()); ok {
			return v
		}
	}
	return nil
}
