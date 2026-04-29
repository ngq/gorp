package otel

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ngq/gorp/framework/contract"
	"github.com/stretchr/testify/require"
)

type testTextMapCarrier struct {
	values map[string]string
}

func (c *testTextMapCarrier) Get(key string) string { return c.values[key] }
func (c *testTextMapCarrier) Set(key string, value string) { c.values[key] = value }
func (c *testTextMapCarrier) Keys() []string {
	keys := make([]string, 0, len(c.values))
	for k := range c.values {
		keys = append(keys, k)
	}
	return keys
}

func TestTracerProviderAndTracer_StartInjectExtractAndShutdown(t *testing.T) {
	provider, err := NewTracerProvider(&contract.TracingConfig{
		Enabled:          true,
		ExporterType:     "stdout",
		ServiceName:      "order-service",
		SamplingRate:     1,
		Propagators:      []string{"tracecontext", "baggage"},
		BatchTimeout:     1,
		MaxQueueSize:     16,
		MaxExportBatchSize: 8,
	})
	require.NoError(t, err)

	tracer := NewTracer(provider, &contract.TracingConfig{ServiceName: "order-service"})
	ctx, span := tracer.StartSpan(context.Background(), "op", func(cfg *contract.SpanConfig) {
		cfg.Kind = contract.SpanKindServer
	})
	require.True(t, span.IsRecording())

	carrier := &testTextMapCarrier{values: map[string]string{}}
	require.NoError(t, tracer.Inject(ctx, carrier))
	require.NotEmpty(t, carrier.Get("traceparent"))

	extracted, err := tracer.Extract(context.Background(), carrier)
	require.NoError(t, err)
	extractedSpan := tracer.SpanFromContext(extracted)
	require.NotEqual(t, contract.SpanContext{}, extractedSpan.SpanContext())

	span.AddEvent("done", map[string]any{"ok": true})
	span.SetTag("service", "order")
	span.SetError(errors.New("boom"))
	span.End()

	require.NoError(t, provider.ForceFlush(context.Background()))
	require.NoError(t, provider.Shutdown(context.Background()))
	require.NoError(t, provider.Shutdown(context.Background()))
}

func TestCreateSamplerAndNoopSpan(t *testing.T) {
	require.NotNil(t, createSampler(1))
	require.NotNil(t, createSampler(0))
	require.NotNil(t, createSampler(0.5))

	span := (&noopSpan{})
	span.SetError(errors.New("ignored"))
	span.SetStatus(contract.SpanStatusCodeError, "bad")
	span.End(func(cfg *contract.SpanEndConfig) {
		cfg.EndTime = time.Now()
	})
	require.False(t, span.IsRecording())
}
