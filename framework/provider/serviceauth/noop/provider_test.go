package noop

import (
	"testing"

	"github.com/ngq/gorp/framework/contract"
	"github.com/stretchr/testify/require"
)

func TestProviderContract(t *testing.T) {
	p := NewProvider()
	require.Equal(t, "serviceauth.noop", p.Name())
	require.True(t, p.IsDefer())
	require.ElementsMatch(t, []string{contract.ServiceAuthKey, contract.ServiceIdentityKey}, p.Provides())
}
