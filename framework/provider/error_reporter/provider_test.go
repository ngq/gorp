package error_reporter

import (
	"context"
	"errors"
	"testing"

	"github.com/ngq/gorp/framework/contract"
	"github.com/stretchr/testify/assert"
)

// mockLogger 用于测试的模拟日志器
type mockLogger struct {
	lastMessage string
	lastFields  []contract.Field
}

func (m *mockLogger) Debug(msg string, fields ...contract.Field) {}
func (m *mockLogger) Info(msg string, fields ...contract.Field)  {}
func (m *mockLogger) Warn(msg string, fields ...contract.Field)  {}
func (m *mockLogger) Error(msg string, fields ...contract.Field) {
	m.lastMessage = msg
	m.lastFields = fields
}
func (m *mockLogger) With(fields ...contract.Field) contract.Logger {
	return m
}

func TestLogReporter_ReportSync(t *testing.T) {
	logger := &mockLogger{}
	reporter := NewLogReporter(logger)

	report := &contract.ErrorReport{
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

func TestLogReporter_ReportSync_NilLogger(t *testing.T) {
	reporter := NewLogReporter(nil)

	report := &contract.ErrorReport{
		Error:   errors.New("test error"),
		Message: "test message",
	}

	err := reporter.ReportSync(context.Background(), report)
	assert.NoError(t, err) // nil logger 时静默返回
}

func TestLogReporter_ReportAsync(t *testing.T) {
	logger := &mockLogger{}
	reporter := NewLogReporter(logger)

	report := &contract.ErrorReport{
		Error:   errors.New("test error"),
		Message: "test message",
	}

	reporter.ReportAsync(context.Background(), report)
	// 异步执行，不检查结果
}

func TestLogReporter_Flush(t *testing.T) {
	reporter := NewLogReporter(&mockLogger{})
	reporter.Flush() // 空操作
}

func TestSentryAdapter_Disabled(t *testing.T) {
	cfg := contract.ErrorReporterConfig{
		Enabled: false,
		DSN:     "",
	}
	adapter := NewSentryAdapter(cfg)

	report := &contract.ErrorReport{
		Error:   errors.New("test error"),
		Message: "test message",
	}

	// 禁用时返回 nil
	err := adapter.ReportSync(context.Background(), report)
	assert.NoError(t, err)
}

func TestSentryAdapter_EnabledButNotImplemented(t *testing.T) {
	cfg := contract.ErrorReporterConfig{
		Enabled: true,
		DSN:     "https://test@sentry.io/123",
	}
	adapter := NewSentryAdapter(cfg)

	report := &contract.ErrorReport{
		Error:   errors.New("test error"),
		Message: "test message",
	}

	// 启用但未实现时返回错误提示
	err := adapter.ReportSync(context.Background(), report)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sentry-go")
}

func TestSentryAdapter_ReportAsync(t *testing.T) {
	cfg := contract.ErrorReporterConfig{
		Enabled: true,
		DSN:     "https://test@sentry.io/123",
	}
	adapter := NewSentryAdapter(cfg)

	report := &contract.ErrorReport{
		Error:   errors.New("test error"),
		Message: "test message",
	}

	adapter.ReportAsync(context.Background(), report)
	// 异步执行，不检查结果
}

func TestSentryAdapter_Flush(t *testing.T) {
	adapter := NewSentryAdapter(contract.ErrorReporterConfig{})
	adapter.Flush() // 空操作
}

func TestCaptureError(t *testing.T) {
	logger := &mockLogger{}
	reporter := NewLogReporter(logger)

	testErr := errors.New("captured error")
	CaptureError(context.Background(), testErr, reporter)

	// 异步执行，等待一下
	// 注意：CaptureError 是异步的，不检查结果
}

func TestCaptureError_NilError(t *testing.T) {
	reporter := NewLogReporter(&mockLogger{})
	CaptureError(context.Background(), nil, reporter)
	// nil error 时静默返回
}

func TestCaptureError_NilReporter(t *testing.T) {
	CaptureError(context.Background(), errors.New("test"), nil)
	// nil reporter 时静默返回
}

func TestProvider_Name(t *testing.T) {
	p := NewProvider(contract.ErrorReporterConfig{})
	assert.Equal(t, "error_reporter", p.Name())
	assert.False(t, p.IsDefer())
	assert.ElementsMatch(t, []string{contract.ErrorReporterKey}, p.Provides())
}