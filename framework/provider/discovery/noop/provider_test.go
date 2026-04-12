package noop

import (
	"context"
	"testing"

	"github.com/ngq/gorp/framework/contract"
	"github.com/stretchr/testify/assert"
)

func TestNoopRegistry(t *testing.T) {
	registry := &noopRegistry{}

	// 测试 Register 无错误
	err := registry.Register(context.Background(), "user-service", "localhost:8080", nil)
	assert.NoError(t, err)

	// 测试 Deregister 无错误
	err = registry.Deregister(context.Background(), "user-service", "localhost:8080")
	assert.NoError(t, err)

	// 测试 Discover 返回空列表
	instances, err := registry.Discover(context.Background(), "user-service")
	assert.NoError(t, err)
	assert.Empty(t, instances)

	// 测试 Close 无错误
	err = registry.Close()
	assert.NoError(t, err)
}

func TestProvider(t *testing.T) {
	p := NewProvider()

	assert.Equal(t, "discovery.noop", p.Name())
	assert.True(t, p.IsDefer())
	assert.ElementsMatch(t, []string{contract.RPCRegistryKey}, p.Provides())
}