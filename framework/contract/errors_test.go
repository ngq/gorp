package contract

import (
	"errors"
	"testing"
)

func TestNewError(t *testing.T) {
	err := NewError(400, ErrorReasonBadRequest, "invalid parameter")

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	status := err.GetStatus()
	if status.Code != 400 {
		t.Errorf("expected code 400, got: %d", status.Code)
	}
	if status.Reason != ErrorReasonBadRequest {
		t.Errorf("expected BAD_REQUEST, got: %s", status.Reason)
	}
	if status.Message != "invalid parameter" {
		t.Errorf("unexpected message: %s", status.Message)
	}
}

func TestNewErrorf(t *testing.T) {
	err := NewErrorf(404, ErrorReasonNotFound, "user %s not found", "john")

	status := err.GetStatus()
	if status.Message != "user john not found" {
		t.Errorf("expected 'user john not found', got: %s", status.Message)
	}
}

func TestBadRequest(t *testing.T) {
	err := BadRequest(ErrorReasonBadRequest, "bad request")

	if Code(err) != 400 {
		t.Errorf("expected code 400")
	}
}

func TestUnauthorized(t *testing.T) {
	err := Unauthorized("please login")

	if Code(err) != 401 {
		t.Errorf("expected code 401")
	}
}

func TestForbidden(t *testing.T) {
	err := Forbidden("access denied")

	if Code(err) != 403 {
		t.Errorf("expected code 403")
	}
}

func TestNotFound(t *testing.T) {
	err := NotFound("resource not found")

	if Code(err) != 404 {
		t.Errorf("expected code 404")
	}
}

func TestInternalError(t *testing.T) {
	err := InternalError("internal error")

	if Code(err) != 500 {
		t.Errorf("expected code 500")
	}
}

func TestServiceUnavailable(t *testing.T) {
	err := ServiceUnavailable("service unavailable")

	if Code(err) != 503 {
		t.Errorf("expected code 503")
	}
}

func TestError_WithCause(t *testing.T) {
	cause := errors.New("underlying error")
	err := NewError(500, ErrorReasonInternal, "internal error").WithCause(cause)

	unwrapped := err.Unwrap()
	if unwrapped != cause {
		t.Error("expected to unwrap to cause")
	}
}

func TestError_WithMetadata(t *testing.T) {
	err := NewError(400, ErrorReasonBadRequest, "bad request").
		WithMetadata(map[string]string{"field": "email"})

	status := err.GetStatus()
	if status.Metadata["field"] != "email" {
		t.Error("expected metadata field=email")
	}
}

func TestError_Is(t *testing.T) {
	err1 := NewError(404, ErrorReasonNotFound, "not found 1")
	err2 := NewError(404, ErrorReasonNotFound, "not found 2")

	// Is 匹配 code 和 reason，不匹配 message
	if !errors.Is(err1, err2) {
		t.Error("errors with same code/reason should match")
	}

	err3 := NewError(500, ErrorReasonInternal, "internal")
	if errors.Is(err1, err3) {
		t.Error("errors with different code/reason should not match")
	}
}

func TestError_Error(t *testing.T) {
	err := NewError(404, ErrorReasonNotFound, "user not found")

	errStr := err.Error()
	if errStr == "" {
		t.Error("error string should not be empty")
	}
}

func TestFromError(t *testing.T) {
	// 测试已经是 AppError 的情况
	appErr := NewError(400, ErrorReasonBadRequest, "bad request")
	result := FromError(appErr)
	if result != appErr {
		t.Error("should return same error")
	}

	// 测试普通 error 的情况
	stdErr := errors.New("standard error")
	result = FromError(stdErr)
	if Code(result) != 500 {
		t.Error("standard error should be wrapped as 500")
	}

	// 测试 nil
	result = FromError(nil)
	if result != nil {
		t.Error("nil should return nil")
	}
}

func TestCode(t *testing.T) {
	// nil 返回 200
	if Code(nil) != 200 {
		t.Error("nil should return 200")
	}

	// AppError 返回实际 code
	err := NewError(404, ErrorReasonNotFound, "not found")
	if Code(err) != 404 {
		t.Error("should return 404")
	}

	// 普通 error 返回 500
	stdErr := errors.New("error")
	if Code(stdErr) != 500 {
		t.Error("standard error should return 500")
	}
}

func TestReason(t *testing.T) {
	err := NewError(404, ErrorReasonNotFound, "not found")
	if Reason(err) != ErrorReasonNotFound {
		t.Error("should return NOT_FOUND")
	}

	if Reason(nil) != ErrorReasonUnknown {
		t.Error("nil should return UNKNOWN")
	}
}

func TestErrorMessage(t *testing.T) {
	err := NewError(404, ErrorReasonNotFound, "user not found")
	if ErrorMessage(err) != "user not found" {
		t.Error("should return message")
	}

	if ErrorMessage(nil) != "" {
		t.Error("nil should return empty string")
	}
}