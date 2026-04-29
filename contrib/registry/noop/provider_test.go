package noop

import (
	"context"
	"testing"

	"github.com/ngq/gorp/framework/contract"
	"github.com/stretchr/testify/assert"
)

func TestNoopRegistry(t *testing.T) {
	registry := &noopRegistry{}

	err := registry.Register(context.Background(), "user-service", "localhost:8080", nil)
	assert.NoError(t, err)

	err = registry.Deregister(context.Background(), "user-service", "localhost:8080")
	assert.NoError(t, err)

	instances, err := registry.Discover(context.Background(), "user-service")
	assert.NoError(t, err)
	assert.Empty(t, instances)

	err = registry.Close()
	assert.NoError(t, err)
}

func TestProviderMeta(t *testing.T) {
	p := NewProvider()
	assert.Equal(t, "discovery.noop", p.Name())
	assert.True(t, p.IsDefer())
	assert.ElementsMatch(t, []string{contract.RPCRegistryKey}, p.Provides())
}
