package observability

import (
	"context"
	"time"
)

const (
	TracerKey         = "framework.tracer"
	TracerProviderKey = "framework.tracer.provider"
)

type TracerProvider interface {
	Tracer(name string, options ...TracerOption) Tracer
	Shutdown(ctx context.Context) error
	ForceFlush(ctx context.Context) error
}

type TracerOption func(*TracerConfig)

type TracerConfig struct {
	SchemaURL  string
	Version    string
	Attributes map[string]interface{}
}

type Tracer interface {
	StartSpan(ctx context.Context, name string, opts ...SpanOption) (context.Context, Span)
	SpanFromContext(ctx context.Context) Span
	Inject(ctx context.Context, carrier TextMapCarrier) error
	Extract(ctx context.Context, carrier TextMapCarrier) (context.Context, error)
}

type SpanOption func(*SpanConfig)

type SpanConfig struct {
	Kind       SpanKind
	Attributes map[string]interface{}
	StartTime  time.Time
	Links      []SpanLink
}

type SpanKind int

const (
	SpanKindUnspecified SpanKind = iota
	SpanKindInternal
	SpanKindServer
	SpanKindClient
	SpanKindProducer
	SpanKindConsumer
)

type SpanLink struct {
	SpanContext SpanContext
	Attributes  map[string]interface{}
}

type SpanContext struct {
	TraceID    string
	SpanID     string
	TraceFlags TraceFlags
	Remote     bool
}

type TraceFlags byte

const (
	TraceFlagsSampled TraceFlags = 0x01
)

type TextMapCarrier interface {
	Get(key string) string
	Set(key string, value string)
	Keys() []string
}

type Span interface {
	End(options ...SpanEndOption)
	AddEvent(name string, attributes map[string]interface{})
	SetTag(key string, value interface{})
	SetAttributes(attributes map[string]interface{})
	SetError(err error)
	SetStatus(code SpanStatusCode, description string)
	SpanContext() SpanContext
	IsRecording() bool
	Context() context.Context
}

type SpanEndOption func(*SpanEndConfig)

type SpanEndConfig struct {
	EndTime time.Time
}

type SpanStatusCode int

const (
	SpanStatusCodeUnset SpanStatusCode = iota
	SpanStatusCodeOk
	SpanStatusCodeError
)

type TracingConfig struct {
	Enabled            bool
	ServiceName        string
	Environment        string
	Version            string
	ExporterType       string
	ExporterEndpoint   string
	SamplingRate       float64
	Propagators        []string
	ResourceAttributes map[string]string
	BatchTimeout       int
	MaxQueueSize       int
	MaxExportBatchSize int
}
