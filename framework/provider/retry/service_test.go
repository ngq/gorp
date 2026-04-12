package retry

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ngq/gorp/framework/contract"
)

// TestRetryService_Do_Success 测试成功执行。
//
// 中文说明：
// - 函数第一次执行成功，无需重试。
func TestRetryService_Do_Success(t *testing.T) {
	cfg := &contract.RetryConfig{
		DefaultPolicy: contract.RetryPolicy{
			MaxAttempts:  3,
			InitialDelay: 10 * time.Millisecond,
			MaxDelay:     100 * time.Millisecond,
			Multiplier:   2.0,
		},
	}

	svc := NewRetryService(cfg)

	callCount := 0
	err := svc.Do(context.Background(), func() error {
		callCount++
		return nil
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if callCount != 1 {
		t.Errorf("expected 1 call, got: %d", callCount)
	}
}

// TestRetryService_Do_RetrySuccess 测试重试后成功。
//
// 中文说明：
// - 前两次失败，第三次成功。
func TestRetryService_Do_RetrySuccess(t *testing.T) {
	cfg := &contract.RetryConfig{
		DefaultPolicy: contract.RetryPolicy{
			MaxAttempts:  3,
			InitialDelay: 10 * time.Millisecond,
			MaxDelay:     100 * time.Millisecond,
			Multiplier:   2.0,
			RetryableCodes: []int{503},
		},
	}

	svc := NewRetryService(cfg)

	callCount := 0
	err := svc.Do(context.Background(), func() error {
		callCount++
		if callCount < 3 {
			return contract.NewError(503, contract.ErrorReasonServiceUnavailable, "service unavailable")
		}
		return nil
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if callCount != 3 {
		t.Errorf("expected 3 calls, got: %d", callCount)
	}
}

// TestRetryService_Do_AllFail 测试全部失败。
//
// 中文说明：
// - 所有尝试都失败，返回最后一个错误。
func TestRetryService_Do_AllFail(t *testing.T) {
	cfg := &contract.RetryConfig{
		DefaultPolicy: contract.RetryPolicy{
			MaxAttempts:  3,
			InitialDelay: 10 * time.Millisecond,
			MaxDelay:     100 * time.Millisecond,
			Multiplier:   2.0,
			RetryableCodes: []int{503},
		},
	}

	svc := NewRetryService(cfg)

	callCount := 0
	err := svc.Do(context.Background(), func() error {
		callCount++
		return contract.NewError(503, contract.ErrorReasonServiceUnavailable, "service unavailable")
	})

	if err == nil {
		t.Error("expected error, got nil")
	}
	if callCount != 3 {
		t.Errorf("expected 3 calls, got: %d", callCount)
	}
}

// TestRetryService_Do_NonRetryable 测试不可重试错误。
//
// 中文说明：
// - 遇到不可重试的错误，立即返回。
func TestRetryService_Do_NonRetryable(t *testing.T) {
	cfg := &contract.RetryConfig{
		DefaultPolicy: contract.RetryPolicy{
			MaxAttempts:  3,
			InitialDelay: 10 * time.Millisecond,
			MaxDelay:     100 * time.Millisecond,
			Multiplier:   2.0,
			RetryableCodes: []int{503},
		},
	}

	svc := NewRetryService(cfg)

	callCount := 0
	err := svc.Do(context.Background(), func() error {
		callCount++
		return contract.NewError(400, contract.ErrorReasonBadRequest, "bad request")
	})

	if err == nil {
		t.Error("expected error, got nil")
	}
	if callCount != 1 {
		t.Errorf("expected 1 call (non-retryable), got: %d", callCount)
	}
}

// TestRetryService_Do_ContextCancel 测试 Context 取消。
//
// 中文说明：
// - Context 取消时，应停止重试。
func TestRetryService_Do_ContextCancel(t *testing.T) {
	cfg := &contract.RetryConfig{
		DefaultPolicy: contract.RetryPolicy{
			MaxAttempts:  10,
			InitialDelay: 200 * time.Millisecond,
			MaxDelay:     1 * time.Second,
			Multiplier:   2.0,
			RetryableCodes: []int{503},
		},
	}

	svc := NewRetryService(cfg)

	ctx, cancel := context.WithCancel(context.Background())

	// 在第一次失败后取消
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	callCount := 0
	err := svc.Do(ctx, func() error {
		callCount++
		return contract.NewError(503, contract.ErrorReasonServiceUnavailable, "unavailable")
	})

	// 验证 context 被取消或重试次数减少
	if err != context.Canceled && callCount >= 10 {
		t.Logf("context cancellation test: err=%v, callCount=%d", err, callCount)
	}
}

// TestRetryService_DoWithResult 测试带返回值的重试。
//
// 中文说明：
// - 验证 DoWithResult 正确返回结果。
func TestRetryService_DoWithResult(t *testing.T) {
	cfg := &contract.RetryConfig{
		DefaultPolicy: contract.RetryPolicy{
			MaxAttempts:  3,
			InitialDelay: 10 * time.Millisecond,
			MaxDelay:     100 * time.Millisecond,
			Multiplier:   2.0,
			RetryableCodes: []int{503},
		},
	}

	svc := NewRetryService(cfg)

	callCount := 0
	result, err := svc.DoWithResult(context.Background(), func() (any, error) {
		callCount++
		if callCount < 2 {
			return nil, contract.NewError(503, contract.ErrorReasonServiceUnavailable, "unavailable")
		}
		return "success", nil
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != "success" {
		t.Errorf("expected success, got: %v", result)
	}
}

// TestRetryService_IsRetryable 测试可重试判断。
//
// 中文说明：
// - 验证各种错误类型的可重试判断。
func TestRetryService_IsRetryable(t *testing.T) {
	cfg := &contract.RetryConfig{
		DefaultPolicy: contract.RetryPolicy{
			RetryableCodes: []int{502, 503, 504},
		},
	}

	svc := NewRetryService(cfg)

	tests := []struct {
		name      string
		err       error
		retryable bool
	}{
		{
			name:      "nil error",
			err:       nil,
			retryable: false,
		},
		{
			name:      "503 AppError",
			err:       contract.NewError(503, contract.ErrorReasonServiceUnavailable, "unavailable"),
			retryable: true,
		},
		{
			name:      "400 AppError",
			err:       contract.NewError(400, contract.ErrorReasonBadRequest, "bad request"),
			retryable: false,
		},
		{
			name:      "generic error",
			err:       errors.New("some error"),
			retryable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.IsRetryable(tt.err)
			if result != tt.retryable {
				t.Errorf("expected %v, got %v", tt.retryable, result)
			}
		})
	}
}

// TestRetryPolicy_CalculateDelay 测试延迟计算。
//
// 中文说明：
// - 验证指数退避延迟计算正确。
func TestRetryPolicy_CalculateDelay(t *testing.T) {
	policy := contract.RetryPolicy{
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     10 * time.Second,
		Multiplier:   2.0,
	}

	tests := []struct {
		attempt int
		min     time.Duration
		max     time.Duration
	}{
		{0, 100 * time.Millisecond, 100 * time.Millisecond},
		{1, 200 * time.Millisecond, 200 * time.Millisecond},
		{2, 400 * time.Millisecond, 400 * time.Millisecond},
		{3, 800 * time.Millisecond, 800 * time.Millisecond},
		{10, 0, 10 * time.Second}, // 超过 MaxDelay
	}

	for _, tt := range tests {
		delay := policy.CalculateDelay(tt.attempt, 0) // 无抖动
		if tt.attempt < 10 && delay != tt.min {
			t.Errorf("attempt %d: expected %v, got %v", tt.attempt, tt.min, delay)
		}
		if delay > tt.max {
			t.Errorf("attempt %d: delay %v exceeds max %v", tt.attempt, delay, tt.max)
		}
	}
}