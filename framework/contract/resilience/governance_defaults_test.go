// Package resilience_test provides unit tests for governance feature defaults stability.
//
// 适用场景：
// - 验证 DefaultGovernanceFeatureSet 在不同 governance mode 下的默认能力集稳定。
// - 确保各模式的 request_identity、logging、recovery、timeout、metrics 等基础能力默认启用。
package resilience

import "testing"

func TestDefaultGovernanceFeatureSetRemainStable(t *testing.T) {
	monolith := DefaultGovernanceFeatureSet(GovernanceModeMonolith)
	if !monolith.RequestIdentity || !monolith.Logging || !monolith.Recovery || !monolith.Timeout || !monolith.Metrics {
		t.Fatalf("expected monolith base governance defaults enabled, got %+v", monolith)
	}
	if monolith.MetadataPropagation || monolith.Tracing || monolith.Selector || monolith.ServiceAuth || monolith.CircuitBreaker || monolith.Retry || monolith.LoadShedding || monolith.Discovery {
		t.Fatalf("expected monolith advanced governance defaults disabled, got %+v", monolith)
	}

	ginFirst := DefaultGovernanceFeatureSet(GovernanceModeGinFirst)
	if ginFirst != monolith {
		t.Fatalf("expected gin-first governance defaults to match monolith for now, got %+v vs %+v", ginFirst, monolith)
	}

	microservice := DefaultGovernanceFeatureSet(GovernanceModeMicroservice)
	if !microservice.RequestIdentity || !microservice.Logging || !microservice.Recovery || !microservice.Timeout || !microservice.Metrics {
		t.Fatalf("expected microservice base governance defaults enabled, got %+v", microservice)
	}
	if !microservice.MetadataPropagation || !microservice.Tracing || !microservice.Selector || !microservice.ServiceAuth || !microservice.CircuitBreaker || !microservice.LoadShedding {
		t.Fatalf("expected microservice advanced governance defaults enabled (including loadshedding), got %+v", microservice)
	}
	if microservice.Retry || microservice.Discovery {
		t.Fatalf("expected retry/discovery to stay disabled until fully promoted, got %+v", microservice)
	}
}
