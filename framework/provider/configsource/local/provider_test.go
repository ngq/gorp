package local

import (
	"testing"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	"github.com/stretchr/testify/require"
)

func TestProviderContract(t *testing.T) {
	p := NewProvider()
	require.Equal(t, "configsource.local", p.Name())
	require.True(t, p.IsDefer())
	require.Equal(t, []string{datacontract.ConfigSourceKey}, p.Provides())
}
