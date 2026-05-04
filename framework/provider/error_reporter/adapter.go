package error_reporter

import (
	"context"
	"fmt"
	"runtime"
	"sync"

	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
)

type LogReporter struct {
	logger observabilitycontract.Logger
	mu     sync.Mutex
}

func NewLogReporter(logger observabilitycontract.Logger) *LogReporter {
	return &LogReporter{logger: logger}
}

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

func (r *LogReporter) ReportAsync(ctx context.Context, report *resiliencecontract.ErrorReport) {
	go r.ReportSync(ctx, report)
}

func (r *LogReporter) Flush() {}

type SentryAdapter struct {
	config  resiliencecontract.ErrorReporterConfig
	enabled bool
	mu      sync.Mutex
}

func NewSentryAdapter(config resiliencecontract.ErrorReporterConfig) *SentryAdapter {
	return &SentryAdapter{
		config:  config,
		enabled: config.Enabled && config.DSN != "",
	}
}

func (a *SentryAdapter) ReportSync(ctx context.Context, report *resiliencecontract.ErrorReport) error {
	if !a.enabled {
		return nil
	}

	return fmt.Errorf("sentry adapter not fully implemented, please import github.com/getsentry/sentry-go")
}

func (a *SentryAdapter) ReportAsync(ctx context.Context, report *resiliencecontract.ErrorReport) {
	go a.ReportSync(ctx, report)
}

func (a *SentryAdapter) Flush() {}

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
