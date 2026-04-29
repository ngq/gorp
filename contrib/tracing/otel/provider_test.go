package otel

import (
	"testing"

	"github.com/ngq/gorp/framework/contract"
	"github.com/stretchr/testify/require"
)

func TestProviderContract(t *testing.T) {
	p := NewProvider()
	require.Equal(t, "tracing.otel", p.Name())
	require.True(t, p.IsDefer())
	require.Equal(t, []string{contract.TracerKey, contract.TracerProviderKey}, p.Provides())
}
