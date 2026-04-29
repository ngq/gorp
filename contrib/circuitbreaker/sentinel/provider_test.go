package sentinel

import (
	"testing"

	"github.com/ngq/gorp/framework/contract"
	"github.com/stretchr/testify/require"
)

func TestProviderContract(t *testing.T) {
	p := NewProvider()
	require.Equal(t, "circuitbreaker.sentinel", p.Name())
	require.True(t, p.IsDefer())
	require.Equal(t, []string{contract.CircuitBreakerKey, contract.RateLimiterKey}, p.Provides())
}
