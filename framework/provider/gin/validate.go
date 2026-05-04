package gin

import (
	"reflect"

	"github.com/gin-gonic/gin"
	datacontract "github.com/ngq/gorp/framework/contract/data"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	supportcontract "github.com/ngq/gorp/framework/contract/support"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
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
		c.Request = c.Request.WithContext(supportcontract.NewValidatedBodyContext(c.Request.Context(), obj))
	}
}

func storeValidator(c *gin.Context, validator datacontract.Validator) {
	if c == nil {
		return
	}
	c.Set(validatorGinKey, validator)
}

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

func ValidateBodyMiddleware(validator datacontract.Validator, objType interface{}) transportcontract.HTTPMiddleware {
	return func(next transportcontract.HTTPHandler) transportcontract.HTTPHandler {
		return func(c transportcontract.HTTPContext) {
			if validator == nil || objType == nil {
				if next != nil {
					next(c)
				}
				return
			}

			binder, ok := clonePointerValue(objType)
			if !ok {
				appErr := resiliencecontract.BadRequest(resiliencecontract.ErrorReasonBadRequest, "validate middleware requires pointer objType")
				respondWithError(c, appErr)
				return
			}
			if err := BindAndValidateJSON(c, validator, binder); err != nil {
				return
			}
			if gc, ok := unwrapGinContext(c); ok {
				storeValidatedBody(gc, binder)
			}
			if next != nil {
				next(c)
			}
		}
	}
}

func ValidateMiddleware(validator datacontract.Validator) transportcontract.HTTPMiddleware {
	return func(next transportcontract.HTTPHandler) transportcontract.HTTPHandler {
		return func(c transportcontract.HTTPContext) {
			if gc, ok := unwrapGinContext(c); ok {
				storeValidator(gc, validator)
			}
			if next != nil {
				next(c)
			}
		}
	}
}

func ValidateBody(c *gin.Context, validator datacontract.Validator, obj interface{}) error {
	if err := BindAndValidateJSON(newHTTPContext(c), validator, obj); err != nil {
		return err
	}
	storeValidatedBody(c, obj)
	return nil
}

func ValidateQuery(c *gin.Context, validator datacontract.Validator, obj interface{}) error {
	return BindAndValidateQuery(newHTTPContext(c), validator, obj)
}

func ValidateForm(c *gin.Context, validator datacontract.Validator, obj interface{}) error {
	return BindAndValidate(newHTTPContext(c), validator, obj)
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

func GetValidator(c *gin.Context) datacontract.Validator {
	if v, exists := c.Get(validatorGinKey); exists {
		if validator, ok := v.(datacontract.Validator); ok {
			return validator
		}
	}
	return nil
}

func GetValidatedBody(c *gin.Context) interface{} {
	if v, exists := c.Get(validatedBodyGinKey); exists {
		return v
	}
	if c != nil && c.Request != nil {
		if v, ok := supportcontract.FromValidatedBodyContext(c.Request.Context()); ok {
			return v
		}
	}
	return nil
}
