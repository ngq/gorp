package error_reporter

import (
	"context"
	"fmt"
	"runtime"
	"sync"

	"github.com/ngq/gorp/framework/contract"
)

// LogReporter 日志错误上报器（fallback 实现）。
//
// 中文说明：
// - 当没有配置 Sentry 或其他后端时，使用日志作为 fallback；
// - 适合开发环境和测试环境；
// - 生产环境建议替换为 Sentry 实现。
type LogReporter struct {
	logger contract.Logger
	mu     sync.Mutex
}

// NewLogReporter 创建日志错误上报器。
func NewLogReporter(logger contract.Logger) *LogReporter {
	return &LogReporter{logger: logger}
}

// ReportSync 同步记录错误日志。
func (r *LogReporter) ReportSync(ctx context.Context, report *contract.ErrorReport) error {
	if r.logger == nil {
		return nil
	}

	fields := []contract.Field{
		{Key: "error", Value: report.Error.Error()},
		{Key: "message", Value: report.Message},
	}

	for k, v := range report.Tags {
		fields = append(fields, contract.Field{Key: "tag_" + k, Value: v})
	}

	for k, v := range report.Context {
		fields = append(fields, contract.Field{Key: "ctx_" + k, Value: v})
	}

	if report.StackTrace != "" {
		fields = append(fields, contract.Field{Key: "stack", Value: report.StackTrace})
	}

	r.logger.Error("error report", fields...)
	return nil
}

// ReportAsync 异步记录错误日志。
func (r *LogReporter) ReportAsync(ctx context.Context, report *contract.ErrorReport) {
	go r.ReportSync(ctx, report)
}

// Flush 刷新缓冲区（日志实现不需要）。
func (r *LogReporter) Flush() {}

// SentryAdapter Sentry 适配器。
//
// 中文说明：
// - 提供与 Sentry SDK 的桥接；
// - 由于 Sentry SDK 依赖较重，这里只定义接口位；
// - 实际集成时需要引入 github.com/getsentry/sentry-go。
type SentryAdapter struct {
	config  contract.ErrorReporterConfig
	enabled bool
	mu      sync.Mutex
}

// NewSentryAdapter 创建 Sentry 适配器。
//
// 中文说明：
// - config: Sentry 配置；
// - 注意：此实现为占位实现，实际使用时需要引入 Sentry SDK。
func NewSentryAdapter(config contract.ErrorReporterConfig) *SentryAdapter {
	return &SentryAdapter{
		config:  config,
		enabled: config.Enabled && config.DSN != "",
	}
}

// ReportSync 同步上报错误到 Sentry。
func (a *SentryAdapter) ReportSync(ctx context.Context, report *contract.ErrorReport) error {
	if !a.enabled {
		return nil
	}

	// 占位实现
	// 实际实现需要：
	// 1. sentry.CaptureException(report.Error)
	// 2. 设置 tags、context、user 等
	// 3. 调用 sentry.Flush() 确保发送

	return fmt.Errorf("sentry adapter not fully implemented, please import github.com/getsentry/sentry-go")
}

// ReportAsync 异步上报错误到 Sentry。
func (a *SentryAdapter) ReportAsync(ctx context.Context, report *contract.ErrorReport) {
	go a.ReportSync(ctx, report)
}

// Flush 刷新 Sentry 缓冲区。
func (a *SentryAdapter) Flush() {
	// 占位实现
	// 实际实现：sentry.Flush(2 * time.Second)
}

// CaptureError 从 panic 或 error 中捕获错误。
//
// 中文说明：
// - 便捷函数，用于快速捕获错误；
// - 自动附加堆栈信息。
func CaptureError(ctx context.Context, err error, reporter contract.ErrorReporter) {
	if err == nil || reporter == nil {
		return
	}

	// 获取堆栈信息
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	stack := ""
	if n > 0 {
		frames := runtime.CallersFrames(pcs[:n])
		for {
			frame, more := frames.Next()
			stack += fmt.Sprintf("%s\n\t%s:%d\n", frame.Function, frame.File, frame.Line)
			if !more {
				break
			}
		}
	}

	report := &contract.ErrorReport{
		Error:      err,
		Message:    err.Error(),
		StackTrace: stack,
		Tags:       make(map[string]string),
		Context:    make(map[string]any),
	}

	reporter.ReportAsync(ctx, report)
}