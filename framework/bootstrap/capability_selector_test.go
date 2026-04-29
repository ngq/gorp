package bootstrap

import (
	"context"
	"errors"
	"testing"

	"github.com/ngq/gorp/framework"
	"github.com/ngq/gorp/framework/contract"
)

type selectorConfigStub struct {
	values map[string]any
}

func (s *selectorConfigStub) Env() string        { return "test" }
func (s *selectorConfigStub) Get(key string) any { return s.values[key] }
func (s *selectorConfigStub) GetString(key string) string {
	if v, ok := s.values[key].(string); ok {
		return v
	}
	return ""
}
func (s *selectorConfigStub) GetInt(key string) int {
	if v, ok := s.values[key].(int); ok {
		return v
	}
	return 0
}
func (s *selectorConfigStub) GetBool(key string) bool {
	if v, ok := s.values[key].(bool); ok {
		return v
	}
	return false
}
func (s *selectorConfigStub) GetFloat(key string) float64 {
	if v, ok := s.values[key].(float64); ok {
		return v
	}
	return 0
}
func (s *selectorConfigStub) Unmarshal(key string, out any) error { return nil }
func (s *selectorConfigStub) Watch(ctx context.Context, key string) (contract.ConfigWatcher, error) {
	return nil, nil
}
func (s *selectorConfigStub) Reload(ctx context.Context) error { return nil }

func TestSelectConfigSourceProvider_PrefersBackendKey(t *testing.T) {
	cfg := &selectorConfigStub{values: map[string]any{"configsource.backend": "nacos"}}
	if got := SelectConfigSourceProvider(cfg).Name(); got != "configsource.nacos" {
		t.Fatalf("expected configsource.nacos, got %s", got)
	}
}

func TestSelectDiscoveryProvider_PrefersBackendKey(t *testing.T) {
	cfg := &selectorConfigStub{values: map[string]any{"discovery.backend": "eureka"}}
	if got := SelectDiscoveryProvider(cfg).Name(); got != "discovery.eureka" {
		t.Fatalf("expected discovery.eureka, got %s", got)
	}
}

func TestSelectRPCProvider_DefaultsToNoop(t *testing.T) {
	cfg := &selectorConfigStub{values: map[string]any{}}
	if got := SelectRPCProvider(cfg).Name(); got != "rpc.noop" {
		t.Fatalf("expected rpc.noop, got %s", got)
	}
}

func TestSelectConfigSourceProvider_FallsBackToLocal(t *testing.T) {
	cfg := &selectorConfigStub{values: map[string]any{"configsource.backend": "unknown"}}
	if got := SelectConfigSourceProvider(cfg).Name(); got != "configsource.local" {
		t.Fatalf("expected configsource.local fallback, got %s", got)
	}
}

func TestSelectDiscoveryProvider_FallsBackToNoop(t *testing.T) {
	cfg := &selectorConfigStub{values: map[string]any{"discovery.backend": "unknown"}}
	if got := SelectDiscoveryProvider(cfg).Name(); got != "discovery.noop" {
		t.Fatalf("expected discovery.noop fallback, got %s", got)
	}
}

func TestSelectMessageQueueProvider_FallsBackToNoop(t *testing.T) {
	cfg := &selectorConfigStub{values: map[string]any{"message_queue.backend": "unknown"}}
	if got := SelectMessageQueueProvider(cfg).Name(); got != "messagequeue.noop" {
		t.Fatalf("expected messagequeue.noop fallback, got %s", got)
	}
}

func TestSelectDistributedLockProvider_FallsBackToNoop(t *testing.T) {
	cfg := &selectorConfigStub{values: map[string]any{"distributed_lock.backend": "unknown"}}
	if got := SelectDistributedLockProvider(cfg).Name(); got != "dlock.noop" {
		t.Fatalf("expected dlock.noop fallback, got %s", got)
	}
}

func TestSelectCircuitBreakerProvider_AcceptsBackendAndEnabled(t *testing.T) {
	backendCfg := &selectorConfigStub{values: map[string]any{"circuit_breaker.backend": "sentinel"}}
	if got := SelectCircuitBreakerProvider(backendCfg).Name(); got != "circuitbreaker.sentinel" {
		t.Fatalf("expected circuitbreaker.sentinel, got %s", got)
	}

	enabledCfg := &selectorConfigStub{values: map[string]any{"circuit_breaker.enabled": true}}
	if got := SelectCircuitBreakerProvider(enabledCfg).Name(); got != "circuitbreaker.sentinel" {
		t.Fatalf("expected enabled config to select sentinel, got %s", got)
	}

	noopCfg := &selectorConfigStub{values: map[string]any{"circuit_breaker.backend": "noop"}}
	if got := SelectCircuitBreakerProvider(noopCfg).Name(); got != "circuitbreaker.noop" {
		t.Fatalf("expected circuitbreaker.noop, got %s", got)
	}
}

func TestSelectDTMProvider_AcceptsBackendDriverAndEnabled(t *testing.T) {
	backendCfg := &selectorConfigStub{values: map[string]any{"dtm.backend": "dtmsdk"}}
	if got := SelectDTMProvider(backendCfg).Name(); got != "dtm.sdk" {
		t.Fatalf("expected dtm.sdk, got %s", got)
	}

	driverCfg := &selectorConfigStub{values: map[string]any{"dtm.driver": "sdk"}}
	if got := SelectDTMProvider(driverCfg).Name(); got != "dtm.sdk" {
		t.Fatalf("expected dtm.sdk via driver key, got %s", got)
	}

	enabledCfg := &selectorConfigStub{values: map[string]any{"dtm.enabled": true}}
	if got := SelectDTMProvider(enabledCfg).Name(); got != "dtm.sdk" {
		t.Fatalf("expected enabled dtm to select sdk provider, got %s", got)
	}

	noopCfg := &selectorConfigStub{values: map[string]any{"dtm.backend": "noop"}}
	if got := SelectDTMProvider(noopCfg).Name(); got != "dtm.noop" {
		t.Fatalf("expected dtm.noop, got %s", got)
	}
}

func TestSelectTracingProvider_AcceptsEnabledAndBackends(t *testing.T) {
	backendCases := map[string]string{
		"otel":   "tracing.otel",
		"otlp":   "tracing.otel",
		"grpc":   "tracing.otel",
		"http":   "tracing.otel",
		"stdout": "tracing.otel",
		"noop":   "tracing.noop",
	}
	for backend, expected := range backendCases {
		cfg := &selectorConfigStub{values: map[string]any{"tracing.backend": backend}}
		if got := SelectTracingProvider(cfg).Name(); got != expected {
			t.Fatalf("backend %s expected %s, got %s", backend, expected, got)
		}
	}

	enabledCfg := &selectorConfigStub{values: map[string]any{"tracing.enabled": true}}
	if got := SelectTracingProvider(enabledCfg).Name(); got != "tracing.otel" {
		t.Fatalf("expected enabled tracing to select tracing.otel, got %s", got)
	}
}

func TestSelectMetadataProvider_AcceptsEnabledAndPrefix(t *testing.T) {
	enabledCfg := &selectorConfigStub{values: map[string]any{"metadata.enabled": true}}
	if got := SelectMetadataProvider(enabledCfg).Name(); got != "metadata.default" {
		t.Fatalf("expected enabled metadata to select metadata.default, got %s", got)
	}

	prefixCfg := &selectorConfigStub{values: map[string]any{"metadata.propagate_prefix": "x-"}}
	if got := SelectMetadataProvider(prefixCfg).Name(); got != "metadata.default" {
		t.Fatalf("expected propagate_prefix to select metadata.default, got %s", got)
	}

	noopCfg := &selectorConfigStub{values: map[string]any{"metadata.backend": "noop"}}
	if got := SelectMetadataProvider(noopCfg).Name(); got != "metadata.noop" {
		t.Fatalf("expected metadata.noop, got %s", got)
	}
}

func TestSelectServiceAuthProvider_AcceptsEnabledAndMode(t *testing.T) {
	enabledCfg := &selectorConfigStub{values: map[string]any{"service_auth.enabled": true}}
	if got := SelectServiceAuthProvider(enabledCfg).Name(); got != "serviceauth.token" {
		t.Fatalf("expected enabled serviceauth to select serviceauth.token, got %s", got)
	}

	mtlsCfg := &selectorConfigStub{values: map[string]any{"service_auth.mode": "mtls"}}
	if got := SelectServiceAuthProvider(mtlsCfg).Name(); got != "serviceauth.mtls" {
		t.Fatalf("expected serviceauth.mtls, got %s", got)
	}

	noopCfg := &selectorConfigStub{values: map[string]any{"service_auth.backend": "noop"}}
	if got := SelectServiceAuthProvider(noopCfg).Name(); got != "serviceauth.noop" {
		t.Fatalf("expected serviceauth.noop, got %s", got)
	}
}

func TestSelectMessageQueueProvider_AcceptsEnabledAndBackend(t *testing.T) {
	enabledCfg := &selectorConfigStub{values: map[string]any{"message_queue.enabled": true}}
	if got := SelectMessageQueueProvider(enabledCfg).Name(); got != "messagequeue.redis" {
		t.Fatalf("expected enabled message queue to select messagequeue.redis, got %s", got)
	}

	noopCfg := &selectorConfigStub{values: map[string]any{"message_queue.backend": "noop"}}
	if got := SelectMessageQueueProvider(noopCfg).Name(); got != "messagequeue.noop" {
		t.Fatalf("expected messagequeue.noop, got %s", got)
	}
}

func TestSelectDistributedLockProvider_AcceptsEnabledAndBackend(t *testing.T) {
	enabledCfg := &selectorConfigStub{values: map[string]any{"distributed_lock.enabled": true}}
	if got := SelectDistributedLockProvider(enabledCfg).Name(); got != "dlock.redis" {
		t.Fatalf("expected enabled distributed lock to select dlock.redis, got %s", got)
	}

	noopCfg := &selectorConfigStub{values: map[string]any{"distributed_lock.backend": "noop"}}
	if got := SelectDistributedLockProvider(noopCfg).Name(); got != "dlock.noop" {
		t.Fatalf("expected dlock.noop, got %s", got)
	}
}

type reloadingConfigStub struct {
		selectorConfigStub
		reloads int
		valuesAfterReload map[string]any
		reloadErr error
}

func (s *reloadingConfigStub) Reload(ctx context.Context) error {
		s.reloads++
		if s.reloadErr != nil {
			return s.reloadErr
		}
		for key, value := range s.valuesAfterReload {
			s.values[key] = value
		}
		return nil
}

func TestRegisterSelectedMicroserviceProviders_SkipsWithoutConfigBinding(t *testing.T) {
	app := framework.NewApplication()
	c := app.Container()
	if err := RegisterSelectedMicroserviceProviders(c); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestRegisterSelectedMicroserviceProviders_ReloadsRemoteConfigSourceBeforeSelectingOthers(t *testing.T) {
	app := framework.NewApplication()
	c := app.Container()
	cfg := &reloadingConfigStub{
		selectorConfigStub: selectorConfigStub{values: map[string]any{
			"configsource.backend": "consul",
		}},
		valuesAfterReload: map[string]any{
			"discovery.backend":        "consul",
			"tracing.enabled":          true,
			"service_auth.enabled":     true,
			"message_queue.enabled":    true,
			"distributed_lock.enabled": true,
		},
	}
	c.Bind(contract.ConfigKey, func(contract.Container) (any, error) {
		return cfg, nil
	}, true)

	if err := RegisterSelectedMicroserviceProviders(c); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if cfg.reloads != 1 {
		t.Fatalf("expected reload once, got %d", cfg.reloads)
	}

	assertBoundKey(t, c, contract.RPCRegistryKey)
	assertBoundKey(t, c, contract.TracerKey)
	assertBoundKey(t, c, contract.ServiceAuthKey)
	assertKeyRegistered(t, c, contract.MessagePublisherKey)
	assertKeyRegistered(t, c, contract.DistributedLockKey)
}

func TestRegisterSelectedMicroserviceProviders_DoesNotReloadLocalOrNoopConfigSource(t *testing.T) {
	for _, backend := range []string{"local", "noop"} {
		app := framework.NewApplication()
		c := app.Container()
		cfg := &reloadingConfigStub{selectorConfigStub: selectorConfigStub{values: map[string]any{
			"configsource.backend": backend,
		}}}
		c.Bind(contract.ConfigKey, func(contract.Container) (any, error) {
			return cfg, nil
		}, true)
		if err := RegisterSelectedMicroserviceProviders(c); err != nil {
			t.Fatalf("backend %s expected nil error, got %v", backend, err)
		}
		if cfg.reloads != 0 {
			t.Fatalf("backend %s expected no reload, got %d", backend, cfg.reloads)
		}
	}
}

func TestRegisterSelectedMicroserviceProviders_PropagatesReloadError(t *testing.T) {
	app := framework.NewApplication()
	c := app.Container()
	cfg := &reloadingConfigStub{
		selectorConfigStub: selectorConfigStub{values: map[string]any{"configsource.backend": "consul"}},
		reloadErr:          errors.New("reload failed"),
	}
	c.Bind(contract.ConfigKey, func(contract.Container) (any, error) {
		return cfg, nil
	}, true)

	err := RegisterSelectedMicroserviceProviders(c)
	if err == nil || err.Error() != "reload failed" {
		t.Fatalf("expected reload failed error, got %v", err)
	}
}

func TestRegistryFactories_ProvideExpectedFallbacks(t *testing.T) {
	assertProviderName(t, providerFromMap(configSourceProviderFactories, "", "local"), "configsource.local")
	assertProviderName(t, providerFromMap(discoveryProviderFactories, "", "noop"), "discovery.noop")
	assertProviderName(t, providerFromMap(selectorProviderFactories, "", "noop"), "selector.noop")
	assertProviderName(t, providerFromMap(rpcProviderFactories, "", "noop"), "rpc.noop")
	assertProviderName(t, providerFromMap(tracingProviderFactories, "", "noop"), "tracing.noop")
	assertProviderName(t, providerFromMap(metadataProviderFactories, "", "noop"), "metadata.noop")
	assertProviderName(t, providerFromMap(serviceAuthProviderFactories, "", "noop"), "serviceauth.noop")
	assertProviderName(t, providerFromMap(circuitBreakerProviderFactories, "", "noop"), "circuitbreaker.noop")
	assertProviderName(t, providerFromMap(dtmProviderFactories, "", "noop"), "dtm.noop")
	assertProviderName(t, providerFromMap(messageQueueProviderFactories, "", "noop"), "messagequeue.noop")
	assertProviderName(t, providerFromMap(distributedLockProviderFactories, "", "noop"), "dlock.noop")
}

func TestSelectedMicroserviceProviders_DefaultsMatchBootstrapExpectations(t *testing.T) {
	cfg := &selectorConfigStub{values: map[string]any{}}
	providers := SelectedMicroserviceProviders(cfg)
	if len(providers) != 10 {
		t.Fatalf("expected 10 selected providers, got %d", len(providers))
	}
	assertProviderName(t, providers[0], "discovery.noop")
	assertProviderName(t, providers[1], "selector.noop")
	assertProviderName(t, providers[2], "rpc.noop")
	assertProviderName(t, providers[3], "tracing.noop")
	assertProviderName(t, providers[4], "metadata.noop")
	assertProviderName(t, providers[5], "serviceauth.noop")
	assertProviderName(t, providers[6], "circuitbreaker.noop")
	assertProviderName(t, providers[7], "dtm.noop")
	assertProviderName(t, providers[8], "messagequeue.noop")
	assertProviderName(t, providers[9], "dlock.noop")
}

func assertKeyRegistered(t *testing.T, c contract.Container, key string) {
	t.Helper()
	if !c.IsBind(key) {
		t.Fatalf("expected key %s to be registered", key)
	}
}

func assertBoundKey(t *testing.T, c contract.Container, key string) {
	t.Helper()
	if !c.IsBind(key) {
		t.Fatalf("expected key %s to be bound", key)
	}
	if _, err := c.Make(key); err != nil {
		t.Fatalf("expected key %s to be resolvable: %v", key, err)
	}
}

func assertProviderName(t *testing.T, provider contract.ServiceProvider, expected string) {
	t.Helper()
	if provider == nil {
		t.Fatalf("expected provider %s, got nil", expected)
	}
	if got := provider.Name(); got != expected {
		t.Fatalf("expected provider %s, got %s", expected, got)
	}
}
