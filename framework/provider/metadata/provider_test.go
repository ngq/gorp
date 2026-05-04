package metadata

import (
	"testing"

	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	"github.com/stretchr/testify/require"
)

func TestProviderContract(t *testing.T) {
	p := NewProvider()
	require.Equal(t, "metadata.default", p.Name())
	require.True(t, p.IsDefer())
	require.ElementsMatch(t, []string{transportcontract.MetadataKey, transportcontract.MetadataPropagatorKey}, p.Provides())
}
