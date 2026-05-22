package etcd

import (
	"testing"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	"github.com/stretchr/testify/require"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func TestProviderContract(t *testing.T) {
	p := NewProvider()
	require.Equal(t, "configsource.etcd", p.Name())
	require.True(t, p.IsDefer())
	require.Equal(t, []string{datacontract.ConfigSourceKey}, p.Provides())
}

func TestSourceUnderlyingAndAs(t *testing.T) {
	client := &clientv3.Client{}
	source := &Source{client: client}

	require.Same(t, client, source.Underlying())

	var projected *clientv3.Client
	require.True(t, source.As(&projected))
	require.Same(t, client, projected)
}
