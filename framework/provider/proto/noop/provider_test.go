package noop

import (
	"testing"

	"github.com/ngq/gorp/framework/contract"
	"github.com/stretchr/testify/require"
)

func TestProviderContract(t *testing.T) {
	p := NewProvider()
	require.Equal(t, "proto.noop", p.Name())
	require.True(t, p.IsDefer())
	require.Equal(t, []string{contract.ProtoGeneratorKey}, p.Provides())
}
