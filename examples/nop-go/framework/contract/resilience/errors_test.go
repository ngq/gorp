// Package resilience_test provides unit tests for error handling.
//
// 适用场景：
// - 验证 AppError 的创建和属性访问。
// - 验证错误包装和元数据操作。
// - 验证 FromError 等辅助函数。
package resilience

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

// ============================================================
// NewError tests
// ============================================================

func TestNewError(t *testing.T) {
	err := NewError(404, ErrorReasonNotFound, "user not found")

	require.NotNil(t, err)
	require.Equal(t, "error: code=404 reason=NOT_FOUND message=user not found", err.Error())
}

func TestNewErrorf(t *testing.T) {
	err := NewErrorf(500, ErrorReasonInternal, "failed to process: %s", "timeout")

	require.NotNil(t, err)
	require.Contains(t, err.Error(), "failed to process: timeout")
}

// ============================================================
// Convenience constructors tests
// ============================================================

func TestBadRequest(t *testing.T) {
	err := BadRequest(ErrorReasonBadRequest, "invalid input")

	require.Equal(t, int32(400), err.GetStatus().Code)
	require.Equal(t, ErrorReasonBadRequest, err.GetStatus().Reason)
}

func TestUnauthorized(t *testing.T) {
	err := Unauthorized("token expired")

	require.Equal(t, int32(401), err.GetStatus().Code)
	require.Equal(t, ErrorReasonUnauthorized, err.GetStatus().Reason)
}

func TestForbidden(t *testing.T) {
	err := Forbidden("access denied")

	require.Equal(t, int32(403), err.GetStatus().Code)
}

func TestNotFound(t *testing.T) {
	err := NotFound("resource not found")

	require.Equal(t, int32(404), err.GetStatus().Code)
}

func TestConflict(t *testing.T) {
	err := Conflict("duplicate entry")

	require.Equal(t, int32(409), err.GetStatus().Code)
}

func TestRateLimited(t *testing.T) {
	err := RateLimited("too many requests")

	require.Equal(t, int32(429), err.GetStatus().Code)
}

func TestInternalError(t *testing.T) {
	err := InternalError("database connection failed")

	require.Equal(t, int32(500), err.GetStatus().Code)
}

func TestServiceUnavailable(t *testing.T) {
	err := ServiceUnavailable("service is down")

	require.Equal(t, int32(503), err.GetStatus().Code)
}

func TestTimeout(t *testing.T) {
	err := Timeout("request timed out")

	require.Equal(t, int32(504), err.GetStatus().Code)
}

// ============================================================
// FromError tests
// ============================================================

func TestFromError_AppError(t *testing.T) {
	original := NewError(404, ErrorReasonNotFound, "not found")

	appErr := FromError(original)
	require.NotNil(t, appErr)
	require.Equal(t, int32(404), appErr.GetStatus().Code)
}

func TestFromError_StandardError(t *testing.T) {
	standardErr := errors.New("standard error")

	appErr := FromError(standardErr)
	require.NotNil(t, appErr)
	require.Equal(t, int32(500), appErr.GetStatus().Code)
	require.Contains(t, appErr.GetStatus().Message, "standard error")
}

func TestFromError_Nil(t *testing.T) {
	appErr := FromError(nil)
	require.Nil(t, appErr)
}

// ============================================================
// Code, Reason, ErrorMessage tests
// ============================================================

func TestCode(t *testing.T) {
	err := NewError(404, ErrorReasonNotFound, "not found")
	require.Equal(t, 404, Code(err))

	require.Equal(t, 200, Code(nil))

	standardErr := errors.New("standard")
	require.Equal(t, 500, Code(standardErr))
}

func TestReason(t *testing.T) {
	err := NewError(404, ErrorReasonNotFound, "not found")
	require.Equal(t, ErrorReasonNotFound, Reason(err))

	require.Equal(t, ErrorReasonUnknown, Reason(nil))
}

func TestErrorMessage(t *testing.T) {
	err := NewError(404, ErrorReasonNotFound, "user not found")
	require.Equal(t, "user not found", ErrorMessage(err))

	require.Equal(t, "", ErrorMessage(nil))

	standardErr := errors.New("standard error")
	require.Equal(t, "standard error", ErrorMessage(standardErr))
}

// ============================================================
// WithCause tests
// ============================================================

func TestWithCause(t *testing.T) {
	cause := errors.New("underlying error")
	err := NewError(500, ErrorReasonInternal, "internal error").WithCause(cause)

	require.NotNil(t, err)
	require.Equal(t, cause, errors.Unwrap(err))
	require.Contains(t, err.Error(), "cause=underlying error")
}

// ============================================================
// WithMetadata tests
// ============================================================

func TestWithMetadata(t *testing.T) {
	err := NewError(404, ErrorReasonNotFound, "not found").
		WithMetadata(map[string]string{"key": "value"})

	require.NotNil(t, err)
	require.Equal(t, "value", err.GetStatus().Metadata["key"])
}

func TestWithMetadata_Merge(t *testing.T) {
	err := NewError(404, ErrorReasonNotFound, "not found").
		WithMetadata(map[string]string{"key1": "value1"}).
		WithMetadata(map[string]string{"key2": "value2"})

	require.Equal(t, "value1", err.GetStatus().Metadata["key1"])
	require.Equal(t, "value2", err.GetStatus().Metadata["key2"])
}

// ============================================================
// Is tests
// ============================================================

func TestIs(t *testing.T) {
	err1 := NewError(404, ErrorReasonNotFound, "not found")
	err2 := NewError(404, ErrorReasonNotFound, "different message")

	require.True(t, errors.Is(err1, err2))

	err3 := NewError(500, ErrorReasonInternal, "internal")
	require.False(t, errors.Is(err1, err3))
}

// ============================================================
// Unwrap tests
// ============================================================

func TestUnwrap(t *testing.T) {
	cause := errors.New("underlying")
	err := NewError(500, ErrorReasonInternal, "internal").WithCause(cause)

	require.Equal(t, cause, errors.Unwrap(err))
}

func TestUnwrap_NoCause(t *testing.T) {
	err := NewError(404, ErrorReasonNotFound, "not found")

	require.Nil(t, errors.Unwrap(err))
}

// ============================================================
// GRPCStatus tests
// ============================================================

func TestGRPCStatus(t *testing.T) {
	err := NewError(404, ErrorReasonNotFound, "not found")

	// Default implementation returns nil
	require.Nil(t, err.GRPCStatus())
}

// ============================================================
// GetStatus tests
// ============================================================

func TestGetStatus(t *testing.T) {
	err := NewError(404, ErrorReasonNotFound, "user not found")
	status := err.GetStatus()

	require.NotNil(t, status)
	require.Equal(t, int32(404), status.Code)
	require.Equal(t, ErrorReasonNotFound, status.Reason)
	require.Equal(t, "user not found", status.Message)
}
