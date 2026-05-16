package resilience

import "context"

const ErrorReporterKey = "framework.error_reporter"

// ErrorReport 错误报告结构。
type ErrorReport struct {
	Error      error
	Message    string
	Tags       map[string]string
	Context    map[string]any
	StackTrace string
	User       *ErrorUser
	Request    *ErrorRequest
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
type ErrorReporter interface {
	ReportSync(ctx context.Context, report *ErrorReport) error
	ReportAsync(ctx context.Context, report *ErrorReport)
	Flush()
}

// ErrorReporterConfig 错误上报配置。
type ErrorReporterConfig struct {
	Enabled     bool
	DSN         string
	Environment string
	Release     string
	SampleRate  float64
	Tags        map[string]string
}
