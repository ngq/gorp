package noop

import (
	"context"
	"testing"

	"github.com/ngq/gorp/framework/contract"
	"github.com/stretchr/testify/assert"
)

func TestNoopTracerProvider(t *testing.T) {
	p := NewProvider()
	assert.Equal(t, "tracing.noop", p.Name())
	assert.True(t, p.IsDefer())
	assert.ElementsMatch(t, []string{contract.TracerKey, contract.TracerProviderKey}, p.Provides())

	provider := &noopTracerProvider{}
	tracer := provider.Tracer("demo")
	ctx, span := tracer.StartSpan(context.Background(), "op")
	assert.NotNil(t, ctx)
	assert.NotNil(t, span)
	assert.False(t, span.IsRecording())
}
