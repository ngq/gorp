// Package bootstrap_test provides integration and boundary tests for HTTP service runtime bootstrapping.
//
// 适用场景：
// - 验证 HTTP Service runtime 的初始化、governance override 和 provider 注册行为。
// - 验证 pprof / governance inspect 端点的路由注册与响应格式。
// - 验证 governance summary 与 diagnostic 的构建与格式化逻辑。
package bootstrap

import (
	"testing"

	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Governance Summary 与 Diagnostic 格式与构建
// =============================================================================

func TestBuildGovernanceSummaryReportsOverridesAndProviders(t *testing.T) {
	cfg := &selectorConfigStub{values: map[string]any{
		"governance.mode":                  "microservice",
		"governance.disable":               []string{"tracing", "selector"},
		"governance.providers.serviceauth": "mtls",
	}}

	summary := BuildGovernanceSummary(cfg, resiliencecontract.GovernanceModeMicroservice)
	require.Equal(t, resiliencecontract.GovernanceModeMicroservice, summary.Mode)
	require.Equal(t, "config_override", summary.ModeSource)
	require.Contains(t, summary.DisabledByOverride, "selector")
	require.Contains(t, summary.DisabledByOverride, "tracing")
	require.Contains(t, summary.DisabledByConfig, "selector")
	require.Contains(t, summary.DisabledByConfig, "tracing")
	require.Contains(t, summary.EnabledFeatures, "serviceauth")
	require.False(t, summary.FeatureDecisions["selector"].Enabled)
	require.Equal(t, "config_override", summary.FeatureDecisions["selector"].Source)
	require.Contains(t, summary.FeatureDecisions["retry"].Reason, "not part of the current governance mode defaults")
	require.Equal(t, "mtls", summary.ProviderBackends["serviceauth"])
	require.Equal(t, "noop", summary.ProviderBackends["selector"])
	require.Equal(t, "config_override", summary.ProviderDecisions["serviceauth"].Source)
	require.Contains(t, summary.ProviderDecisions["serviceauth"].Reason, "governance.providers.serviceauth")
	require.Equal(t, "governance.providers.serviceauth", summary.ProviderDecisions["serviceauth"].ConfigKey)
	require.Contains(t, summary.ConfigSnapshot, "governance.mode")
	require.Contains(t, summary.ConfigSnapshot, "governance.providers")
}

func TestFormatGovernanceSummaryIncludesModeEnabledAndDisabled(t *testing.T) {
	summary := GovernanceSummary{
		Mode:               resiliencecontract.GovernanceModeMicroservice,
		EnabledFeatures:    []string{"logging", "metrics", "serviceauth"},
		DisabledByOverride: []string{"selector", "tracing"},
		ProviderBackends: map[string]string{
			"selector":    "noop",
			"serviceauth": "token",
		},
	}

	formatted := FormatGovernanceSummary(summary)
	require.Contains(t, formatted, "mode=microservice")
	require.Contains(t, formatted, "enabled=[logging, metrics, serviceauth]")
	require.Contains(t, formatted, "disabled_by_override=[selector, tracing]")
	require.Contains(t, formatted, "selector=noop")
	require.Contains(t, formatted, "serviceauth=token")
}

func TestFormatGovernanceDiagnosticGroupsFeaturesProvidersAndSnapshot(t *testing.T) {
	summary := GovernanceSummary{
		Mode:            resiliencecontract.GovernanceModeMonolith,
		ModeSource:      "implicit_default",
		ModeReason:      "mode fell back to monolith because no config key was set",
		ResolutionOrder: []string{"code_explicit_override", "config_explicit_override", "mode_defaults", "provider_fallback"},
		FeatureDecisions: map[string]GovernanceFeatureDecision{
			"logging": {Enabled: true, Source: "mode_default", Reason: "logging is enabled by the current governance mode defaults"},
			"tracing": {Enabled: false, Source: "mode_default", Reason: "tracing is not part of the current governance mode defaults"},
		},
		ProviderDecisions: map[string]GovernanceProviderDecision{
			"tracing": {
				Backend:         "noop",
				Source:          "provider_fallback",
				Reason:          "provider backend fell back to noop tracing",
				FallbackBackend: "noop",
			},
		},
		ConfigSnapshot: map[string]any{
			"governance.mode": "monolith",
		},
	}

	diagnostic := FormatGovernanceDiagnostic(summary)
	require.Contains(t, diagnostic, "Governance Summary")
	require.Contains(t, diagnostic, "Features")
	require.Contains(t, diagnostic, "- logging: enabled")
	require.Contains(t, diagnostic, "- tracing: disabled")
	require.Contains(t, diagnostic, "Providers")
	require.Contains(t, diagnostic, "- tracing: noop")
	require.Contains(t, diagnostic, "Config Snapshot")
	require.Contains(t, diagnostic, "- governance.mode: monolith")
}

func TestFormatGovernanceDiagnosticViewSupportsBriefProvidersAndFeatures(t *testing.T) {
	summary := GovernanceSummary{
		Mode:               resiliencecontract.GovernanceModeMicroservice,
		ModeSource:         "config_override",
		ModeReason:         "mode selected from config key governance.mode",
		EnabledFeatures:    []string{"logging"},
		DisabledByOverride: []string{"tracing"},
		FeatureDecisions: map[string]GovernanceFeatureDecision{
			"logging": {Enabled: true, Source: "mode_default", Reason: "logging is enabled by the current governance mode defaults"},
		},
		ProviderBackends: map[string]string{
			"selector": "noop",
		},
		ProviderDecisions: map[string]GovernanceProviderDecision{
			"selector": {Backend: "noop", Source: "provider_fallback", Reason: "provider backend fell back to noop selector"},
		},
	}

	brief := FormatGovernanceDiagnosticView(summary, "brief")
	require.Contains(t, brief, "Governance Brief")
	require.NotContains(t, brief, "Config Snapshot")

	providers := FormatGovernanceDiagnosticView(summary, "providers")
	require.Contains(t, providers, "Governance Providers")
	require.NotContains(t, providers, "Features")

	features := FormatGovernanceDiagnosticView(summary, "features")
	require.Contains(t, features, "Governance Features")
	require.NotContains(t, features, "Providers")

	config := FormatGovernanceDiagnosticView(summary, "config")
	require.Contains(t, config, "Governance Config Snapshot")

	full := FormatGovernanceDiagnosticView(summary, "full")
	require.Contains(t, full, "Governance Summary")
}

func TestBuildGovernanceSummaryWithModeOverrideReportsCodeAndConfigSources(t *testing.T) {
	cfg := overlayGovernanceConfig(
		&selectorConfigStub{values: map[string]any{
			"governance.mode":               "monolith",
			"governance.disable":            []string{"tracing"},
			"governance.providers.selector": "random",
			"tracing.enabled":               true,
			"service_auth.enabled":          true,
			"message_queue.enabled":         true,
			"distributed_lock.enabled":      true,
			"circuit_breaker.enabled":       true,
		}},
		[]string{"selector"},
		nil,
		map[string]string{"serviceauth": "mtls"},
	)

	summary := BuildGovernanceSummaryWithModeOverride(cfg, resiliencecontract.GovernanceModeMicroservice, "microservice")
	require.Equal(t, "code_override", summary.ModeSource)
	require.Contains(t, summary.DisabledByConfig, "tracing")
	require.Contains(t, summary.DisabledByCode, "selector")
	require.Equal(t, "mtls", summary.ProviderBackends["serviceauth"])
	require.Equal(t, "code_override", summary.ProviderDecisions["serviceauth"].Source)
	require.Contains(t, summary.ProviderDecisions["serviceauth"].Reason, "startup option")
	require.Equal(t, "noop", summary.ProviderBackends["selector"])
	require.Equal(t, "code_override", summary.ProviderDecisions["selector"].Source)
	require.Contains(t, summary.ProviderDecisions["selector"].Reason, "disabled by startup option")
	require.Equal(t, "noop", summary.ProviderBackends["tracing"])
	require.Equal(t, "config_override", summary.ProviderDecisions["tracing"].Source)
	require.Contains(t, summary.ProviderDecisions["tracing"].Reason, "governance.disable")
	require.Equal(t, "redis", summary.ProviderBackends["message_queue"])
	require.Equal(t, "config_override", summary.ProviderDecisions["message_queue"].Source)
	require.Contains(t, summary.ProviderDecisions["message_queue"].Reason, "message_queue.enabled")
	require.Contains(t, summary.ConfigSnapshot, "startup.governance_mode_override")
	require.Contains(t, summary.ConfigSnapshot, "startup.governance_disable")
	require.Contains(t, summary.ConfigSnapshot, "startup.governance_providers")
}

func TestBuildGovernanceSummaryProviderDecisionFallsBackWhenRequestedBackendUnknown(t *testing.T) {
	cfg := &selectorConfigStub{values: map[string]any{
		"governance.providers.selector": "unknown",
	}}

	summary := BuildGovernanceSummary(cfg, resiliencecontract.GovernanceModeMicroservice)
	require.Equal(t, "noop", summary.ProviderBackends["selector"])
	require.Equal(t, "config_override", summary.ProviderDecisions["selector"].Source)
	require.Contains(t, summary.ProviderDecisions["selector"].Reason, "requested backend unknown was unavailable")
	require.Equal(t, "unknown", summary.ProviderDecisions["selector"].RequestedBackend)
	require.Equal(t, "noop", summary.ProviderDecisions["selector"].FallbackBackend)
}
