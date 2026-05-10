// Package integration provides tracing integration tests.
//
// 本包提供 tracing 集成测试。
package integration

import (
	"context"
	"testing"
	"time"

	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	otelprovider "github.com/ngq/gorp/contrib/tracing/otel"
)

// TestTracingSpanCreation tests OTel tracing span creation.
//
// TestTracingSpanCreation 测试 OTel tracing span 创建。
func TestTracingSpanCreation(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test requires OTel backend")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 1. Create OTel tracer
	// 创建 OTel tracer
	cfg := &observabilitycontract.TracingConfig{
		ServiceName:      "test-service-integration",
		ExporterEndpoint: getEnvOrDefault("GORP_TEST_JAEGER_ADDR", "localhost:4317"),
		ExporterType:     "otlp",
		Enabled:          true,
		Propagators:      []string{"tracecontext", "baggage"},
		SamplingRate:     1.0,
	}

	tracerProvider, err := otelprovider.NewTracerProvider(cfg)
	if err != nil {
		t.Fatalf("failed to create tracer provider: %v", err)
	}

	tracer := otelprovider.NewTracer(tracerProvider, cfg)

	t.Logf("tracer created: service=%s", cfg.ServiceName)

	// 2. Create a span
	// 创建 span
	spanName := "test-span-integration"
	ctx, span := tracer.StartSpan(ctx, spanName, func(cfg *observabilitycontract.SpanConfig) {
		cfg.Kind = observabilitycontract.SpanKindServer
	})

	// 3. Set attributes
	// 设置 attributes
	span.SetAttributes(map[string]interface{}{
		"test.key":    "test-value",
		"test.number": 123,
	})

	t.Logf("span started: name=%s", spanName)

	// 4. Simulate work
	// 模拟工作
	time.Sleep(100 * time.Millisecond)

	// 5. End span
	// 结束 span
	span.SetStatus(observabilitycontract.SpanStatusCodeOk, "test completed")
	span.End()

	t.Log("span ended successfully")

	// 6. Shutdown tracer
	// 关闭 tracer
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	tracerProvider.Shutdown(shutdownCtx)
}

// TestTracingErrorRecording tests span error recording.
//
// TestTracingErrorRecording 测试 span 错误记录。
func TestTracingErrorRecording(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test requires OTel backend")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cfg := &observabilitycontract.TracingConfig{
		ServiceName:      "test-service-error",
		ExporterEndpoint: getEnvOrDefault("GORP_TEST_JAEGER_ADDR", "localhost:4317"),
		ExporterType:     "otlp",
		Enabled:          true,
		Propagators:      []string{"tracecontext", "baggage"},
	}

	tracerProvider, err := otelprovider.NewTracerProvider(cfg)
	if err != nil {
		t.Fatalf("failed to create tracer provider: %v", err)
	}

	tracer := otelprovider.NewTracer(tracerProvider, cfg)

	// Create span with error
	// 创建带错误的 span
	ctx, span := tracer.StartSpan(ctx, "test-span-error", func(cfg *observabilitycontract.SpanConfig) {
		cfg.Kind = observabilitycontract.SpanKindServer
	})

	// Simulate error using resiliencecontract.NewError
	// 使用 resiliencecontract.NewError 模拟错误
	testErr := resiliencecontract.NewError(
		resiliencecontract.ErrorCodeInternalServerError,
		resiliencecontract.ErrorReasonInternal,
		"test error for tracing",
	)

	span.SetError(testErr)
	span.SetStatus(observabilitycontract.SpanStatusCodeError, testErr.GetStatus().Message)
	span.End()

	t.Logf("error span recorded: error=%s", testErr.GetStatus().Message)

	shutdownCtx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	tracerProvider.Shutdown(shutdownCtx)
}

// TestTracingNestedSpans tests nested span hierarchy.
//
// TestTracingNestedSpans 测试嵌套 span 层级。
func TestTracingNestedSpans(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test requires OTel backend")
	}

	ctx := context.Background()

	cfg := &observabilitycontract.TracingConfig{
		ServiceName:      "test-service-nested",
		ExporterEndpoint: getEnvOrDefault("GORP_TEST_JAEGER_ADDR", "localhost:4317"),
		ExporterType:     "otlp",
		Enabled:          true,
		Propagators:      []string{"tracecontext", "baggage"},
	}

	tracerProvider, err := otelprovider.NewTracerProvider(cfg)
	if err != nil {
		t.Fatalf("failed to create tracer provider: %v", err)
	}

	tracer := otelprovider.NewTracer(tracerProvider, cfg)

	// Create parent span
	// 创建父 span
	ctx, parentSpan := tracer.StartSpan(ctx, "parent-span", func(cfg *observabilitycontract.SpanConfig) {
		cfg.Kind = observabilitycontract.SpanKindServer
	})

	parentSpan.SetAttributes(map[string]interface{}{
		"level": "parent",
	})

	t.Log("parent span started")

	// Create child span
	// 创建子 span
	ctx, childSpan := tracer.StartSpan(ctx, "child-span", func(cfg *observabilitycontract.SpanConfig) {
		cfg.Kind = observabilitycontract.SpanKindInternal
	})

	childSpan.SetAttributes(map[string]interface{}{
		"level": "child",
	})

	t.Log("child span started")

	// End spans in order (child first, then parent)
	// 按顺序结束 span（子先，父后）
	childSpan.SetStatus(observabilitycontract.SpanStatusCodeOk, "child done")
	childSpan.End()

	parentSpan.SetStatus(observabilitycontract.SpanStatusCodeOk, "parent done")
	parentSpan.End()

	t.Log("nested spans ended successfully")

	shutdownCtx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	tracerProvider.Shutdown(shutdownCtx)
}

// TestTracingShutdown tests tracer provider shutdown.
//
// TestTracingShutdown 测试 tracer provider 关闭。
func TestTracingShutdown(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test requires OTel backend")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cfg := &observabilitycontract.TracingConfig{
		ServiceName:      "test-service-shutdown",
		ExporterEndpoint: getEnvOrDefault("GORP_TEST_JAEGER_ADDR", "localhost:4317"),
		ExporterType:     "otlp",
		Enabled:          true,
		Propagators:      []string{"tracecontext", "baggage"},
	}

	tracerProvider, err := otelprovider.NewTracerProvider(cfg)
	if err != nil {
		t.Fatalf("failed to create tracer provider: %v", err)
	}

	tracer := otelprovider.NewTracer(tracerProvider, cfg)

	// Create and end some spans before shutdown
	// 关闭前创建和结束一些 spans
	for i := 0; i < 3; i++ {
		_, span := tracer.StartSpan(ctx, "shutdown-test-span", func(cfg *observabilitycontract.SpanConfig) {
			cfg.Kind = observabilitycontract.SpanKindInternal
		})
		span.End()
	}

	// Shutdown with timeout
	// 带超时关闭
	err = tracerProvider.Shutdown(ctx)
	if err != nil {
		t.Logf("tracer shutdown: %v (timeout acceptable)", err)
	} else {
		t.Log("tracer shutdown completed")
	}
}