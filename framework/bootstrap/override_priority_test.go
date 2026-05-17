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
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	securitycontract "github.com/ngq/gorp/framework/contract/security"
)

// =============================================================================
// Override 优先级链路
// =============================================================================

// TestCodeDisableOverridesConfigEnableForSameFeature 验证代码侧 WithGovernanceDisabled 与配置启用冲突时，代码关闭优先
func TestCodeDisableOverridesConfigEnableForSameFeature(t *testing.T) {
	app := framework.NewApplication()
	c := app.Container()
	// 配置启用了 tracing，但代码侧显式关闭了 tracing
	cfg := &selectorConfigStub{values: map[string]any{
		"governance.mode":      "micro",
		"tracing.enabled":      true,
		"tracing.backend":      "otel",
		"service_auth.enabled": true,
	}}
	c.Bind(datacontract.ConfigKey, func(runtimecontract.Container) (any, error) {
		return cfg, nil
	}, true)

	// 代码侧显式关闭 tracing（模拟 WithGovernanceDisabled("tracing")）
	if err := registerSelectedMicroserviceProvidersWithOptions(c, "micro", []string{"tracing"}, nil, nil); err != nil {
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
		"governance.mode":                  "micro",
		"governance.providers.serviceauth": "mtls",
		"service_auth.enabled":             true,
	}}
	c.Bind(datacontract.ConfigKey, func(runtimecontract.Container) (any, error) {
		return cfg, nil
	}, true)

	// 代码侧 WithGovernanceProvider("serviceauth", "noop") 覆盖配置的 mtls
	if err := registerSelectedMicroserviceProvidersWithOptions(c, "micro", nil, nil, map[string]string{"serviceauth": "noop"}); err != nil {
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
	noopCfg := &selectorConfigStub{values: map[string]any{"governance.mode": "mono"}}
	if got := SelectTracingProviderWithMode(noopCfg, resiliencecontract.GovernanceModeMono).Name(); got != "tracing.noop" {
		t.Fatalf("priority 4 (provider fallback): expected tracing.noop, got %s", got)
	}

	// 级别3：模式默认值 —— microservice 模式下 tracing 默认为 otel
	modeCfg := &selectorConfigStub{values: map[string]any{"governance.mode": "micro"}}
	if got := SelectTracingProviderWithMode(modeCfg, resiliencecontract.GovernanceModeMicro).Name(); got != "tracing.otel" {
		t.Fatalf("priority 3 (mode default): expected tracing.otel, got %s", got)
	}

	// 级别2：配置显式覆盖 —— 配置中 governance.providers.tracing = noop 优先于模式默认
	configOverrideCfg := &selectorConfigStub{values: map[string]any{
		"governance.mode":              "micro",
		"governance.providers.tracing": "noop",
	}}
	if got := SelectTracingProviderWithMode(configOverrideCfg, resiliencecontract.GovernanceModeMicro).Name(); got != "tracing.noop" {
		t.Fatalf("priority 2 (config override): expected tracing.noop, got %s", got)
	}

	// 级别1：代码显式覆盖 —— 通过 overlay 注入的代码覆盖优先于配置
	overlayCfg := overlayGovernanceConfig(configOverrideCfg, nil, nil, map[string]string{"tracing": "otel"})
	if got := SelectTracingProviderWithMode(overlayCfg, resiliencecontract.GovernanceModeMicro).Name(); got != "tracing.otel" {
		t.Fatalf("priority 1 (code override): expected tracing.otel, got %s", got)
	}
}
