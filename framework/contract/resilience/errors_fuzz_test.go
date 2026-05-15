// Package resilience_test provides fuzz tests for error handling.
//
// 适用场景：
// - 验证 AppError 对各种边界输入的处理稳定性。
// - 发现潜在的 panic 或异常行为。
package resilience

import (
	"testing"
)

// FuzzNewError fuzzes error creation with arbitrary inputs.
//
// FuzzNewError 对错误创建进行模糊测试。
func FuzzNewError(f *testing.F) {
	f.Add(404, "NOT_FOUND", "user not found")
	f.Add(500, "INTERNAL", "")
	f.Add(0, "", "error")
	f.Add(999, "UNKNOWN", "test error with special chars: \n\t")

	f.Fuzz(func(t *testing.T, code int, reason, message string) {
		err := NewError(code, ErrorReason(reason), message)
		if err == nil {
			t.Fatal("NewError returned nil")
		}

		// Verify Error() doesn't panic
		errStr := err.Error()
		if errStr == "" {
			t.Error("Error() returned empty string")
		}

		// Verify GetStatus doesn't panic
		status := err.GetStatus()
		if status == nil {
			t.Fatal("GetStatus returned nil")
		}
	})
}

// FuzzFromError fuzzes error conversion with arbitrary errors.
//
// FuzzFromError 对错误转换进行模糊测试。
func FuzzFromError(f *testing.F) {
	f.Add("standard error")
	f.Add("")
	f.Add("error with\nnewline")
	f.Add("error with\ttab")
	f.Add("error with 中文")

	f.Fuzz(func(t *testing.T, errMsg string) {
		// Test with standard error
		appErr := FromError(newError(errMsg))
		if appErr == nil && errMsg != "" {
			t.Error("FromError returned nil for non-empty error")
		}

		// Test with nil
		nilErr := FromError(nil)
		if nilErr != nil {
			t.Error("FromError(nil) returned non-nil")
		}
	})
}

// newError creates a simple error for testing
type testError struct {
	msg string
}

func (e *testError) Error() string { return e.msg }

func newError(msg string) error {
	if msg == "" {
		return nil
	}
	return &testError{msg: msg}
}

// FuzzWithCause fuzzes error wrapping with arbitrary causes.
//
// FuzzWithCause 对错误包装进行模糊测试。
func FuzzWithCause(f *testing.F) {
	f.Add(500, "INTERNAL", "internal error", "underlying cause")
	f.Add(404, "NOT_FOUND", "not found", "")

	f.Fuzz(func(t *testing.T, code int, reason, message, causeMsg string) {
		err := NewError(code, ErrorReason(reason), message)

		var cause error
		if causeMsg != "" {
			cause = &testError{msg: causeMsg}
		}

		wrapped := err.WithCause(cause)
		if wrapped == nil {
			t.Fatal("WithCause returned nil")
		}

		// Verify Error() includes cause info
		errStr := wrapped.Error()
		if errStr == "" {
			t.Error("Error() returned empty string")
		}
	})
}

// FuzzWithMetadata fuzzes metadata attachment with arbitrary maps.
//
// FuzzWithMetadata 对元数据附加进行模糊测试。
func FuzzWithMetadata(f *testing.F) {
	f.Add(404, "NOT_FOUND", "not found", "key1", "value1")
	f.Add(500, "INTERNAL", "internal", "", "")

	f.Fuzz(func(t *testing.T, code int, reason, message, k1, v1 string) {
		err := NewError(code, ErrorReason(reason), message)

		md := make(map[string]string)
		if k1 != "" {
			md[k1] = v1
		}

		withMd := err.WithMetadata(md)
		if withMd == nil {
			t.Fatal("WithMetadata returned nil")
		}

		status := withMd.GetStatus()
		if status == nil {
			t.Fatal("GetStatus returned nil")
		}
	})
}

// FuzzCodeReasonMessage fuzzes helper functions with arbitrary errors.
//
// FuzzCodeReasonMessage 对 Code/Reason/ErrorMessage 辅助函数进行模糊测试。
func FuzzCodeReasonMessage(f *testing.F) {
	f.Add(404, "NOT_FOUND", "not found")
	f.Add(500, "INTERNAL", "internal error")

	f.Fuzz(func(t *testing.T, code int, reason, message string) {
		err := NewError(code, ErrorReason(reason), message)

		// These should not panic
		gotCode := Code(err)
		gotReason := Reason(err)
		gotMessage := ErrorMessage(err)

		// Verify consistency
		status := err.GetStatus()
		if status != nil {
			if int(status.Code) != gotCode {
				t.Errorf("Code() = %d, want %d", gotCode, status.Code)
			}
			if status.Reason != gotReason {
				t.Errorf("Reason() = %s, want %s", gotReason, status.Reason)
			}
			if status.Message != gotMessage {
				t.Errorf("ErrorMessage() = %s, want %s", gotMessage, status.Message)
			}
		}
	})
}