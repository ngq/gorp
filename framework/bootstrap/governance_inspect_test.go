// Package bootstrap_test provides integration and boundary tests for HTTP service runtime bootstrapping.
//
// 适用场景：
// - 验证 HTTP Service runtime 的初始化、governance override 和 provider 注册行为。
// - 验证 pprof / governance inspect 端点的路由注册与响应格式。
// - 验证 governance summary 与 diagnostic 的构建与格式化逻辑。
package bootstrap

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Governance Inspect 端点注册与视图
// =============================================================================

func TestRegisterGovernanceInspectEndpointsUsesGET(t *testing.T) {
	router := &recordingRouter{}
	summary := GovernanceSummary{
		Mode:            resiliencecontract.GovernanceModeMicroservice,
		EnabledFeatures: []string{"logging", "metrics"},
	}

	RegisterGovernanceInspectEndpoints(router, summary)

	require.Contains(t, router.gets, "/debug/governance")
	require.Contains(t, router.gets, "/doctor/governance")
}

func TestRegisterGovernanceInspectEndpointsServesSummaryJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	router := &ginTestRouter{engine: engine}
	summary := GovernanceSummary{
		Mode:               resiliencecontract.GovernanceModeMicroservice,
		EnabledFeatures:    []string{"logging", "metrics", "serviceauth"},
		DisabledByOverride: []string{"tracing"},
		ProviderBackends: map[string]string{
			"serviceauth": "mtls",
			"selector":    "p2c_ewma",
		},
		ResolutionOrder: []string{"code_explicit_override", "config_explicit_override", "mode_defaults", "provider_fallback"},
	}

	RegisterGovernanceInspectEndpoints(router, summary)

	req := httptest.NewRequest(http.MethodGet, "/debug/governance", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var got GovernanceSummary
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	require.Equal(t, summary.Mode, got.Mode)
	require.Equal(t, summary.EnabledFeatures, got.EnabledFeatures)
	require.Equal(t, summary.DisabledByOverride, got.DisabledByOverride)
	require.Equal(t, summary.ProviderBackends, got.ProviderBackends)
	require.Equal(t, summary.ResolutionOrder, got.ResolutionOrder)
}

func TestRegisterGovernanceInspectEndpointsServesDiagnosticText(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	router := &ginTestRouter{engine: engine}
	summary := GovernanceSummary{
		Mode:            resiliencecontract.GovernanceModeMicroservice,
		ModeSource:      "code_override",
		ModeReason:      "mode selected by startup option",
		EnabledFeatures: []string{"logging", "metrics", "serviceauth"},
		ResolutionOrder: []string{"code_explicit_override", "config_explicit_override", "mode_defaults", "provider_fallback"},
		FeatureDecisions: map[string]GovernanceFeatureDecision{
			"logging": {Enabled: true, Source: "mode_default", Reason: "logging is enabled by the current governance mode defaults"},
			"retry":   {Enabled: false, Source: "mode_default", Reason: "retry is not part of the current governance mode defaults"},
		},
		ProviderDecisions: map[string]GovernanceProviderDecision{
			"serviceauth": {
				Backend:          "mtls",
				Source:           "code_override",
				Reason:           "provider backend came from startup option governance.providers.serviceauth",
				RequestedBackend: "mtls",
			},
			"selector": {
				Backend:          "noop",
				Source:           "config_override",
				Reason:           "provider backend came from config key governance.providers.selector; requested backend unknown was unavailable and fell back to noop",
				RequestedBackend: "unknown",
				FallbackBackend:  "noop",
				ConfigKey:        "governance.providers.selector",
			},
		},
		ConfigSnapshot: map[string]any{
			"governance.mode":                  "microservice",
			"startup.governance_mode_override": "microservice",
			"governance.providers":             map[string]string{"selector": "unknown"},
		},
	}

	RegisterGovernanceInspectEndpoints(router, summary)

	req := httptest.NewRequest(http.MethodGet, "/doctor/governance?format=text", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "text/plain; charset=utf-8", w.Header().Get("Content-Type"))
	require.Contains(t, w.Body.String(), "Governance Summary")
	require.Contains(t, w.Body.String(), "Features")
	require.Contains(t, w.Body.String(), "- logging: enabled")
	require.Contains(t, w.Body.String(), "- retry: disabled")
	require.Contains(t, w.Body.String(), "Providers")
	require.Contains(t, w.Body.String(), "- serviceauth: mtls")
	require.Contains(t, w.Body.String(), "requested=unknown")
	require.Contains(t, w.Body.String(), "fallback=noop")
	require.Contains(t, w.Body.String(), "Config Snapshot")
	require.Contains(t, w.Body.String(), "governance.mode: microservice")
}

func TestRegisterGovernanceInspectEndpointsSupportsBriefView(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	router := &ginTestRouter{engine: engine}
	summary := GovernanceSummary{
		Mode:               resiliencecontract.GovernanceModeMicroservice,
		ModeSource:         "code_override",
		ModeReason:         "mode selected by startup option",
		EnabledFeatures:    []string{"logging", "metrics"},
		DisabledByOverride: []string{"tracing"},
		ProviderBackends: map[string]string{
			"serviceauth": "mtls",
			"selector":    "noop",
		},
	}

	RegisterGovernanceInspectEndpoints(router, summary)

	req := httptest.NewRequest(http.MethodGet, "/doctor/governance?view=brief", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "text/plain; charset=utf-8", w.Header().Get("Content-Type"))
	require.Contains(t, w.Body.String(), "Governance Brief")
	require.Contains(t, w.Body.String(), "Enabled Features: logging, metrics")
	require.Contains(t, w.Body.String(), "Providers: selector=noop, serviceauth=mtls")
	require.NotContains(t, w.Body.String(), "Config Snapshot")
}

func TestRegisterGovernanceInspectEndpointsSupportsProvidersView(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	router := &ginTestRouter{engine: engine}
	summary := GovernanceSummary{
		Mode: resiliencecontract.GovernanceModeMicroservice,
		ProviderDecisions: map[string]GovernanceProviderDecision{
			"selector": {
				Backend:          "noop",
				Source:           "config_override",
				Reason:           "provider backend came from config key governance.providers.selector; requested backend unknown was unavailable and fell back to noop",
				RequestedBackend: "unknown",
				FallbackBackend:  "noop",
				ConfigKey:        "governance.providers.selector",
			},
		},
	}

	RegisterGovernanceInspectEndpoints(router, summary)

	req := httptest.NewRequest(http.MethodGet, "/doctor/governance?view=providers", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), "Governance Providers")
	require.Contains(t, w.Body.String(), "- selector: noop")
	require.Contains(t, w.Body.String(), "requested=unknown")
	require.Contains(t, w.Body.String(), "fallback=noop")
	require.NotContains(t, w.Body.String(), "Features")
}

func TestRegisterGovernanceInspectEndpointsSupportsFeaturesView(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	router := &ginTestRouter{engine: engine}
	summary := GovernanceSummary{
		Mode: resiliencecontract.GovernanceModeMonolith,
		FeatureDecisions: map[string]GovernanceFeatureDecision{
			"logging": {Enabled: true, Source: "mode_default", Reason: "logging is enabled by the current governance mode defaults"},
			"retry":   {Enabled: false, Source: "mode_default", Reason: "retry is not part of the current governance mode defaults"},
		},
	}

	RegisterGovernanceInspectEndpoints(router, summary)

	req := httptest.NewRequest(http.MethodGet, "/doctor/governance?view=features", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), "Governance Features")
	require.Contains(t, w.Body.String(), "- logging: enabled")
	require.Contains(t, w.Body.String(), "- retry: disabled")
	require.NotContains(t, w.Body.String(), "Providers")
}

func TestRegisterGovernanceInspectEndpointsSupportsConfigView(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	router := &ginTestRouter{engine: engine}
	summary := GovernanceSummary{
		Mode:       resiliencecontract.GovernanceModeMicroservice,
		ModeSource: "config_override",
		ConfigSnapshot: map[string]any{
			"governance.mode":      "microservice",
			"governance.disable":   []string{"tracing"},
			"governance.providers": map[string]string{"serviceauth": "mtls"},
		},
	}

	RegisterGovernanceInspectEndpoints(router, summary)

	req := httptest.NewRequest(http.MethodGet, "/doctor/governance?view=config", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), "Governance Config Snapshot")
	require.Contains(t, w.Body.String(), "- governance.mode: microservice")
	require.Contains(t, w.Body.String(), "- governance.disable: [tracing]")
	require.Contains(t, w.Body.String(), "- governance.providers: {serviceauth=mtls}")
	require.NotContains(t, w.Body.String(), "Providers")
	require.NotContains(t, w.Body.String(), "Features")
}

func TestRegisterGovernanceInspectEndpointsSupportsFullView(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	router := &ginTestRouter{engine: engine}
	summary := GovernanceSummary{
		Mode:            resiliencecontract.GovernanceModeMicroservice,
		ModeSource:      "code_override",
		ModeReason:      "mode selected by startup option",
		ResolutionOrder: []string{"code_explicit_override", "config_explicit_override", "mode_defaults", "provider_fallback"},
		FeatureDecisions: map[string]GovernanceFeatureDecision{
			"logging": {Enabled: true, Source: "mode_default", Reason: "logging is enabled by the current governance mode defaults"},
		},
		ProviderDecisions: map[string]GovernanceProviderDecision{
			"selector": {Backend: "noop", Source: "provider_fallback", Reason: "provider backend fell back to noop selector"},
		},
		ConfigSnapshot: map[string]any{
			"governance.mode": "microservice",
		},
	}

	RegisterGovernanceInspectEndpoints(router, summary)

	req := httptest.NewRequest(http.MethodGet, "/doctor/governance?view=full", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), "Governance Summary")
	require.Contains(t, w.Body.String(), "Features")
	require.Contains(t, w.Body.String(), "Providers")
	require.Contains(t, w.Body.String(), "Config Snapshot")
}
