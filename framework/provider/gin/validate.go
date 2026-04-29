package gin

import (
	"github.com/gin-gonic/gin"

	"github.com/ngq/gorp/framework/contract"
)

// ValidateBodyMiddleware 创建请求体验证中间件。
//
// 中文说明：
// - 验证 JSON 请求体；
// - 使用 Validator 接口进行验证；
// - 验证失败返回统一 AppError 格式；
// - 验证成功将解析后的对象存入 context。
//
// 注意：
// - 当前默认业务主线优先使用 `ValidateBody` / `ValidateQuery` / `ValidateForm` 这类 handler 内 helper；
// - `ValidateBodyMiddleware` 仅适合明确需要中间件式校验的场景。
func ValidateBodyMiddleware(validator contract.Validator, objType interface{}) gin.HandlerFunc {
	_ = objType
	// 当前中间件版本不根据 objType 自动创建实例；
	// 如需显式控制请求结构体，优先在 handler 中使用 ValidateBody。

	return func(c *gin.Context) {
		// 创建目标对象
		// objType 应该是一个结构体指针类型
		var obj interface{}

		// 尝试绑定 JSON
		if err := c.ShouldBindJSON(obj); err != nil {
			// 绑定失败（JSON 格式错误或 Content-Type 不支持）
			appErr := contract.BadRequest(contract.ErrorReasonBadRequest, "invalid request body: "+err.Error())
			respondWithError(c, appErr)
			c.Abort()
			return
		}

		// 执行验证
		if err := validator.Validate(c.Request.Context(), obj); err != nil {
			// 验证失败
			appErr, ok := err.(contract.AppError)
			if !ok {
				appErr = contract.BadRequest(contract.ErrorReasonBadRequest, err.Error())
			}
			respondWithError(c, appErr)
			c.Abort()
			return
		}

		// 验证成功，存入 context
		c.Set("validated_body", obj)
		c.Next()
	}
}

// ValidateMiddleware 创建通用验证中间件。
//
// 中文说明：
// - 提供验证器实例供 handler 使用；
// - 不自动验证，需要 handler 调用 ValidateBody 辅助函数。
//
// 使用示例：
//
//	router.Use(gin.ValidateMiddleware(validator))
//	router.POST("/login", func(c *gin.Context) {
//	    var req LoginRequest
//	    if err := gin.ValidateBody(c, validator, &req); err != nil {
//	        return // 错误已处理
//	    }
//	    // 使用 req ...
//	})
func ValidateMiddleware(validator contract.Validator) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("validator", validator)
		c.Next()
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
		respondWithError(c, appErr)
		return appErr
	}

	// 执行验证
	if err := validator.Validate(c.Request.Context(), obj); err != nil {
		appErr, ok := err.(contract.AppError)
		if !ok {
			appErr = contract.BadRequest(contract.ErrorReasonBadRequest, err.Error())
		}
		respondWithError(c, appErr)
		return appErr
	}

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
		respondWithError(c, appErr)
		return appErr
	}

	// 执行验证
	if err := validator.Validate(c.Request.Context(), obj); err != nil {
		appErr, ok := err.(contract.AppError)
		if !ok {
			appErr = contract.BadRequest(contract.ErrorReasonBadRequest, err.Error())
		}
		respondWithError(c, appErr)
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
		respondWithError(c, appErr)
		return appErr
	}

	// 执行验证
	if err := validator.Validate(c.Request.Context(), obj); err != nil {
		appErr, ok := err.(contract.AppError)
		if !ok {
			appErr = contract.BadRequest(contract.ErrorReasonBadRequest, err.Error())
		}
		respondWithError(c, appErr)
		return appErr
	}

	return nil
}

// respondWithError 统一错误响应。
//
// 中文说明：
// - 使用统一的响应格式；
// - 包含 code、reason、message 字段。
func respondWithError(c *gin.Context, err contract.AppError) {
	status := err.GetStatus()

	response := ValidateErrorResponse{
		Code:    int(status.Code),
		Reason:  string(status.Reason),
		Message: status.Message,
	}

	// 添加验证错误详情
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
	if v, exists := c.Get("validator"); exists {
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
	if v, exists := c.Get("validated_body"); exists {
		return v
	}
	return nil
}