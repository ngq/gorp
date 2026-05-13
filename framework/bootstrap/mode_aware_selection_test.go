// Package bootstrap_test provides integration tests for governance mode and capability provider selection.
//
// 适用场景：
// - 验证 governance mode 的检测、标准化与模式感知选择逻辑。
// - 验证各 provider backend 的 Select 优先級（backend key > config > code disable > default）。
// - 验证 RegisterSelectedMicroserviceProviders 的重载、传播与降级行为。
// - 验证 governance override 链路的优先级顺序。
package bootstrap

import (
	"testing"

	"github.com/ngq/gorp/framework"
	datacontract "github.com/ngq/gorp/framework/contract/data"
	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	"github.com/ngq/gorp/framework/contract/runtime"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	securitycontract "github.com/ngq/gorp/framework/contract/security"
)

// =============================================================================
// Mode-Aware Provider 选择逻辑
// =============================================================================

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
	if got := SelectLoadSheddingProvider(cfg).Name(); got != "loadshedding.semaphore" {
		t.Fatalf("expected loadshedding.semaphore, got %s", got)
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
	if got := SelectLoadSheddingProvider(cfg).Name(); got != "loadshedding.noop" {
		t.Fatalf("expected loadshedding.noop, got %s", got)
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
	if got := SelectLoadSheddingProvider(cfg).Name(); got != "loadshedding.noop" {
		t.Fatalf("expected loadshedding.noop, got %s", got)
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
	if len(providers) != 11 {
		t.Fatalf("expected 11 selected providers, got %d", len(providers))
	}
	assertProviderName(t, providers[0], "discovery.noop")
	assertProviderName(t, providers[1], "selector.p2c")
	assertProviderName(t, providers[2], "rpc.noop")
	assertProviderName(t, providers[3], "tracing.otel")
	assertProviderName(t, providers[4], "metadata.default")
	assertProviderName(t, providers[5], "serviceauth.token")
	assertProviderName(t, providers[6], "circuitbreaker.sentinel")
	assertProviderName(t, providers[7], "loadshedding.semaphore")
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

func assertProviderName(t *testing.T, provider runtime.ServiceProvider, expected string) {
	t.Helper()
	if provider == nil {
		t.Fatalf("expected provider %s, got nil", expected)
	}
	if got := provider.Name(); got != expected {
		t.Fatalf("expected provider %s, got %s", expected, got)
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
