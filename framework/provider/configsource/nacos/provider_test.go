package nacos

import (
	"testing"

	"github.com/ngq/gorp/framework/contract"
	"github.com/stretchr/testify/require"
)

func TestProviderContract(t *testing.T) {
	p := NewProvider()
	require.Equal(t, "configsource.nacos", p.Name())
	require.True(t, p.IsDefer())
	require.Equal(t, []string{contract.ConfigSourceKey}, p.Provides())
}
