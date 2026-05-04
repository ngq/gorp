package sentinel

import (
	"testing"

	base "github.com/alibaba/sentinel-golang/core/base"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	"github.com/stretchr/testify/require"
)

func TestProviderContract(t *testing.T) {
	p := NewProvider()
	require.Equal(t, "circuitbreaker.sentinel", p.Name())
	require.True(t, p.IsDefer())
	require.Equal(t, []string{resiliencecontract.CircuitBreakerKey, resiliencecontract.RateLimiterKey}, p.Provides())
}

func TestInterfaceCompatibility(t *testing.T) {
	var cb resiliencecontract.CircuitBreaker = NewSentinelCircuitBreaker(&resiliencecontract.CircuitBreakerConfig{})
	var rl resiliencecontract.RateLimiter = NewSentinelRateLimiter(&resiliencecontract.CircuitBreakerConfig{})
	require.NotNil(t, cb)
	require.NotNil(t, rl)
}

func TestNativeEscapeHatch(t *testing.T) {
	cb := NewSentinelCircuitBreaker(&resiliencecontract.CircuitBreakerConfig{})
	rl := NewSentinelRateLimiter(&resiliencecontract.CircuitBreakerConfig{})

	require.NotNil(t, cb.Underlying())
	require.NotNil(t, rl.Underlying())

	var cbChain *base.SlotChain
	require.True(t, cb.As(&cbChain))
	require.NotNil(t, cbChain)

	var rlChain *base.SlotChain
	require.True(t, rl.As(&rlChain))
	require.NotNil(t, rlChain)
}
