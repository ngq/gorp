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
// Governance Mode 检测与标准化
// =============================================================================

func TestDetectGovernanceModeDefaultsToMonolith(t *testing.T) {
	cfg := &selectorConfigStub{values: map[string]any{}}
	if got := DetectGovernanceMode(cfg); got != resilience.GovernanceModeMonolith {
		t.Fatalf("expected monolith mode, got %q", got)
	}
}

func TestDetectGovernanceModeAcceptsLegacyServiceModeKey(t *testing.T) {
	cfg := &selectorConfigStub{values: map[string]any{"service.mode": "microservice"}}
	if got := DetectGovernanceMode(cfg); got != resilience.GovernanceModeMicroservice {
		t.Fatalf("expected microservice mode from service.mode, got %q", got)
	}
}

func TestDetectGovernanceModeAcceptsGinFirst(t *testing.T) {
	cfg := &selectorConfigStub{values: map[string]any{"governance.mode": "gin-first"}}
	if got := DetectGovernanceMode(cfg); got != resilience.GovernanceModeGinFirst {
		t.Fatalf("expected gin-first mode from governance.mode, got %q", got)
	}
}

func TestNormalizeGovernanceModeFallsBackToMonolith(t *testing.T) {
	if got := NormalizeGovernanceMode(""); got != resilience.GovernanceModeMonolith {
		t.Fatalf("expected empty mode to normalize to monolith, got %q", got)
	}
	if got := NormalizeGovernanceMode("unknown"); got != resilience.GovernanceModeMonolith {
		t.Fatalf("expected unknown mode to normalize to monolith, got %q", got)
	}
}

func TestNormalizeGovernanceModePreservesGinFirst(t *testing.T) {
	if got := NormalizeGovernanceMode(resilience.GovernanceModeGinFirst); got != resilience.GovernanceModeGinFirst {
		t.Fatalf("expected gin-first mode preserved, got %q", got)
	}
	if !IsGinFirstMode(resilience.GovernanceModeGinFirst) {
		t.Fatal("expected IsGinFirstMode to report true for gin-first")
	}
	if IsMicroserviceMode(resilience.GovernanceModeGinFirst) {
		t.Fatal("expected gin-first not to be treated as microservice mode")
	}
}
