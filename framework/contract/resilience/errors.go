package resilience

import (
	"errors"
	"fmt"
)

const (
	ErrorsKey = "framework.errors"

	ErrorCodeOK                   = 200
	ErrorCodeBadRequest           = 400
	ErrorCodeUnauthorized         = 401
	ErrorCodeForbidden            = 403
	ErrorCodeNotFound             = 404
	ErrorCodeConflict             = 409
	ErrorCodeTooManyRequests      = 429
	ErrorCodeInternalServerError  = 500
	ErrorCodeServiceUnavailable   = 503
	ErrorCodeGatewayTimeout       = 504
)

type ErrorReason string

const (
	ErrorReasonUnknown            ErrorReason = "UNKNOWN"
	ErrorReasonBadRequest         ErrorReason = "BAD_REQUEST"
	ErrorReasonUnauthorized       ErrorReason = "UNAUTHORIZED"
	ErrorReasonForbidden          ErrorReason = "FORBIDDEN"
	ErrorReasonNotFound           ErrorReason = "NOT_FOUND"
	ErrorReasonConflict           ErrorReason = "CONFLICT"
	ErrorReasonRateLimited        ErrorReason = "RATE_LIMITED"
	ErrorReasonInternal           ErrorReason = "INTERNAL"
	ErrorReasonServiceUnavailable ErrorReason = "SERVICE_UNAVAILABLE"
	ErrorReasonTimeout            ErrorReason = "TIMEOUT"
)

type Status struct {
	Code     int32             `json:"code"`
	Reason   ErrorReason       `json:"reason"`
	Message  string            `json:"message"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

type AppError interface {
	Error() string
	Unwrap() error
	Is(target error) bool
	GetStatus() *Status
	WithCause(cause error) AppError
	WithMetadata(md map[string]string) AppError
	GRPCStatus() any
}

func NewError(code int, reason ErrorReason, message string) AppError {
	return &appError{
		Status: &Status{
			Code:    int32(code),
			Reason:  reason,
			Message: message,
		},
	}
}

func NewErrorf(code int, reason ErrorReason, format string, args ...any) AppError {
	return NewError(code, reason, fmt.Sprintf(format, args...))
}

func BadRequest(reason ErrorReason, message string) AppError {
	return NewError(ErrorCodeBadRequest, reason, message)
}

func Unauthorized(message string) AppError {
	return NewError(ErrorCodeUnauthorized, ErrorReasonUnauthorized, message)
}

func Forbidden(message string) AppError {
	return NewError(ErrorCodeForbidden, ErrorReasonForbidden, message)
}

func NotFound(message string) AppError {
	return NewError(ErrorCodeNotFound, ErrorReasonNotFound, message)
}

func Conflict(message string) AppError {
	return NewError(ErrorCodeConflict, ErrorReasonConflict, message)
}

func RateLimited(message string) AppError {
	return NewError(ErrorCodeTooManyRequests, ErrorReasonRateLimited, message)
}

func InternalError(message string) AppError {
	return NewError(ErrorCodeInternalServerError, ErrorReasonInternal, message)
}

func ServiceUnavailable(message string) AppError {
	return NewError(ErrorCodeServiceUnavailable, ErrorReasonServiceUnavailable, message)
}

func Timeout(message string) AppError {
	return NewError(ErrorCodeGatewayTimeout, ErrorReasonTimeout, message)
}

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

type appError struct {
	*Status
	cause error
}

func (e *appError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("error: code=%d reason=%s message=%s cause=%v", e.Code, e.Reason, e.Message, e.cause)
	}
	return fmt.Sprintf("error: code=%d reason=%s message=%s", e.Code, e.Reason, e.Message)
}

func (e *appError) Unwrap() error { return e.cause }

func (e *appError) Is(target error) bool {
	t, ok := target.(*appError)
	if !ok {
		return false
	}
	return e.Code == t.Code && e.Reason == t.Reason
}

func (e *appError) GetStatus() *Status { return e.Status }

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

func (e *appError) GRPCStatus() any { return nil }
