package noop

import (
	"context"

	"github.com/ngq/gorp/framework/contract"
)

// Provider 提供 noop 链路追踪实现。
//
// 中文说明：
// - 单服务 / 单体场景默认使用此 provider；
// - 所有 span 操作均为空实现，不产生任何追踪数据；
// - 需要真实链路追踪时，注册 contrib/tracing/otel 中的实现替换本 provider。
type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string { return "tracing.noop" }
func (p *Provider) IsDefer() bool { return true }
func (p *Provider) Provides() []string {
	return []string{contract.TracerKey, contract.TracerProviderKey}
}

func (p *Provider) Register(c contract.Container) error {
	c.Bind(contract.TracerKey, func(c contract.Container) (any, error) {
		return &noopTracer{}, nil
	}, true)
	c.Bind(contract.TracerProviderKey, func(c contract.Container) (any, error) {
		return &noopTracerProvider{}, nil
	}, true)
	return nil
}

func (p *Provider) Boot(contract.Container) error { return nil }

// ── noopTracer ────────────────────────────────────────────────────────────────

type noopTracer struct{}

func (t *noopTracer) StartSpan(ctx context.Context, _ string, _ ...contract.SpanOption) (context.Context, contract.Span) {
	return ctx, &noopSpan{}
}

func (t *noopTracer) SpanFromContext(_ context.Context) contract.Span {
	return &noopSpan{}
}

func (t *noopTracer) Inject(_ context.Context, _ contract.TextMapCarrier) error {
	return nil
}

func (t *noopTracer) Extract(ctx context.Context, _ contract.TextMapCarrier) (context.Context, error) {
	return ctx, nil
}

// ── noopTracerProvider ────────────────────────────────────────────────────────

type noopTracerProvider struct{}

func (p *noopTracerProvider) Tracer(_ string, _ ...contract.TracerOption) contract.Tracer {
	return &noopTracer{}
}

func (p *noopTracerProvider) Shutdown(_ context.Context) error { return nil }

// ── noopSpan ──────────────────────────────────────────────────────────────────

type noopSpan struct{}

func (s *noopSpan) End(_ ...contract.SpanEndOption)             {}
func (s *noopSpan) AddEvent(_ string, _ map[string]interface{}) {}
func (s *noopSpan) SetTag(_ string, _ interface{})              {}
func (s *noopSpan) SetAttributes(_ map[string]interface{})      {}
func (s *noopSpan) SetError(_ error)                            {}
func (s *noopSpan) SetStatus(_ contract.SpanStatusCode, _ string) {}
func (s *noopSpan) SpanContext() contract.SpanContext            { return contract.SpanContext{} }
func (s *noopSpan) IsRecording() bool                           { return false }
func (s *noopSpan) Context() context.Context                    { return context.Background() }
