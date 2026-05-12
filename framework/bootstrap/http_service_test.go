// Package bootstrap_test provides integration and boundary tests for HTTP service runtime bootstrapping.
//
// 适用场景：
// - 验证 HTTP Service runtime 的初始化、governance override 和 provider 注册行为。
// - 验证 pprof / governance inspect 端点的路由注册与响应格式。
// - 验证 governance summary 与 diagnostic 的构建与格式化逻辑。
package bootstrap

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	"github.com/stretchr/testify/require"
)

func TestNewHTTPServiceRuntimeLeavesGovernanceOverrideEmptyByDefault(t *testing.T) {
	originProviders := buildHTTPProvidersFunc
	origin := registerSelectedMicroserviceProvidersWithMode
	originWithOptions := registerSelectedMicroserviceProvidersWithOptionsFunc
	defer func() {
		buildHTTPProvidersFunc = originProviders
		registerSelectedMicroserviceProvidersWithMode = origin
		registerSelectedMicroserviceProvidersWithOptionsFunc = originWithOptions
	}()

	var gotMode string
	sentinel := errors.New("stop after capture")
	buildHTTPProvidersFunc = func(opts HTTPServiceOptions) []runtimecontract.ServiceProvider {
		return nil
	}
	registerSelectedMicroserviceProvidersWithOptionsFunc = func(c runtimecontract.Container, modeOverride string, disabled []string, enabled []string, providers map[string]string) error {
		gotMode = modeOverride
		return sentinel
	}

	_, err := NewHTTPServiceRuntime("demo", HTTPServiceOptions{})
	require.Error(t, err)
	require.ErrorIs(t, err, sentinel)
	require.Empty(t, gotMode)
}

func TestNewHTTPServiceRuntimeForwardsMicroserviceGovernanceOverride(t *testing.T) {
	originProviders := buildHTTPProvidersFunc
	origin := registerSelectedMicroserviceProvidersWithMode
	originWithOptions := registerSelectedMicroserviceProvidersWithOptionsFunc
	defer func() {
		buildHTTPProvidersFunc = originProviders
		registerSelectedMicroserviceProvidersWithMode = origin
		registerSelectedMicroserviceProvidersWithOptionsFunc = originWithOptions
	}()

	var gotMode string
	sentinel := errors.New("stop after capture")
	buildHTTPProvidersFunc = func(opts HTTPServiceOptions) []runtimecontract.ServiceProvider {
		return nil
	}
	registerSelectedMicroserviceProvidersWithOptionsFunc = func(c runtimecontract.Container, modeOverride string, disabled []string, enabled []string, providers map[string]string) error {
		gotMode = modeOverride
		return sentinel
	}

	_, err := NewHTTPServiceRuntime("demo", HTTPServiceOptions{
		GovernanceMode: string(resiliencecontract.GovernanceModeMicroservice),
	})
	require.Error(t, err)
	require.ErrorIs(t, err, sentinel)
	require.Equal(t, string(resiliencecontract.GovernanceModeMicroservice), gotMode)
}

func TestNewHTTPServiceRuntimeForwardsGovernanceDisableAndProviderOverrides(t *testing.T) {
	originProviders := buildHTTPProvidersFunc
	originWithOptions := registerSelectedMicroserviceProvidersWithOptionsFunc
	defer func() {
		buildHTTPProvidersFunc = originProviders
		registerSelectedMicroserviceProvidersWithOptionsFunc = originWithOptions
	}()

	var (
		gotDisabled []string
		gotProviders map[string]string
	)
	sentinel := errors.New("stop after capture")
	buildHTTPProvidersFunc = func(opts HTTPServiceOptions) []runtimecontract.ServiceProvider { return nil }
	registerSelectedMicroserviceProvidersWithOptionsFunc = func(c runtimecontract.Container, modeOverride string, disabled []string, enabled []string, providers map[string]string) error {
		gotDisabled = append([]string(nil), disabled...)
		gotProviders = providers
		return sentinel
	}

	_, err := NewHTTPServiceRuntime("demo", HTTPServiceOptions{
		GovernanceDisable:  []string{"tracing"},
		GovernanceProviders: map[string]string{"serviceauth": "mtls"},
	})
	require.Error(t, err)
	require.ErrorIs(t, err, sentinel)
	require.Equal(t, []string{"tracing"}, gotDisabled)
	require.Equal(t, "mtls", gotProviders["serviceauth"])
}

// =============================================================================
// pprof 端点注册与行为
// =============================================================================

func TestAutoMigrateModelsNilRuntime(t *testing.T) {
	require.NoError(t, AutoMigrateModels(nil, struct{}{}))
}

func TestAutoMigrateModelsNilDB(t *testing.T) {
	rt := &HTTPServiceRuntime{}
	require.NoError(t, AutoMigrateModels(rt, struct{}{}))
}

func TestRegisterPprofEndpointsUsesMount(t *testing.T) {
	router := &recordingRouter{}

	RegisterPprofEndpoints(router)

	require.Contains(t, router.mounted, "/debug/pprof/")
	require.Contains(t, router.mounted, "/debug/pprof/cmdline")
	require.Contains(t, router.mounted, "/debug/pprof/profile")
	require.Contains(t, router.mounted, "/debug/pprof/symbol")
	require.Contains(t, router.mounted, "/debug/pprof/trace")
}

func TestRegisterPprofEndpointsServesIndex(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	router := &ginTestRouter{engine: engine}

	RegisterPprofEndpoints(router)

	req := httptest.NewRequest(http.MethodGet, "/debug/pprof/", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.True(t, strings.Contains(w.Body.String(), "profile") || strings.Contains(w.Body.String(), "pprof"))
}

func TestRegisterPprofEndpointsRejectsPost(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	router := &ginTestRouter{engine: engine}

	RegisterPprofEndpoints(router)

	req := httptest.NewRequest(http.MethodPost, "/debug/pprof/", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
}

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
		Mode:                resiliencecontract.GovernanceModeMicroservice,
		ModeSource:          "code_override",
		ModeReason:          "mode selected by startup option",
		EnabledFeatures:     []string{"logging", "metrics", "serviceauth"},
		ResolutionOrder:     []string{"code_explicit_override", "config_explicit_override", "mode_defaults", "provider_fallback"},
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
			"governance.mode":                "microservice",
			"startup.governance_mode_override": "microservice",
			"governance.providers":           map[string]string{"selector": "unknown"},
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
		Mode:                resiliencecontract.GovernanceModeMicroservice,
		EnabledFeatures:     []string{"logging", "metrics", "serviceauth"},
		DisabledByOverride:  []string{"selector", "tracing"},
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
				Backend:          "noop",
				Source:           "provider_fallback",
				Reason:           "provider backend fell back to noop tracing",
				FallbackBackend:  "noop",
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
			"governance.mode":                "monolith",
			"governance.disable":             []string{"tracing"},
			"governance.providers.selector":  "random",
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


type recordingRouter struct {
	mounted []string
	gets    []string
}

func (r *recordingRouter) Use(middleware ...transportcontract.HTTPMiddleware) {}
func (r *recordingRouter) Group(prefix string, middleware ...transportcontract.HTTPMiddleware) transportcontract.HTTPRouter {
	return r
}
func (r *recordingRouter) Handle(method, path string, handler transportcontract.HTTPHandler) {}
func (r *recordingRouter) HandleFunc(method, path string, handlerFunc transportcontract.HTTPHandler) {
}
func (r *recordingRouter) GET(path string, handler transportcontract.HTTPHandler) {
	r.gets = append(r.gets, path)
}
func (r *recordingRouter) POST(path string, handler transportcontract.HTTPHandler)   {}
func (r *recordingRouter) PUT(path string, handler transportcontract.HTTPHandler)    {}
func (r *recordingRouter) DELETE(path string, handler transportcontract.HTTPHandler) {}
func (r *recordingRouter) Mount(path string, handler http.Handler) {
	r.mounted = append(r.mounted, path)
}

type ginTestRouter struct {
	engine *gin.Engine
}

func (r *ginTestRouter) Use(middleware ...transportcontract.HTTPMiddleware) {
	adapted := make([]gin.HandlerFunc, 0, len(middleware))
	for _, mw := range middleware {
		mw := mw
		adapted = append(adapted, func(c *gin.Context) {
			httpCtx := transportcontract.NewDefaultHTTPContext(c.Request.Context(), c.Request)
			httpCtx.SetParamFunc(c.Param)
			httpCtx.SetQueryFunc(c.Query)
			httpCtx.SetDefaultQueryFunc(c.DefaultQuery)
			httpCtx.SetHeaderFuncs(c.GetHeader, c.Header)
			httpCtx.SetBindFuncs(c.ShouldBindJSON, c.ShouldBindQuery, c.ShouldBind)
			httpCtx.SetResponseFuncs(c.JSON, func(code int, body string) { c.String(code, body) }, c.XML, c.Data, c.Redirect, c.Status, func() int { return c.Writer.Status() })
			httpCtx.SetRoutePathFunc(c.FullPath)
			if wrapped := mw(func(inner transportcontract.HTTPContext) {
				if inner != nil && inner.Request() != nil {
					c.Request = inner.Request()
				}
				c.Next()
			}); wrapped != nil {
				wrapped(httpCtx)
			}
		})
	}
	r.engine.Use(adapted...)
}
func (r *ginTestRouter) Group(prefix string, middleware ...transportcontract.HTTPMiddleware) transportcontract.HTTPRouter {
	group := r.engine.Group(prefix)
	wrapped := &ginGroupTestRouter{group: group}
	wrapped.Use(middleware...)
	return wrapped
}
func (r *ginTestRouter) Handle(method, path string, handler transportcontract.HTTPHandler) {
	r.engine.Handle(method, path, func(c *gin.Context) {
		httpCtx := transportcontract.NewDefaultHTTPContext(c.Request.Context(), c.Request)
		httpCtx.SetParamFunc(c.Param)
		httpCtx.SetQueryFunc(c.Query)
		httpCtx.SetDefaultQueryFunc(c.DefaultQuery)
		httpCtx.SetHeaderFuncs(c.GetHeader, c.Header)
		httpCtx.SetBindFuncs(c.ShouldBindJSON, c.ShouldBindQuery, c.ShouldBind)
		httpCtx.SetResponseFuncs(c.JSON, func(code int, body string) { c.String(code, body) }, c.XML, c.Data, c.Redirect, c.Status, func() int { return c.Writer.Status() })
		httpCtx.SetRoutePathFunc(c.FullPath)
		handler(httpCtx)
	})
}
func (r *ginTestRouter) HandleFunc(method, path string, handlerFunc transportcontract.HTTPHandler) {
	r.Handle(method, path, handlerFunc)
}
func (r *ginTestRouter) GET(path string, handler transportcontract.HTTPHandler) {
	r.Handle(http.MethodGet, path, handler)
}
func (r *ginTestRouter) POST(path string, handler transportcontract.HTTPHandler) {
	r.Handle(http.MethodPost, path, handler)
}
func (r *ginTestRouter) PUT(path string, handler transportcontract.HTTPHandler) {
	r.Handle(http.MethodPut, path, handler)
}
func (r *ginTestRouter) DELETE(path string, handler transportcontract.HTTPHandler) {
	r.Handle(http.MethodDelete, path, handler)
}
func (r *ginTestRouter) Mount(path string, handler http.Handler) {
	h := func(c *gin.Context) {
		handler.ServeHTTP(c.Writer, c.Request)
	}
	r.engine.Handle(http.MethodGet, path, h)
	r.engine.Handle(http.MethodHead, path, h)
}

type ginGroupTestRouter struct {
	group *gin.RouterGroup
}

func (r *ginGroupTestRouter) Use(middleware ...transportcontract.HTTPMiddleware) {
	adapted := make([]gin.HandlerFunc, 0, len(middleware))
	for _, mw := range middleware {
		mw := mw
		adapted = append(adapted, func(c *gin.Context) {
			httpCtx := transportcontract.NewDefaultHTTPContext(c.Request.Context(), c.Request)
			httpCtx.SetParamFunc(c.Param)
			httpCtx.SetQueryFunc(c.Query)
			httpCtx.SetDefaultQueryFunc(c.DefaultQuery)
			httpCtx.SetHeaderFuncs(c.GetHeader, c.Header)
			httpCtx.SetBindFuncs(c.ShouldBindJSON, c.ShouldBindQuery, c.ShouldBind)
			httpCtx.SetResponseFuncs(c.JSON, func(code int, body string) { c.String(code, body) }, c.XML, c.Data, c.Redirect, c.Status, func() int { return c.Writer.Status() })
			httpCtx.SetRoutePathFunc(c.FullPath)
			if wrapped := mw(func(inner transportcontract.HTTPContext) {
				if inner != nil && inner.Request() != nil {
					c.Request = inner.Request()
				}
				c.Next()
			}); wrapped != nil {
				wrapped(httpCtx)
			}
		})
	}
	r.group.Use(adapted...)
}
func (r *ginGroupTestRouter) Group(prefix string, middleware ...transportcontract.HTTPMiddleware) transportcontract.HTTPRouter {
	group := &ginGroupTestRouter{group: r.group.Group(prefix)}
	group.Use(middleware...)
	return group
}
func (r *ginGroupTestRouter) Handle(method, path string, handler transportcontract.HTTPHandler) {
	r.group.Handle(method, path, func(c *gin.Context) {
		httpCtx := transportcontract.NewDefaultHTTPContext(c.Request.Context(), c.Request)
		httpCtx.SetParamFunc(c.Param)
		httpCtx.SetQueryFunc(c.Query)
		httpCtx.SetDefaultQueryFunc(c.DefaultQuery)
		httpCtx.SetHeaderFuncs(c.GetHeader, c.Header)
		httpCtx.SetBindFuncs(c.ShouldBindJSON, c.ShouldBindQuery, c.ShouldBind)
		httpCtx.SetResponseFuncs(c.JSON, func(code int, body string) { c.String(code, body) }, c.XML, c.Data, c.Redirect, c.Status, func() int { return c.Writer.Status() })
		httpCtx.SetRoutePathFunc(c.FullPath)
		handler(httpCtx)
	})
}
func (r *ginGroupTestRouter) HandleFunc(method, path string, handlerFunc transportcontract.HTTPHandler) {
	r.Handle(method, path, handlerFunc)
}
func (r *ginGroupTestRouter) GET(path string, handler transportcontract.HTTPHandler) {
	r.group.Handle(http.MethodGet, path, func(c *gin.Context) {
		httpCtx := transportcontract.NewDefaultHTTPContext(c.Request.Context(), c.Request)
		httpCtx.SetParamFunc(c.Param)
		httpCtx.SetQueryFunc(c.Query)
		httpCtx.SetDefaultQueryFunc(c.DefaultQuery)
		httpCtx.SetHeaderFuncs(c.GetHeader, c.Header)
		httpCtx.SetBindFuncs(c.ShouldBindJSON, c.ShouldBindQuery, c.ShouldBind)
		httpCtx.SetResponseFuncs(c.JSON, func(code int, body string) { c.String(code, body) }, c.XML, c.Data, c.Redirect, c.Status, func() int { return c.Writer.Status() })
		httpCtx.SetRoutePathFunc(c.FullPath)
		handler(httpCtx)
	})
}
func (r *ginGroupTestRouter) POST(path string, handler transportcontract.HTTPHandler) {
	r.Handle(http.MethodPost, path, handler)
}
func (r *ginGroupTestRouter) PUT(path string, handler transportcontract.HTTPHandler) {
	r.Handle(http.MethodPut, path, handler)
}
func (r *ginGroupTestRouter) DELETE(path string, handler transportcontract.HTTPHandler) {
	r.Handle(http.MethodDelete, path, handler)
}
func (r *ginGroupTestRouter) Mount(path string, handler http.Handler) {
	h := func(c *gin.Context) {
		handler.ServeHTTP(c.Writer, c.Request)
	}
	r.group.Handle(http.MethodGet, path, h)
	r.group.Handle(http.MethodHead, path, h)
}
