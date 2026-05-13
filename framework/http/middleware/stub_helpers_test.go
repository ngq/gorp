// Package middleware_test provides shared test helpers for middleware unit tests.
//
// 适用场景：
// - 测试辅助类型和函数
// - stubLogger、stubValidator、stubCircuitBreaker 等测试替身
package middleware

import (
	"context"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	prometheus "github.com/prometheus/client_golang/prometheus"
)

// logEntry is a lightweight in-memory log entry for testing.
//
// logEntry 是一个轻量内存 log entry，用于测试。
type logEntry struct {
	level  string
	msg    string
	fields []observabilitycontract.Field
}

// stubLogger is a lightweight in-memory logger used for logging, audit, and recovery assertions.
//
// stubLogger 是一个轻量内存 logger，用于 logging、audit 和 recovery 断言。
type stubLogger struct {
	fields  []observabilitycontract.Field
	entries *[]logEntry
}

func newStubLogger() *stubLogger {
	entries := make([]logEntry, 0, 8)
	return &stubLogger{entries: &entries}
}

func (l *stubLogger) Debug(msg string, fields ...observabilitycontract.Field) {
	l.append("debug", msg, fields...)
}

func (l *stubLogger) Info(msg string, fields ...observabilitycontract.Field) {
	l.append("info", msg, fields...)
}

func (l *stubLogger) Warn(msg string, fields ...observabilitycontract.Field) {
	l.append("warn", msg, fields...)
}

func (l *stubLogger) Error(msg string, fields ...observabilitycontract.Field) {
	l.append("error", msg, fields...)
}

func (l *stubLogger) With(fields ...observabilitycontract.Field) observabilitycontract.Logger {
	combined := append(append([]observabilitycontract.Field{}, l.fields...), fields...)
	return &stubLogger{fields: combined, entries: l.entries}
}

func (l *stubLogger) append(level string, msg string, fields ...observabilitycontract.Field) {
	combined := append(append([]observabilitycontract.Field{}, l.fields...), fields...)
	*l.entries = append(*l.entries, logEntry{level: level, msg: msg, fields: combined})
}

func (l *stubLogger) Entries() []logEntry {
	return append([]logEntry(nil), (*l.entries)...)
}

// stubValidator is a test validator implementation.
type stubValidator struct {
	validateFn func(context.Context, any) error
}

func (v *stubValidator) Validate(ctx context.Context, obj any) error {
	if v.validateFn != nil {
		return v.validateFn(ctx, obj)
	}
	return nil
}

func (v *stubValidator) ValidateVar(context.Context, any, string) error { return nil }
func (v *stubValidator) RegisterCustom(string, datacontract.CustomValidateFunc) error {
	return nil
}
func (v *stubValidator) SetLocale(string) error { return nil }
func (v *stubValidator) TranslateError(err error) resiliencecontract.AppError {
	if appErr, ok := err.(resiliencecontract.AppError); ok {
		return appErr
	}
	return resiliencecontract.BadRequest(resiliencecontract.ErrorReasonBadRequest, err.Error())
}

// stubCircuitBreaker is a test circuit breaker implementation.
type stubCircuitBreaker struct {
	allowErr         error
	allowedResources []string
	successResources []string
	failureResources []string
}

func (s *stubCircuitBreaker) Allow(context.Context, string) error { return s.allowErr }

func (s *stubCircuitBreaker) RecordSuccess(_ context.Context, resource string) {
	s.successResources = append(s.successResources, resource)
}

func (s *stubCircuitBreaker) RecordFailure(_ context.Context, resource string, _ error) {
	s.failureResources = append(s.failureResources, resource)
}

func (s *stubCircuitBreaker) Do(context.Context, string, func() error) error { return nil }

func (s *stubCircuitBreaker) State(context.Context, string) resiliencecontract.CircuitBreakerState {
	return resiliencecontract.CircuitBreakerStateClosed
}

// fieldValue extracts a field value by key from observability fields.
func fieldValue(fields []observabilitycontract.Field, key string) any {
	for _, field := range fields {
		if field.Key == key {
			return field.Value
		}
	}
	return nil
}

// counterValue retrieves a Prometheus counter value by metric name and labels.
func counterValue(metricName string, labels map[string]string) float64 {
	families, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		return 0
	}
	for _, family := range families {
		if family.GetName() != metricName {
			continue
		}
		for _, metric := range family.GetMetric() {
			matched := true
			for _, pair := range metric.GetLabel() {
				expected, ok := labels[pair.GetName()]
				if !ok || expected != pair.GetValue() {
					matched = false
					break
				}
			}
			if matched && len(metric.GetLabel()) == len(labels) && metric.GetCounter() != nil {
				return metric.GetCounter().GetValue()
			}
		}
	}
	return 0
}
