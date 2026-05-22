package token

import (
	"testing"

	securitycontract "github.com/ngq/gorp/framework/contract/security"
	"github.com/stretchr/testify/require"
)

func TestProviderContract(t *testing.T) {
	p := NewProvider()
	require.Equal(t, "serviceauth.token", p.Name())
	require.True(t, p.IsDefer())
	require.Equal(t, []string{securitycontract.ServiceAuthKey, securitycontract.ServiceIdentityKey}, p.Provides())
}
