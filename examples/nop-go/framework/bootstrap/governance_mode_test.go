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
	if got := DetectGovernanceMode(cfg); got != resilience.GovernanceModeMono {
		t.Fatalf("expected monolith mode, got %q", got)
	}
}

func TestDetectGovernanceModeAcceptsLegacyServiceModeKey(t *testing.T) {
	cfg := &selectorConfigStub{values: map[string]any{"service.mode": "micro"}}
	if got := DetectGovernanceMode(cfg); got != resilience.GovernanceModeMicro {
		t.Fatalf("expected microservice mode from service.mode, got %q", got)
	}
}

func TestNormalizeGovernanceModeFallsBackToMonolith(t *testing.T) {
	if got := NormalizeGovernanceMode(""); got != resilience.GovernanceModeMono {
		t.Fatalf("expected empty mode to normalize to monolith, got %q", got)
	}
	if got := NormalizeGovernanceMode("unknown"); got != resilience.GovernanceModeMono {
		t.Fatalf("expected unknown mode to normalize to monolith, got %q", got)
	}
}

func TestNormalizeHTTPModeFallsBackToContract(t *testing.T) {
	if got := NormalizeHTTPMode(""); got != resilience.HTTPModeContract {
		t.Fatalf("expected empty HTTP mode to normalize to contract, got %q", got)
	}
	if got := NormalizeHTTPMode("unknown"); got != resilience.HTTPModeContract {
		t.Fatalf("expected unknown HTTP mode to normalize to contract, got %q", got)
	}
}

func TestNormalizeHTTPModePreservesGin(t *testing.T) {
	if got := NormalizeHTTPMode(resilience.HTTPModeGin); got != resilience.HTTPModeGin {
		t.Fatalf("expected gin HTTP mode preserved, got %q", got)
	}
	if !IsGinHTTPMode(resilience.HTTPModeGin) {
		t.Fatal("expected IsGinHTTPMode to report true for gin")
	}
	if IsGinHTTPMode(resilience.HTTPModeContract) {
		t.Fatal("expected IsGinHTTPMode to report false for contract")
	}
}
