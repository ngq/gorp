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
	"sort"
	"sync"
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

// collector 缓存：同一 name 重复调用 promauto.New*Vec 会向默认注册表重复
// 注册并 panic，必须按 name 复用已创建的 collector。
var (
	counterVecs   sync.Map // name -> *prometheus.CounterVec
	gaugeVecs     sync.Map // name -> *prometheus.GaugeVec
	histogramVecs sync.Map // name -> *prometheus.HistogramVec
)

func (m *PrometheusMetrics) Counter(name string, labels map[string]string, delta float64) {
	keys, values := labelKeysAndValues(labels)
	counter, ok := counterVecs.Load(name)
	if !ok {
		counter = promauto.NewCounterVec(prometheus.CounterOpts{
			Name: name,
			Help: name,
		}, keys)
		counter, _ = counterVecs.LoadOrStore(name, counter)
	}
	counter.(*prometheus.CounterVec).WithLabelValues(values...).Add(delta)
}

func (m *PrometheusMetrics) Gauge(name string, labels map[string]string, value float64) {
	keys, values := labelKeysAndValues(labels)
	gauge, ok := gaugeVecs.Load(name)
	if !ok {
		gauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: name,
			Help: name,
		}, keys)
		gauge, _ = gaugeVecs.LoadOrStore(name, gauge)
	}
	gauge.(*prometheus.GaugeVec).WithLabelValues(values...).Set(value)
}

func (m *PrometheusMetrics) Histogram(name string, labels map[string]string, value float64) {
	keys, values := labelKeysAndValues(labels)
	histogram, ok := histogramVecs.Load(name)
	if !ok {
		histogram = promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    name,
			Help:    name,
			Buckets: prometheus.DefBuckets,
		}, keys)
		histogram, _ = histogramVecs.LoadOrStore(name, histogram)
	}
	histogram.(*prometheus.HistogramVec).WithLabelValues(values...).Observe(value)
}

func (m *PrometheusMetrics) Timing(name string, labels map[string]string, duration time.Duration) {
	m.Histogram(name+"_seconds", labels, duration.Seconds())
}

// labelKeysAndValues 在单次确定序（按 key 排序）中同时产出 keys 与 values。
// 旧实现 labelKeys/labelValues 各自遍历 map，Go map 顺序随机，
// 两次结果不一致时标签值会张冠李戴（如 path=200, status="/a"）。
func labelKeysAndValues(labels map[string]string) ([]string, []string) {
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	values := make([]string, len(keys))
	for i, k := range keys {
		values[i] = labels[k]
	}
	return keys, values
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
