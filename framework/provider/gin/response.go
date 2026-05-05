package gin

import (
	"net/http"

	"github.com/gin-gonic/gin"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	supportcontract "github.com/ngq/gorp/framework/contract/support"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

type DefaultResponder struct{}

var DefaultResponderInstance transportcontract.HTTPResponder = DefaultResponder{}

func NewDefaultResponder() transportcontract.HTTPResponder {
	return DefaultResponderInstance
}

func responderFor(c transportcontract.HTTPContext) transportcontract.HTTPResponder {
	if c != nil && c.Context() != nil {
		if containerAny, ok := supportcontract.FromContainerContext(c.Context()); ok {
			if container, ok := containerAny.(runtimecontract.Container); ok && container != nil && container.IsBind(transportcontract.HTTPResponderKey) {
				if responderAny, err := container.Make(transportcontract.HTTPResponderKey); err == nil {
					if responder, ok := responderAny.(transportcontract.HTTPResponder); ok && responder != nil {
						return responder
					}
				}
			}
		}
	}
	return DefaultResponderInstance
}

func writeResponseHeaders(c transportcontract.HTTPContext) {
	requestID, _ := supportcontract.FromRequestIDContext(c.Context())
	traceID, _ := supportcontract.FromTraceIDContext(c.Context())

	if requestID != "" {
		c.Header("X-Request-Id", requestID)
	}
	if traceID != "" {
		c.Header("X-Trace-Id", traceID)
	}
}

func Success(c transportcontract.HTTPContext, data any) {
	DefaultResponderInstance.Success(c, data)
}

func (DefaultResponder) Success(c transportcontract.HTTPContext, data any) {
	writeResponseHeaders(c)
	resp := Response{
		Code:    0,
		Message: "success",
		Data:    data,
	}
	c.JSON(http.StatusOK, resp)
}

func SuccessWithMessage(c transportcontract.HTTPContext, message string, data any) {
	DefaultResponderInstance.SuccessWithMessage(c, message, data)
}

func (DefaultResponder) SuccessWithMessage(c transportcontract.HTTPContext, message string, data any) {
	writeResponseHeaders(c)
	resp := Response{
		Code:    0,
		Message: message,
		Data:    data,
	}
	c.JSON(http.StatusOK, resp)
}

func SuccessWithStatus(c transportcontract.HTTPContext, status int, data any) {
	DefaultResponderInstance.SuccessWithStatus(c, status, data)
}

func (DefaultResponder) SuccessWithStatus(c transportcontract.HTTPContext, status int, data any) {
	writeResponseHeaders(c)
	resp := Response{
		Code:    0,
		Message: "success",
		Data:    data,
	}
	c.JSON(status, resp)
}

func Error(c transportcontract.HTTPContext, err error) {
	DefaultResponderInstance.Error(c, err)
}

func (DefaultResponder) Error(c transportcontract.HTTPContext, err error) {
	writeResponseHeaders(c)
	code, message := parseError(err)
	httpStatus := codeToHTTPStatus(code)

	resp := Response{
		Code:    code,
		Message: message,
		Data:    nil,
	}
	c.JSON(httpStatus, resp)
}

func ErrorWithData(c transportcontract.HTTPContext, err error, data any) {
	writeResponseHeaders(c)
	code, message := parseError(err)
	httpStatus := codeToHTTPStatus(code)

	resp := Response{
		Code:    code,
		Message: message,
		Data:    data,
	}
	c.JSON(httpStatus, resp)
}

func ErrorWithStatus(c transportcontract.HTTPContext, status int, err error) {
	writeResponseHeaders(c)
	code, message := parseError(err)

	resp := Response{
		Code:    code,
		Message: message,
		Data:    nil,
	}
	c.JSON(status, resp)
}

func BadRequest(c transportcontract.HTTPContext, message string) {
	DefaultResponderInstance.BadRequest(c, message)
}

func (DefaultResponder) BadRequest(c transportcontract.HTTPContext, message string) {
	writeResponseHeaders(c)
	resp := Response{Code: CodeBadRequest, Message: message}
	c.JSON(http.StatusBadRequest, resp)
}

func Unauthorized(c transportcontract.HTTPContext, message string) {
	writeResponseHeaders(c)
	resp := Response{Code: CodeUnauthorized, Message: message}
	c.JSON(http.StatusUnauthorized, resp)
}

func Forbidden(c transportcontract.HTTPContext, message string) {
	writeResponseHeaders(c)
	resp := Response{Code: CodeForbidden, Message: message}
	c.JSON(http.StatusForbidden, resp)
}

func NotFound(c transportcontract.HTTPContext, message string) {
	writeResponseHeaders(c)
	resp := Response{Code: CodeNotFound, Message: message}
	c.JSON(http.StatusNotFound, resp)
}

func InternalError(c transportcontract.HTTPContext, message string) {
	DefaultResponderInstance.InternalError(c, message)
}

func (DefaultResponder) InternalError(c transportcontract.HTTPContext, message string) {
	writeResponseHeaders(c)
	resp := Response{Code: CodeInternalError, Message: message}
	c.JSON(http.StatusInternalServerError, resp)
}

type PaginatedData struct {
	Items    any   `json:"items"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}

func SuccessPaginated(c transportcontract.HTTPContext, items any, total int64, page, pageSize int) {
	data := PaginatedData{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}
	Success(c, data)
}

const (
	CodeSuccess = 0

	CodeBadRequest         = 1001
	CodeUnauthorized       = 1002
	CodeForbidden          = 1003
	CodeNotFound           = 1004
	CodeInternalError      = 1005
	CodeServiceUnavailable = 1006
	CodeTooManyRequests    = 1007
	CodeConflict           = 1008
	CodeValidationFailed   = 1009

	CodeUserNotFound      = 10001
	CodeUserAlreadyExists = 10002
	CodeUserPasswordWrong = 10003
	CodeUserDisabled      = 10004
	CodeUserTokenExpired  = 10005
	CodeUserTokenInvalid  = 10006

	CodeOrderNotFound      = 20001
	CodeOrderStatusInvalid = 20002
	CodeOrderAlreadyPaid   = 20003
	CodeOrderCanceled      = 20004

	CodeProductNotFound   = 30001
	CodeProductOutOfStock = 30002
	CodeProductDisabled   = 30003

	CodePaymentFailed   = 40001
	CodePaymentTimeout  = 40002
	CodePaymentCanceled = 40003
)

func writeGinResponseHeaders(c *gin.Context) {
	if c == nil {
		return
	}
	writeResponseHeaders(newHTTPContext(c))
}

func parseError(err error) (int, string) {
	if err == nil {
		return CodeSuccess, "success"
	}

	if bizErr, ok := err.(BusinessError); ok {
		return bizErr.Code(), bizErr.Message()
	}

	return CodeInternalError, err.Error()
}

func codeToHTTPStatus(code int) int {
	switch code {
	case CodeSuccess:
		return http.StatusOK
	case CodeBadRequest, CodeValidationFailed:
		return http.StatusBadRequest
	case CodeUnauthorized, CodeUserTokenExpired, CodeUserTokenInvalid:
		return http.StatusUnauthorized
	case CodeForbidden, CodeUserDisabled:
		return http.StatusForbidden
	case CodeNotFound, CodeUserNotFound, CodeOrderNotFound, CodeProductNotFound:
		return http.StatusNotFound
	case CodeConflict, CodeUserAlreadyExists, CodeOrderAlreadyPaid:
		return http.StatusConflict
	case CodeTooManyRequests:
		return http.StatusTooManyRequests
	case CodeServiceUnavailable:
		return http.StatusServiceUnavailable
	default:
		if code >= 1001 && code <= 9999 {
			return http.StatusBadRequest
		}
		return http.StatusInternalServerError
	}
}

type BusinessError interface {
	error
	Code() int
	Message() string
}

type BizError struct {
	code    int
	message string
}

func NewBizError(code int, message string) *BizError {
	return &BizError{code: code, message: message}
}

func (e *BizError) Error() string   { return e.message }
func (e *BizError) Code() int       { return e.code }
func (e *BizError) Message() string { return e.message }

func ErrBadRequest(message string) *BizError {
	return NewBizError(CodeBadRequest, message)
}

func ErrUnauthorized(message string) *BizError {
	return NewBizError(CodeUnauthorized, message)
}

func ErrForbidden(message string) *BizError {
	return NewBizError(CodeForbidden, message)
}

func ErrNotFound(message string) *BizError {
	return NewBizError(CodeNotFound, message)
}

func ErrInternal(message string) *BizError {
	return NewBizError(CodeInternalError, message)
}

func ErrUserNotFound() *BizError {
	return NewBizError(CodeUserNotFound, "用户不存在")
}

func ErrOrderNotFound() *BizError {
	return NewBizError(CodeOrderNotFound, "订单不存在")
}

func ErrProductNotFound() *BizError {
	return NewBizError(CodeProductNotFound, "商品不存在")
}

func ErrOutOfStock() *BizError {
	return NewBizError(CodeProductOutOfStock, "商品库存不足")
}
