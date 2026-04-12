package noop

import (
	"context"
	"testing"

	"github.com/ngq/gorp/framework/contract"
	"github.com/stretchr/testify/assert"
)

func TestNoopSource(t *testing.T) {
	source := &noopSource{}

	// 测试 Load 返回空 map
	cfg, err := source.Load(context.Background())
	assert.NoError(t, err)
	assert.Empty(t, cfg)

	// 测试 Get 返回错误
	val, err := source.Get(context.Background(), "test.key")
	assert.Nil(t, val)
	assert.ErrorIs(t, err, ErrNoopConfigSource)

	// 测试 Set 返回错误
	err = source.Set(context.Background(), "test.key", "value")
	assert.ErrorIs(t, err, ErrNoopConfigSource)

	// 测试 Watch 返回错误
	watcher, err := source.Watch(context.Background(), "test.key")
	assert.Nil(t, watcher)
	assert.ErrorIs(t, err, ErrNoopConfigSource)

	// 测试 Close 无错误
	err = source.Close()
	assert.NoError(t, err)
}

func TestProvider(t *testing.T) {
	p := NewProvider()

	assert.Equal(t, "configsource.noop", p.Name())
	assert.True(t, p.IsDefer())
	assert.ElementsMatch(t, []string{
		contract.ConfigSourceKey,
		contract.ConfigWatcherKey,
	}, p.Provides())
}