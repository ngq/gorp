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
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
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

type PrometheusTracer struct {
	mu    sync.Mutex
	spans []*RecordedSpan // 环形缓冲，最多 maxRecordedSpans 条
}

// RecordedSpan 是本地记录的已完成 span。
type RecordedSpan struct {
	TraceID      string
	SpanID       string
	ParentSpanID string
	Name         string
	Kind         observabilitycontract.SpanKind
	StartTime    time.Time
	EndTime      time.Time
	Duration     time.Duration
	Status       observabilitycontract.SpanStatusCode
	Attributes   map[string]interface{}
	Error        string
}

// maxRecordedSpans 是有界环形缓冲的上限，防止无界增长。
const maxRecordedSpans = 1000

// NewPrometheusTracer 创建无外部依赖的本地 span 记录 tracer。
func NewPrometheusTracer() *PrometheusTracer { return &PrometheusTracer{} }

// RecordedSpans 返回已记录 span 的快照（新到旧）。
func (t *PrometheusTracer) RecordedSpans() []*RecordedSpan {
	t.mu.Lock()
	defer t.mu.Unlock()
	out := make([]*RecordedSpan, 0, len(t.spans))
	for i := len(t.spans) - 1; i >= 0; i-- {
		out = append(out, cloneRecordedSpan(t.spans[i]))
	}
	return out
}

func (t *PrometheusTracer) record(span *prometheusSpan) {
	rec := &RecordedSpan{
		TraceID:      span.spanCtx.TraceID,
		SpanID:       span.spanCtx.SpanID,
		ParentSpanID: span.parentSpanID,
		Name:         span.name,
		Kind:         span.kind,
		StartTime:    span.startTime,
		EndTime:      span.endTime,
		Duration:     span.duration(),
		Status:       span.status,
		Attributes:   span.attrs,
		Error:        span.errMsg,
	}
	t.mu.Lock()
	t.spans = append(t.spans, rec)
	if len(t.spans) > maxRecordedSpans {
		t.spans = t.spans[len(t.spans)-maxRecordedSpans:]
	}
	t.mu.Unlock()
}

func cloneRecordedSpan(s *RecordedSpan) *RecordedSpan {
	c := *s
	if s.Attributes != nil {
		c.Attributes = make(map[string]interface{}, len(s.Attributes))
		for k, v := range s.Attributes {
			c.Attributes[k] = v
		}
	}
	return &c
}

// StartSpan 生成真实 trace/span ID，从 context 继承父 span（如果存在），
// 并在 End 时记录进本地缓冲。返回的 context 携带当前 span。
func (t *PrometheusTracer) StartSpan(ctx context.Context, name string, opts ...observabilitycontract.SpanOption) (context.Context, observabilitycontract.Span) {
	cfg := &observabilitycontract.SpanConfig{Kind: observabilitycontract.SpanKindInternal, StartTime: time.Now()}
	for _, opt := range opts {
		opt(cfg)
	}
	span := &prometheusSpan{
		tracer:    t,
		name:      name,
		kind:      cfg.Kind,
		startTime: cfg.StartTime,
		attrs:     make(map[string]interface{}, len(cfg.Attributes)),
		status:    observabilitycontract.SpanStatusCodeUnset,
		spanCtx: observabilitycontract.SpanContext{
			TraceID:    newTraceID(),
			SpanID:     newSpanID(),
			TraceFlags: observabilitycontract.TraceFlagsSampled,
		},
	}
	for k, v := range cfg.Attributes {
		span.attrs[k] = v
	}
	if parent := spanFromContext(ctx); parent != nil {
		span.spanCtx.TraceID = parent.SpanContext().TraceID
		span.parentSpanID = parent.SpanContext().SpanID
	}
	span.ctx = withSpan(ctx, span)
	return span.ctx, span
}

func (t *PrometheusTracer) SpanFromContext(ctx context.Context) observabilitycontract.Span {
	if s := spanFromContext(ctx); s != nil {
		return s
	}
	return &prometheusSpan{attrs: map[string]interface{}{}, status: observabilitycontract.SpanStatusCodeUnset}
}

// Inject 把当前 span 上下文写入 carrier（W3C traceparent 格式）。
func (t *PrometheusTracer) Inject(ctx context.Context, carrier observabilitycontract.TextMapCarrier) error {
	if carrier == nil {
		return nil
	}
	span := spanFromContext(ctx)
	if span == nil {
		return nil
	}
	sc := span.SpanContext()
	if sc.TraceID == "" || sc.SpanID == "" {
		return nil
	}
	flags := "00"
	if sc.TraceFlags&observabilitycontract.TraceFlagsSampled != 0 {
		flags = "01"
	}
	carrier.Set(traceParentHeader, fmt.Sprintf("00-%s-%s-%s", sc.TraceID, sc.SpanID, flags))
	return nil
}

// Extract 从 carrier 解析 W3C traceparent，返回携带远端父 span 的 context。
func (t *PrometheusTracer) Extract(ctx context.Context, carrier observabilitycontract.TextMapCarrier) (context.Context, error) {
	if carrier == nil {
		return ctx, nil
	}
	traceID, spanID, ok := parseTraceParent(carrier.Get(traceParentHeader))
	if !ok {
		return ctx, nil
	}
	remote := &remoteSpan{
		sc: observabilitycontract.SpanContext{
			TraceID:    traceID,
			SpanID:     spanID,
			TraceFlags: observabilitycontract.TraceFlagsSampled,
			Remote:     true,
		},
	}
	return withSpan(ctx, remote), nil
}

// traceParentHeader 是 W3C traceparent 请求头名。
const traceParentHeader = "traceparent"

// parseTraceParent 解析 "00-{traceID}-{spanID}-{flags}"。
func parseTraceParent(tp string) (traceID, spanID string, ok bool) {
	parts := strings.Split(strings.TrimSpace(tp), "-")
	if len(parts) != 4 || parts[0] != "00" {
		return "", "", false
	}
	if len(parts[1]) != 32 || len(parts[2]) != 16 {
		return "", "", false
	}
	return parts[1], parts[2], true
}

// prometheusSpan 实现 observabilitycontract.Span。
type prometheusSpan struct {
	tracer       *PrometheusTracer
	spanCtx      observabilitycontract.SpanContext
	parentSpanID string
	name         string
	kind         observabilitycontract.SpanKind
	startTime    time.Time
	endTime      time.Time
	status       observabilitycontract.SpanStatusCode
	attrs        map[string]interface{}
	errMsg       string
	ctx          context.Context
	ended        bool
}

func (s *prometheusSpan) End(options ...observabilitycontract.SpanEndOption) {
	if s == nil || s.tracer == nil || s.ended {
		return
	}
	s.ended = true
	s.endTime = time.Now()
	for _, opt := range options {
		cfg := &observabilitycontract.SpanEndConfig{}
		opt(cfg)
		if !cfg.EndTime.IsZero() {
			s.endTime = cfg.EndTime
		}
	}
	s.tracer.record(s)
}

func (s *prometheusSpan) duration() time.Duration {
	if s.endTime.IsZero() {
		return time.Since(s.startTime)
	}
	return s.endTime.Sub(s.startTime)
}

func (s *prometheusSpan) AddEvent(name string, attributes map[string]interface{}) {}

func (s *prometheusSpan) SetTag(key string, value interface{}) {
	if s == nil {
		return
	}
	if s.attrs == nil {
		s.attrs = make(map[string]interface{})
	}
	s.attrs[key] = value
}

func (s *prometheusSpan) SetAttributes(attributes map[string]interface{}) {
	if s == nil {
		return
	}
	if s.attrs == nil {
		s.attrs = make(map[string]interface{}, len(attributes))
	}
	for k, v := range attributes {
		s.attrs[k] = v
	}
}

func (s *prometheusSpan) SetError(err error) {
	if s == nil || err == nil {
		return
	}
	s.errMsg = err.Error()
	s.status = observabilitycontract.SpanStatusCodeError
}

func (s *prometheusSpan) SetStatus(code observabilitycontract.SpanStatusCode, description string) {
	if s == nil {
		return
	}
	s.status = code
	if code == observabilitycontract.SpanStatusCodeError && description != "" {
		s.errMsg = description
	}
}

func (s *prometheusSpan) SpanContext() observabilitycontract.SpanContext { return s.spanCtx }
func (s *prometheusSpan) IsRecording() bool                              { return s.tracer != nil }
func (s *prometheusSpan) Context() context.Context {
	if s.ctx != nil {
		return s.ctx
	}
	return context.Background()
}

// remoteSpan 是 Extract 出来的远端父 span（只携带上下文，不记录）。
type remoteSpan struct {
	sc observabilitycontract.SpanContext
}

func (s *remoteSpan) End(options ...observabilitycontract.SpanEndOption) {}
func (s *remoteSpan) AddEvent(name string, attributes map[string]interface{}) {
}
func (s *remoteSpan) SetTag(key string, value interface{}) {}
func (s *remoteSpan) SetAttributes(attributes map[string]interface{}) {
}
func (s *remoteSpan) SetError(err error) {}
func (s *remoteSpan) SetStatus(code observabilitycontract.SpanStatusCode, description string) {
}
func (s *remoteSpan) SpanContext() observabilitycontract.SpanContext { return s.sc }
func (s *remoteSpan) IsRecording() bool                              { return false }
func (s *remoteSpan) Context() context.Context                       { return context.Background() }

// spanContextKey 是 context 中当前 span 的类型化 key。
type spanContextKey struct{}

func withSpan(ctx context.Context, span observabilitycontract.Span) context.Context {
	return context.WithValue(ctx, spanContextKey{}, span)
}

func spanFromContext(ctx context.Context) observabilitycontract.Span {
	if ctx == nil {
		return nil
	}
	span, _ := ctx.Value(spanContextKey{}).(observabilitycontract.Span)
	return span
}

// newTraceID 生成 32 位十六进制 traceID。
func newTraceID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return hex.EncodeToString([]byte(time.Now().Format("20060102150405.000000000")))
	}
	return hex.EncodeToString(b[:])
}

// newSpanID 生成 16 位十六进制 spanID。
func newSpanID() string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return hex.EncodeToString([]byte(time.Now().Format("20060102150405")))
	}
	return hex.EncodeToString(b[:])
}
