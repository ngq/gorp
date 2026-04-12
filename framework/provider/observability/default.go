package observability

import (
	"context"
	"time"

	"github.com/ngq/gorp/framework/contract"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// DefaultObservability 默认观测实现。
//
// 中文说明：
// - 整合 Metrics、Tracing、Logging、ErrorReporting；
// - Metrics 使用 Prometheus；
// - Tracing 当前为占位实现，可后续对接 OpenTelemetry；
// - 提供统一的观测入口。
type DefaultObservability struct {
	metrics       contract.Metrics
	tracer        contract.Tracer
	logger        contract.Logger
	errorReporter contract.ErrorReporter
}

// NewDefaultObservability 创建默认观测实现。
func NewDefaultObservability(
	metrics contract.Metrics,
	tracer contract.Tracer,
	logger contract.Logger,
	errorReporter contract.ErrorReporter,
) *DefaultObservability {
	return &DefaultObservability{
		metrics:       metrics,
		tracer:        tracer,
		logger:        logger,
		errorReporter: errorReporter,
	}
}

func (o *DefaultObservability) Metrics() contract.Metrics       { return o.metrics }
func (o *DefaultObservability) Tracer() contract.Tracer         { return o.tracer }
func (o *DefaultObservability) Logger() contract.Logger         { return o.logger }
func (o *DefaultObservability) ErrorReporter() contract.ErrorReporter {
	return o.errorReporter
}

// PrometheusMetrics Prometheus 指标实现。
type PrometheusMetrics struct{}

// NewPrometheusMetrics 创建 Prometheus 指标实现。
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

// NoopTracer 空追踪器实现。
//
// 中文说明：
// - 当未启用链路追踪时使用；
// - 所有操作都是空操作。
type NoopTracer struct{}

// NewNoopTracer 创建空追踪器。
func NewNoopTracer() *NoopTracer { return &NoopTracer{} }

func (t *NoopTracer) StartSpan(ctx context.Context, name string, opts ...contract.SpanOption) (context.Context, contract.Span) {
	return ctx, &NoopSpan{}
}

func (t *NoopTracer) SpanFromContext(ctx context.Context) contract.Span {
	return &NoopSpan{}
}

func (t *NoopTracer) Inject(ctx context.Context, carrier contract.TextMapCarrier) error {
	return nil
}

func (t *NoopTracer) Extract(ctx context.Context, carrier contract.TextMapCarrier) (context.Context, error) {
	return ctx, nil
}

// NoopSpan 空 Span 实现。
//
// 中文说明：
// - 所有方法空操作；
// - IsRecording 返回 false，表示不记录；
// - 用于避免不必要的属性计算开销。
type NoopSpan struct{}

func (s *NoopSpan) End(options ...contract.SpanEndOption) {}
func (s *NoopSpan) AddEvent(name string, attributes map[string]interface{}) {}
func (s *NoopSpan) SetTag(key string, value interface{}) {}
func (s *NoopSpan) SetAttributes(attributes map[string]interface{}) {}
func (s *NoopSpan) SetError(err error) {}
func (s *NoopSpan) SetStatus(code contract.SpanStatusCode, description string) {}
func (s *NoopSpan) SpanContext() contract.SpanContext { return contract.SpanContext{} }
func (s *NoopSpan) IsRecording() bool { return false }
func (s *NoopSpan) Context() context.Context { return context.Background() }