package sentinel

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ngq/gorp/framework/contract"
	"github.com/stretchr/testify/require"
)

type cbConfigStub struct{}

func (cbConfigStub) Env() string                                   { return "test" }
func (cbConfigStub) Get(string) any                                { return nil }
func (cbConfigStub) GetString(string) string                       { return "" }
func (cbConfigStub) GetInt(string) int                             { return 0 }
func (cbConfigStub) GetBool(string) bool                           { return false }
func (cbConfigStub) GetFloat(string) float64                       { return 0 }
func (cbConfigStub) Unmarshal(string, any) error                   { return nil }
func (cbConfigStub) Watch(context.Context, string) (contract.ConfigWatcher, error) { return nil, nil }
func (cbConfigStub) Reload(context.Context) error                  { return nil }

type invalidConfigContainerStub struct{}

func (invalidConfigContainerStub) Bind(string, contract.Factory, bool)           {}
func (invalidConfigContainerStub) IsBind(string) bool                            { return true }
func (invalidConfigContainerStub) Make(string) (any, error)                      { return 1, nil }
func (invalidConfigContainerStub) MustMake(string) any                           { return 1 }
func (invalidConfigContainerStub) RegisterProvider(contract.ServiceProvider) error { return nil }
func (invalidConfigContainerStub) RegisterProviders(...contract.ServiceProvider) error { return nil }

func TestGetCircuitBreakerConfigDefaults(t *testing.T) {
	cfg := cbConfigStub{}
	container := sentinelContainerStub{cfg: cfg}
	cbCfg, err := getCircuitBreakerConfig(container)
	require.NoError(t, err)
	require.True(t, cbCfg.Enabled)
	require.Equal(t, "sentinel", cbCfg.Strategy)
	require.NotNil(t, cbCfg.ResourceConfigs)
	require.Equal(t, 10*time.Second, cbCfg.DefaultConfig.Timeout)
}

func TestGetCircuitBreakerConfigRejectsInvalidConfigService(t *testing.T) {
	_, err := getCircuitBreakerConfig(invalidConfigContainerStub{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid config service")
}

func TestSentinelCircuitBreaker_StateAndDo(t *testing.T) {
	cb := NewSentinelCircuitBreaker(&contract.CircuitBreakerConfig{})
	require.Equal(t, contract.CircuitBreakerStateClosed, cb.State(context.Background(), "resource"))

	errBoom := errors.New("boom")
	err := cb.Do(context.Background(), "resource", func() error { return errBoom })
	require.ErrorIs(t, err, errBoom)
}

func TestSentinelRateLimiter_AllowNAndReserve(t *testing.T) {
	rl := NewSentinelRateLimiter(&contract.CircuitBreakerConfig{})
	require.NoError(t, rl.AllowN(context.Background(), "resource", 0))

	res := rl.Reserve(context.Background(), "resource")
	require.NotNil(t, res)
	require.True(t, res.OK())
	require.Zero(t, res.Delay())
	require.NotPanics(t, func() { res.Cancel(); res.CancelAt(time.Now()) })
}

type sentinelContainerStub struct {
	cfg contract.Config
}

func (s sentinelContainerStub) Bind(string, contract.Factory, bool) {}
func (s sentinelContainerStub) IsBind(string) bool                  { return true }
func (s sentinelContainerStub) Make(key string) (any, error) {
	if key == contract.ConfigKey {
		return s.cfg, nil
	}
	return nil, errors.New("not found")
}
func (s sentinelContainerStub) MustMake(string) any { return s.cfg }
func (s sentinelContainerStub) RegisterProvider(contract.ServiceProvider) error { return nil }
func (s sentinelContainerStub) RegisterProviders(...contract.ServiceProvider) error { return nil }
