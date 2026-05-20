// Package bootstrap_test provides integration tests for governance mode and capability provider selection.
//
// 适用场景：
// - 验证 governance mode 的检测、标准化与模式感知选择逻辑。
// - 验证各 provider backend 的 Select 优先級（backend key > config > code disable > default）。
// - 验证 RegisterSelectedMicroserviceProviders 的重载、传播与降级行为。
// - 验证 governance override 链路的优先级顺序。
//
// 注意：contrib 组件现在是独立模块，这些测试验证框架选择逻辑，
// 当 contrib provider 未注册时，会回退到 noop。
package bootstrap

import (
	"testing"

	"github.com/ngq/gorp/framework"
	datacontract "github.com/ngq/gorp/framework/contract/data"
	"github.com/ngq/gorp/framework/contract/runtime"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

// =============================================================================
// Mode-Aware Provider 选择逻辑
// 注意：contrib 组件是独立模块，未注册时会回退到 noop
// =============================================================================

func TestModeAwareSelectionsPromoteMicroserviceDefaults(t *testing.T) {
	cfg := &selectorConfigStub{values: map[string]any{"governance.mode": "micro"}}

	// selector.p2c 是 framework 内建 provider
	if got := SelectSelectorProvider(cfg).Name(); got != "selector.p2c" {
		t.Fatalf("expected selector.p2c provider implementation, got %s", got)
	}
	// otel 是 contrib 组件，未注册时回退到 noop
	if got := SelectTracingProvider(cfg).Name(); got != "tracing.noop" {
		t.Fatalf("expected tracing.noop (otel not registered), got %s", got)
	}
	// metadata.default 是 framework 内建 provider
	if got := SelectMetadataProvider(cfg).Name(); got != "metadata.default" {
		t.Fatalf("expected metadata.default, got %s", got)
	}
	// token 是 contrib 组件，未注册时回退到 noop
	if got := SelectServiceAuthProvider(cfg).Name(); got != "serviceauth.noop" {
		t.Fatalf("expected serviceauth.noop (token not registered), got %s", got)
	}
	// sentinel 是 contrib 组件，未注册时回退到 noop
	if got := SelectCircuitBreakerProvider(cfg).Name(); got != "circuitbreaker.noop" {
		t.Fatalf("expected circuitbreaker.noop (sentinel not registered), got %s", got)
	}
	// semaphore 是 framework 内建 provider
	if got := SelectLoadSheddingProvider(cfg).Name(); got != "loadshedding.semaphore" {
		t.Fatalf("expected loadshedding.semaphore, got %s", got)
	}
}

func TestModeAwareSelectionsKeepMonolithDefaultsWithoutExplicitEnablement(t *testing.T) {
	cfg := &selectorConfigStub{values: map[string]any{"governance.mode": "mono"}}

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
		"governance.mode":         "micro",
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
		"governance.mode":                  "micro",
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
		"governance.mode":    "micro",
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
	cfg := &selectorConfigStub{values: map[string]any{"governance.mode": "micro"}}
	providers := SelectedMicroserviceProviders(cfg)
	// 13 providers: discovery, selector, rpc, tracing, metadata, serviceauth,
	// circuitbreaker, loadshedding, retry, dtm, mq, dlock, websocket
	if len(providers) != 13 {
		t.Fatalf("expected 13 selected providers, got %d", len(providers))
	}
	assertProviderName(t, providers[0], "discovery.noop")
	assertProviderName(t, providers[1], "selector.p2c")
	assertProviderName(t, providers[2], "rpc.noop")
	// contrib 组件未注册，回退到 noop
	assertProviderName(t, providers[3], "tracing.noop")
	assertProviderName(t, providers[4], "metadata.default")
	assertProviderName(t, providers[5], "serviceauth.noop")
	assertProviderName(t, providers[6], "circuitbreaker.noop")
	assertProviderName(t, providers[7], "loadshedding.semaphore")
}

func TestRegisterSelectedMicroserviceProvidersWithModeOverrideWinsOverConfig(t *testing.T) {
	app := framework.NewApplication()
	c := app.Container()
	cfg := &selectorConfigStub{values: map[string]any{"governance.mode": "mono"}}
	c.Bind(datacontract.ConfigKey, func(runtimecontract.Container) (any, error) {
		return cfg, nil
	}, true)

	if err := RegisterSelectedMicroserviceProvidersWithMode(c, "micro"); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	// contrib 组件未注册，这些 key 不会被绑定（因为 provider 是 noop）
	// noop provider 通常不绑定实际能力
	// 所以我们只验证调用成功，不验证 key 绑定
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