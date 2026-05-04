package noop

import (
	"context"

	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string { return "tracing.noop" }

func (p *Provider) IsDefer() bool { return true }

func (p *Provider) Provides() []string {
	return []string{observabilitycontract.TracerKey, observabilitycontract.TracerProviderKey}
}

func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(observabilitycontract.TracerKey, func(c runtimecontract.Container) (any, error) {
		return &noopTracer{}, nil
	}, true)
	c.Bind(observabilitycontract.TracerProviderKey, func(c runtimecontract.Container) (any, error) {
		return &noopTracerProvider{}, nil
	}, true)
	return nil
}

func (p *Provider) Boot(runtimecontract.Container) error { return nil }

type noopTracer struct{}

func (t *noopTracer) StartSpan(ctx context.Context, _ string, _ ...observabilitycontract.SpanOption) (context.Context, observabilitycontract.Span) {
	return ctx, &noopSpan{}
}

func (t *noopTracer) SpanFromContext(_ context.Context) observabilitycontract.Span {
	return &noopSpan{}
}

func (t *noopTracer) Inject(_ context.Context, _ observabilitycontract.TextMapCarrier) error {
	return nil
}

func (t *noopTracer) Extract(ctx context.Context, _ observabilitycontract.TextMapCarrier) (context.Context, error) {
	return ctx, nil
}

type noopTracerProvider struct{}

func (p *noopTracerProvider) Tracer(_ string, _ ...observabilitycontract.TracerOption) observabilitycontract.Tracer {
	return &noopTracer{}
}

func (p *noopTracerProvider) Shutdown(_ context.Context) error { return nil }

func (p *noopTracerProvider) ForceFlush(_ context.Context) error { return nil }

type noopSpan struct{}

func (s *noopSpan) End(_ ...observabilitycontract.SpanEndOption)               {}
func (s *noopSpan) AddEvent(_ string, _ map[string]interface{})                {}
func (s *noopSpan) SetTag(_ string, _ interface{})                             {}
func (s *noopSpan) SetAttributes(_ map[string]interface{})                     {}
func (s *noopSpan) SetError(_ error)                                           {}
func (s *noopSpan) SetStatus(_ observabilitycontract.SpanStatusCode, _ string) {}
func (s *noopSpan) SpanContext() observabilitycontract.SpanContext {
	return observabilitycontract.SpanContext{}
}
func (s *noopSpan) IsRecording() bool        { return false }
func (s *noopSpan) Context() context.Context { return context.Background() }
