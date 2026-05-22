package observability

import (
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	"time"
)

const ObservabilityKey = "framework.observability"

type Observability interface {
	Metrics() Metrics
	Tracer() Tracer
	Logger() Logger
	ErrorReporter() resiliencecontract.ErrorReporter
}

type ObservabilityConfig struct {
	MetricsEnabled        bool
	TracingEnabled        bool
	ErrorReportingEnabled bool
	ServiceName           string
	Environment           string
	Version               string
	SamplingRate          float64
}

type RequestContext struct {
	TraceID   string
	RequestID string
	Span      Span
	StartTime time.Time
	Labels    map[string]string
}
