// Package redis_test provides unit tests for the redis provider.
//
// 适用场景：
// - 验证 Redis provider 的注册与契约实现。
package redis

import (
	"testing"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	"github.com/stretchr/testify/require"
)

func TestProviderContract(t *testing.T) {
	p := NewProvider()
	require.Equal(t, "redis", p.Name())
	require.False(t, p.IsDefer())
	require.Equal(t, []string{datacontract.RedisKey}, p.Provides())
}
