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

// storeValidatedBodyOnHTTPContext writes the validated request object into request context.
//
// storeValidatedBodyOnHTTPContext 将已校验的请求对象写回请求上下文。
func storeValidatedBodyOnHTTPContext(c transportcontract.HTTPContext, obj any) {
	if c == nil || obj == nil {
		return
	}
	req := c.Request()
	if req == nil {
		return
	}
	c.SetRequest(req.WithContext(supportcontract.NewValidatedBodyContext(req.Context(), obj)))
}

// validateBoundValue runs validator logic and emits a unified error response on validation failure.
//
// validateBoundValue 执行校验逻辑，并在校验失败时输出统一错误响应。
func validateBoundValue(c transportcontract.HTTPContext, validator datacontract.Validator, obj any) error {
	if validator == nil {
		return nil
	}
	if err := validator.Validate(c.Context(), obj); err != nil {
		appErr, ok := err.(resiliencecontract.AppError)
		if !ok {
			appErr = resiliencecontract.BadRequest(resiliencecontract.ErrorReasonBadRequest, err.Error())
		}
		respondWithError(c, appErr)
		return appErr
	}
	storeValidatedBodyOnHTTPContext(c, obj)
	return nil
}

// BindAndValidateJSON binds a JSON payload and validates the bound object.
//
// BindAndValidateJSON 绑定 JSON 载荷并校验绑定后的对象。
func BindAndValidateJSON(c transportcontract.HTTPContext, validator datacontract.Validator, obj any) error {
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
func BindAndValidateQuery(c transportcontract.HTTPContext, validator datacontract.Validator, obj any) error {
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
func BindAndValidate(c transportcontract.HTTPContext, validator datacontract.Validator, obj any) error {
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
func respondWithError(c transportcontract.HTTPContext, err resiliencecontract.AppError) {
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
