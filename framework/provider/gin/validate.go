package gin

import (
	datacontract "github.com/ngq/gorp/framework/contract/data"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	supportcontract "github.com/ngq/gorp/framework/contract/support"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

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

func BindAndValidateJSON(c transportcontract.HTTPContext, validator datacontract.Validator, obj any) error {
	if err := c.BindJSON(obj); err != nil {
		appErr := resiliencecontract.BadRequest(resiliencecontract.ErrorReasonBadRequest, "invalid request body: "+err.Error())
		respondWithError(c, appErr)
		return appErr
	}
	return validateBoundValue(c, validator, obj)
}

func BindAndValidateQuery(c transportcontract.HTTPContext, validator datacontract.Validator, obj any) error {
	if err := c.BindQuery(obj); err != nil {
		appErr := resiliencecontract.BadRequest(resiliencecontract.ErrorReasonBadRequest, "invalid query parameters: "+err.Error())
		respondWithError(c, appErr)
		return appErr
	}
	return validateBoundValue(c, validator, obj)
}

func BindAndValidate(c transportcontract.HTTPContext, validator datacontract.Validator, obj any) error {
	if err := c.Bind(obj); err != nil {
		appErr := resiliencecontract.BadRequest(resiliencecontract.ErrorReasonBadRequest, "invalid form data: "+err.Error())
		respondWithError(c, appErr)
		return appErr
	}
	return validateBoundValue(c, validator, obj)
}

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

type ValidateErrorResponse struct {
	Code    int    `json:"code"`
	Reason  string `json:"reason"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}
