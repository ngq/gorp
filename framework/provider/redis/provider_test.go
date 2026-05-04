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
