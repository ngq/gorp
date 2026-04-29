package redis

import (
	"testing"

	"github.com/ngq/gorp/framework/contract"
	"github.com/stretchr/testify/require"
)

func TestProviderContract(t *testing.T) {
	p := NewProvider()
	require.Equal(t, "redis", p.Name())
	require.False(t, p.IsDefer())
	require.Equal(t, []string{contract.RedisKey}, p.Provides())
}
