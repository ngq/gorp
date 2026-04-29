package metadata

import (
	"testing"

	"github.com/ngq/gorp/framework/contract"
	"github.com/stretchr/testify/require"
)

func TestProviderContract(t *testing.T) {
	p := NewProvider()
	require.Equal(t, "metadata.default", p.Name())
	require.True(t, p.IsDefer())
	require.ElementsMatch(t, []string{contract.MetadataKey, contract.MetadataPropagatorKey}, p.Provides())
}
