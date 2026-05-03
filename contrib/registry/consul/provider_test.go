package consul

import (
	"testing"

	"github.com/hashicorp/consul/api"
	"github.com/ngq/gorp/framework/contract"
	"github.com/stretchr/testify/require"
)

func TestProviderContract(t *testing.T) {
	p := NewProvider()
	require.Equal(t, "registry.consul", p.Name())
	require.True(t, p.IsDefer())
	require.Equal(t, []string{contract.RPCRegistryKey}, p.Provides())
}

func TestRegistryUnderlyingAndAs(t *testing.T) {
	client, err := api.NewClient(api.DefaultConfig())
	require.NoError(t, err)

	registry := &Registry{client: client}
	require.Same(t, client, registry.Underlying())

	var projected *api.Client
	require.True(t, registry.As(&projected))
	require.Same(t, client, projected)
}
