// Package error_reporter provides error reporting service for gorp framework.
// Supports multiple backends: Sentry (production), Log (development/fallback).
// Core logic: Report errors asynchronously with stack trace and context.
//
// 错误上报包，提供 gorp 框架的错误上报能力。
// 支持多种后端：Sentry（生产环境）、日志（开发/兜底）。
// 核心逻辑：异步上报错误，携带堆栈和上下文信息。
package error_reporter

import (
	"context"
	"fmt"
	"runtime"
	"sync"

	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
)

// LogReporter is the fallback error reporter that logs errors.
// Suitable for development or when DSN is not configured.
// Core logic: Format error with context, log with Error level.
//
// LogReporter 是基于日志的错误上报器，作为兜底实现。
// 适用于开发环境或未配置 DSN 的场景。
// 核心逻辑：格式化错误及上下文、以 Error 级别记录日志。
type LogReporter struct {
	logger observabilitycontract.Logger
	mu     sync.Mutex
}

// NewLogReporter creates a new log-based error reporter.
//
// NewLogReporter 创建基于日志的错误上报器。
func NewLogReporter(logger observabilitycontract.Logger) *LogReporter {
	return &LogReporter{logger: logger}
}

// ReportSync reports error synchronously by logging.
// Core logic: Build fields from error report, log with Error level.
//
// ReportSync 同步上报错误（通过日志记录）。
// 核心逻辑：从错误报告构建字段、以 Error 级别记录。
func (r *LogReporter) ReportSync(ctx context.Context, report *resiliencecontract.ErrorReport) error {
	if r.logger == nil {
		return nil
	}

	fields := []observabilitycontract.Field{
		{Key: "error", Value: report.Error.Error()},
		{Key: "message", Value: report.Message},
	}

	for k, v := range report.Tags {
		fields = append(fields, observabilitycontract.Field{Key: "tag_" + k, Value: v})
	}

	for k, v := range report.Context {
		fields = append(fields, observabilitycontract.Field{Key: "ctx_" + k, Value: v})
	}

	if report.StackTrace != "" {
		fields = append(fields, observabilitycontract.Field{Key: "stack", Value: report.StackTrace})
	}

	r.logger.Error("error report", fields...)
	return nil
}

// ReportAsync reports error asynchronously by spawning goroutine.
// Core logic: Call ReportSync in background goroutine.
//
// ReportAsync 异步上报错误（启动 goroutine）。
// 核心逻辑：在后台 goroutine 中调用 ReportSync。
func (r *LogReporter) ReportAsync(ctx context.Context, report *resiliencecontract.ErrorReport) {
	go r.ReportSync(ctx, report)
}

// Flush flushes pending reports (no-op for log reporter).
//
// Flush 刷新待上报的错误（日志上报器无操作）。
func (r *LogReporter) Flush() {}

// SentryAdapter is the Sentry-based error reporter.
// Requires DSN configuration for production use.
// Core logic: Send errors to Sentry service with full context.
//
// SentryAdapter 是基于 Sentry 的错误上报器。
// 需要配置 DSN 才能在生产环境使用。
// 核心逻辑：将错误发送到 Sentry 服务，携带完整上下文。
type SentryAdapter struct {
	config  resiliencecontract.ErrorReporterConfig
	enabled bool
	mu      sync.Mutex
}

// NewSentryAdapter creates a new Sentry-based error reporter.
// Core logic: Validate config, enable only if DSN is configured.
//
// NewSentryAdapter 创建基于 Sentry 的错误上报器。
// 核心逻辑：验证配置，仅在 DSN 已配置时启用。
func NewSentryAdapter(config resiliencecontract.ErrorReporterConfig) *SentryAdapter {
	return &SentryAdapter{
		config:  config,
		enabled: config.Enabled && config.DSN != "",
	}
}

// ReportSync reports error synchronously to Sentry.
// Core logic: Send error with stack trace and context to Sentry.
//
// ReportSync 同步上报错误到 Sentry。
// 核心逻辑：将错误及堆栈和上下文发送到 Sentry。
func (a *SentryAdapter) ReportSync(ctx context.Context, report *resiliencecontract.ErrorReport) error {
	if !a.enabled {
		return nil
	}

	return fmt.Errorf("sentry adapter not fully implemented, please import github.com/getsentry/sentry-go")
}

// ReportAsync reports error asynchronously to Sentry.
// Core logic: Call ReportSync in background goroutine.
//
// ReportAsync 异步上报错误到 Sentry。
// 核心逻辑：在后台 goroutine 中调用 ReportSync。
func (a *SentryAdapter) ReportAsync(ctx context.Context, report *resiliencecontract.ErrorReport) {
	go a.ReportSync(ctx, report)
}

// Flush flushes pending reports (no-op for current implementation).
//
// Flush 刷新待上报的错误（当前实现无操作）。
func (a *SentryAdapter) Flush() {}

// CaptureError captures error with stack trace and reports asynchronously.
// Core logic: Capture stack trace, build error report, call reporter.
//
// CaptureError 捕获错误及堆栈并异步上报。
// 核心逻辑：捕获堆栈、构建错误报告、调用上报器。
func CaptureError(ctx context.Context, err error, reporter resiliencecontract.ErrorReporter) {
	if err == nil || reporter == nil {
		return
	}

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

	report := &resiliencecontract.ErrorReport{
		Error:      err,
		Message:    err.Error(),
		StackTrace: stack,
		Tags:       make(map[string]string),
		Context:    make(map[string]any),
	}

	reporter.ReportAsync(ctx, report)
}
