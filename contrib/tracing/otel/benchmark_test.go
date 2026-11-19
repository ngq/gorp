package otel

import (
	"context"
	"testing"

	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
)

// BenchmarkTracingSpanCreation 测试 OTel span 创建开销
func BenchmarkTracingSpanCreation(b *testing.B) {
	cfg := &observabilitycontract.TracingConfig{
		ServiceName:      "benchmark-service",
		ExporterEndpoint: "",
		ExporterType:     "stdout",
		Enabled:          false,
		SamplingRate:     1.0,
	}

	tracerProvider, _ := NewTracerProvider(cfg)
	tracer := NewTracer(tracerProvider, cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx := context.Background()
		_, span := tracer.StartSpan(ctx, "benchmark-span", func(cfg *observabilitycontract.SpanConfig) {
			cfg.Kind = observabilitycontract.SpanKindServer
		})
		span.End()
	}
}

// BenchmarkTracingSpanCreationWithAttributes 测试带 attributes 的 span 创建
func BenchmarkTracingSpanCreationWithAttributes(b *testing.B) {
	cfg := &observabilitycontract.TracingConfig{
		ServiceName:      "benchmark-service",
		ExporterEndpoint: "",
		ExporterType:     "stdout",
		Enabled:          false,
		SamplingRate:     1.0,
	}

	tracerProvider, _ := NewTracerProvider(cfg)
	tracer := NewTracer(tracerProvider, cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx := context.Background()
		_, span := tracer.StartSpan(ctx, "benchmark-span", func(cfg *observabilitycontract.SpanConfig) {
			cfg.Kind = observabilitycontract.SpanKindServer
			cfg.Attributes = map[string]interface{}{
				"user.id":    12345,
				"request.id": "req-123",
				"method":     "GET",
			}
		})
		span.SetAttributes(map[string]interface{}{
			"result": "success",
			"code":   200,
		})
		span.End()
	}
}

// BenchmarkTracingSpanCreation_Nested 测试嵌套 span 创建
func BenchmarkTracingSpanCreation_Nested(b *testing.B) {
	cfg := &observabilitycontract.TracingConfig{
		ServiceName:      "benchmark-service",
		ExporterEndpoint: "",
		ExporterType:     "stdout",
		Enabled:          false,
		SamplingRate:     1.0,
	}

	tracerProvider, _ := NewTracerProvider(cfg)
	tracer := NewTracer(tracerProvider, cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx := context.Background()

		// parent span
		ctx, parentSpan := tracer.StartSpan(ctx, "parent-span", func(cfg *observabilitycontract.SpanConfig) {
			cfg.Kind = observabilitycontract.SpanKindServer
		})

		// child span
		_, childSpan := tracer.StartSpan(ctx, "child-span", func(cfg *observabilitycontract.SpanConfig) {
			cfg.Kind = observabilitycontract.SpanKindInternal
		})
		childSpan.End()

		parentSpan.End()
	}
}
