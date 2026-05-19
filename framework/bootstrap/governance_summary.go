// Package bootstrap provides framework bootstrap and assembly helpers for gorp.
// This file builds governance summary for startup logs, tests, and inspect tooling.
// Makes “what is enabled/disabled” and “provider backends” observable in one place.
//
// Bootstrap 包提供 gorp 框架的启动装配辅助能力。
// 本文件生成统一的治理生效摘要，供启动日志、测试和后续 inspect 工具复用。
// 将启用了什么、关闭了什么、当前 provider backend 集中展示。
package bootstrap

import (
	"fmt"
	"sort"
	"strings"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
)

// GovernanceSummary describes the effective governance result for one runtime mode.
//
// GovernanceSummary 描述某个运行模式下最终生效的治理结果。
type GovernanceSummary struct {
	Mode                 resiliencecontract.GovernanceMode           `json:"mode"`
	ModeSource           string                                      `json:"mode_source"`
	ModeReason           string                                      `json:"mode_reason"`
	EnabledFeatures      []string                                    `json:"enabled_features"`
	ModeDefaultFeatures  []string                                    `json:"mode_default_features"`
	DisabledByOverride   []string                                    `json:"disabled_by_override"`
	DisabledByConfig     []string                                    `json:"disabled_by_config"`
	DisabledByCode       []string                                    `json:"disabled_by_code"`
	EnabledByOverride    []string                                    `json:"enabled_by_override"`
	EnabledByConfig      []string                                    `json:"enabled_by_config"`
	EnabledByCode        []string                                    `json:"enabled_by_code"`
	FeatureDecisions     map[string]GovernanceFeatureDecision        `json:"feature_decisions"`
	ProviderBackends     map[string]string                           `json:"provider_backends"`
	ProviderDecisions    map[string]GovernanceProviderDecision       `json:"provider_decisions"`
	ConfigSnapshot       map[string]any                              `json:"config_snapshot"`
	ResolutionOrder      []string                                    `json:"resolution_order"`
	MiddlewareChainOrder []string                                    `json:"middleware_chain_order"`
	RPCClientChainOrder  []string                                    `json:"rpc_client_chain_order"`
	Defaults             *resiliencecontract.GovernanceDefaultsTable `json:"defaults,omitempty"`
}

// GovernanceFeatureDecision explains why one governance feature is enabled or disabled.
//
// GovernanceFeatureDecision 解释某个治理 feature 为什么启用或未启用。
type GovernanceFeatureDecision struct {
	Enabled bool   `json:"enabled"`
	Source  string `json:"source"`
	Reason  string `json:"reason"`
}

// GovernanceProviderDecision explains why one governance provider backend is currently active.
//
// GovernanceProviderDecision 解释某个治理 provider backend 为什么最终生效。
type GovernanceProviderDecision struct {
	Backend          string `json:"backend"`
	Source           string `json:"source"`
	Reason           string `json:"reason"`
	RequestedBackend string `json:"requested_backend,omitempty"`
	FallbackBackend  string `json:"fallback_backend,omitempty"`
	ConfigKey        string `json:"config_key,omitempty"`
}

// BuildGovernanceSummary builds the effective governance summary for one runtime mode and config snapshot.
//
// BuildGovernanceSummary 根据运行模式与配置快照构建最终生效治理摘要。
func BuildGovernanceSummary(cfg datacontract.Config, mode resiliencecontract.GovernanceMode) GovernanceSummary {
	return BuildGovernanceSummaryWithModeOverride(cfg, mode, "")
}

// BuildGovernanceSummaryWithModeOverride builds the effective governance summary and records whether mode came from code override.
//
// BuildGovernanceSummaryWithModeOverride 构建最终治理生效摘要，并记录 mode 是否来自代码显式覆盖。
func BuildGovernanceSummaryWithModeOverride(cfg datacontract.Config, mode resiliencecontract.GovernanceMode, modeOverride string) GovernanceSummary {
	mode = NormalizeGovernanceMode(mode)
	defaultFeatures := resiliencecontract.DefaultGovernanceFeatureSet(mode)
	configOverrides, codeOverrides, configView := splitGovernanceOverrides(cfg)
	mergedOverrides := governanceOverrides{
		Disabled:         mergeGovernanceDisabled(configOverrides.Disabled, codeOverrides.Disabled),
		Enabled:          mergeGovernanceEnabled(configOverrides.Enabled, codeOverrides.Enabled),
		ProviderBackends: mergeGovernanceProviderBackends(configOverrides.ProviderBackends, codeOverrides.ProviderBackends),
	}
	// 先开启再关闭：同一 feature 同时在 enable 和 disable 中时，disable 生效。
	// 这与 buildGovernanceFeatureDecisions 的优先级一致（disable > enable）。
	features := applyGovernanceFeatureEnables(defaultFeatures, mergedOverrides.Enabled)
	features = applyGovernanceFeatureDisables(features, mergedOverrides.Disabled)
	modeSource, modeReason := governanceModeDecision(configView, modeOverride)
	featureDecisions := buildGovernanceFeatureDecisions(mode, defaultFeatures, configOverrides.Disabled, codeOverrides.Disabled, configOverrides.Enabled, codeOverrides.Enabled)
	providerDecisions := buildGovernanceProviderDecisions(configView, mode, configOverrides, codeOverrides)
	configSnapshot := buildGovernanceConfigSnapshot(configView, modeOverride, configOverrides, codeOverrides)

	return GovernanceSummary{
		Mode:                 mode,
		ModeSource:           modeSource,
		ModeReason:           modeReason,
		EnabledFeatures:      governanceFeatureNames(features),
		ModeDefaultFeatures:  governanceFeatureNames(defaultFeatures),
		DisabledByOverride:   governanceDisabledNames(mergedOverrides.Disabled),
		DisabledByConfig:     governanceDisabledNames(configOverrides.Disabled),
		DisabledByCode:       governanceDisabledNames(codeOverrides.Disabled),
		EnabledByOverride:    governanceEnabledNames(mergedOverrides.Enabled),
		EnabledByConfig:      governanceEnabledNames(configOverrides.Enabled),
		EnabledByCode:        governanceEnabledNames(codeOverrides.Enabled),
		FeatureDecisions:     featureDecisions,
		ProviderBackends:     governanceProviderBackendsFromDecisions(providerDecisions),
		ProviderDecisions:    providerDecisions,
		ConfigSnapshot:       configSnapshot,
		ResolutionOrder:      governanceResolutionOrder(),
		MiddlewareChainOrder: governanceMiddlewareChainOrder(),
		RPCClientChainOrder:  governanceRPCClientChainOrder(),
	}
}

// FormatGovernanceSummary renders the governance summary into one stable log-friendly line.
//
// FormatGovernanceSummary 将治理摘要渲染成稳定的日志字符串。
func FormatGovernanceSummary(summary GovernanceSummary) string {
	enabled := strings.Join(summary.EnabledFeatures, ", ")
	if enabled == "" {
		enabled = "none"
	}

	disabled := strings.Join(summary.DisabledByOverride, ", ")
	if disabled == "" {
		disabled = "none"
	}

	enabledBy := strings.Join(summary.EnabledByOverride, ", ")

	providers := make([]string, 0, len(summary.ProviderBackends))
	for key, value := range summary.ProviderBackends {
		providers = append(providers, fmt.Sprintf("%s=%s", key, value))
	}
	sort.Strings(providers)

	if enabledBy != "" {
		return fmt.Sprintf(
			"governance mode=%s enabled=[%s] disabled_by_override=[%s] enabled_by_override=[%s] providers=[%s]",
			summary.Mode,
			enabled,
			disabled,
			enabledBy,
			strings.Join(providers, ", "),
		)
	}
	return fmt.Sprintf(
		"governance mode=%s enabled=[%s] disabled_by_override=[%s] providers=[%s]",
		summary.Mode,
		enabled,
		disabled,
		strings.Join(providers, ", "),
	)
}

// FormatGovernanceDiagnostic renders one human-readable governance diagnostic report.
//
// FormatGovernanceDiagnostic 将治理结果渲染成人类可读的诊断文本。
func FormatGovernanceDiagnostic(summary GovernanceSummary) string {
	var b strings.Builder

	writeLine := func(format string, args ...any) {
		if b.Len() > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(fmt.Sprintf(format, args...))
	}

	writeLine("Governance Summary")
	writeLine("Mode: %s", summary.Mode)
	writeLine("Mode Source: %s", summary.ModeSource)
	writeLine("Mode Reason: %s", summary.ModeReason)
	writeLine("Resolution Order: %s", strings.Join(summary.ResolutionOrder, " > "))
	if len(summary.MiddlewareChainOrder) > 0 {
		writeLine("HTTP Middleware Chain: %s", strings.Join(summary.MiddlewareChainOrder, " → "))
	}
	if len(summary.RPCClientChainOrder) > 0 {
		writeLine("RPC Client Chain: %s", strings.Join(summary.RPCClientChainOrder, " → "))
	}

	writeLine("")
	writeLine("Features")
	for _, name := range governanceAllFeatureNames() {
		decision, ok := summary.FeatureDecisions[name]
		if !ok {
			continue
		}
		state := "disabled"
		if decision.Enabled {
			state = "enabled"
		}
		writeLine("- %s: %s", name, state)
		writeLine("  source=%s", decision.Source)
		writeLine("  reason=%s", decision.Reason)
	}

	writeLine("")
	writeLine("Providers")
	for _, name := range governanceProviderNames() {
		decision, ok := summary.ProviderDecisions[name]
		if !ok {
			continue
		}
		writeLine("- %s: %s", name, decision.Backend)
		writeLine("  source=%s", decision.Source)
		writeLine("  reason=%s", decision.Reason)
		if decision.RequestedBackend != "" {
			writeLine("  requested=%s", decision.RequestedBackend)
		}
		if decision.FallbackBackend != "" {
			writeLine("  fallback=%s", decision.FallbackBackend)
		}
		if decision.ConfigKey != "" {
			writeLine("  config_key=%s", decision.ConfigKey)
		}
	}

	if len(summary.ConfigSnapshot) > 0 {
		writeLine("")
		writeLine("Config Snapshot")
		for _, key := range governanceSortedSnapshotKeys(summary.ConfigSnapshot) {
			writeLine("- %s: %s", key, governanceFormatSnapshotValue(summary.ConfigSnapshot[key]))
		}
	}

	return b.String()
}

// FormatGovernanceDiagnosticView renders one selected human-readable governance diagnostic view.
//
// FormatGovernanceDiagnosticView 渲染指定视角的人类可读治理诊断文本。
func FormatGovernanceDiagnosticView(summary GovernanceSummary, view string) string {
	switch strings.ToLower(strings.TrimSpace(view)) {
	case "brief":
		return formatGovernanceBriefDiagnostic(summary)
	case "providers":
		return formatGovernanceProvidersDiagnostic(summary)
	case "features":
		return formatGovernanceFeaturesDiagnostic(summary)
	case "config":
		return formatGovernanceConfigDiagnostic(summary)
	case "defaults":
		return formatGovernanceDefaultsDiagnostic(summary)
	case "full":
		return FormatGovernanceDiagnostic(summary)
	default:
		return FormatGovernanceDiagnostic(summary)
	}
}

func formatGovernanceBriefDiagnostic(summary GovernanceSummary) string {
	var b strings.Builder
	writeGovernanceLine(&b, "Governance Brief")
	writeGovernanceLine(&b, "Mode: %s", summary.Mode)
	writeGovernanceLine(&b, "Mode Source: %s", summary.ModeSource)
	writeGovernanceLine(&b, "Mode Reason: %s", summary.ModeReason)
	writeGovernanceLine(&b, "Enabled Features: %s", governanceJoinOrNone(summary.EnabledFeatures))
	writeGovernanceLine(&b, "Disabled By Override: %s", governanceJoinOrNone(summary.DisabledByOverride))
	if len(summary.EnabledByOverride) > 0 {
		writeGovernanceLine(&b, "Enabled By Override: %s", governanceJoinOrNone(summary.EnabledByOverride))
	}
	if len(summary.MiddlewareChainOrder) > 0 {
		writeGovernanceLine(&b, "HTTP Chain: %s", strings.Join(summary.MiddlewareChainOrder, " → "))
	}
	if len(summary.RPCClientChainOrder) > 0 {
		writeGovernanceLine(&b, "RPC Chain: %s", strings.Join(summary.RPCClientChainOrder, " → "))
	}

	providers := make([]string, 0, len(summary.ProviderBackends))
	for _, name := range governanceProviderNames() {
		if backend := summary.ProviderBackends[name]; backend != "" {
			providers = append(providers, fmt.Sprintf("%s=%s", name, backend))
		}
	}
	writeGovernanceLine(&b, "Providers: %s", governanceJoinOrNone(providers))
	return b.String()
}

func formatGovernanceFeaturesDiagnostic(summary GovernanceSummary) string {
	var b strings.Builder
	writeGovernanceLine(&b, "Governance Features")
	writeGovernanceLine(&b, "Mode: %s", summary.Mode)
	for _, name := range governanceAllFeatureNames() {
		decision, ok := summary.FeatureDecisions[name]
		if !ok {
			continue
		}
		state := "disabled"
		if decision.Enabled {
			state = "enabled"
		}
		writeGovernanceLine(&b, "- %s: %s", name, state)
		writeGovernanceLine(&b, "  source=%s", decision.Source)
		writeGovernanceLine(&b, "  reason=%s", decision.Reason)
	}
	return b.String()
}

func formatGovernanceProvidersDiagnostic(summary GovernanceSummary) string {
	var b strings.Builder
	writeGovernanceLine(&b, "Governance Providers")
	writeGovernanceLine(&b, "Mode: %s", summary.Mode)
	for _, name := range governanceProviderNames() {
		decision, ok := summary.ProviderDecisions[name]
		if !ok {
			continue
		}
		writeGovernanceLine(&b, "- %s: %s", name, decision.Backend)
		writeGovernanceLine(&b, "  source=%s", decision.Source)
		writeGovernanceLine(&b, "  reason=%s", decision.Reason)
		if decision.RequestedBackend != "" {
			writeGovernanceLine(&b, "  requested=%s", decision.RequestedBackend)
		}
		if decision.FallbackBackend != "" {
			writeGovernanceLine(&b, "  fallback=%s", decision.FallbackBackend)
		}
		if decision.ConfigKey != "" {
			writeGovernanceLine(&b, "  config_key=%s", decision.ConfigKey)
		}
	}
	return b.String()
}

func formatGovernanceConfigDiagnostic(summary GovernanceSummary) string {
	var b strings.Builder
	writeGovernanceLine(&b, "Governance Config Snapshot")
	writeGovernanceLine(&b, "Mode: %s", summary.Mode)
	writeGovernanceLine(&b, "Mode Source: %s", summary.ModeSource)
	if len(summary.ConfigSnapshot) == 0 {
		writeGovernanceLine(&b, "Config Snapshot: none")
		return b.String()
	}
	for _, key := range governanceSortedSnapshotKeys(summary.ConfigSnapshot) {
		writeGovernanceLine(&b, "- %s: %s", key, governanceFormatSnapshotValue(summary.ConfigSnapshot[key]))
	}
	return b.String()
}

// formatGovernanceDefaultsDiagnostic renders the governance defaults table as human-readable text.
// This view shows what the framework would use if no overrides were applied.
//
// formatGovernanceDefaultsDiagnostic 将治理默认值表渲染成人类可读文本。
// 此视图展示框架在没有任何覆盖时的默认值。
func formatGovernanceDefaultsDiagnostic(summary GovernanceSummary) string {
	var b strings.Builder
	writeGovernanceLine(&b, "Governance Defaults")
	writeGovernanceLine(&b, "Mode: %s", summary.Mode)

	if summary.Defaults == nil {
		writeGovernanceLine(&b, "(defaults table not loaded)")
		return b.String()
	}
	d := summary.Defaults

	// 特性默认值
	writeGovernanceLine(&b, "")
	writeGovernanceLine(&b, "Feature Defaults")
	for _, name := range governanceAllFeatureNames() {
		if enabled, ok := d.FeatureDefaults[name]; ok {
			state := "disabled"
			if enabled {
				state = "enabled"
			}
			writeGovernanceLine(&b, "- %s: %s", name, state)
		}
	}

	// Provider 默认值
	writeGovernanceLine(&b, "")
	writeGovernanceLine(&b, "Provider Defaults")
	for _, name := range governanceProviderNames() {
		if backend, ok := d.ProviderDefaults[name]; ok {
			writeGovernanceLine(&b, "- %s: %s", name, backend)
		}
	}

	// HTTP 中间件默认值
	writeGovernanceLine(&b, "")
	writeGovernanceLine(&b, "Middleware Defaults")
	writeGovernanceLine(&b, "- timeout: %s", d.MiddlewareDefaults.Timeout)
	writeGovernanceLine(&b, "- body_limit: %s", d.MiddlewareDefaults.BodyLimit)
	writeGovernanceLine(&b, "- max_concurrent: %d", d.MiddlewareDefaults.MaxConcurrent)
	writeGovernanceLine(&b, "- enable_metrics: %t", d.MiddlewareDefaults.EnableMetrics)
	writeGovernanceLine(&b, "- enable_compression: %t", d.MiddlewareDefaults.EnableCompression)

	// CORS 默认值
	cors := d.MiddlewareDefaults.CORS
	writeGovernanceLine(&b, "- cors:")
	writeGovernanceLine(&b, "  allow_origins: %s", strings.Join(cors.AllowOrigins, ", "))
	writeGovernanceLine(&b, "  max_age_seconds: %d", cors.MaxAgeSeconds)

	// 安全头默认值
	sec := d.MiddlewareDefaults.SecurityHeaders
	writeGovernanceLine(&b, "- security_headers:")
	writeGovernanceLine(&b, "  x_frame_options: %s", sec.XFrameOptions)
	writeGovernanceLine(&b, "  x_content_type_options: %s", sec.XContentTypeOptions)
	writeGovernanceLine(&b, "  referrer_policy: %s", sec.ReferrerPolicy)

	// 本地化默认值
	loc := d.MiddlewareDefaults.Locale
	writeGovernanceLine(&b, "- locale:")
	writeGovernanceLine(&b, "  supported: %s", strings.Join(loc.Supported, ", "))
	writeGovernanceLine(&b, "  default: %s", loc.Default)
	writeGovernanceLine(&b, "  query_keys: %s", strings.Join(loc.QueryKeys, ", "))

	// RPC 客户端默认值
	writeGovernanceLine(&b, "")
	writeGovernanceLine(&b, "RPC Client Defaults")
	writeGovernanceLine(&b, "- timeout: %s", d.RPCClientDefaults.Timeout)

	return b.String()
}

func writeGovernanceLine(b *strings.Builder, format string, args ...any) {
	if b.Len() > 0 {
		b.WriteByte('\n')
	}
	b.WriteString(fmt.Sprintf(format, args...))
}

func governanceJoinOrNone(items []string) string {
	if len(items) == 0 {
		return "none"
	}
	return strings.Join(items, ", ")
}

func governanceFeatureNames(features resiliencecontract.GovernanceFeatureSet) []string {
	names := make([]string, 0, 13)
	appendIf := func(enabled bool, name string) {
		if enabled {
			names = append(names, name)
		}
	}

	appendIf(features.RequestIdentity, "request_identity")
	appendIf(features.Logging, "logging")
	appendIf(features.Recovery, "recovery")
	appendIf(features.Timeout, "timeout")
	appendIf(features.Metrics, "metrics")
	appendIf(features.MetadataPropagation, "metadata")
	appendIf(features.Tracing, "tracing")
	appendIf(features.Selector, "selector")
	appendIf(features.ServiceAuth, "serviceauth")
	appendIf(features.CircuitBreaker, "circuitbreaker")
	appendIf(features.Retry, "retry")
	appendIf(features.LoadShedding, "loadshedding")
	appendIf(features.Discovery, "discovery")
	return names
}

func governanceDisabledNames(disabled map[string]struct{}) []string {
	if len(disabled) == 0 {
		return nil
	}
	names := make([]string, 0, len(disabled))
	for name := range disabled {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// governanceEnabledNames 从 enabled 集合中提取排序后的名称列表。
func governanceEnabledNames(enabled map[string]struct{}) []string {
	if len(enabled) == 0 {
		return nil
	}
	names := make([]string, 0, len(enabled))
	for name := range enabled {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func governanceProviderMap(defaults GovernanceProviderDefaults) map[string]string {
	return map[string]string{
		"configsource":     defaults.ConfigSource,
		"discovery":        defaults.Discovery,
		"selector":         defaults.Selector,
		"rpc":              defaults.RPC,
		"tracing":          defaults.Tracing,
		"metadata":         defaults.Metadata,
		"serviceauth":      defaults.ServiceAuth,
		"circuitbreaker":   defaults.CircuitBreaker,
		"loadshedding":     defaults.LoadShedder,
		"dtm":              defaults.DTM,
		"message_queue":    defaults.MessageQueue,
		"distributed_lock": defaults.DistributedLock,
		"websocket":        defaults.WebSocket,
	}
}

func governanceProviderBackendsFromDecisions(decisions map[string]GovernanceProviderDecision) map[string]string {
	if len(decisions) == 0 {
		return nil
	}
	backends := make(map[string]string, len(decisions))
	for key, decision := range decisions {
		backends[key] = decision.Backend
	}
	return backends
}

func governanceModeDecision(cfg datacontract.Config, modeOverride string) (string, string) {
	if strings.TrimSpace(modeOverride) != "" {
		return "code_override", "mode selected by startup option"
	}
	if cfg == nil {
		return "implicit_default", "mode fell back to monolith because no config key was set"
	}
	for _, key := range []string{
		"governance.mode",
		"app.mode",
		"runtime.mode",
		"service.mode",
	} {
		if value := strings.TrimSpace(cfg.GetString(key)); value != "" {
			return "config_override", fmt.Sprintf("mode selected from config key %s", key)
		}
	}
	return "implicit_default", "mode fell back to monolith because no config key was set"
}

type governanceProviderSpec struct {
	Name               string
	DisabledCapability string
	Factories          providerFactoryRegistry
	Fallback           string
	ModeDefault        string
	ConfigKeys         []string
	ConfigEnabled      func(datacontract.Config) (string, bool)
	ConfigEnabledValue string
	FallbackReason     string
}

func buildGovernanceProviderDecisions(cfg datacontract.Config, mode resiliencecontract.GovernanceMode, configOverrides governanceOverrides, codeOverrides governanceOverrides) map[string]GovernanceProviderDecision {
	defaults := DefaultGovernanceProviderDefaults(mode)
	specs := []governanceProviderSpec{
		{Name: "configsource", Factories: configSourceProviderFactories, Fallback: "local", ModeDefault: defaults.ConfigSource, ConfigKeys: []string{"configsource.backend", "configsource.type", "config_source.backend", "config_source.type"}, FallbackReason: "provider backend fell back to registry default"},
		{Name: "discovery", DisabledCapability: "discovery", Factories: discoveryProviderFactories, Fallback: "noop", ModeDefault: defaults.Discovery, ConfigKeys: []string{"discovery.backend", "discovery.type"}, FallbackReason: "provider backend fell back to noop discovery"},
		{Name: "selector", DisabledCapability: "selector", Factories: selectorProviderFactories, Fallback: "noop", ModeDefault: defaults.Selector, ConfigKeys: []string{"selector.backend", "selector.algorithm", "selector.type"}, FallbackReason: "provider backend fell back to noop selector"},
		{Name: "rpc", Factories: rpcProviderFactories, Fallback: "noop", ModeDefault: defaults.RPC, ConfigKeys: []string{"rpc.mode"}, FallbackReason: "provider backend fell back to noop rpc"},
		{Name: "tracing", DisabledCapability: "tracing", Factories: tracingProviderFactories, Fallback: "noop", ModeDefault: defaults.Tracing, ConfigKeys: []string{"tracing.backend", "tracing.type"}, ConfigEnabled: tracingEnabledFromConfig, ConfigEnabledValue: "otel", FallbackReason: "provider backend fell back to noop tracing"},
		{Name: "metadata", DisabledCapability: "metadata", Factories: metadataProviderFactories, Fallback: "noop", ModeDefault: defaults.Metadata, ConfigKeys: []string{"metadata.mode", "metadata.backend"}, ConfigEnabled: metadataEnabledFromConfig, ConfigEnabledValue: "default", FallbackReason: "provider backend fell back to noop metadata"},
		{Name: "serviceauth", DisabledCapability: "serviceauth", Factories: serviceAuthProviderFactories, Fallback: "noop", ModeDefault: defaults.ServiceAuth, ConfigKeys: []string{"service_auth.backend", "service_auth.mode"}, ConfigEnabled: serviceAuthEnabledFromConfig, ConfigEnabledValue: "token", FallbackReason: "provider backend fell back to noop serviceauth"},
		{Name: "circuitbreaker", DisabledCapability: "circuitbreaker", Factories: circuitBreakerProviderFactories, Fallback: "noop", ModeDefault: defaults.CircuitBreaker, ConfigKeys: []string{"circuit_breaker.backend", "circuit_breaker.type"}, ConfigEnabled: circuitBreakerEnabledFromConfig, ConfigEnabledValue: "sentinel", FallbackReason: "provider backend fell back to noop circuitbreaker"},
		{Name: "loadshedding", DisabledCapability: "loadshedding", Factories: loadShedderProviderFactories, Fallback: "noop", ModeDefault: defaults.LoadShedder, ConfigKeys: []string{"load_shedding.backend", "load_shedding.type"}, ConfigEnabled: loadSheddingEnabledFromConfig, ConfigEnabledValue: "semaphore", FallbackReason: "provider backend fell back to noop loadshedding"},
		{Name: "retry", DisabledCapability: "retry", Factories: retryProviderFactories, Fallback: "noop", ModeDefault: defaults.Retry, ConfigKeys: []string{"retry.backend", "retry.type"}, ConfigEnabled: retryEnabledFromConfig, ConfigEnabledValue: "default", FallbackReason: "provider backend fell back to noop retry"},
		{Name: "dtm", Factories: dtmProviderFactories, Fallback: "noop", ModeDefault: defaults.DTM, ConfigKeys: []string{"dtm.backend", "dtm.type", "dtm.driver"}, ConfigEnabled: dtmEnabledFromConfig, ConfigEnabledValue: "dtmsdk", FallbackReason: "provider backend fell back to noop dtm"},
		{Name: "message_queue", Factories: messageQueueProviderFactories, Fallback: "noop", ModeDefault: defaults.MessageQueue, ConfigKeys: []string{"message_queue.backend", "message_queue.type"}, ConfigEnabled: messageQueueEnabledFromConfig, ConfigEnabledValue: "redis", FallbackReason: "provider backend fell back to noop message queue"},
		{Name: "distributed_lock", Factories: distributedLockProviderFactories, Fallback: "noop", ModeDefault: defaults.DistributedLock, ConfigKeys: []string{"distributed_lock.backend", "distributed_lock.type"}, ConfigEnabled: distributedLockEnabledFromConfig, ConfigEnabledValue: "redis", FallbackReason: "provider backend fell back to noop distributed lock"},
		{Name: "websocket", Factories: webSocketProviderFactories, Fallback: "noop", ModeDefault: defaults.WebSocket, ConfigKeys: []string{"websocket.backend", "websocket.type"}, ConfigEnabled: webSocketEnabledFromConfig, ConfigEnabledValue: "gws", FallbackReason: "provider backend fell back to noop websocket"},
	}

	decisions := make(map[string]GovernanceProviderDecision, len(specs))
	for _, spec := range specs {
		decisions[spec.Name] = resolveGovernanceProviderDecision(cfg, spec, configOverrides, codeOverrides)
	}
	return decisions
}

func resolveGovernanceProviderDecision(cfg datacontract.Config, spec governanceProviderSpec, configOverrides governanceOverrides, codeOverrides governanceOverrides) GovernanceProviderDecision {
	if spec.DisabledCapability != "" {
		if _, ok := codeOverrides.Disabled[spec.DisabledCapability]; ok {
			backend := canonicalGovernanceBackend(spec.Factories, "noop", spec.Fallback)
			return GovernanceProviderDecision{
				Backend:         backend,
				Source:          "code_override",
				Reason:          fmt.Sprintf("%s was disabled by startup option", spec.DisabledCapability),
				FallbackBackend: backend,
			}
		}
		if _, ok := configOverrides.Disabled[spec.DisabledCapability]; ok {
			backend := canonicalGovernanceBackend(spec.Factories, "noop", spec.Fallback)
			return GovernanceProviderDecision{
				Backend:         backend,
				Source:          "config_override",
				Reason:          fmt.Sprintf("%s was disabled by governance.disable", spec.DisabledCapability),
				FallbackBackend: backend,
				ConfigKey:       "governance.disable",
			}
		}
	}

	if backend := strings.TrimSpace(codeOverrides.ProviderBackends[spec.Name]); backend != "" {
		return governanceProviderDecisionFromRequestedBackend(spec, backend, "code_override", fmt.Sprintf("provider backend came from startup option governance.providers.%s", spec.Name), "")
	}
	if backend := strings.TrimSpace(configOverrides.ProviderBackends[spec.Name]); backend != "" {
		return governanceProviderDecisionFromRequestedBackend(spec, backend, "config_override", fmt.Sprintf("provider backend came from config key governance.providers.%s", spec.Name), fmt.Sprintf("governance.providers.%s", spec.Name))
	}

	if backend, key := firstNonEmptyConfigValue(cfg, spec.ConfigKeys...); backend != "" {
		return governanceProviderDecisionFromRequestedBackend(spec, backend, "config_override", fmt.Sprintf("provider backend came from config key %s", key), key)
	}

	if spec.ConfigEnabled != nil {
		if key, ok := spec.ConfigEnabled(cfg); ok {
			return governanceProviderDecisionFromRequestedBackend(spec, spec.ConfigEnabledValue, "config_override", fmt.Sprintf("provider backend was activated by config key %s", key), key)
		}
	}

	if spec.ModeDefault != "" {
		return governanceProviderDecisionFromRequestedBackend(spec, spec.ModeDefault, "mode_default", fmt.Sprintf("provider backend came from current governance mode defaults"), "")
	}

	return governanceProviderDecisionFromRequestedBackend(spec, spec.Fallback, "provider_fallback", spec.FallbackReason, "")
}

func governanceProviderDecisionFromRequestedBackend(spec governanceProviderSpec, requested string, source string, reason string, configKey string) GovernanceProviderDecision {
	resolved := canonicalGovernanceBackend(spec.Factories, requested, spec.Fallback)
	fallbackBackend := ""
	if resolved != requested && strings.TrimSpace(requested) != "" {
		reason = fmt.Sprintf("%s; requested backend %s was unavailable and fell back to %s", reason, requested, resolved)
		fallbackBackend = resolved
	}
	return GovernanceProviderDecision{
		Backend:          resolved,
		Source:           source,
		Reason:           reason,
		RequestedBackend: requested,
		FallbackBackend:  fallbackBackend,
		ConfigKey:        configKey,
	}
}

func canonicalGovernanceBackend(factories providerFactoryRegistry, key string, fallback string) string {
	if _, ok := factories[key]; ok {
		return key
	}
	if _, ok := factories[fallback]; ok {
		return fallback
	}
	return ""
}

func firstNonEmptyConfigValue(cfg datacontract.Config, keys ...string) (string, string) {
	if cfg == nil {
		return "", ""
	}
	for _, key := range keys {
		if value := strings.TrimSpace(cfg.GetString(key)); value != "" {
			return value, key
		}
	}
	return "", ""
}

func tracingEnabledFromConfig(cfg datacontract.Config) (string, bool) {
	return "tracing.enabled", cfg != nil && cfg.GetBool("tracing.enabled")
}

func metadataEnabledFromConfig(cfg datacontract.Config) (string, bool) {
	if cfg == nil {
		return "", false
	}
	if cfg.Get("metadata.propagate_prefix") != nil {
		return "metadata.propagate_prefix", true
	}
	return "metadata.enabled", cfg.GetBool("metadata.enabled")
}

func serviceAuthEnabledFromConfig(cfg datacontract.Config) (string, bool) {
	return "service_auth.enabled", cfg != nil && cfg.GetBool("service_auth.enabled")
}

func circuitBreakerEnabledFromConfig(cfg datacontract.Config) (string, bool) {
	return "circuit_breaker.enabled", cfg != nil && cfg.GetBool("circuit_breaker.enabled")
}

func loadSheddingEnabledFromConfig(cfg datacontract.Config) (string, bool) {
	return "load_shedding.enabled", cfg != nil && cfg.GetBool("load_shedding.enabled")
}

func dtmEnabledFromConfig(cfg datacontract.Config) (string, bool) {
	return "dtm.enabled", cfg != nil && cfg.GetBool("dtm.enabled")
}

func messageQueueEnabledFromConfig(cfg datacontract.Config) (string, bool) {
	return "message_queue.enabled", cfg != nil && cfg.GetBool("message_queue.enabled")
}

func distributedLockEnabledFromConfig(cfg datacontract.Config) (string, bool) {
	return "distributed_lock.enabled", cfg != nil && cfg.GetBool("distributed_lock.enabled")
}
func webSocketEnabledFromConfig(cfg datacontract.Config) (string, bool) {
	return "websocket.enabled", cfg != nil && cfg.GetBool("websocket.enabled")
}
func retryEnabledFromConfig(cfg datacontract.Config) (string, bool) {
	return "retry.enabled", cfg != nil && cfg.GetBool("retry.enabled")
}

// buildGovernanceFeatureDecisions 构建每个 feature 的启用/关闭决策记录。
// 优先级：code disable > config disable > code enable > config enable > mode default。
// 同一 feature 同时被 disable 和 enable 时，disable 生效。
func buildGovernanceFeatureDecisions(mode resiliencecontract.GovernanceMode, defaults resiliencecontract.GovernanceFeatureSet, configDisabled map[string]struct{}, codeDisabled map[string]struct{}, configEnabled map[string]struct{}, codeEnabled map[string]struct{}) map[string]GovernanceFeatureDecision {
	decisions := make(map[string]GovernanceFeatureDecision, 13)
	defaultNames := governanceFeatureNames(defaults)
	defaultSet := make(map[string]struct{}, len(defaultNames))
	for _, name := range defaultNames {
		defaultSet[name] = struct{}{}
	}

	for _, name := range governanceAllFeatureNames() {
		// 最高优先级：代码显式关闭
		if _, ok := codeDisabled[name]; ok {
			decisions[name] = GovernanceFeatureDecision{
				Enabled: false,
				Source:  "code_override",
				Reason:  fmt.Sprintf("%s was explicitly disabled by startup option", name),
			}
			continue
		}
		// 次高优先级：配置显式关闭
		if _, ok := configDisabled[name]; ok {
			decisions[name] = GovernanceFeatureDecision{
				Enabled: false,
				Source:  "config_override",
				Reason:  fmt.Sprintf("%s was explicitly disabled by governance.disable", name),
			}
			continue
		}
		// 代码显式开启（仅在非模式默认时记录来源）
		if _, ok := codeEnabled[name]; ok {
			if _, isDefault := defaultSet[name]; isDefault {
				decisions[name] = GovernanceFeatureDecision{
					Enabled: true,
					Source:  "mode_default",
					Reason:  fmt.Sprintf("%s is enabled by the current governance mode defaults", name),
				}
			} else {
				decisions[name] = GovernanceFeatureDecision{
					Enabled: true,
					Source:  "code_override",
					Reason:  fmt.Sprintf("%s was explicitly enabled by startup option", name),
				}
			}
			continue
		}
		// 配置显式开启（仅在非模式默认时记录来源）
		if _, ok := configEnabled[name]; ok {
			if _, isDefault := defaultSet[name]; isDefault {
				decisions[name] = GovernanceFeatureDecision{
					Enabled: true,
					Source:  "mode_default",
					Reason:  fmt.Sprintf("%s is enabled by the current governance mode defaults", name),
				}
			} else {
				decisions[name] = GovernanceFeatureDecision{
					Enabled: true,
					Source:  "config_override",
					Reason:  fmt.Sprintf("%s was explicitly enabled by governance.enable", name),
				}
			}
			continue
		}
		// 模式默认
		if _, ok := defaultSet[name]; ok {
			decisions[name] = GovernanceFeatureDecision{
				Enabled: true,
				Source:  "mode_default",
				Reason:  fmt.Sprintf("%s is enabled by the current governance mode defaults", name),
			}
			continue
		}
		decisions[name] = GovernanceFeatureDecision{
			Enabled: false,
			Source:  "mode_default",
			Reason:  fmt.Sprintf("%s is not part of the current governance mode defaults", name),
		}
	}
	return decisions
}

func governanceAllFeatureNames() []string {
	return []string{
		"request_identity",
		"logging",
		"recovery",
		"timeout",
		"metrics",
		"metadata",
		"tracing",
		"selector",
		"serviceauth",
		"circuitbreaker",
		"retry",
		"loadshedding",
		"discovery",
	}
}

func governanceProviderNames() []string {
	return []string{
		"configsource",
		"discovery",
		"selector",
		"rpc",
		"tracing",
		"metadata",
		"serviceauth",
		"circuitbreaker",
		"loadshedding",
		"retry",
		"dtm",
		"message_queue",
		"distributed_lock",
		"websocket",
	}
}

func buildGovernanceConfigSnapshot(cfg datacontract.Config, modeOverride string, configOverrides governanceOverrides, codeOverrides governanceOverrides) map[string]any {
	snapshot := make(map[string]any)
	for _, key := range governanceTrackedConfigKeys() {
		if value := governanceSnapshotValue(cfg, key); value != nil {
			snapshot[key] = value
		}
	}
	if strings.TrimSpace(modeOverride) != "" {
		snapshot["startup.governance_mode_override"] = modeOverride
	}
	if len(codeOverrides.Disabled) > 0 {
		snapshot["startup.governance_disable"] = governanceDisabledNames(codeOverrides.Disabled)
	}
	if len(codeOverrides.ProviderBackends) > 0 {
		snapshot["startup.governance_providers"] = cloneGovernanceProviderMap(codeOverrides.ProviderBackends)
	}
	if len(configOverrides.ProviderBackends) > 0 {
		snapshot["governance.providers"] = cloneGovernanceProviderMap(configOverrides.ProviderBackends)
	}
	if len(configOverrides.Disabled) == 0 {
		delete(snapshot, "governance.disable")
	}
	if len(configOverrides.ProviderBackends) == 0 {
		delete(snapshot, "governance.providers")
	}
	return snapshot
}

func governanceTrackedConfigKeys() []string {
	return []string{
		"governance.mode",
		"app.mode",
		"runtime.mode",
		"service.mode",
		"governance.disable",
		"tracing.enabled",
		"tracing.backend",
		"tracing.type",
		"metadata.enabled",
		"metadata.mode",
		"metadata.backend",
		"metadata.propagate_prefix",
		"service_auth.enabled",
		"service_auth.backend",
		"service_auth.mode",
		"selector.backend",
		"selector.algorithm",
		"selector.type",
		"discovery.backend",
		"discovery.type",
		"rpc.mode",
		"circuit_breaker.enabled",
		"circuit_breaker.backend",
		"circuit_breaker.type",
		"load_shedding.enabled",
		"load_shedding.backend",
		"load_shedding.type",
		"retry.enabled",
		"retry.backend",
		"retry.type",
		"dtm.enabled",
		"dtm.backend",
		"dtm.type",
		"dtm.driver",
		"message_queue.enabled",
		"message_queue.backend",
		"message_queue.type",
		"distributed_lock.enabled",
		"distributed_lock.backend",
		"distributed_lock.type",
		"websocket.enabled",
		"websocket.backend",
	}
}

func governanceSnapshotValue(cfg datacontract.Config, key string) any {
	if cfg == nil {
		return nil
	}
	value := cfg.Get(key)
	switch typed := value.(type) {
	case nil:
		return nil
	case string:
		if strings.TrimSpace(typed) == "" {
			return nil
		}
		return typed
	case []string:
		if len(typed) == 0 {
			return nil
		}
		return append([]string(nil), typed...)
	case []any:
		if len(typed) == 0 {
			return nil
		}
		return append([]any(nil), typed...)
	case bool:
		if !typed {
			return nil
		}
		return typed
	default:
		return typed
	}
}

func governanceSortedSnapshotKeys(snapshot map[string]any) []string {
	keys := make([]string, 0, len(snapshot))
	for key := range snapshot {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func governanceFormatSnapshotValue(value any) string {
	switch typed := value.(type) {
	case nil:
		return "null"
	case string:
		return typed
	case []string:
		return "[" + strings.Join(typed, ", ") + "]"
	case []any:
		parts := make([]string, 0, len(typed))
		for _, item := range typed {
			parts = append(parts, governanceFormatSnapshotValue(item))
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case map[string]string:
		keys := make([]string, 0, len(typed))
		for key := range typed {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		parts := make([]string, 0, len(keys))
		for _, key := range keys {
			parts = append(parts, fmt.Sprintf("%s=%s", key, typed[key]))
		}
		return "{" + strings.Join(parts, ", ") + "}"
	case bool:
		if typed {
			return "true"
		}
		return "false"
	default:
		return fmt.Sprintf("%v", typed)
	}
}

func governanceResolutionOrder() []string {
	return []string{
		"code_explicit_override",
		"config_explicit_override",
		"mode_defaults",
		"provider_fallback",
	}
}

// governanceMiddlewareChainOrder 返回中间件链的正式逻辑顺序。
// 入站方向按此顺序执行，出站方向逆序执行。
func governanceMiddlewareChainOrder() []string {
	return []string{
		"request_identity",
		"logging",
		"recovery",
		"cors",
		"security_headers",
		"timeout",
		"load_shedding",
		"rate_limit",
		"circuit_breaker",
		"body_limit",
		"locale",
		"metrics",
		"compression",
	}
}

// governanceRPCClientChainOrder 返回 RPC 出站调用治理链的正式逻辑顺序。
func governanceRPCClientChainOrder() []string {
	return []string{
		"selector",
		"timeout",
		"tracing",
		"metadata",
		"serviceauth",
		"breaker",
		"retry",
	}
}
