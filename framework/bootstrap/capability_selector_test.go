package bootstrap

import (
	"context"
	"errors"
	"testing"

	"github.com/ngq/gorp/framework"
	datacontract "github.com/ngq/gorp/framework/contract/data"
	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	securitycontract "github.com/ngq/gorp/framework/contract/security"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
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
func (s *selectorConfigStub) Watch(ctx context.Context, key string) (datacontract.ConfigWatcher, error) {
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
	if got := SelectDiscoveryProvider(cfg).Name(); got != "registry.eureka" {
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
	reloads           int
	valuesAfterReload map[string]any
	reloadErr         error
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
	c.Bind(datacontract.ConfigKey, func(runtimecontract.Container) (any, error) {
		return cfg, nil
	}, true)

	if err := RegisterSelectedMicroserviceProviders(c); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if cfg.reloads != 1 {
		t.Fatalf("expected reload once, got %d", cfg.reloads)
	}

	assertBoundKey(t, c, transportcontract.RPCRegistryKey)
	assertBoundKey(t, c, observabilitycontract.TracerKey)
	assertBoundKey(t, c, securitycontract.ServiceAuthKey)
	assertKeyRegistered(t, c, integrationcontract.MessagePublisherKey)
	assertKeyRegistered(t, c, datacontract.DistributedLockKey)
}

func TestRegisterSelectedMicroserviceProviders_DoesNotReloadLocalOrNoopConfigSource(t *testing.T) {
	for _, backend := range []string{"local", "noop"} {
		app := framework.NewApplication()
		c := app.Container()
		cfg := &reloadingConfigStub{selectorConfigStub: selectorConfigStub{values: map[string]any{
			"configsource.backend": backend,
		}}}
		c.Bind(datacontract.ConfigKey, func(runtimecontract.Container) (any, error) {
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
	c.Bind(datacontract.ConfigKey, func(runtimecontract.Container) (any, error) {
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

func TestDefaultGovernanceProviderDefaultsRemainStable(t *testing.T) {
	monolith := DefaultGovernanceProviderDefaults(resiliencecontract.GovernanceModeMonolith)
	if monolith.ConfigSource != "local" || monolith.Discovery != "noop" || monolith.Selector != "noop" {
		t.Fatalf("unexpected monolith defaults: %+v", monolith)
	}
	if monolith.Tracing != "noop" || monolith.Metadata != "noop" || monolith.ServiceAuth != "noop" || monolith.CircuitBreaker != "noop" {
		t.Fatalf("unexpected monolith governance protection defaults: %+v", monolith)
	}

	ginFirst := DefaultGovernanceProviderDefaults(resiliencecontract.GovernanceModeGinFirst)
	if ginFirst != monolith {
		t.Fatalf("expected gin-first defaults to match monolith for now, got %+v vs %+v", ginFirst, monolith)
	}

	microservice := DefaultGovernanceProviderDefaults(resiliencecontract.GovernanceModeMicroservice)
	if microservice.ConfigSource != "local" || microservice.Discovery != "noop" || microservice.RPC != "noop" {
		t.Fatalf("unexpected microservice transport defaults: %+v", microservice)
	}
	if microservice.Selector != "p2c_ewma" || microservice.Tracing != "otel" || microservice.Metadata != "default" || microservice.ServiceAuth != "token" || microservice.CircuitBreaker != "sentinel" {
		t.Fatalf("unexpected microservice governance defaults: %+v", microservice)
	}
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

func TestDetectGovernanceModeDefaultsToMonolith(t *testing.T) {
	cfg := &selectorConfigStub{values: map[string]any{}}
	if got := DetectGovernanceMode(cfg); got != resiliencecontract.GovernanceModeMonolith {
		t.Fatalf("expected monolith mode, got %q", got)
	}
}

func TestDetectGovernanceModeAcceptsLegacyServiceModeKey(t *testing.T) {
	cfg := &selectorConfigStub{values: map[string]any{"service.mode": "microservice"}}
	if got := DetectGovernanceMode(cfg); got != resiliencecontract.GovernanceModeMicroservice {
		t.Fatalf("expected microservice mode from service.mode, got %q", got)
	}
}

func TestDetectGovernanceModeAcceptsGinFirst(t *testing.T) {
	cfg := &selectorConfigStub{values: map[string]any{"governance.mode": "gin-first"}}
	if got := DetectGovernanceMode(cfg); got != resiliencecontract.GovernanceModeGinFirst {
		t.Fatalf("expected gin-first mode from governance.mode, got %q", got)
	}
}

func TestNormalizeGovernanceModeFallsBackToMonolith(t *testing.T) {
	if got := NormalizeGovernanceMode(""); got != resiliencecontract.GovernanceModeMonolith {
		t.Fatalf("expected empty mode to normalize to monolith, got %q", got)
	}
	if got := NormalizeGovernanceMode("unknown"); got != resiliencecontract.GovernanceModeMonolith {
		t.Fatalf("expected unknown mode to normalize to monolith, got %q", got)
	}
}

func TestNormalizeGovernanceModePreservesGinFirst(t *testing.T) {
	if got := NormalizeGovernanceMode(resiliencecontract.GovernanceModeGinFirst); got != resiliencecontract.GovernanceModeGinFirst {
		t.Fatalf("expected gin-first mode preserved, got %q", got)
	}
	if !IsGinFirstMode(resiliencecontract.GovernanceModeGinFirst) {
		t.Fatal("expected IsGinFirstMode to report true for gin-first")
	}
	if IsMicroserviceMode(resiliencecontract.GovernanceModeGinFirst) {
		t.Fatal("expected gin-first not to be treated as microservice mode")
	}
}

func TestModeAwareSelectionsPromoteMicroserviceDefaults(t *testing.T) {
	cfg := &selectorConfigStub{values: map[string]any{"governance.mode": "microservice"}}

	if got := SelectSelectorProvider(cfg).Name(); got != "selector.p2c" {
		t.Fatalf("expected selector.p2c provider implementation, got %s", got)
	}
	if got := SelectTracingProvider(cfg).Name(); got != "tracing.otel" {
		t.Fatalf("expected tracing.otel, got %s", got)
	}
	if got := SelectMetadataProvider(cfg).Name(); got != "metadata.default" {
		t.Fatalf("expected metadata.default, got %s", got)
	}
	if got := SelectServiceAuthProvider(cfg).Name(); got != "serviceauth.token" {
		t.Fatalf("expected serviceauth.token, got %s", got)
	}
	if got := SelectCircuitBreakerProvider(cfg).Name(); got != "circuitbreaker.sentinel" {
		t.Fatalf("expected circuitbreaker.sentinel, got %s", got)
	}
}

func TestModeAwareSelectionsKeepMonolithDefaultsWithoutExplicitEnablement(t *testing.T) {
	cfg := &selectorConfigStub{values: map[string]any{"governance.mode": "monolith"}}

	if got := SelectSelectorProvider(cfg).Name(); got != "selector.noop" {
		t.Fatalf("expected selector.noop, got %s", got)
	}
	if got := SelectTracingProvider(cfg).Name(); got != "tracing.noop" {
		t.Fatalf("expected tracing.noop, got %s", got)
	}
	if got := SelectMetadataProvider(cfg).Name(); got != "metadata.noop" {
		t.Fatalf("expected metadata.noop, got %s", got)
	}
	if got := SelectServiceAuthProvider(cfg).Name(); got != "serviceauth.noop" {
		t.Fatalf("expected serviceauth.noop, got %s", got)
	}
	if got := SelectCircuitBreakerProvider(cfg).Name(); got != "circuitbreaker.noop" {
		t.Fatalf("expected circuitbreaker.noop, got %s", got)
	}
}

func TestModeAwareSelectionsKeepGinFirstOnLightweightDefaults(t *testing.T) {
	cfg := &selectorConfigStub{values: map[string]any{"governance.mode": "gin-first"}}

	if got := SelectSelectorProvider(cfg).Name(); got != "selector.noop" {
		t.Fatalf("expected selector.noop, got %s", got)
	}
	if got := SelectTracingProvider(cfg).Name(); got != "tracing.noop" {
		t.Fatalf("expected tracing.noop, got %s", got)
	}
	if got := SelectMetadataProvider(cfg).Name(); got != "metadata.noop" {
		t.Fatalf("expected metadata.noop, got %s", got)
	}
	if got := SelectServiceAuthProvider(cfg).Name(); got != "serviceauth.noop" {
		t.Fatalf("expected serviceauth.noop, got %s", got)
	}
	if got := SelectCircuitBreakerProvider(cfg).Name(); got != "circuitbreaker.noop" {
		t.Fatalf("expected circuitbreaker.noop, got %s", got)
	}
}

func TestModeAwareSelectionsRespectExplicitBackendsOverGovernanceDefaults(t *testing.T) {
	cfg := &selectorConfigStub{values: map[string]any{
		"governance.mode":         "microservice",
		"selector.backend":        "noop",
		"tracing.backend":         "noop",
		"metadata.backend":        "noop",
		"service_auth.backend":    "noop",
		"circuit_breaker.backend": "noop",
	}}

	if got := SelectSelectorProvider(cfg).Name(); got != "selector.noop" {
		t.Fatalf("expected explicit selector.noop, got %s", got)
	}
	if got := SelectTracingProvider(cfg).Name(); got != "tracing.noop" {
		t.Fatalf("expected explicit tracing.noop, got %s", got)
	}
	if got := SelectMetadataProvider(cfg).Name(); got != "metadata.noop" {
		t.Fatalf("expected explicit metadata.noop, got %s", got)
	}
	if got := SelectServiceAuthProvider(cfg).Name(); got != "serviceauth.noop" {
		t.Fatalf("expected explicit serviceauth.noop, got %s", got)
	}
	if got := SelectCircuitBreakerProvider(cfg).Name(); got != "circuitbreaker.noop" {
		t.Fatalf("expected explicit circuitbreaker.noop, got %s", got)
	}
}

func TestModeAwareSelectionsRespectGovernanceProviderOverrides(t *testing.T) {
	cfg := &selectorConfigStub{values: map[string]any{
		"governance.mode":                  "microservice",
		"governance.providers.selector":    "noop",
		"governance.providers.tracing":     "noop",
		"governance.providers.metadata":    "noop",
		"governance.providers.serviceauth": "noop",
	}}

	if got := SelectSelectorProvider(cfg).Name(); got != "selector.noop" {
		t.Fatalf("expected governance override selector.noop, got %s", got)
	}
	if got := SelectTracingProvider(cfg).Name(); got != "tracing.noop" {
		t.Fatalf("expected governance override tracing.noop, got %s", got)
	}
	if got := SelectMetadataProvider(cfg).Name(); got != "metadata.noop" {
		t.Fatalf("expected governance override metadata.noop, got %s", got)
	}
	if got := SelectServiceAuthProvider(cfg).Name(); got != "serviceauth.noop" {
		t.Fatalf("expected governance override serviceauth.noop, got %s", got)
	}
}

func TestModeAwareSelectionsRespectGovernanceDisableList(t *testing.T) {
	cfg := &selectorConfigStub{values: map[string]any{
		"governance.mode":    "microservice",
		"governance.disable": []string{"tracing", "selector", "metadata", "serviceauth", "circuitbreaker"},
	}}

	if got := SelectSelectorProvider(cfg).Name(); got != "selector.noop" {
		t.Fatalf("expected disabled selector to fall back to noop, got %s", got)
	}
	if got := SelectTracingProvider(cfg).Name(); got != "tracing.noop" {
		t.Fatalf("expected disabled tracing to fall back to noop, got %s", got)
	}
	if got := SelectMetadataProvider(cfg).Name(); got != "metadata.noop" {
		t.Fatalf("expected disabled metadata to fall back to noop, got %s", got)
	}
	if got := SelectServiceAuthProvider(cfg).Name(); got != "serviceauth.noop" {
		t.Fatalf("expected disabled serviceauth to fall back to noop, got %s", got)
	}
	if got := SelectCircuitBreakerProvider(cfg).Name(); got != "circuitbreaker.noop" {
		t.Fatalf("expected disabled circuit breaker to fall back to noop, got %s", got)
	}
}

func TestSelectedMicroserviceProvidersPromoteMicroserviceDefaults(t *testing.T) {
	cfg := &selectorConfigStub{values: map[string]any{"governance.mode": "microservice"}}
	providers := SelectedMicroserviceProviders(cfg)
	if len(providers) != 10 {
		t.Fatalf("expected 10 selected providers, got %d", len(providers))
	}
	assertProviderName(t, providers[0], "discovery.noop")
	assertProviderName(t, providers[1], "selector.p2c")
	assertProviderName(t, providers[2], "rpc.noop")
	assertProviderName(t, providers[3], "tracing.otel")
	assertProviderName(t, providers[4], "metadata.default")
	assertProviderName(t, providers[5], "serviceauth.token")
	assertProviderName(t, providers[6], "circuitbreaker.sentinel")
}

func TestRegisterSelectedMicroserviceProvidersWithModeOverrideWinsOverConfig(t *testing.T) {
	app := framework.NewApplication()
	c := app.Container()
	cfg := &selectorConfigStub{values: map[string]any{"governance.mode": "monolith"}}
	c.Bind(datacontract.ConfigKey, func(runtimecontract.Container) (any, error) {
		return cfg, nil
	}, true)

	if err := RegisterSelectedMicroserviceProvidersWithMode(c, "microservice"); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	assertBoundKey(t, c, observabilitycontract.TracerKey)
	assertBoundKey(t, c, securitycontract.ServiceAuthKey)
	assertBoundKey(t, c, resiliencecontract.CircuitBreakerKey)
}

// TestCodeDisableOverridesConfigEnableForSameFeature 验证代码侧 WithGovernanceDisabled 与配置启用冲突时，代码关闭优先
func TestCodeDisableOverridesConfigEnableForSameFeature(t *testing.T) {
	app := framework.NewApplication()
	c := app.Container()
	// 配置启用了 tracing，但代码侧显式关闭了 tracing
	cfg := &selectorConfigStub{values: map[string]any{
		"governance.mode":  "microservice",
		"tracing.enabled":  true,
		"tracing.backend":  "otel",
		"service_auth.enabled": true,
	}}
	c.Bind(datacontract.ConfigKey, func(runtimecontract.Container) (any, error) {
		return cfg, nil
	}, true)

	// 代码侧显式关闭 tracing（模拟 WithGovernanceDisabled("tracing")）
	if err := registerSelectedMicroserviceProvidersWithOptions(c, "microservice", []string{"tracing"}, nil); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	// tracing 应该被关闭（noop），即使配置中 enabled=true
	tracer, err := c.Make(observabilitycontract.TracerKey)
	if err != nil {
		t.Fatalf("expected tracer to be bound, got error: %v", err)
	}
	if tracerName, ok := tracer.(interface{ Name() string }); ok {
		if tracerName.Name() != "noop" {
			t.Fatalf("expected noop tracer when code disables tracing, got %s", tracerName.Name())
		}
	}

	// serviceauth 没有被代码关闭，应该仍走配置的 token 模式
	assertBoundKey(t, c, securitycontract.ServiceAuthKey)
}

// TestCodeProviderOverrideWinsOverConfigProviderOverrideForSameKey 验证代码侧 WithGovernanceProvider 与配置冲突时，代码覆盖优先
func TestCodeProviderOverrideWinsOverConfigProviderOverrideForSameKey(t *testing.T) {
	app := framework.NewApplication()
	c := app.Container()
	// 配置中 governance.providers.serviceauth 设为 mtls
	cfg := &selectorConfigStub{values: map[string]any{
		"governance.mode":                   "microservice",
		"governance.providers.serviceauth":   "mtls",
		"service_auth.enabled":               true,
	}}
	c.Bind(datacontract.ConfigKey, func(runtimecontract.Container) (any, error) {
		return cfg, nil
	}, true)

	// 代码侧 WithGovernanceProvider("serviceauth", "noop") 覆盖配置的 mtls
	if err := registerSelectedMicroserviceProvidersWithOptions(c, "microservice", nil, map[string]string{"serviceauth": "noop"}); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	// serviceauth 应该走代码侧的 noop，而非配置的 mtls
	auth, err := c.Make(securitycontract.ServiceAuthKey)
	if err != nil {
		t.Fatalf("expected serviceauth to be bound, got error: %v", err)
	}
	if authName, ok := auth.(interface{ Name() string }); ok {
		if authName.Name() != "noop" {
			t.Fatalf("expected noop serviceauth when code overrides to noop, got %s", authName.Name())
		}
	}
}

// TestOverridePriorityChainForSingleFeature 验证单个治理能力上的完整优先级链：
// 代码显式覆盖 > 配置显式覆盖 > 模式默认值 > provider 兜底
func TestOverridePriorityChainForSingleFeature(t *testing.T) {
	// 级别4：provider 兜底 —— monolith 模式下 tracing 默认为 noop
	noopCfg := &selectorConfigStub{values: map[string]any{"governance.mode": "monolith"}}
	if got := SelectTracingProviderWithMode(noopCfg, resiliencecontract.GovernanceModeMonolith).Name(); got != "tracing.noop" {
		t.Fatalf("priority 4 (provider fallback): expected tracing.noop, got %s", got)
	}

	// 级别3：模式默认值 —— microservice 模式下 tracing 默认为 otel
	modeCfg := &selectorConfigStub{values: map[string]any{"governance.mode": "microservice"}}
	if got := SelectTracingProviderWithMode(modeCfg, resiliencecontract.GovernanceModeMicroservice).Name(); got != "tracing.otel" {
		t.Fatalf("priority 3 (mode default): expected tracing.otel, got %s", got)
	}

	// 级别2：配置显式覆盖 —— 配置中 governance.providers.tracing = noop 优先于模式默认
	configOverrideCfg := &selectorConfigStub{values: map[string]any{
		"governance.mode":                "microservice",
		"governance.providers.tracing":   "noop",
	}}
	if got := SelectTracingProviderWithMode(configOverrideCfg, resiliencecontract.GovernanceModeMicroservice).Name(); got != "tracing.noop" {
		t.Fatalf("priority 2 (config override): expected tracing.noop, got %s", got)
	}

	// 级别1：代码显式覆盖 —— 通过 overlay 注入的代码覆盖优先于配置
	overlayCfg := overlayGovernanceConfig(configOverrideCfg, nil, map[string]string{"tracing": "otel"})
	if got := SelectTracingProviderWithMode(overlayCfg, resiliencecontract.GovernanceModeMicroservice).Name(); got != "tracing.otel" {
		t.Fatalf("priority 1 (code override): expected tracing.otel, got %s", got)
	}
}

func assertKeyRegistered(t *testing.T, c runtimecontract.Container, key string) {
	t.Helper()
	if !c.IsBind(key) {
		t.Fatalf("expected key %s to be registered", key)
	}
}

func assertBoundKey(t *testing.T, c runtimecontract.Container, key string) {
	t.Helper()
	if !c.IsBind(key) {
		t.Fatalf("expected key %s to be bound", key)
	}
	if _, err := c.Make(key); err != nil {
		t.Fatalf("expected key %s to be resolvable: %v", key, err)
	}
}

func assertProviderName(t *testing.T, provider runtimecontract.ServiceProvider, expected string) {
	t.Helper()
	if provider == nil {
		t.Fatalf("expected provider %s, got nil", expected)
	}
	if got := provider.Name(); got != expected {
		t.Fatalf("expected provider %s, got %s", expected, got)
	}
}
