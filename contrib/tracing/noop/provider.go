package noop

import (
	"context"

	"github.com/ngq/gorp/framework/contract"
)

type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string     { return "tracing.noop" }
func (p *Provider) IsDefer() bool    { return true }
func (p *Provider) Provides() []string { return []string{contract.TracerKey, contract.TracerProviderKey} }

func (p *Provider) Register(c contract.Container) error {
	c.Bind(contract.TracerProviderKey, func(c contract.Container) (any, error) {
		return &noopTracerProvider{}, nil
	}, true)
	c.Bind(contract.TracerKey, func(c contract.Container) (any, error) {
		return &noopTracer{}, nil
	}, true)
	return nil
}

func (p *Provider) Boot(c contract.Container) error { return nil }

type noopTracerProvider struct{}

func (p *noopTracerProvider) Tracer(name string, options ...contract.TracerOption) contract.Tracer { return &noopTracer{} }
func (p *noopTracerProvider) Shutdown(ctx context.Context) error { return nil }
func (p *noopTracerProvider) ForceFlush(ctx context.Context) error { return nil }

type noopTracer struct{}

func (t *noopTracer) StartSpan(ctx context.Context, name string, opts ...contract.SpanOption) (context.Context, contract.Span) {
	return ctx, &noopSpan{}
}
func (t *noopTracer) SpanFromContext(ctx context.Context) contract.Span { return &noopSpan{} }
func (t *noopTracer) Inject(ctx context.Context, carrier contract.TextMapCarrier) error { return nil }
func (t *noopTracer) Extract(ctx context.Context, carrier contract.TextMapCarrier) (context.Context, error) { return ctx, nil }

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
