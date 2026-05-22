// Package otel provides OpenTelemetry tracing implementation for gorp.
//
// OpenTelemetry 链路追踪 Provider，实现 observabilitycontract.Tracer 契约。
// 支持 OTLP gRPC/HTTP 导出、Stdout 导出、Span 创建和属性设置。
//
// 使用示例：
//
//	cfg := &TracingConfig{
//	    ServiceName: "my-service",
//	    Exporter:    "otlp-grpc",
//	    Endpoint:    "localhost:4317",
//	}
//	tracer, err := NewTracer(cfg)
//	if err != nil {
//	    panic(err)
//	}
//	defer tracer.Close()
//
//	ctx, span := tracer.Start(ctx, "operation-name")
//	defer span.End()
//
// 配置路径：tracing.*
package otel

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	configprovider "github.com/ngq/gorp/framework/provider/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string  { return "tracing.otel" }
func (p *Provider) IsDefer() bool { return true }
func (p *Provider) Provides() []string {
	return []string{observabilitycontract.TracerKey, observabilitycontract.TracerProviderKey}
}

// DependsOn returns the keys this provider depends on.
// OTel tracing depends on Config for tracing configuration.
//
// DependsOn 返回该 provider 依赖的 key。
// OTel tracing 依赖 Config 获取追踪配置。
func (p *Provider) DependsOn() []string { return []string{datacontract.ConfigKey} }

func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(observabilitycontract.TracerProviderKey, func(c runtimecontract.Container) (any, error) {
		cfg, err := getTracingConfig(c)
		if err != nil {
			return nil, err
		}
		tp, err := NewTracerProvider(cfg)
		if err != nil {
			return nil, err
		}
		c.RegisterCloser(observabilitycontract.TracerProviderKey, &tracerProviderCloser{provider: tp})
		return tp, nil
	}, true)

	c.Bind(observabilitycontract.TracerKey, func(c runtimecontract.Container) (any, error) {
		tpAny, err := c.Make(observabilitycontract.TracerProviderKey)
		if err != nil {
			return nil, err
		}
		tp, ok := tpAny.(*TracerProviderWrapper)
		if !ok {
			return nil, fmt.Errorf("tracing.otel: expected *TracerProviderWrapper, got %T", tpAny)
		}
		return NewTracer(tp, tp.cfg), nil
	}, true)

	return nil
}

func (p *Provider) Boot(c runtimecontract.Container) error { return nil }

func getTracingConfig(c runtimecontract.Container) (*observabilitycontract.TracingConfig, error) {
	cfgAny, err := c.Make(datacontract.ConfigKey)
	if err != nil {
		return nil, err
	}
	cfg, ok := cfgAny.(datacontract.Config)
	if !ok {
		return nil, errors.New("tracing.otel: invalid config service")
	}

	tracingCfg := &observabilitycontract.TracingConfig{
		Enabled:            true,
		ExporterType:       "otlp",
		ExporterEndpoint:   "localhost:4317",
		SamplingRate:       1.0,
		Propagators:        []string{"tracecontext", "baggage"},
		BatchTimeout:       5,
		MaxQueueSize:       2048,
		MaxExportBatchSize: 512,
	}
	if name := configprovider.GetStringAny(cfg, "tracing.service_name", "tracing.service.name", "service.name"); name != "" {
		tracingCfg.ServiceName = name
	}
	if env := configprovider.GetStringAny(cfg, "tracing.environment", "service.environment"); env != "" {
		tracingCfg.Environment = env
	}
	if version := configprovider.GetStringAny(cfg, "tracing.version"); version != "" {
		tracingCfg.Version = version
	}
	if enabled, ok := configprovider.GetBoolAny(cfg, "tracing.enabled"); ok {
		tracingCfg.Enabled = enabled
	}
	if exporter := configprovider.GetStringAny(cfg, "tracing.backend", "tracing.type", "tracing.exporter_type"); exporter != "" {
		tracingCfg.ExporterType = exporter
	}
	if endpoint := configprovider.GetStringAny(cfg, "tracing.otel.endpoint", "tracing.exporter_endpoint"); endpoint != "" {
		tracingCfg.ExporterEndpoint = endpoint
	}
	if rate := configprovider.GetFloatAny(cfg, "tracing.sampling_rate"); rate > 0 {
		tracingCfg.SamplingRate = rate
	}
	if timeout := configprovider.GetIntAny(cfg, "tracing.batch_timeout"); timeout > 0 {
		tracingCfg.BatchTimeout = timeout
	}
	if size := configprovider.GetIntAny(cfg, "tracing.max_queue_size"); size > 0 {
		tracingCfg.MaxQueueSize = size
	}
	if size := configprovider.GetIntAny(cfg, "tracing.max_export_batch_size"); size > 0 {
		tracingCfg.MaxExportBatchSize = size
	}
	return tracingCfg, nil
}

type TracerProviderWrapper struct {
	provider *sdktrace.TracerProvider
	cfg      *observabilitycontract.TracingConfig
	mu       sync.Mutex
	closed   bool
}

func NewTracerProvider(cfg *observabilitycontract.TracingConfig) (*TracerProviderWrapper, error) {
	if !cfg.Enabled {
		return &TracerProviderWrapper{provider: sdktrace.NewTracerProvider(), cfg: cfg}, nil
	}
	exporter, err := createExporter(cfg)
	if err != nil {
		return nil, fmt.Errorf("tracing.otel: create exporter failed: %w", err)
	}
	res := createResource(cfg)
	sampler := createSampler(cfg.SamplingRate)
	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter,
			sdktrace.WithBatchTimeout(time.Duration(cfg.BatchTimeout)*time.Second),
			sdktrace.WithMaxQueueSize(cfg.MaxQueueSize),
			sdktrace.WithMaxExportBatchSize(cfg.MaxExportBatchSize),
		),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	)
	otel.SetTracerProvider(provider)
	setPropagators(cfg.Propagators)
	return &TracerProviderWrapper{provider: provider, cfg: cfg}, nil
}

func (p *TracerProviderWrapper) Tracer(name string, options ...observabilitycontract.TracerOption) observabilitycontract.Tracer {
	cfg := &observabilitycontract.TracerConfig{}
	for _, opt := range options {
		opt(cfg)
	}
	opts := []trace.TracerOption{}
	if cfg.SchemaURL != "" {
		opts = append(opts, trace.WithSchemaURL(cfg.SchemaURL))
	}
	tracer := p.provider.Tracer(name, opts...)
	return &TracerWrapper{tracer: tracer, cfg: p.cfg}
}

// tracerProviderCloser wraps TracerProviderWrapper.Shutdown as io.Closer for Destroy lifecycle.
type tracerProviderCloser struct {
	provider *TracerProviderWrapper
}

func (c *tracerProviderCloser) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return c.provider.Shutdown(ctx)
}

func (p *TracerProviderWrapper) Shutdown(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return nil
	}
	p.closed = true
	return p.provider.Shutdown(ctx)
}

func (p *TracerProviderWrapper) ForceFlush(ctx context.Context) error {
	return p.provider.ForceFlush(ctx)
}

func (p *TracerProviderWrapper) Underlying() any {
	return p.provider
}

func (p *TracerProviderWrapper) As(target any) bool {
	return As(p.provider, target)
}

type TracerWrapper struct {
	tracer trace.Tracer
	cfg    *observabilitycontract.TracingConfig
}

func NewTracer(provider *TracerProviderWrapper, cfg *observabilitycontract.TracingConfig) *TracerWrapper {
	return &TracerWrapper{tracer: provider.provider.Tracer(cfg.ServiceName), cfg: cfg}
}

func (t *TracerWrapper) StartSpan(ctx context.Context, name string, opts ...observabilitycontract.SpanOption) (context.Context, observabilitycontract.Span) {
	cfg := &observabilitycontract.SpanConfig{}
	for _, opt := range opts {
		opt(cfg)
	}
	otelOpts := []trace.SpanStartOption{}
	switch cfg.Kind {
	case observabilitycontract.SpanKindServer:
		otelOpts = append(otelOpts, trace.WithSpanKind(trace.SpanKindServer))
	case observabilitycontract.SpanKindClient:
		otelOpts = append(otelOpts, trace.WithSpanKind(trace.SpanKindClient))
	case observabilitycontract.SpanKindProducer:
		otelOpts = append(otelOpts, trace.WithSpanKind(trace.SpanKindProducer))
	case observabilitycontract.SpanKindConsumer:
		otelOpts = append(otelOpts, trace.WithSpanKind(trace.SpanKindConsumer))
	case observabilitycontract.SpanKindInternal:
		otelOpts = append(otelOpts, trace.WithSpanKind(trace.SpanKindInternal))
	}
	if len(cfg.Attributes) > 0 {
		attrs := make([]attribute.KeyValue, 0, len(cfg.Attributes))
		for k, v := range cfg.Attributes {
			attrs = append(attrs, toAttribute(k, v))
		}
		otelOpts = append(otelOpts, trace.WithAttributes(attrs...))
	}
	if !cfg.StartTime.IsZero() {
		otelOpts = append(otelOpts, trace.WithTimestamp(cfg.StartTime))
	}
	ctx, otelSpan := t.tracer.Start(ctx, name, otelOpts...)
	return ctx, &SpanWrapper{span: otelSpan}
}

func (t *TracerWrapper) SpanFromContext(ctx context.Context) observabilitycontract.Span {
	otelSpan := trace.SpanFromContext(ctx)
	if !otelSpan.SpanContext().IsValid() {
		return &noopSpan{}
	}
	return &SpanWrapper{span: otelSpan}
}

func (t *TracerWrapper) Inject(ctx context.Context, carrier observabilitycontract.TextMapCarrier) error {
	otel.GetTextMapPropagator().Inject(ctx, &textMapCarrierWrapper{carrier: carrier})
	return nil
}

func (t *TracerWrapper) Extract(ctx context.Context, carrier observabilitycontract.TextMapCarrier) (context.Context, error) {
	ctx = otel.GetTextMapPropagator().Extract(ctx, &textMapCarrierWrapper{carrier: carrier})
	return ctx, nil
}

func (t *TracerWrapper) Underlying() any {
	return t.tracer
}

func (t *TracerWrapper) As(target any) bool {
	return As(t.tracer, target)
}

type SpanWrapper struct {
	span trace.Span
}

func (s *SpanWrapper) End(options ...observabilitycontract.SpanEndOption) {
	cfg := &observabilitycontract.SpanEndConfig{}
	for _, opt := range options {
		opt(cfg)
	}
	if cfg.EndTime.IsZero() {
		s.span.End()
	} else {
		s.span.End(trace.WithTimestamp(cfg.EndTime))
	}
}

func (s *SpanWrapper) AddEvent(name string, attributes map[string]any) {
	if len(attributes) > 0 {
		attrs := make([]attribute.KeyValue, 0, len(attributes))
		for k, v := range attributes {
			attrs = append(attrs, toAttribute(k, v))
		}
		s.span.AddEvent(name, trace.WithAttributes(attrs...))
	} else {
		s.span.AddEvent(name)
	}
}

func (s *SpanWrapper) SetTag(key string, value any) { s.span.SetAttributes(toAttribute(key, value)) }
func (s *SpanWrapper) SetAttributes(attributes map[string]any) {
	attrs := make([]attribute.KeyValue, 0, len(attributes))
	for k, v := range attributes {
		attrs = append(attrs, toAttribute(k, v))
	}
	s.span.SetAttributes(attrs...)
}
func (s *SpanWrapper) SetError(err error) {
	s.span.RecordError(err)
	s.span.SetStatus(codes.Error, err.Error())
}
func (s *SpanWrapper) SetStatus(code observabilitycontract.SpanStatusCode, description string) {
	var otelCode codes.Code
	switch code {
	case observabilitycontract.SpanStatusCodeOk:
		otelCode = codes.Ok
	case observabilitycontract.SpanStatusCodeError:
		otelCode = codes.Error
	default:
		otelCode = codes.Unset
	}
	s.span.SetStatus(otelCode, description)
}
func (s *SpanWrapper) SpanContext() observabilitycontract.SpanContext {
	sc := s.span.SpanContext()
	return observabilitycontract.SpanContext{
		TraceID:    sc.TraceID().String(),
		SpanID:     sc.SpanID().String(),
		TraceFlags: observabilitycontract.TraceFlags(sc.TraceFlags()),
		Remote:     sc.IsRemote(),
	}
}
func (s *SpanWrapper) IsRecording() bool { return s.span.IsRecording() }
func (s *SpanWrapper) Context() context.Context {
	return trace.ContextWithSpan(context.Background(), s.span)
}

func (s *SpanWrapper) Underlying() any {
	return s.span
}

func (s *SpanWrapper) As(target any) bool {
	return As(s.span, target)
}

type noopSpan struct{}

func (s *noopSpan) End(options ...observabilitycontract.SpanEndOption)      {}
func (s *noopSpan) AddEvent(name string, attributes map[string]interface{}) {}
func (s *noopSpan) SetTag(key string, value interface{})                    {}
func (s *noopSpan) SetAttributes(attributes map[string]interface{})         {}
func (s *noopSpan) SetError(err error)                                      {}
func (s *noopSpan) SetStatus(code observabilitycontract.SpanStatusCode, description string) {
}
func (s *noopSpan) SpanContext() observabilitycontract.SpanContext {
	return observabilitycontract.SpanContext{}
}
func (s *noopSpan) IsRecording() bool        { return false }
func (s *noopSpan) Context() context.Context { return context.Background() }

type textMapCarrierWrapper struct {
	carrier observabilitycontract.TextMapCarrier
}

func (w *textMapCarrierWrapper) Get(key string) string        { return w.carrier.Get(key) }
func (w *textMapCarrierWrapper) Set(key string, value string) { w.carrier.Set(key, value) }
func (w *textMapCarrierWrapper) Keys() []string               { return w.carrier.Keys() }

func createExporter(cfg *observabilitycontract.TracingConfig) (sdktrace.SpanExporter, error) {
	switch cfg.ExporterType {
	case "otlp", "grpc":
		return otlptracegrpc.New(context.Background(), otlptracegrpc.WithEndpoint(cfg.ExporterEndpoint))
	case "otlphttp", "http":
		return otlptracehttp.New(context.Background(), otlptracehttp.WithEndpoint(cfg.ExporterEndpoint))
	case "stdout":
		return stdouttrace.New()
	default:
		return otlptracegrpc.New(context.Background(), otlptracegrpc.WithEndpoint(cfg.ExporterEndpoint))
	}
}

func createResource(cfg *observabilitycontract.TracingConfig) *resource.Resource {
	attrs := []attribute.KeyValue{semconv.ServiceName(cfg.ServiceName)}
	if cfg.Environment != "" {
		attrs = append(attrs, semconv.DeploymentEnvironment(cfg.Environment))
	}
	if cfg.Version != "" {
		attrs = append(attrs, semconv.ServiceVersion(cfg.Version))
	}
	for k, v := range cfg.ResourceAttributes {
		attrs = append(attrs, attribute.String(k, v))
	}
	return resource.NewWithAttributes(semconv.SchemaURL, attrs...)
}

func createSampler(rate float64) sdktrace.Sampler {
	if rate >= 1.0 {
		return sdktrace.AlwaysSample()
	}
	if rate <= 0 {
		return sdktrace.NeverSample()
	}
	return sdktrace.TraceIDRatioBased(rate)
}

func setPropagators(propagators []string) {
	var props []propagation.TextMapPropagator
	for _, p := range propagators {
		switch p {
		case "tracecontext":
			props = append(props, propagation.TraceContext{})
		case "baggage":
			props = append(props, propagation.Baggage{})
		case "b3":
		}
	}
	if len(props) > 0 {
		otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(props...))
	}
}

func toAttribute(key string, value any) attribute.KeyValue {
	switch v := value.(type) {
	case string:
		return attribute.String(key, v)
	case int:
		return attribute.Int(key, v)
	case int64:
		return attribute.Int64(key, v)
	case float64:
		return attribute.Float64(key, v)
	case bool:
		return attribute.Bool(key, v)
	default:
		return attribute.String(key, fmt.Sprintf("%v", v))
	}
}
