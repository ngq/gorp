// Application scenarios:
// - Bind and validate request bodies, query parameters, and form payloads with a unified flow.
// - Return stable validation error responses to HTTP callers.
// - Store validated request objects back into request context for downstream reuse.
//
// 适用场景：
// - 以统一流程绑定并校验请求体、查询参数和表单载荷。
// - 向 HTTP 调用方返回稳定的参数校验错误响应。
// - 将已校验的请求对象回写到请求上下文，供下游复用。
package middleware

import (
	"errors"
	"reflect"

	"github.com/gin-gonic/gin"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	supportcontract "github.com/ngq/gorp/framework/contract/support"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

type ValidateErrorResponse struct {
	Code    int    `json:"code"`
	Reason  string `json:"reason"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// storeValidatedBody writes the validated request object into request context.
//
// storeValidatedBody 将已校验的请求对象写回请求上下文。
func storeValidatedBody(c transportcontract.Context, obj any) {
	if c == nil || obj == nil {
		return
	}
	// Store validated body in context using Set
	c.Set("validated_body", obj)
	// Also update gin.Request.Context for context.Context value propagation
	if gc, ok := unwrapGinContext(c); ok && gc.Request != nil {
		gc.Request = gc.Request.WithContext(supportcontract.NewValidatedBodyContext(gc.Request.Context(), obj))
	}
}

// validateBoundValue runs validator logic and emits a unified error response on validation failure.
//
// validateBoundValue 执行校验逻辑，并在校验失败时输出统一错误响应。
func validateBoundValue(c transportcontract.Context, validator datacontract.Validator, obj any) error {
	if validator == nil {
		return nil
	}
	if err := validator.Validate(c.Context(), obj); err != nil {
		var appErr resiliencecontract.AppError
		if !errors.As(err, &appErr) {
			appErr = resiliencecontract.BadRequest(resiliencecontract.ErrorReasonBadRequest, err.Error())
		}
		respondWithError(c, appErr)
		return appErr
	}
	storeValidatedBody(c, obj)
	return nil
}

// BindAndValidateJSON binds a JSON payload and validates the bound object.
//
// BindAndValidateJSON 绑定 JSON 载荷并校验绑定后的对象。
func BindAndValidateJSON(c transportcontract.Context, validator datacontract.Validator, obj any) error {
	if err := c.BindJSON(obj); err != nil {
		appErr := resiliencecontract.BadRequest(resiliencecontract.ErrorReasonBadRequest, "invalid request body: "+err.Error())
		respondWithError(c, appErr)
		return appErr
	}
	return validateBoundValue(c, validator, obj)
}

// BindAndValidateQuery binds query parameters and validates the bound object.
//
// BindAndValidateQuery 绑定查询参数并校验绑定后的对象。
func BindAndValidateQuery(c transportcontract.Context, validator datacontract.Validator, obj any) error {
	if err := c.BindQuery(obj); err != nil {
		appErr := resiliencecontract.BadRequest(resiliencecontract.ErrorReasonBadRequest, "invalid query parameters: "+err.Error())
		respondWithError(c, appErr)
		return appErr
	}
	return validateBoundValue(c, validator, obj)
}

// BindAndValidate binds a generic request payload and validates the bound object.
//
// BindAndValidate 绑定通用请求载荷并校验绑定后的对象。
func BindAndValidate(c transportcontract.Context, validator datacontract.Validator, obj any) error {
	if err := c.Bind(obj); err != nil {
		appErr := resiliencecontract.BadRequest(resiliencecontract.ErrorReasonBadRequest, "invalid form data: "+err.Error())
		respondWithError(c, appErr)
		return appErr
	}
	return validateBoundValue(c, validator, obj)
}

// respondWithError writes a normalized validation error response.
//
// respondWithError 输出归一化后的校验错误响应。
func respondWithError(c transportcontract.Context, err resiliencecontract.AppError) {
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

// ValidateBodyMiddleware creates a Gin middleware that automatically binds and validates
// the JSON request body into the given prototype object.
// The prototype must be a pointer to a struct; each request gets a new zero-value copy.
// On bind failure or validation failure, the middleware writes a unified error response and aborts.
// On success, the validated object is stored in request context for downstream retrieval.
//
// ValidateBodyMiddleware 创建自动绑定并校验 JSON 请求体的 Gin 中间件。
// prototype 必须是结构体指针；每次请求会创建一个新的零值拷贝。
// 绑定或校验失败时，中间件输出统一错误响应并中断请求链。
// 校验成功后，已校验对象存入请求上下文供下游复用。
//
// Example:
//
//	type CreateUserReq struct {
//	    Name  string `json:"name" validate:"required,min=3"`
//	    Email string `json:"email" validate:"required,email"`
//	}
//	router.POST("/users", httpmiddleware.ValidateBodyMiddleware(validator, &CreateUserReq{}), createUserHandler)
func ValidateBodyMiddleware(validator datacontract.Validator, prototype any) func(*gin.Context) {
	// 在中间件注册时验证 prototype 类型，尽早暴露配置错误
	protoType := reflect.TypeOf(prototype)
	if protoType == nil || protoType.Kind() != reflect.Ptr || protoType.Elem().Kind() != reflect.Struct {
		panic("ValidateBodyMiddleware: prototype must be a pointer to a struct")
	}
	elemType := protoType.Elem()

	return func(c *gin.Context) {
		// 每次请求创建新的零值结构体，避免复用导致数据泄漏
		obj := reflect.New(elemType).Interface()

		httpCtx := newContext(c)
		if err := BindAndValidateJSON(httpCtx, validator, obj); err != nil {
			c.Abort()
			return
		}
		c.Next()
	}
}

// ValidateQueryMiddleware creates a Gin middleware that automatically binds and validates
// query parameters into the given prototype object.
// The prototype must be a pointer to a struct; each request gets a new zero-value copy.
// On bind failure or validation failure, the middleware writes a unified error response and aborts.
// On success, the validated object is stored in request context for downstream retrieval.
//
// ValidateQueryMiddleware 创建自动绑定并校验查询参数的 Gin 中间件。
// prototype 必须是结构体指针；每次请求会创建一个新的零值拷贝。
// 绑定或校验失败时，中间件输出统一错误响应并中断请求链。
// 校验成功后，已校验对象存入请求上下文供下游复用。
func ValidateQueryMiddleware(validator datacontract.Validator, prototype any) func(*gin.Context) {
	protoType := reflect.TypeOf(prototype)
	if protoType == nil || protoType.Kind() != reflect.Ptr || protoType.Elem().Kind() != reflect.Struct {
		panic("ValidateQueryMiddleware: prototype must be a pointer to a struct")
	}
	elemType := protoType.Elem()

	return func(c *gin.Context) {
		obj := reflect.New(elemType).Interface()

		httpCtx := newContext(c)
		if err := BindAndValidateQuery(httpCtx, validator, obj); err != nil {
			c.Abort()
			return
		}
		c.Next()
	}
}
