// Package httpx provides HTTP utilities for gorp framework.
// This file exposes the standard HTTP response protocol and error types.
// Business handlers use these types to build consistent HTTP responses.
//
// httpx 包提供 gorp 框架的 HTTP 工具。
// 本文件暴露标准 HTTP 响应协议和错误类型。
// 业务 handler 使用这些类型构建一致的 HTTP 响应。
package httpx

import (
	"errors"
	"net/http"

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

// PaginatedData is the standard pagination payload shape.
//
// PaginatedData 是标准的分页载荷结构。
type PaginatedData struct {
	Items    any   `json:"items"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}

// CodeSuccess is the success response code.
//
// CodeSuccess 是成功响应码。
const CodeSuccess = 0

// 业务错误码常量，与 HTTP 状态码对应但独立编码。
// Business error code constants, mapped to HTTP status codes but independently encoded.
const (
	CodeBadRequest    = 1001 // 请求参数错误
	CodeUnauthorized  = 1002 // 未认证
	CodeForbidden     = 1003 // 无权限
	CodeNotFound      = 1004 // 资源不存在
	CodeInternalError = 1005 // 内部错误
)

// BusinessError describes an error carrying business code and message semantics.
//
// BusinessError 描述带有业务错误码与消息语义的错误。
type BusinessError interface {
	error
	Code() int
	Message() string
}

// BizError is the default business error implementation.
//
// BizError 是默认的业务错误实现。
type BizError struct {
	code    int
	message string
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

// Success writes a standard success response.
//
// Success 输出标准成功响应。
func Success(c transportcontract.Context, data any) {
	c.JSON(http.StatusOK, Response{Code: CodeSuccess, Message: "success", Data: data})
}

// SuccessWithMessage writes a success response with a custom message.
//
// SuccessWithMessage 输出带自定义消息的成功响应。
func SuccessWithMessage(c transportcontract.Context, message string, data any) {
	c.JSON(http.StatusOK, Response{Code: CodeSuccess, Message: message, Data: data})
}

// SuccessWithStatus writes a success response with a custom HTTP status.
//
// SuccessWithStatus 输出带自定义 HTTP 状态码的成功响应。
func SuccessWithStatus(c transportcontract.Context, status int, data any) {
	c.JSON(status, Response{Code: CodeSuccess, Message: "success", Data: data})
}

// Error writes an error response.
//
// Error 输出错误响应。
func Error(c transportcontract.Context, err error) {
	code, message := parseError(err)
	c.JSON(codeToHTTPStatus(code), Response{Code: code, Message: message})
}

// ErrorWithData writes an error response and attaches extra response data.
//
// ErrorWithData 输出错误响应，并附带额外数据。
func ErrorWithData(c transportcontract.Context, err error, data any) {
	code, message := parseError(err)
	c.JSON(codeToHTTPStatus(code), Response{Code: code, Message: message, Data: data})
}

// ErrorWithStatus writes an error response using a caller-provided HTTP status.
//
// ErrorWithStatus 使用调用方指定的 HTTP 状态码输出错误响应。
func ErrorWithStatus(c transportcontract.Context, status int, err error) {
	code, message := parseError(err)
	c.JSON(status, Response{Code: code, Message: message})
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

// BadRequest writes a 400 Bad Request response with a business error code and message.
//
// BadRequest 输出 400 Bad Request 响应，附带业务错误码和消息。
func BadRequest(c transportcontract.Context, message string) {
	c.JSON(http.StatusBadRequest, Response{Code: CodeBadRequest, Message: message})
}

// Unauthorized writes a 401 Unauthorized response with a business error code and message.
//
// Unauthorized 输出 401 Unauthorized 响应，附带业务错误码和消息。
func Unauthorized(c transportcontract.Context, message string) {
	c.JSON(http.StatusUnauthorized, Response{Code: CodeUnauthorized, Message: message})
}

// Forbidden writes a 403 Forbidden response with a business error code and message.
//
// Forbidden 输出 403 Forbidden 响应，附带业务错误码和消息。
func Forbidden(c transportcontract.Context, message string) {
	c.JSON(http.StatusForbidden, Response{Code: CodeForbidden, Message: message})
}

// NotFound writes a 404 Not Found response with a business error code and message.
//
// NotFound 输出 404 Not Found 响应，附带业务错误码和消息。
func NotFound(c transportcontract.Context, message string) {
	c.JSON(http.StatusNotFound, Response{Code: CodeNotFound, Message: message})
}

// InternalError writes a 500 Internal Server Error response with a business error code and message.
//
// InternalError 输出 500 Internal Server Error 响应，附带业务错误码和消息。
func InternalError(c transportcontract.Context, message string) {
	c.JSON(http.StatusInternalServerError, Response{Code: CodeInternalError, Message: message})
}

// parseError extracts code and message from an error.
// If the error implements BusinessError, use its code and message.
// Otherwise, return CodeSuccess for nil error or hide internal error details.
//
// parseError 从错误中提取错误码和消息。
// 如果错误实现了 BusinessError，使用其错误码和消息。
// 否则，nil 错误返回成功码，其他错误隐藏内部细节。
func parseError(err error) (int, string) {
	if err == nil {
		return CodeSuccess, "success"
	}
	var bizErr BusinessError
	if errors.As(err, &bizErr) {
		return bizErr.Code(), bizErr.Message()
	}
	// 非 BusinessError 不暴露内部错误细节，返回通用消息
	return CodeSuccess + 1, "internal server error"
}

// codeToHTTPStatus maps business error codes to HTTP status codes.
// Business should override this function or provide custom mapping.
//
// codeToHTTPStatus 将业务错误码映射到 HTTP 状态码。
// 业务可以覆盖此函数或提供自定义映射。
func codeToHTTPStatus(code int) int {
	if code == CodeSuccess {
		return http.StatusOK
	}
	// 默认：业务错误码 >= 1000 视为客户端错误，其他视为服务端错误
	// Default: business codes >= 1000 are client errors, others are server errors
	if code >= 1000 && code < 5000 {
		return http.StatusBadRequest
	}
	return http.StatusInternalServerError
}
