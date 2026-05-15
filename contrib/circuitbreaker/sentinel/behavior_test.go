package sentinel

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	sentinelapi "github.com/alibaba/sentinel-golang/api"
	base "github.com/alibaba/sentinel-golang/core/base"
	sentinelcb "github.com/alibaba/sentinel-golang/core/circuitbreaker"
	"github.com/alibaba/sentinel-golang/core/isolation"
	datacontract "github.com/ngq/gorp/framework/contract/data"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	"github.com/stretchr/testify/require"
)

type cbConfigStub struct{}

func (cbConfigStub) Env() string                 { return "test" }
func (cbConfigStub) Get(string) any              { return nil }
func (cbConfigStub) GetString(string) string     { return "" }
func (cbConfigStub) GetInt(string) int           { return 0 }
func (cbConfigStub) GetBool(string) bool         { return false }
func (cbConfigStub) GetFloat(string) float64     { return 0 }
func (cbConfigStub) Unmarshal(string, any) error { return nil }
func (cbConfigStub) Watch(context.Context, string) (datacontract.ConfigWatcher, error) {
	return nil, nil
}
func (cbConfigStub) Reload(context.Context) error { return nil }

type invalidConfigContainerStub struct{}

func (invalidConfigContainerStub) Bind(string, runtimecontract.Factory, bool)              {}
func (invalidConfigContainerStub) NamedBind(string, string, runtimecontract.Factory, bool) {}
func (invalidConfigContainerStub) IsBind(string) bool                                      { return true }
func (invalidConfigContainerStub) IsBindNamed(string, string) bool                         { return false }
func (invalidConfigContainerStub) Make(string) (any, error)                                { return 1, nil }
func (invalidConfigContainerStub) MakeNamed(string, string) (any, error)                   { return nil, nil }
func (invalidConfigContainerStub) MustMake(string) any                                     { return 1 }
func (invalidConfigContainerStub) MustMakeNamed(string, string) any                        { return nil }
func (invalidConfigContainerStub) RegisterCloser(string, io.Closer)                        {}
func (invalidConfigContainerStub) Destroy() error                                          { return nil }
func (invalidConfigContainerStub) RegisterProvider(runtimecontract.ServiceProvider) error {
	return nil
}
func (invalidConfigContainerStub) RegisterProviders(...runtimecontract.ServiceProvider) error {
	return nil
}
func (invalidConfigContainerStub) RegisteredProviders() []runtimecontract.ProviderInfo { return nil }
func (invalidConfigContainerStub) DebugPrint() string                                  { return "" }
func (invalidConfigContainerStub) ProviderDAG() runtimecontract.ProviderDAG            { return runtimecontract.ProviderDAG{} }

type sentinelContainerStub struct {
	cfg datacontract.Config
}

func (s sentinelContainerStub) Bind(string, runtimecontract.Factory, bool)              {}
func (s sentinelContainerStub) NamedBind(string, string, runtimecontract.Factory, bool) {}
func (s sentinelContainerStub) IsBind(string) bool                                      { return true }
func (s sentinelContainerStub) IsBindNamed(string, string) bool                         { return false }
func (s sentinelContainerStub) Make(key string) (any, error) {
	if key == datacontract.ConfigKey {
		return s.cfg, nil
	}
	return nil, errors.New("not found")
}
func (s sentinelContainerStub) MakeNamed(string, string) (any, error) { return nil, nil }
func (s sentinelContainerStub) MustMake(string) any                   { return s.cfg }
func (s sentinelContainerStub) MustMakeNamed(string, string) any      { return nil }
func (s sentinelContainerStub) RegisterCloser(string, io.Closer)      {}
func (s sentinelContainerStub) Destroy() error                        { return nil }
func (s sentinelContainerStub) RegisterProvider(runtimecontract.ServiceProvider) error {
	return nil
}
func (s sentinelContainerStub) RegisterProviders(...runtimecontract.ServiceProvider) error {
	return nil
}
func (s sentinelContainerStub) RegisteredProviders() []runtimecontract.ProviderInfo { return nil }
func (s sentinelContainerStub) DebugPrint() string                                  { return "" }
func (s sentinelContainerStub) ProviderDAG() runtimecontract.ProviderDAG            { return runtimecontract.ProviderDAG{} }

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

func TestSentinelCircuitBreakerDoTracksFailureAndSuccess(t *testing.T) {
	cb := NewSentinelCircuitBreaker(&resiliencecontract.CircuitBreakerConfig{
		DefaultConfig: resiliencecontract.ResourceConfig{Timeout: 5 * time.Millisecond},
	})

	errBoom := errors.New("boom")
	err := cb.Do(context.Background(), "resource", func() error { return errBoom })
	require.ErrorIs(t, err, errBoom)
	require.Equal(t, resiliencecontract.CircuitBreakerStateOpen, cb.State(context.Background(), "resource"))

	time.Sleep(10 * time.Millisecond)
	cb.RecordSuccess(context.Background(), "resource")
	require.Equal(t, resiliencecontract.CircuitBreakerStateClosed, cb.State(context.Background(), "resource"))
}

func TestSentinelCircuitBreakerAllowMarksOpenOnBlock(t *testing.T) {
	original := sentinelEntry
	sentinelEntry = func(resource string, opts ...sentinelapi.EntryOption) (*base.SentinelEntry, *base.BlockError) {
		return nil, base.NewBlockErrorWithMessage(base.BlockTypeFlow, "blocked")
	}
	defer func() { sentinelEntry = original }()

	cb := NewSentinelCircuitBreaker(&resiliencecontract.CircuitBreakerConfig{})
	err := cb.Allow(context.Background(), "blocked-resource")
	require.Error(t, err)
	require.Equal(t, resiliencecontract.CircuitBreakerStateOpen, cb.State(context.Background(), "blocked-resource"))
}

func TestSentinelRateLimiterReserveWaitAndTimeout(t *testing.T) {
	rl := NewSentinelRateLimiter(&resiliencecontract.CircuitBreakerConfig{})
	require.NoError(t, rl.AllowN(context.Background(), "resource", 0))

	res := rl.Reserve(context.Background(), "resource")
	require.NotNil(t, res)
	require.True(t, res.OK())
	require.Zero(t, res.Delay())
	require.NotPanics(t, func() { res.Cancel(); res.CancelAt(time.Now()) })

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	require.ErrorIs(t, rl.WaitTimeout(ctx, "resource", time.Millisecond), context.Canceled)
}

func TestInitSentinelLoadsRulesFromConfig(t *testing.T) {
	original := sentinelInitDefault
	sentinelInitDefault = func() error { return nil }
	defer func() { sentinelInitDefault = original }()

	_, _ = isolation.LoadRules(nil)
	_, _ = sentinelcb.LoadRules(nil)

	err := initSentinel(&resiliencecontract.CircuitBreakerConfig{
		ResourceConfigs: map[string]resiliencecontract.ResourceConfig{
			"order.create": {
				Threshold:             0.7,
				MinRequestCount:       20,
				MaxConcurrentRequests: 8,
				Timeout:               3 * time.Second,
				Interval:              2 * time.Second,
			},
		},
		DefaultConfig: resiliencecontract.ResourceConfig{
			Threshold:             0.5,
			MinRequestCount:       10,
			MaxConcurrentRequests: 5,
			Timeout:               5 * time.Second,
			Interval:              1 * time.Second,
		},
	})
	require.NoError(t, err)

	isoRules := isolation.GetRules()
	require.NotEmpty(t, isoRules)
	require.Equal(t, "order.create", isoRules[0].Resource)
	require.EqualValues(t, 8, isoRules[0].Threshold)

	cbRules := sentinelcb.GetRules()
	require.NotEmpty(t, cbRules)
	require.Equal(t, "order.create", cbRules[0].Resource)
	require.EqualValues(t, 20, cbRules[0].MinRequestAmount)
	require.EqualValues(t, 3000, cbRules[0].RetryTimeoutMs)
}

func TestBuildCircuitBreakerRuleFallsBackToDefaultConfig(t *testing.T) {
	rule := buildCircuitBreakerRule("order.pay", resiliencecontract.ResourceConfig{}, resiliencecontract.ResourceConfig{
		Threshold:       0.6,
		MinRequestCount: 12,
		Timeout:         4 * time.Second,
		Interval:        1500 * time.Millisecond,
	})
	require.NotNil(t, rule)
	require.Equal(t, "order.pay", rule.Resource)
	require.EqualValues(t, 12, rule.MinRequestAmount)
	require.EqualValues(t, 4000, rule.RetryTimeoutMs)
	require.EqualValues(t, 1500, rule.StatIntervalMs)
	require.Equal(t, 0.6, rule.Threshold)
}

func TestInitSentinelReturnsInitError(t *testing.T) {
	original := sentinelInitDefault
	sentinelInitDefault = func() error { return errors.New("init failed") }
	defer func() { sentinelInitDefault = original }()

	err := initSentinel(&resiliencecontract.CircuitBreakerConfig{
		ResourceConfigs: map[string]resiliencecontract.ResourceConfig{},
		DefaultConfig:   resiliencecontract.ResourceConfig{},
	})
	require.EqualError(t, err, "init failed")
}

func TestBuildCircuitBreakerRuleReturnsNilWithoutEffectiveThresholds(t *testing.T) {
	rule := buildCircuitBreakerRule("order.empty", resiliencecontract.ResourceConfig{}, resiliencecontract.ResourceConfig{})
	require.Nil(t, rule)
}

func TestBuildCircuitBreakerRuleUsesDefaultThresholdWhenOnlyThresholdMissing(t *testing.T) {
	rule := buildCircuitBreakerRule("order.partial", resiliencecontract.ResourceConfig{
		MinRequestCount: 9,
		Timeout:         2 * time.Second,
		Interval:        time.Second,
	}, resiliencecontract.ResourceConfig{
		Threshold: 0.55,
	})
	require.NotNil(t, rule)
	require.Equal(t, 0.55, rule.Threshold)
	require.EqualValues(t, 9, rule.MinRequestAmount)
}
