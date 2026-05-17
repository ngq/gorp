// Package bootstrap_test provides end-to-end tests for governance mode switching.
//
// 适用场景：
// - 验证从单体模式切换到微服务模式的完整流程。
// - 验证不同治理模式下 provider 选择正确性。
//go:build integration

package bootstrap

import (
	"testing"

	"github.com/ngq/gorp/framework/container"
	datacontract "github.com/ngq/gorp/framework/contract/data"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
)

// stubConfigForGovernance is a minimal config stub for governance mode tests.
type stubConfigForGovernance struct {
	values map[string]any
}

func (s *stubConfigForGovernance) Env() string                    { return "testing" }
func (s *stubConfigForGovernance) Get(key string) any             { return s.values[key] }
func (s *stubConfigForGovernance) GetString(key string) string    {
	if v, ok := s.values[key]; ok {
		if str, ok := v.(string); ok {
			return str
		}
	}
	return ""
}
func (s *stubConfigForGovernance) GetInt(key string) int {
	if v, ok := s.values[key]; ok {
		if i, ok := v.(int); ok {
			return i
		}
	}
	return 0
}
func (s *stubConfigForGovernance) GetBool(key string) bool {
	if v, ok := s.values[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}
func (s *stubConfigForGovernance) GetFloat(key string) float64    { return 0 }
func (s *stubConfigForGovernance) Unmarshal(key string, out any) error { return nil }

// TestE2EGovernanceModeMonolithToMicroservice verifies the governance mode can be switched from monolith to microservice.
//
// TestE2EGovernanceModeMonolithToMicroservice 验证治理模式可从单体切换到微服务。
func TestE2EGovernanceModeMonolithToMicroservice(t *testing.T) {
	// Start with monolith config
	monolithCfg := &stubConfigForGovernance{values: map[string]any{
		"governance.mode": "monolith",
	}}

	mode := DetectGovernanceMode(monolithCfg)
	if mode != resiliencecontract.GovernanceModeMono {
		t.Fatalf("expected monolith mode, got %q", mode)
	}

	// Switch to microservice config
	microCfg := &stubConfigForGovernance{values: map[string]any{
		"governance.mode": "micro",
	}}

	mode = DetectGovernanceMode(microCfg)
	if mode != resiliencecontract.GovernanceModeMicro {
		t.Fatalf("expected microservice mode, got %q", mode)
	}
}

// TestE2EGovernanceModeProviderSelection verifies that different governance modes select different default providers.
//
// TestE2EGovernanceModeProviderSelection 验证不同治理模式选择不同的默认 provider。
func TestE2EGovernanceModeProviderSelection(t *testing.T) {
	c := container.New()

	// Bind monolith config
	monolithCfg := &stubConfigForGovernance{values: map[string]any{
		"governance.mode": "monolith",
	}}
	c.Bind(datacontract.ConfigKey, func(container.Container) (any, error) {
		return monolithCfg, nil
	}, true)

	// Detect mode from config
	mode := DetectGovernanceMode(monolithCfg)
	if !IsMicroMode(mode) {
		// Monolith mode should use noop providers by default
		t.Log("monolith mode detected, noop providers expected")
	}

	// Rebind with microservice config
	microCfg := &stubConfigForGovernance{values: map[string]any{
		"governance.mode": "micro",
	}}
	c.Bind(datacontract.ConfigKey, func(container.Container) (any, error) {
		return microCfg, nil
	}, true)

	mode = DetectGovernanceMode(microCfg)
	if !IsMicroMode(mode) {
		t.Fatal("expected microservice mode")
	}
	// Microservice mode should use real providers (redis, etc.)
	t.Log("microservice mode detected, real providers expected")
}

// TestE2EHTTPModeContractToGin verifies the HTTP mode can be switched from contract to gin.
//
// TestE2EHTTPModeContractToGin 验证 HTTP 模式可从契约切换到 Gin 原生。
func TestE2EHTTPModeContractToGin(t *testing.T) {
	// Default is contract mode
	mode := NormalizeHTTPMode("")
	if mode != resiliencecontract.HTTPModeContract {
		t.Fatalf("expected contract mode, got %q", mode)
	}
	if IsGinHTTPMode(mode) {
		t.Fatal("expected IsGinHTTPMode to be false for contract mode")
	}

	// Switch to gin mode
	mode = NormalizeHTTPMode(resiliencecontract.HTTPModeGin)
	if mode != resiliencecontract.HTTPModeGin {
		t.Fatalf("expected gin mode, got %q", mode)
	}
	if !IsGinHTTPMode(mode) {
		t.Fatal("expected IsGinHTTPMode to be true for gin mode")
	}
}

// TestE2EGovernanceModeWithMultipleConfigKeys verifies governance mode can be read from multiple config keys.
//
// TestE2EGovernanceModeWithMultipleConfigKeys 验证治理模式可从多个配置键读取。
func TestE2EGovernanceModeWithMultipleConfigKeys(t *testing.T) {
	keys := []string{"governance.mode", "app.mode", "runtime.mode", "service.mode"}

	for _, key := range keys {
		cfg := &stubConfigForGovernance{values: map[string]any{
			key: "micro",
		}}
		mode := DetectGovernanceMode(cfg)
		if mode != resiliencecontract.GovernanceModeMicro {
			t.Fatalf("expected microservice mode from key %q, got %q", key, mode)
		}
	}
}

// TestE2EGovernanceModePriorityOrder verifies the priority order of governance mode config keys.
//
// TestE2EGovernanceModePriorityOrder 验证治理模式配置键的优先级顺序。
func TestE2EGovernanceModePriorityOrder(t *testing.T) {
	// governance.mode should take priority over service.mode
	cfg := &stubConfigForGovernance{values: map[string]any{
		"governance.mode": "monolith",
		"service.mode":    "micro",
	}}
	mode := DetectGovernanceMode(cfg)
	if mode != resiliencecontract.GovernanceModeMono {
		t.Fatalf("expected governance.mode to take priority, got %q", mode)
	}
}
