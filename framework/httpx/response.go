// Package httpx provides HTTP response helpers and error mappings for gorp framework.
package httpx

import (
	"errors"
	"net/http"

	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
)

// JSONResponder represents any context capable of rendering a JSON response,
// seamlessly supporting both *gin.Context and transportcontract.Context.
type JSONResponder interface {
	JSON(code int, obj any)
}

// Response is the standard HTTP response envelope used by the framework.
type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// PaginatedData is the standard pagination payload shape.
type PaginatedData struct {
	Items    any   `json:"items"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}

// CodeSuccess is the success response code.
const CodeSuccess = 0

// Business error code constants.
const (
	CodeBadRequest         = 400
	CodeUnauthorized       = 401
	CodeForbidden          = 403
	CodeNotFound           = 404
	CodeInternalError      = 500
	CodeServiceUnavailable = 503
	CodeTooManyRequests    = 429
	CodeConflict           = 409
	CodeValidationFailed   = 422
)

// BusinessError describes an error carrying business code and message semantics.
type BusinessError interface {
	error
	Code() int
	Message() string
}

// BizError is the default business error implementation.
type BizError struct {
	code    int
	message string
}

// NewBizError creates a new business error with the given code and message.
func NewBizError(code int, message string) *BizError {
	return &BizError{code: code, message: message}
}

func (e *BizError) Error() string   { return e.message }
func (e *BizError) Code() int       { return e.code }
func (e *BizError) Message() string { return e.message }

// Success writes a standard success response.
func Success(c JSONResponder, data any) {
	c.JSON(http.StatusOK, Response{Code: CodeSuccess, Message: "success", Data: data})
}

// OK is an alias for Success.
func OK(c JSONResponder, data any) {
	Success(c, data)
}

// SuccessWithMessage writes a success response with a custom message.
func SuccessWithMessage(c JSONResponder, message string, data any) {
	c.JSON(http.StatusOK, Response{Code: CodeSuccess, Message: message, Data: data})
}

// SuccessWithStatus writes a success response with a custom HTTP status.
func SuccessWithStatus(c JSONResponder, status int, data any) {
	c.JSON(status, Response{Code: CodeSuccess, Message: "success", Data: data})
}

// Error writes an error response, automatically resolving AppError and BizError.
func Error(c JSONResponder, err error) {
	code, message := parseError(err)
	c.JSON(codeToHTTPStatus(code), Response{Code: code, Message: message})
}

// Fail is an alias for Error.
func Fail(c JSONResponder, err error) {
	Error(c, err)
}

// ErrorWithData writes an error response and attaches extra response data.
func ErrorWithData(c JSONResponder, err error, data any) {
	code, message := parseError(err)
	c.JSON(codeToHTTPStatus(code), Response{Code: code, Message: message, Data: data})
}

// ErrorWithStatus writes an error response using a caller-provided HTTP status.
func ErrorWithStatus(c JSONResponder, status int, err error) {
	code, message := parseError(err)
	c.JSON(status, Response{Code: code, Message: message})
}

// SuccessPaginated writes a standard paginated success response.
func SuccessPaginated(c JSONResponder, items any, total int64, page, pageSize int) {
	Success(c, PaginatedData{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

// BadRequest writes a 400 Bad Request response with a message.
func BadRequest(c JSONResponder, message string) {
	c.JSON(http.StatusBadRequest, Response{Code: CodeBadRequest, Message: message})
}

// Unauthorized writes a 401 Unauthorized response with a message.
func Unauthorized(c JSONResponder, message string) {
	c.JSON(http.StatusUnauthorized, Response{Code: CodeUnauthorized, Message: message})
}

// Forbidden writes a 403 Forbidden response with a message.
func Forbidden(c JSONResponder, message string) {
	c.JSON(http.StatusForbidden, Response{Code: CodeForbidden, Message: message})
}

// NotFound writes a 404 Not Found response with a message.
func NotFound(c JSONResponder, message string) {
	c.JSON(http.StatusNotFound, Response{Code: CodeNotFound, Message: message})
}

// InternalError writes a 500 Internal Server Error response with a message.
func InternalError(c JSONResponder, message string) {
	c.JSON(http.StatusInternalServerError, Response{Code: CodeInternalError, Message: message})
}

func parseError(err error) (int, string) {
	if err == nil {
		return CodeSuccess, "success"
	}
	var bizErr BusinessError
	if errors.As(err, &bizErr) {
		return bizErr.Code(), bizErr.Message()
	}
	var appErr resiliencecontract.AppError
	if errors.As(err, &appErr) {
		if st := appErr.GetStatus(); st != nil {
			return int(st.Code), st.Message
		}
	}
	return CodeInternalError, err.Error()
}

func codeToHTTPStatus(code int) int {
	if code >= 100 && code < 600 {
		return code
	}
	switch code {
	case CodeSuccess:
		return http.StatusOK
	default:
		if code >= 10000 {
			return http.StatusOK
		}
		return http.StatusInternalServerError
	}
}
