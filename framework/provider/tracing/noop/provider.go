package noop

import (
	"context"

	"github.com/ngq/gorp/framework/contract"
)

// Provider 提供 noop 追踪实现。
//
// 中文说明：
// - 单体项目默认使用此 provider；
// - 不引入任何外部依赖（无 OpenTelemetry/Jaeger）；
// - 所有 Span 操作为空实现，不记录任何追踪数据。
type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string     { return "tracing.noop" }
func (p *Provider) IsDefer() bool    { return true }
func (p *Provider) Provides() []string {
	return []string{contract.TracerKey, contract.TracerProviderKey}
}

func (p *Provider) Register(c contract.Container) error {
	c.Bind(contract.TracerProviderKey, func(c contract.Container) (any, error) {
		return &noopTracerProvider{}, nil
	}, true)

	c.Bind(contract.TracerKey, func(c contract.Container) (any, error) {
		return &noopTracer{}, nil
	}, true)

	return nil
}

func (p *Provider) Boot(c contract.Container) error {
	return nil
}

// noopTracerProvider 是 TracerProvider 的空实现。
type noopTracerProvider struct{}

func (p *noopTracerProvider) Tracer(name string, options ...contract.TracerOption) contract.Tracer {
	return &noopTracer{}
}

func (p *noopTracerProvider) Shutdown(ctx context.Context) error {
	return nil
}

func (p *noopTracerProvider) ForceFlush(ctx context.Context) error {
	return nil
}

// noopTracer 是 Tracer 的空实现。
//
// 中文说明：
// - StartSpan 返回 noopSpan；
// - Inject/Extract 空操作；
// - 不传播任何追踪上下文。
type noopTracer struct{}

func (t *noopTracer) StartSpan(ctx context.Context, name string, opts ...contract.SpanOption) (context.Context, contract.Span) {
	return ctx, &noopSpan{}
}

func (t *noopTracer) SpanFromContext(ctx context.Context) contract.Span {
	return &noopSpan{}
}

func (t *noopTracer) Inject(ctx context.Context, carrier contract.TextMapCarrier) error {
	return nil // 空操作
}

func (t *noopTracer) Extract(ctx context.Context, carrier contract.TextMapCarrier) (context.Context, error) {
	return ctx, nil // 返回原始 context
}

// noopSpan 是 Span 的空实现。
//
// 中文说明：
// - 所有方法空操作；
// - IsRecording 返回 false，表示不记录；
// - 用于避免不必要的属性计算开销。
type noopSpan struct{}

func (s *noopSpan) End(options ...contract.SpanEndOption) {}

func (s *noopSpan) AddEvent(name string, attributes map[string]interface{}) {}

func (s *noopSpan) SetTag(key string, value interface{}) {}

func (s *noopSpan) SetAttributes(attributes map[string]interface{}) {}

func (s *noopSpan) SetError(err error) {}

func (s *noopSpan) SetStatus(code contract.SpanStatusCode, description string) {}

func (s *noopSpan) SpanContext() contract.SpanContext {
	return contract.SpanContext{} // 返回空的 SpanContext
}

func (s *noopSpan) IsRecording() bool {
	return false // 不记录
}

func (s *noopSpan) Context() context.Context {
	return context.Background()
}