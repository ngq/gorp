package contract

import "context"

const ErrorReporterKey = "framework.error_reporter"

// ErrorReport 错误报告结构。
//
// 中文说明：
// - 包含错误详情、上下文信息、标签等；
// - 用于统一错误上报接口。
type ErrorReport struct {
	Error       error              // 原始错误
	Message     string             // 错误消息
	Tags        map[string]string  // 标签
	Context     map[string]any     // 上下文信息
	StackTrace  string             // 堆栈信息
	User        *ErrorUser         // 用户信息
	Request     *ErrorRequest      // 请求信息
}

// ErrorUser 用户信息。
type ErrorUser struct {
	ID    string
	Name  string
	Email string
}

// ErrorRequest 请求信息。
type ErrorRequest struct {
	Method  string
	URL     string
	Headers map[string]string
	Body    string
}

// ErrorReporter 错误上报接口。
//
// 中文说明：
// - 定义统一的错误上报接口；
// - 支持不同的后端实现（Sentry、自定义等）；
// - ReportSync 同步上报，ReportAsync 异步上报。
type ErrorReporter interface {
	// ReportSync 同步上报错误
	ReportSync(ctx context.Context, report *ErrorReport) error
	// ReportAsync 异步上报错误
	ReportAsync(ctx context.Context, report *ErrorReport)
	// Flush 刷新缓冲区（通常在应用关闭时调用）
	Flush()
}

// ErrorReporterConfig 错误上报配置。
type ErrorReporterConfig struct {
	// Enabled 是否启用
	Enabled bool
	// DSN Sentry DSN（或其他后端地址）
	DSN string
	// Environment 环境标识
	Environment string
	// Release 版本号
	Release string
	// SampleRate 采样率（0-1）
	SampleRate float64
	// Tags 默认标签
	Tags map[string]string
}