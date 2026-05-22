// Package observability provides observability service implementation for gorp framework.
// Includes metrics, tracing, logging, and error reporting components.
// Prometheus metrics, noop tracer by default, extensible to OpenTelemetry.
//
// 可观测性包提供可观测性服务实现，用于 gorp 框架。
// 包括指标、追踪、日志和错误上报组件。
// Prometheus 指标、默认 noop tracer，可扩展到 OpenTelemetry。
package observability

import (
	"context"
	"time"

	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// DefaultObservability aggregates metrics, tracer, logger, and error reporter.
// Core logic: Provide unified access to all observability components.
//
// DefaultObservability 聚合指标、追踪器、日志器和错误上报器。
// 核心逻辑：提供统一的可观测性组件访问入口。
type DefaultObservability struct {
	metrics       observabilitycontract.Metrics
	tracer        observabilitycontract.Tracer
	logger        observabilitycontract.Logger
	errorReporter resiliencecontract.ErrorReporter
}

// NewDefaultObservability creates observability with all components.
// Core logic: Store all components for unified access.
//
// NewDefaultObservability 创建携带所有组件的可观测性实例。
// 核心逻辑：存储所有组件供统一访问。
func NewDefaultObservability(
	metrics observabilitycontract.Metrics,
	tracer observabilitycontract.Tracer,
	logger observabilitycontract.Logger,
	errorReporter resiliencecontract.ErrorReporter,
) *DefaultObservability {
	return &DefaultObservability{
		metrics:       metrics,
		tracer:        tracer,
		logger:        logger,
		errorReporter: errorReporter,
	}
}

func (o *DefaultObservability) Metrics() observabilitycontract.Metrics { return o.metrics }
func (o *DefaultObservability) Tracer() observabilitycontract.Tracer   { return o.tracer }
func (o *DefaultObservability) Logger() observabilitycontract.Logger   { return o.logger }
func (o *DefaultObservability) ErrorReporter() resiliencecontract.ErrorReporter {
	return o.errorReporter
}

type PrometheusMetrics struct{}

func NewPrometheusMetrics() *PrometheusMetrics {
	return &PrometheusMetrics{}
}

func (m *PrometheusMetrics) Counter(name string, labels map[string]string, delta float64) {
	counter := promauto.NewCounterVec(prometheus.CounterOpts{
		Name: name,
		Help: name,
	}, labelKeys(labels))
	counter.WithLabelValues(labelValues(labels)...).Add(delta)
}

func (m *PrometheusMetrics) Gauge(name string, labels map[string]string, value float64) {
	gauge := promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: name,
		Help: name,
	}, labelKeys(labels))
	gauge.WithLabelValues(labelValues(labels)...).Set(value)
}

func (m *PrometheusMetrics) Histogram(name string, labels map[string]string, value float64) {
	histogram := promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    name,
		Help:    name,
		Buckets: prometheus.DefBuckets,
	}, labelKeys(labels))
	histogram.WithLabelValues(labelValues(labels)...).Observe(value)
}

func (m *PrometheusMetrics) Timing(name string, labels map[string]string, duration time.Duration) {
	m.Histogram(name+"_seconds", labels, duration.Seconds())
}

func labelKeys(labels map[string]string) []string {
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	return keys
}

func labelValues(labels map[string]string) []string {
	keys := labelKeys(labels)
	values := make([]string, len(keys))
	for i, k := range keys {
		values[i] = labels[k]
	}
	return values
}

type NoopTracer struct{}

func NewNoopTracer() *NoopTracer { return &NoopTracer{} }

func (t *NoopTracer) StartSpan(ctx context.Context, name string, opts ...observabilitycontract.SpanOption) (context.Context, observabilitycontract.Span) {
	return ctx, &NoopSpan{}
}

func (t *NoopTracer) SpanFromContext(ctx context.Context) observabilitycontract.Span {
	return &NoopSpan{}
}

func (t *NoopTracer) Inject(ctx context.Context, carrier observabilitycontract.TextMapCarrier) error {
	return nil
}

func (t *NoopTracer) Extract(ctx context.Context, carrier observabilitycontract.TextMapCarrier) (context.Context, error) {
	return ctx, nil
}

type NoopSpan struct{}

func (s *NoopSpan) End(options ...observabilitycontract.SpanEndOption)      {}
func (s *NoopSpan) AddEvent(name string, attributes map[string]interface{}) {}
func (s *NoopSpan) SetTag(key string, value interface{})                    {}
func (s *NoopSpan) SetAttributes(attributes map[string]interface{})         {}
func (s *NoopSpan) SetError(err error)                                      {}
func (s *NoopSpan) SetStatus(code observabilitycontract.SpanStatusCode, description string) {
}
func (s *NoopSpan) SpanContext() observabilitycontract.SpanContext {
	return observabilitycontract.SpanContext{}
}
func (s *NoopSpan) IsRecording() bool        { return false }
func (s *NoopSpan) Context() context.Context { return context.Background() }

type PrometheusTracer struct{}

func NewPrometheusTracer() *PrometheusTracer { return &PrometheusTracer{} }

func (t *PrometheusTracer) StartSpan(ctx context.Context, name string, opts ...observabilitycontract.SpanOption) (context.Context, observabilitycontract.Span) {
	// 简单的实现，实际项目中应该集成 OpenTelemetry 等真正的 tracing 系统
	return ctx, &NoopSpan{}
}

func (t *PrometheusTracer) SpanFromContext(ctx context.Context) observabilitycontract.Span {
	return &NoopSpan{}
}

func (t *PrometheusTracer) Inject(ctx context.Context, carrier observabilitycontract.TextMapCarrier) error {
	return nil
}

func (t *PrometheusTracer) Extract(ctx context.Context, carrier observabilitycontract.TextMapCarrier) (context.Context, error) {
	return ctx, nil
}
