package noop

import (
	"context"
	"testing"

	"github.com/ngq/gorp/framework/contract"
	"github.com/stretchr/testify/assert"
)

func TestNoopSource(t *testing.T) {
	source := &noopSource{}

	cfg, err := source.Load(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	_, err = source.Get(context.Background(), "demo")
	assert.ErrorIs(t, err, ErrNoopConfigSource)

	err = source.Set(context.Background(), "demo", "value")
	assert.ErrorIs(t, err, ErrNoopConfigSource)

	_, err = source.Watch(context.Background(), "demo")
	assert.ErrorIs(t, err, ErrNoopConfigSource)
}

func TestProviderMeta(t *testing.T) {
	p := NewProvider()
	assert.Equal(t, "configsource.noop", p.Name())
	assert.True(t, p.IsDefer())
	assert.ElementsMatch(t, []string{contract.ConfigSourceKey, contract.ConfigWatcherKey}, p.Provides())
}
