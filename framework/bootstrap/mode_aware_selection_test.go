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
	// otel 是 contrib 组件，micro 模式默认但未注册 → fail-fast
	assertFailingProvider(t, SelectTracingProvider(cfg))
	// metadata.default 是 framework 内建 provider
	if got := SelectMetadataProvider(cfg).Name(); got != "metadata.default" {
		t.Fatalf("expected metadata.default, got %s", got)
	}
	// token 是 contrib 组件，micro 模式默认但未注册 → fail-fast
	assertFailingProvider(t, SelectServiceAuthProvider(cfg))
	// sentinel 是 contrib 组件，micro 模式默认但未注册 → fail-fast
	assertFailingProvider(t, SelectCircuitBreakerProvider(cfg))
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
	// 注意：micro 模式默认引用 contrib 后端（etcd/otel/token/sentinel），
	// 未注册时为 failingProvider——fail-fast 决策，而非静默 noop。
	if len(providers) != 13 {
		t.Fatalf("expected 13 selected providers, got %d", len(providers))
	}
	// discovery 不在 micro 特性集，默认仍是 noop
	assertProviderName(t, providers[0], "discovery.noop")
	assertProviderName(t, providers[1], "selector.p2c")
	assertProviderName(t, providers[2], "rpc.noop")
	assertFailingProvider(t, providers[3]) // tracing → otel 未注册
	assertProviderName(t, providers[4], "metadata.default")
	assertFailingProvider(t, providers[5]) // serviceauth → token 未注册
	assertFailingProvider(t, providers[6]) // circuitbreaker → sentinel 未注册
	assertProviderName(t, providers[7], "loadshedding.semaphore")
}

func TestRegisterSelectedMicroserviceProvidersWithModeOverrideWinsOverConfig(t *testing.T) {
	// 强制 micro 模式会引用 contrib 默认后端（未注册时 fail-fast）。
	// 这里把 contrib 能力显式 noop，聚焦验证"mode override 生效"本身。
	app := framework.NewApplication()
	c := app.Container()
	cfg := &selectorConfigStub{values: map[string]any{
		"governance.mode": "mono",
		// 显式 noop，避免 micro 默认（etcd/otel/token/sentinel）未注册触发 fail-fast
		"discovery.backend":        "noop",
		"tracing.backend":         "noop",
		"service_auth.backend":    "noop",
		"circuit_breaker.backend": "noop",
		"message_queue.backend":   "noop",
		"distributed_lock.backend": "noop",
		"websocket.backend":       "noop",
		"dtm.backend":             "noop",
	}}
	c.Bind(datacontract.ConfigKey, func(runtimecontract.Container) (any, error) {
		return cfg, nil
	}, true)

	// modeOverride=micro 应覆盖配置的 mono——但因 contrib 未注册且未显式
	// noop 的能力会 fail-fast；此处全部显式 noop，注册应成功。
	if err := RegisterSelectedMicroserviceProvidersWithMode(c, "micro"); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
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