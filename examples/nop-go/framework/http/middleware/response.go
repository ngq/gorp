// Application scenarios:
// - Standardize HTTP success and error output across middleware and handlers.
// - Keep framework-level business error codes separate from raw HTTP status codes.
// - Provide reusable responder helpers for common transport-layer response scenarios.
//
// 适用场景：
// - 在中间件与处理器之间统一 HTTP 成功与失败输出。
// - 将框架级业务错误码与原始 HTTP 状态码分离。
// - 为常见传输层响应场景提供可复用的 responder 助手。
package middleware

import (
	"errors"
	"net/http"

	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	supportcontract "github.com/ngq/gorp/framework/contract/support"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

// Response is the standard HTTP response envelope used by the framework.
//
// Response 是框架使用的标准 HTTP 响应包裹结构。
type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// DefaultResponder is the default HTTP responder implementation.
//
// DefaultResponder 是默认的 HTTP responder 实现。
type DefaultResponder struct{}

// DefaultResponderInstance is the shared default responder singleton.
//
// DefaultResponderInstance 是共享的默认 responder 单例。
var DefaultResponderInstance transportcontract.HTTPResponder = DefaultResponder{}

// PaginatedData is the standard pagination payload shape.
//
// PaginatedData 是标准的分页载荷结构。
type PaginatedData struct {
	Items    any   `json:"items"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}

const (
	CodeSuccess            = 0
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

// BusinessError describes an error carrying framework business code and message semantics.
//
// BusinessError 描述带有框架业务错误码与消息语义的错误。
type BusinessError interface {
	error
	Code() int
	Message() string
}

// BizError is the default framework business error implementation.
//
// BizError 是框架默认的业务错误实现。
type BizError struct {
	code    int
	message string
}

// NewDefaultResponder returns the shared default HTTP responder.
//
// NewDefaultResponder 返回共享的默认 HTTP responder。
func NewDefaultResponder() transportcontract.HTTPResponder {
	return DefaultResponderInstance
}

// NewBizError creates a new business error with the given code and message.
//
// NewBizError 使用给定错误码和消息创建业务错误。
func NewBizError(code int, message string) *BizError {
	return &BizError{code: code, message: message}
}

func (e *BizError) Error() string   { return e.message }
func (e *BizError) Code() int       { return e.code }
func (e *BizError) Message() string { return e.message }

// Success writes a standard success response using the default responder.
//
// Success 使用默认 responder 输出标准成功响应。
func Success(c transportcontract.Context, data any) {
	DefaultResponderInstance.Success(c, data)
}

// SuccessWithMessage writes a success response with a custom message.
//
// SuccessWithMessage 输出带自定义消息的成功响应。
func SuccessWithMessage(c transportcontract.Context, message string, data any) {
	DefaultResponderInstance.SuccessWithMessage(c, message, data)
}

// SuccessWithStatus writes a success response with a custom HTTP status.
//
// SuccessWithStatus 输出带自定义 HTTP 状态码的成功响应。
func SuccessWithStatus(c transportcontract.Context, status int, data any) {
	DefaultResponderInstance.SuccessWithStatus(c, status, data)
}

// Error writes an error response using the default responder.
//
// Error 使用默认 responder 输出错误响应。
func Error(c transportcontract.Context, err error) {
	DefaultResponderInstance.Error(c, err)
}

// ErrorWithData writes an error response and attaches extra response data.
//
// ErrorWithData 输出错误响应，并附带额外数据。
func ErrorWithData(c transportcontract.Context, err error, data any) {
	writeResponseHeaders(c)
	code, message := parseError(err)
	c.JSON(codeToHTTPStatus(code), Response{Code: code, Message: message, Data: data})
}

// ErrorWithStatus writes an error response using a caller-provided HTTP status.
//
// ErrorWithStatus 使用调用方指定的 HTTP 状态码输出错误响应。
func ErrorWithStatus(c transportcontract.Context, status int, err error) {
	writeResponseHeaders(c)
	code, message := parseError(err)
	c.JSON(status, Response{Code: code, Message: message})
}

// BadRequest writes a standard bad-request response.
//
// BadRequest 输出标准的错误请求响应。
func BadRequest(c transportcontract.Context, message string) {
	DefaultResponderInstance.BadRequest(c, message)
}

// Unauthorized writes a standard unauthorized response.
//
// Unauthorized 输出标准的未认证响应。
func Unauthorized(c transportcontract.Context, message string) {
	writeResponseHeaders(c)
	c.JSON(http.StatusUnauthorized, Response{Code: CodeUnauthorized, Message: message})
}

// Forbidden writes a standard forbidden response.
//
// Forbidden 输出标准的无权限响应。
func Forbidden(c transportcontract.Context, message string) {
	writeResponseHeaders(c)
	c.JSON(http.StatusForbidden, Response{Code: CodeForbidden, Message: message})
}

// NotFound writes a standard not-found response.
//
// NotFound 输出标准的资源不存在响应。
func NotFound(c transportcontract.Context, message string) {
	writeResponseHeaders(c)
	c.JSON(http.StatusNotFound, Response{Code: CodeNotFound, Message: message})
}

// InternalError writes a standard internal-error response.
//
// InternalError 输出标准的内部错误响应。
func InternalError(c transportcontract.Context, message string) {
	DefaultResponderInstance.InternalError(c, message)
}

// SuccessPaginated writes a standard paginated success response.
//
// SuccessPaginated 输出标准的分页成功响应。
func SuccessPaginated(c transportcontract.Context, items any, total int64, page, pageSize int) {
	Success(c, PaginatedData{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

func (DefaultResponder) Success(c transportcontract.Context, data any) {
	writeResponseHeaders(c)
	c.JSON(http.StatusOK, Response{Code: CodeSuccess, Message: "success", Data: data})
}

func (DefaultResponder) SuccessWithMessage(c transportcontract.Context, message string, data any) {
	writeResponseHeaders(c)
	c.JSON(http.StatusOK, Response{Code: CodeSuccess, Message: message, Data: data})
}

func (DefaultResponder) SuccessWithStatus(c transportcontract.Context, status int, data any) {
	writeResponseHeaders(c)
	c.JSON(status, Response{Code: CodeSuccess, Message: "success", Data: data})
}

func (DefaultResponder) Error(c transportcontract.Context, err error) {
	writeResponseHeaders(c)
	code, message := parseError(err)
	c.JSON(codeToHTTPStatus(code), Response{Code: code, Message: message})
}

func (DefaultResponder) BadRequest(c transportcontract.Context, message string) {
	writeResponseHeaders(c)
	c.JSON(http.StatusBadRequest, Response{Code: CodeBadRequest, Message: message})
}

func (DefaultResponder) InternalError(c transportcontract.Context, message string) {
	writeResponseHeaders(c)
	c.JSON(http.StatusInternalServerError, Response{Code: CodeInternalError, Message: message})
}

func responderFor(c transportcontract.Context) transportcontract.HTTPResponder {
	if c != nil {
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

func writeResponseHeaders(c transportcontract.Context) {
	if c == nil {
		return
	}
	if requestID, ok := supportcontract.FromRequestIDContext(c.Context()); ok && requestID != "" {
		c.SetHeader("X-Request-Id", requestID)
	}
	if traceID, ok := supportcontract.FromTraceIDContext(c.Context()); ok && traceID != "" {
		c.SetHeader("X-Trace-Id", traceID)
	}
}

func parseError(err error) (int, string) {
	if err == nil {
		return CodeSuccess, "success"
	}
	var bizErr BusinessError
	if errors.As(err, &bizErr) {
		return bizErr.Code(), bizErr.Message()
	}
	// 非 BusinessError 不暴露内部错误细节，返回通用消息
	return CodeInternalError, "internal server error"
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
		// 未定义的业务码返回 InternalServerError，避免将服务端错误误判为客户端错误
		// Undefined business codes return InternalServerError to avoid misclassifying server errors as client errors
		return http.StatusInternalServerError
	}
}

// ErrBadRequest creates a bad-request business error.
//
// ErrBadRequest 创建错误请求业务错误。
func ErrBadRequest(message string) *BizError {
	return NewBizError(CodeBadRequest, message)
}

// ErrUnauthorized creates an unauthorized business error.
//
// ErrUnauthorized 创建未认证业务错误。
func ErrUnauthorized(message string) *BizError {
	return NewBizError(CodeUnauthorized, message)
}

// ErrForbidden creates a forbidden business error.
//
// ErrForbidden 创建无权限业务错误。
func ErrForbidden(message string) *BizError {
	return NewBizError(CodeForbidden, message)
}

// ErrNotFound creates a not-found business error.
//
// ErrNotFound 创建资源不存在业务错误。
func ErrNotFound(message string) *BizError {
	return NewBizError(CodeNotFound, message)
}

// ErrInternal creates an internal-error business error.
//
// ErrInternal 创建内部错误业务错误。
func ErrInternal(message string) *BizError {
	return NewBizError(CodeInternalError, message)
}

// ErrUserNotFound creates the standard user-not-found business error.
//
// ErrUserNotFound 创建标准的用户不存在业务错误。
func ErrUserNotFound() *BizError {
	return NewBizError(CodeUserNotFound, "user not found")
}

// ErrOrderNotFound creates the standard order-not-found business error.
//
// ErrOrderNotFound 创建标准的订单不存在业务错误。
func ErrOrderNotFound() *BizError {
	return NewBizError(CodeOrderNotFound, "order not found")
}

// ErrProductNotFound creates the standard product-not-found business error.
//
// ErrProductNotFound 创建标准的商品不存在业务错误。
func ErrProductNotFound() *BizError {
	return NewBizError(CodeProductNotFound, "product not found")
}

// ErrOutOfStock creates the standard out-of-stock business error.
//
// ErrOutOfStock 创建标准的库存不足业务错误。
func ErrOutOfStock() *BizError {
	return NewBizError(CodeProductOutOfStock, "product out of stock")
}
