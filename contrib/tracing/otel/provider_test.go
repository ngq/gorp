package otel

import (
	"testing"

	"github.com/ngq/gorp/framework/contract"
	"github.com/stretchr/testify/require"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

func TestProviderContract(t *testing.T) {
	p := NewProvider()
	require.Equal(t, "tracing.otel", p.Name())
	require.True(t, p.IsDefer())
	require.Equal(t, []string{contract.TracerKey, contract.TracerProviderKey}, p.Provides())
}

func TestTracerProviderUnderlyingAndAs(t *testing.T) {
	native := sdktrace.NewTracerProvider()
	wrapper := &TracerProviderWrapper{provider: native}

	require.Equal(t, native, wrapper.Underlying())

	var projected *sdktrace.TracerProvider
	require.True(t, wrapper.As(&projected))
	require.Same(t, native, projected)
}

func TestTracerUnderlyingAndAs(t *testing.T) {
	native := noop.NewTracerProvider().Tracer("test")
	wrapper := &TracerWrapper{tracer: native}

	require.Equal(t, native, wrapper.Underlying())

	var projected trace.Tracer
	require.True(t, wrapper.As(&projected))
	require.Equal(t, native, projected)
}
