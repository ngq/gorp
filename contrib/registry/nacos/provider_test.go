package nacos

import (
	"testing"

	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	"github.com/stretchr/testify/require"
)

func TestProviderContract(t *testing.T) {
	p := NewProvider()
	require.Equal(t, "registry.nacos", p.Name())
	require.True(t, p.IsDefer())
	require.Equal(t, []string{transportcontract.RPCRegistryKey}, p.Provides())
}
