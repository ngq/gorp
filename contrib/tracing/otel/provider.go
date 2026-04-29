package otel

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ngq/gorp/framework/contract"
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

// Provider 提供 OpenTelemetry 追踪实现。
//
// 中文说明：
// - 基于 OpenTelemetry SDK 实现；
// - 支持多种导出器：OTLP（Jaeger/Zipkin/Collector）、stdout；
// - 支持 W3C TraceContext 传播；
// - 支持采样率配置；
// - 当前已从 framework/provider 真实下沉到 contrib 层。
type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string     { return "tracing.otel" }
func (p *Provider) IsDefer() bool    { return true }
func (p *Provider) Provides() []string { return []string{contract.TracerKey, contract.TracerProviderKey} }

func (p *Provider) Register(c contract.Container) error {
	c.Bind(contract.TracerProviderKey, func(c contract.Container) (any, error) {
		cfg, err := getTracingConfig(c)
		if err != nil {
			return nil, err
		}
		return NewTracerProvider(cfg)
	}, true)

	c.Bind(contract.TracerKey, func(c contract.Container) (any, error) {
		cfg, err := getTracingConfig(c)
		if err != nil {
			return nil, err
		}
		provider, err := NewTracerProvider(cfg)
		if err != nil {
			return nil, err
		}
		return NewTracer(provider, cfg), nil
	}, true)

	return nil
}

func (p *Provider) Boot(c contract.Container) error { return nil }

func getTracingConfig(c contract.Container) (*contract.TracingConfig, error) {
	cfgAny, err := c.Make(contract.ConfigKey)
	if err != nil {
		return nil, err
	}
	cfg, ok := cfgAny.(contract.Config)
	if !ok {
		return nil, errors.New("tracing: invalid config service")
	}

	tracingCfg := &contract.TracingConfig{
		Enabled:            true,
		ExporterType:       "otlp",
		ExporterEndpoint:   "localhost:4317",
		SamplingRate:       1.0,
		Propagators:        []string{"tracecontext", "baggage"},
		BatchTimeout:       5,
		MaxQueueSize:       2048,
		MaxExportBatchSize: 512,
	}

	if name := configprovider.GetStringAny(cfg,
		"tracing.service_name",
		"tracing.service.name",
		"service.name",
	); name != "" {
		tracingCfg.ServiceName = name
	}
	if env := configprovider.GetStringAny(cfg,
		"tracing.environment",
		"service.environment",
	); env != "" {
		tracingCfg.Environment = env
	}
	if version := configprovider.GetStringAny(cfg,
		"tracing.version",
	); version != "" {
		tracingCfg.Version = version
	}
	if enabled, ok := configprovider.GetBoolAny(cfg, "tracing.enabled"); ok {
		tracingCfg.Enabled = enabled
	}
	if exporter := configprovider.GetStringAny(cfg,
		"tracing.backend",
		"tracing.type",
		"tracing.exporter_type",
	); exporter != "" {
		tracingCfg.ExporterType = exporter
	}
	if endpoint := configprovider.GetStringAny(cfg,
		"tracing.otel.endpoint",
		"tracing.exporter_endpoint",
	); endpoint != "" {
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
	cfg      *contract.TracingConfig
	mu       sync.Mutex
	closed   bool
}

func NewTracerProvider(cfg *contract.TracingConfig) (*TracerProviderWrapper, error) {
	if !cfg.Enabled {
		return &TracerProviderWrapper{provider: sdktrace.NewTracerProvider(), cfg: cfg}, nil
	}
	exporter, err := createExporter(cfg)
	if err != nil {
		return nil, fmt.Errorf("tracing: create exporter failed: %w", err)
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

func (p *TracerProviderWrapper) Tracer(name string, options ...contract.TracerOption) contract.Tracer {
	cfg := &contract.TracerConfig{}
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

type TracerWrapper struct {
	tracer trace.Tracer
	cfg    *contract.TracingConfig
}

func NewTracer(provider *TracerProviderWrapper, cfg *contract.TracingConfig) *TracerWrapper {
	return &TracerWrapper{tracer: provider.provider.Tracer(cfg.ServiceName), cfg: cfg}
}

func (t *TracerWrapper) StartSpan(ctx context.Context, name string, opts ...contract.SpanOption) (context.Context, contract.Span) {
	cfg := &contract.SpanConfig{}
	for _, opt := range opts {
		opt(cfg)
	}
	otelOpts := []trace.SpanStartOption{}
	switch cfg.Kind {
	case contract.SpanKindServer:
		otelOpts = append(otelOpts, trace.WithSpanKind(trace.SpanKindServer))
	case contract.SpanKindClient:
		otelOpts = append(otelOpts, trace.WithSpanKind(trace.SpanKindClient))
	case contract.SpanKindProducer:
		otelOpts = append(otelOpts, trace.WithSpanKind(trace.SpanKindProducer))
	case contract.SpanKindConsumer:
		otelOpts = append(otelOpts, trace.WithSpanKind(trace.SpanKindConsumer))
	case contract.SpanKindInternal:
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

func (t *TracerWrapper) SpanFromContext(ctx context.Context) contract.Span {
	otelSpan := trace.SpanFromContext(ctx)
	if !otelSpan.SpanContext().IsValid() {
		return &noopSpan{}
	}
	return &SpanWrapper{span: otelSpan}
}

func (t *TracerWrapper) Inject(ctx context.Context, carrier contract.TextMapCarrier) error {
	otel.GetTextMapPropagator().Inject(ctx, &textMapCarrierWrapper{carrier: carrier})
	return nil
}

func (t *TracerWrapper) Extract(ctx context.Context, carrier contract.TextMapCarrier) (context.Context, error) {
	ctx = otel.GetTextMapPropagator().Extract(ctx, &textMapCarrierWrapper{carrier: carrier})
	return ctx, nil
}

type SpanWrapper struct {
	span trace.Span
}

func (s *SpanWrapper) End(options ...contract.SpanEndOption) {
	cfg := &contract.SpanEndConfig{}
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
func (s *SpanWrapper) SetError(err error) { s.span.RecordError(err); s.span.SetStatus(codes.Error, err.Error()) }
func (s *SpanWrapper) SetStatus(code contract.SpanStatusCode, description string) {
	var otelCode codes.Code
	switch code {
	case contract.SpanStatusCodeOk:
		otelCode = codes.Ok
	case contract.SpanStatusCodeError:
		otelCode = codes.Error
	default:
		otelCode = codes.Unset
	}
	s.span.SetStatus(otelCode, description)
}
func (s *SpanWrapper) SpanContext() contract.SpanContext {
	sc := s.span.SpanContext()
	return contract.SpanContext{TraceID: sc.TraceID().String(), SpanID: sc.SpanID().String(), TraceFlags: contract.TraceFlags(sc.TraceFlags()), Remote: sc.IsRemote()}
}
func (s *SpanWrapper) IsRecording() bool { return s.span.IsRecording() }
func (s *SpanWrapper) Context() context.Context { return trace.ContextWithSpan(context.Background(), s.span) }

type noopSpan struct{}

func (s *noopSpan) End(options ...contract.SpanEndOption) {}
func (s *noopSpan) AddEvent(name string, attributes map[string]interface{}) {}
func (s *noopSpan) SetTag(key string, value interface{}) {}
func (s *noopSpan) SetAttributes(attributes map[string]interface{}) {}
func (s *noopSpan) SetError(err error) {}
func (s *noopSpan) SetStatus(code contract.SpanStatusCode, description string) {}
func (s *noopSpan) SpanContext() contract.SpanContext { return contract.SpanContext{} }
func (s *noopSpan) IsRecording() bool { return false }
func (s *noopSpan) Context() context.Context { return context.Background() }

type textMapCarrierWrapper struct { carrier contract.TextMapCarrier }
func (w *textMapCarrierWrapper) Get(key string) string { return w.carrier.Get(key) }
func (w *textMapCarrierWrapper) Set(key string, value string) { w.carrier.Set(key, value) }
func (w *textMapCarrierWrapper) Keys() []string { return w.carrier.Keys() }

func createExporter(cfg *contract.TracingConfig) (sdktrace.SpanExporter, error) {
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

func createResource(cfg *contract.TracingConfig) *resource.Resource {
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
