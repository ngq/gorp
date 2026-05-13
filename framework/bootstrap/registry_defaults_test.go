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

	"github.com/ngq/gorp/framework/contract/resilience"
)

// =============================================================================
// Registry Factory 降级与默认值稳定性
// =============================================================================

func TestRegistryFactories_ProvideExpectedFallbacks(t *testing.T) {
	assertProviderName(t, providerFromMap(configSourceProviderFactories, "", "local"), "configsource.local")
	assertProviderName(t, providerFromMap(discoveryProviderFactories, "", "noop"), "discovery.noop")
	assertProviderName(t, providerFromMap(selectorProviderFactories, "", "noop"), "selector.noop")
	assertProviderName(t, providerFromMap(rpcProviderFactories, "", "noop"), "rpc.noop")
	assertProviderName(t, providerFromMap(tracingProviderFactories, "", "noop"), "tracing.noop")
	assertProviderName(t, providerFromMap(metadataProviderFactories, "", "noop"), "metadata.noop")
	assertProviderName(t, providerFromMap(serviceAuthProviderFactories, "", "noop"), "serviceauth.noop")
	assertProviderName(t, providerFromMap(circuitBreakerProviderFactories, "", "noop"), "circuitbreaker.noop")
	assertProviderName(t, providerFromMap(loadShedderProviderFactories, "", "noop"), "loadshedding.noop")
	assertProviderName(t, providerFromMap(dtmProviderFactories, "", "noop"), "dtm.noop")
	assertProviderName(t, providerFromMap(messageQueueProviderFactories, "", "noop"), "messagequeue.noop")
	assertProviderName(t, providerFromMap(distributedLockProviderFactories, "", "noop"), "dlock.noop")
}

func TestDefaultGovernanceProviderDefaultsRemainStable(t *testing.T) {
	monolith := DefaultGovernanceProviderDefaults(resilience.GovernanceModeMonolith)
	if monolith.ConfigSource != "local" || monolith.Discovery != "noop" || monolith.Selector != "noop" {
		t.Fatalf("unexpected monolith defaults: %+v", monolith)
	}
	if monolith.Tracing != "noop" || monolith.Metadata != "noop" || monolith.ServiceAuth != "noop" || monolith.CircuitBreaker != "noop" {
		t.Fatalf("unexpected monolith governance protection defaults: %+v", monolith)
	}

	ginFirst := DefaultGovernanceProviderDefaults(resilience.GovernanceModeGinFirst)
	if ginFirst != monolith {
		t.Fatalf("expected gin-first defaults to match monolith for now, got %+v vs %+v", ginFirst, monolith)
	}

	microservice := DefaultGovernanceProviderDefaults(resilience.GovernanceModeMicroservice)
	if microservice.ConfigSource != "local" || microservice.Discovery != "noop" || microservice.RPC != "noop" {
		t.Fatalf("unexpected microservice transport defaults: %+v", microservice)
	}
	if microservice.Selector != "p2c_ewma" || microservice.Tracing != "otel" || microservice.Metadata != "default" || microservice.ServiceAuth != "token" || microservice.CircuitBreaker != "sentinel" || microservice.LoadShedder != "semaphore" {
		t.Fatalf("unexpected microservice governance defaults: %+v", microservice)
	}
}

func TestSelectedMicroserviceProviders_DefaultsMatchBootstrapExpectations(t *testing.T) {
	cfg := &selectorConfigStub{values: map[string]any{}}
	providers := SelectedMicroserviceProviders(cfg)
	if len(providers) != 11 {
		t.Fatalf("expected 11 selected providers, got %d", len(providers))
	}
	assertProviderName(t, providers[0], "discovery.noop")
	assertProviderName(t, providers[1], "selector.noop")
	assertProviderName(t, providers[2], "rpc.noop")
	assertProviderName(t, providers[3], "tracing.noop")
	assertProviderName(t, providers[4], "metadata.noop")
	assertProviderName(t, providers[5], "serviceauth.noop")
	assertProviderName(t, providers[6], "circuitbreaker.noop")
	assertProviderName(t, providers[7], "loadshedding.noop")
	assertProviderName(t, providers[8], "dtm.noop")
	assertProviderName(t, providers[9], "messagequeue.noop")
	assertProviderName(t, providers[10], "dlock.noop")
}
