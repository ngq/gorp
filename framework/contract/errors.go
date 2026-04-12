package contract

import (
	"errors"
	"fmt"
)

const (
	// ErrorsKey 是统一错误处理在容器中的绑定 key。
	ErrorsKey = "framework.errors"

	// ErrorCodeOK 成功状态码
	ErrorCodeOK = 200

	// ErrorCodeBadRequest 客户端错误
	ErrorCodeBadRequest = 400

	// ErrorCodeUnauthorized 未授权
	ErrorCodeUnauthorized = 401

	// ErrorCodeForbidden 禁止访问
	ErrorCodeForbidden = 403

	// ErrorCodeNotFound 资源不存在
	ErrorCodeNotFound = 404

	// ErrorCodeConflict 冲突
	ErrorCodeConflict = 409

	// ErrorCodeTooManyRequests 请求过多
	ErrorCodeTooManyRequests = 429

	// ErrorCodeInternalServerError 服务器内部错误
	ErrorCodeInternalServerError = 500

	// ErrorCodeServiceUnavailable 服务不可用
	ErrorCodeServiceUnavailable = 503

	// ErrorCodeGatewayTimeout 网关超时
	ErrorCodeGatewayTimeout = 504
)

// ErrorReason 定义错误原因类型。
//
// 中文说明：
// - 用于区分不同类型的错误；
// - 便于错误分类和处理；
// - 可用于国际化错误消息。
type ErrorReason string

const (
	// ErrorReasonUnknown 未知错误
	ErrorReasonUnknown ErrorReason = "UNKNOWN"

	// ErrorReasonBadRequest 请求参数错误
	ErrorReasonBadRequest ErrorReason = "BAD_REQUEST"

	// ErrorReasonUnauthorized 未授权
	ErrorReasonUnauthorized ErrorReason = "UNAUTHORIZED"

	// ErrorReasonForbidden 禁止访问
	ErrorReasonForbidden ErrorReason = "FORBIDDEN"

	// ErrorReasonNotFound 资源不存在
	ErrorReasonNotFound ErrorReason = "NOT_FOUND"

	// ErrorReasonConflict 冲突
	ErrorReasonConflict ErrorReason = "CONFLICT"

	// ErrorReasonRateLimited 限流
	ErrorReasonRateLimited ErrorReason = "RATE_LIMITED"

	// ErrorReasonInternal 内部错误
	ErrorReasonInternal ErrorReason = "INTERNAL"

	// ErrorReasonServiceUnavailable 服务不可用
	ErrorReasonServiceUnavailable ErrorReason = "SERVICE_UNAVAILABLE"

	// ErrorReasonTimeout 超时
	ErrorReasonTimeout ErrorReason = "TIMEOUT"
)

// Status 定义错误状态。
//
// 中文说明：
// - Code: HTTP 状态码；
// - Reason: 错误原因（大写下划线格式）；
// - Message: 用户友好的错误消息；
// - Metadata: 额外的错误元数据。
type Status struct {
	Code     int32              `json:"code"`
	Reason   ErrorReason         `json:"reason"`
	Message  string              `json:"message"`
	Metadata map[string]string   `json:"metadata,omitempty"`
}

// AppError 定义应用错误接口。
//
// 中文说明：
// - 统一的错误抽象；
// - 支持 HTTP/gRPC 错误转换；
// - 支持错误链；
// - 自己实现，不抄袭 Kratos。
type AppError interface {
	// Error 实现 error 接口
	Error() string

	// Unwrap 返回底层错误
	Unwrap() error

	// Is 检查是否匹配目标错误
	Is(target error) bool

	// GetStatus 获取错误状态
	GetStatus() *Status

	// WithCause 添加底层错误原因
	WithCause(cause error) AppError

	// WithMetadata 添加元数据
	WithMetadata(md map[string]string) AppError

	// GRPCStatus 返回 gRPC 状态（用于 gRPC 错误转换）
	GRPCStatus() any
}

// NewError 创建新的应用错误。
//
// 中文说明：
// - code: HTTP 状态码；
// - reason: 错误原因；
// - message: 错误消息。
func NewError(code int, reason ErrorReason, message string) AppError {
	return &appError{
		Status: &Status{
			Code:    int32(code),
			Reason:  reason,
			Message: message,
		},
	}
}

// NewErrorf 创建带格式化的应用错误。
func NewErrorf(code int, reason ErrorReason, format string, args ...any) AppError {
	return NewError(code, reason, fmt.Sprintf(format, args...))
}

// BadRequest 创建 400 错误。
func BadRequest(reason ErrorReason, message string) AppError {
	return NewError(ErrorCodeBadRequest, reason, message)
}

// Unauthorized 创建 401 错误。
func Unauthorized(message string) AppError {
	return NewError(ErrorCodeUnauthorized, ErrorReasonUnauthorized, message)
}

// Forbidden 创建 403 错误。
func Forbidden(message string) AppError {
	return NewError(ErrorCodeForbidden, ErrorReasonForbidden, message)
}

// NotFound 创建 404 错误。
func NotFound(message string) AppError {
	return NewError(ErrorCodeNotFound, ErrorReasonNotFound, message)
}

// Conflict 创建 409 错误。
func Conflict(message string) AppError {
	return NewError(ErrorCodeConflict, ErrorReasonConflict, message)
}

// RateLimited 创建 429 错误。
func RateLimited(message string) AppError {
	return NewError(ErrorCodeTooManyRequests, ErrorReasonRateLimited, message)
}

// InternalError 创建 500 错误。
func InternalError(message string) AppError {
	return NewError(ErrorCodeInternalServerError, ErrorReasonInternal, message)
}

// ServiceUnavailable 创建 503 错误。
func ServiceUnavailable(message string) AppError {
	return NewError(ErrorCodeServiceUnavailable, ErrorReasonServiceUnavailable, message)
}

// Timeout 创建 504 错误。
func Timeout(message string) AppError {
	return NewError(ErrorCodeGatewayTimeout, ErrorReasonTimeout, message)
}

// FromError 从 error 转换为 AppError。
//
// 中文说明：
// - 如果 err 已经是 AppError，直接返回；
// - 否则包装为 500 内部错误。
func FromError(err error) AppError {
	if err == nil {
		return nil
	}
	var appErr AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return InternalError(err.Error())
}

// Code 获取错误的 HTTP 状态码。
func Code(err error) int {
	if err == nil {
		return ErrorCodeOK
	}
	appErr := FromError(err)
	if appErr == nil {
		return ErrorCodeInternalServerError
	}
	return int(appErr.GetStatus().Code)
}

// Reason 获取错误原因。
func Reason(err error) ErrorReason {
	if err == nil {
		return ErrorReasonUnknown
	}
	appErr := FromError(err)
	if appErr == nil {
		return ErrorReasonUnknown
	}
	return appErr.GetStatus().Reason
}

// ErrorMessage 获取错误消息。
func ErrorMessage(err error) string {
	if err == nil {
		return ""
	}
	appErr := FromError(err)
	if appErr == nil {
		return err.Error()
	}
	return appErr.GetStatus().Message
}

// appError 是 AppError 的默认实现。
type appError struct {
	*Status
	cause error
}

func (e *appError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("error: code=%d reason=%s message=%s cause=%v",
			e.Code, e.Reason, e.Message, e.cause)
	}
	return fmt.Sprintf("error: code=%d reason=%s message=%s",
		e.Code, e.Reason, e.Message)
}

func (e *appError) Unwrap() error {
	return e.cause
}

func (e *appError) Is(target error) bool {
	t, ok := target.(*appError)
	if !ok {
		return false
	}
	return e.Code == t.Code && e.Reason == t.Reason
}

func (e *appError) GetStatus() *Status {
	return e.Status
}

func (e *appError) WithCause(cause error) AppError {
	return &appError{
		Status: &Status{
			Code:     e.Code,
			Reason:   e.Reason,
			Message:  e.Message,
			Metadata: e.Metadata,
		},
		cause: cause,
	}
}

func (e *appError) WithMetadata(md map[string]string) AppError {
	newMd := make(map[string]string, len(e.Metadata)+len(md))
	for k, v := range e.Metadata {
		newMd[k] = v
	}
	for k, v := range md {
		newMd[k] = v
	}
	return &appError{
		Status: &Status{
			Code:     e.Code,
			Reason:   e.Reason,
			Message:  e.Message,
			Metadata: newMd,
		},
		cause: e.cause,
	}
}

func (e *appError) GRPCStatus() any {
	// 返回 nil，gRPC 错误转换由 provider 实现
	return nil
}