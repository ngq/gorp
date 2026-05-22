// Package error_reporter_test provides unit tests for the error reporter provider.
//
// 适用场景：
// - 验证 Error Reporter provider 的注册与错误上报行为。
package error_reporter

import (
	"context"
	"errors"
	"testing"

	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	"github.com/stretchr/testify/assert"
)

// mockLogger 用于测试的模拟日志器
type mockLogger struct {
	lastMessage string
	lastFields  []observabilitycontract.Field
}

func (m *mockLogger) Debug(msg string, fields ...observabilitycontract.Field) {}
func (m *mockLogger) Info(msg string, fields ...observabilitycontract.Field)  {}
func (m *mockLogger) Warn(msg string, fields ...observabilitycontract.Field)  {}
func (m *mockLogger) Error(msg string, fields ...observabilitycontract.Field) {
	m.lastMessage = msg
	m.lastFields = fields
}
func (m *mockLogger) With(fields ...observabilitycontract.Field) observabilitycontract.Logger {
	return m
}

// TestLogReporter_ReportSync verifies synchronous error reporting via logger.
//
// TestLogReporter_ReportSync 验证通过日志记录器进行同步错误上报。
func TestLogReporter_ReportSync(t *testing.T) {
	logger := &mockLogger{}
	reporter := NewLogReporter(logger)

	report := &resiliencecontract.ErrorReport{
		Error:   errors.New("test error"),
		Message: "test message",
		Tags:    map[string]string{"env": "test"},
		Context: map[string]any{"user_id": 123},
	}

	err := reporter.ReportSync(context.Background(), report)
	assert.NoError(t, err)
	assert.Equal(t, "error report", logger.lastMessage)
	assert.NotEmpty(t, logger.lastFields)
}

// TestLogReporter_ReportSync_NilLogger verifies behavior when logger is nil.
//
// TestLogReporter_ReportSync_NilLogger 验证日志记录器为 nil 时的行为。
func TestLogReporter_ReportSync_NilLogger(t *testing.T) {
	reporter := NewLogReporter(nil)

	report := &resiliencecontract.ErrorReport{
		Error:   errors.New("test error"),
		Message: "test message",
	}

	err := reporter.ReportSync(context.Background(), report)
	assert.NoError(t, err) // nil logger 时静默返回
}

// TestLogReporter_ReportAsync verifies asynchronous error reporting via logger.
//
// TestLogReporter_ReportAsync 验证通过日志记录器进行异步错误上报。
func TestLogReporter_ReportAsync(t *testing.T) {
	logger := &mockLogger{}
	reporter := NewLogReporter(logger)

	report := &resiliencecontract.ErrorReport{
		Error:   errors.New("test error"),
		Message: "test message",
	}

	reporter.ReportAsync(context.Background(), report)
	// 异步执行，不检查结果
}

// TestLogReporter_Flush verifies the flush operation on log reporter.
//
// TestLogReporter_Flush 验证日志报告器的刷新操作。
func TestLogReporter_Flush(t *testing.T) {
	reporter := NewLogReporter(&mockLogger{})
	reporter.Flush() // 空操作
}

// TestSentryAdapter_Disabled verifies Sentry adapter when disabled.
//
// TestSentryAdapter_Disabled 验证 Sentry 适配器在禁用状态下的行为。
func TestSentryAdapter_Disabled(t *testing.T) {
	cfg := resiliencecontract.ErrorReporterConfig{
		Enabled: false,
		DSN:     "",
	}
	adapter := NewSentryAdapter(cfg)

	report := &resiliencecontract.ErrorReport{
		Error:   errors.New("test error"),
		Message: "test message",
	}

	// 禁用时返回 nil
	err := adapter.ReportSync(context.Background(), report)
	assert.NoError(t, err)
}

// TestSentryAdapter_EnabledButNotImplemented verifies Sentry adapter when enabled but not implemented.
//
// TestSentryAdapter_EnabledButNotImplemented 验证 Sentry 适配器在启用但未实现时的行为。
func TestSentryAdapter_EnabledButNotImplemented(t *testing.T) {
	cfg := resiliencecontract.ErrorReporterConfig{
		Enabled: true,
		DSN:     "https://test@sentry.io/123",
	}
	adapter := NewSentryAdapter(cfg)

	report := &resiliencecontract.ErrorReport{
		Error:   errors.New("test error"),
		Message: "test message",
	}

	// 启用但未实现时返回错误提示
	err := adapter.ReportSync(context.Background(), report)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sentry-go")
}

// TestSentryAdapter_ReportAsync verifies asynchronous error reporting via Sentry.
//
// TestSentryAdapter_ReportAsync 验证通过 Sentry 进行异步错误上报。
func TestSentryAdapter_ReportAsync(t *testing.T) {
	cfg := resiliencecontract.ErrorReporterConfig{
		Enabled: true,
		DSN:     "https://test@sentry.io/123",
	}
	adapter := NewSentryAdapter(cfg)

	report := &resiliencecontract.ErrorReport{
		Error:   errors.New("test error"),
		Message: "test message",
	}

	adapter.ReportAsync(context.Background(), report)
	// 异步执行，不检查结果
}

// TestSentryAdapter_Flush verifies the flush operation on Sentry adapter.
//
// TestSentryAdapter_Flush 验证 Sentry 适配器的刷新操作。
func TestSentryAdapter_Flush(t *testing.T) {
	adapter := NewSentryAdapter(resiliencecontract.ErrorReporterConfig{})
	adapter.Flush() // 空操作
}

// TestCaptureError verifies error capture and reporting.
//
// TestCaptureError 验证错误捕获与上报功能。
func TestCaptureError(t *testing.T) {
	logger := &mockLogger{}
	reporter := NewLogReporter(logger)

	testErr := errors.New("captured error")
	CaptureError(context.Background(), testErr, reporter)

	// 异步执行，等待一下
	// 注意：CaptureError 是异步的，不检查结果
}

// TestCaptureError_NilError verifies behavior when error is nil.
//
// TestCaptureError_NilError 验证错误为 nil 时的行为。
func TestCaptureError_NilError(t *testing.T) {
	reporter := NewLogReporter(&mockLogger{})
	CaptureError(context.Background(), nil, reporter)
	// nil error 时静默返回
}

// TestCaptureError_NilReporter verifies behavior when reporter is nil.
//
// TestCaptureError_NilReporter 验证报告器为 nil 时的行为。
func TestCaptureError_NilReporter(t *testing.T) {
	CaptureError(context.Background(), errors.New("test"), nil)
	// nil reporter 时静默返回
}

// TestErrorReporterProvider verifies the error reporter provider registration.
//
// TestErrorReporterProvider 验证错误上报服务提供者的注册。
func TestProvider_Name(t *testing.T) {
	p := NewProvider(resiliencecontract.ErrorReporterConfig{})
	assert.Equal(t, "error_reporter", p.Name())
	assert.False(t, p.IsDefer())
	assert.ElementsMatch(t, []string{resiliencecontract.ErrorReporterKey}, p.Provides())
}
