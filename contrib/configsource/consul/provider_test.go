package consul

import (
	"testing"

	"github.com/hashicorp/consul/api"
	datacontract "github.com/ngq/gorp/framework/contract/data"
	"github.com/stretchr/testify/require"
)

func TestProviderContract(t *testing.T) {
	p := NewProvider()
	require.Equal(t, "configsource.consul", p.Name())
	require.True(t, p.IsDefer())
	require.Equal(t, []string{datacontract.ConfigSourceKey}, p.Provides())
}

func TestSourceUnderlyingAndAs(t *testing.T) {
	client, err := api.NewClient(api.DefaultConfig())
	require.NoError(t, err)

	source := &Source{client: client}
	require.Same(t, client, source.Underlying())

	var projected *api.Client
	require.True(t, source.As(&projected))
	require.Same(t, client, projected)
}
