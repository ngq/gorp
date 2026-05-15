// Package noop provides a no-op tracing implementation for monolith scenarios.
// This tracer creates no-op spans that do nothing.
// Note: Distributed tracing is not available in monolith mode.
//
// 空链路追踪实现包，用于单体应用场景。
// 此追踪器创建空 span，不执行任何操作。
// 注意：分布式链路追踪在单体模式下不可用。
package noop

import (
	"context"

	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

// Provider registers no-op tracing contracts.
//
// Provider 注册空链路追踪契约。
type Provider struct{}

// NewProvider creates a new no-op tracing provider instance.
//
// NewProvider 创建新的空链路追踪 Provider 实例。
func NewProvider() *Provider { return &Provider{} }

// Name returns the provider name "tracing.noop".
//
// Name 返回 Provider 名称 "tracing.noop"。
func (p *Provider) Name() string { return "tracing.noop" }

// IsDefer returns true, tracing can be deferred until first use.
//
// IsDefer 返回 true，链路追踪可延迟初始化直到首次使用。
func (p *Provider) IsDefer() bool { return true }

// Provides returns the tracing contract keys.
//
// Provides 返回链路追踪契约键列表。
func (p *Provider) Provides() []string {
	return []string{observabilitycontract.TracerKey, observabilitycontract.TracerProviderKey}
}

// DependsOn returns the keys this provider depends on.
// Noop tracing has no dependencies.
//
// DependsOn 返回该 provider 依赖的 key。
// Noop tracing 无依赖。
func (p *Provider) DependsOn() []string { return nil }

// Register binds the no-op tracer to the container.
//
// Register 将空追踪器绑定到容器。
func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(observabilitycontract.TracerKey, func(c runtimecontract.Container) (any, error) {
		return &noopTracer{}, nil
	}, true)
	c.Bind(observabilitycontract.TracerProviderKey, func(c runtimecontract.Container) (any, error) {
		return &noopTracerProvider{}, nil
	}, true)
	return nil
}

// Boot is a no-op for this provider.
//
// Boot 此 Provider 无启动逻辑。
func (p *Provider) Boot(runtimecontract.Container) error { return nil }

// noopTracer implements observabilitycontract.Tracer with no-op behavior.
//
// noopTracer 使用空行为实现 observabilitycontract.Tracer 接口。
type noopTracer struct{}

// StartSpan returns a no-op span.
//
// StartSpan 返回空 span。
func (t *noopTracer) StartSpan(ctx context.Context, _ string, _ ...observabilitycontract.SpanOption) (context.Context, observabilitycontract.Span) {
	return ctx, &noopSpan{}
}

// SpanFromContext returns a no-op span.
//
// SpanFromContext 返回空 span。
func (t *noopTracer) SpanFromContext(_ context.Context) observabilitycontract.Span {
	return &noopSpan{}
}

// Inject does nothing and returns nil.
//
// Inject 不执行任何操作并返回 nil。
func (t *noopTracer) Inject(_ context.Context, _ observabilitycontract.TextMapCarrier) error {
	return nil
}

// Extract returns the original context.
//
// Extract 返回原始 context。
func (t *noopTracer) Extract(ctx context.Context, _ observabilitycontract.TextMapCarrier) (context.Context, error) {
	return ctx, nil
}

// noopTracerProvider implements observabilitycontract.TracerProvider with no-op behavior.
//
// noopTracerProvider 使用空行为实现 observabilitycontract.TracerProvider 接口。
type noopTracerProvider struct{}

// Tracer returns a no-op tracer.
//
// Tracer 返回空追踪器。
func (p *noopTracerProvider) Tracer(_ string, _ ...observabilitycontract.TracerOption) observabilitycontract.Tracer {
	return &noopTracer{}
}

// Shutdown does nothing and returns nil.
//
// Shutdown 不执行任何操作并返回 nil。
func (p *noopTracerProvider) Shutdown(_ context.Context) error { return nil }

// ForceFlush does nothing and returns nil.
//
// ForceFlush 不执行任何操作并返回 nil。
func (p *noopTracerProvider) ForceFlush(_ context.Context) error { return nil }

// noopSpan implements observabilitycontract.Span with no-op behavior.
//
// noopSpan 使用空行为实现 observabilitycontract.Span 接口。
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