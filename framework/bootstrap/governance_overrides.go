// Application scenarios:
// - Parse explicit governance overrides from config without requiring callers to repeat all defaults.
// - Support low-cost "disable this capability" and "replace this backend" control paths.
// - Keep override semantics reusable across selectors, summaries, and future bootstrap helpers.
//
// 适用场景：
// - 解析显式治理覆盖配置，而不要求调用方展开所有默认值。
// - 支持低成本“关闭某个治理能力”和“替换某个 provider backend”控制路径。
// - 让覆盖语义可供 selector、生效摘要和后续 bootstrap helper 复用。
package bootstrap

import (
	"strings"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
)

type governanceOverrides struct {
	Disabled         map[string]struct{}
	ProviderBackends map[string]string
}

func splitGovernanceOverrides(cfg datacontract.Config) (configOverrides governanceOverrides, codeOverrides governanceOverrides, configView datacontract.Config) {
	configView = cfg
	if overlay, ok := cfg.(*governanceOverlayConfig); ok {
		configView = overlay.base
		codeOverrides = governanceOverrides{
			Disabled:         normalizeGovernanceDisabledList(overlay.governanceDisable),
			ProviderBackends: cloneGovernanceProviderMap(overlay.governanceProviders),
		}
	}

	configOverrides = governanceOverrides{
		Disabled:         loadGovernanceDisabledSet(configView),
		ProviderBackends: loadGovernanceProviderOverrides(configView),
	}
	return configOverrides, codeOverrides, configView
}

func loadGovernanceOverrides(cfg datacontract.Config) governanceOverrides {
	configOverrides, codeOverrides, _ := splitGovernanceOverrides(cfg)
	return governanceOverrides{
		Disabled:         mergeGovernanceDisabled(configOverrides.Disabled, codeOverrides.Disabled),
		ProviderBackends: mergeGovernanceProviderBackends(configOverrides.ProviderBackends, codeOverrides.ProviderBackends),
	}
}

func loadGovernanceDisabledSet(cfg datacontract.Config) map[string]struct{} {
	if cfg == nil {
		return map[string]struct{}{}
	}
	return normalizeGovernanceDisabledValue(cfg.Get("governance.disable"))
}

func normalizeGovernanceDisabledValue(raw any) map[string]struct{} {
	set := make(map[string]struct{})
	switch values := raw.(type) {
	case []string:
		for _, value := range values {
			value = normalizeGovernanceKey(value)
			if value != "" {
				set[value] = struct{}{}
			}
		}
	case []any:
		for _, value := range values {
			if s, ok := value.(string); ok {
				s = normalizeGovernanceKey(s)
				if s != "" {
					set[s] = struct{}{}
				}
			}
		}
	}
	return set
}

func normalizeGovernanceDisabledList(values []string) map[string]struct{} {
	return normalizeGovernanceDisabledValue(values)
}

func loadGovernanceProviderOverrides(cfg datacontract.Config) map[string]string {
	if cfg == nil {
		return nil
	}

	keys := []string{
		"configsource",
		"discovery",
		"selector",
		"rpc",
		"tracing",
		"metadata",
		"serviceauth",
		"circuitbreaker",
		"dtm",
		"message_queue",
		"distributed_lock",
	}

	overrides := make(map[string]string)
	for _, key := range keys {
		if value := strings.TrimSpace(cfg.GetString("governance.providers." + key)); value != "" {
			overrides[key] = value
		}
	}
	return overrides
}

func governanceProviderOverride(cfg datacontract.Config, key string) string {
	if cfg == nil {
		return ""
	}
	return strings.TrimSpace(cfg.GetString("governance.providers." + key))
}

func isGovernanceCapabilityDisabled(cfg datacontract.Config, name string) bool {
	_, ok := loadGovernanceDisabledSet(cfg)[normalizeGovernanceKey(name)]
	return ok
}

func applyGovernanceFeatureDisables(features resiliencecontract.GovernanceFeatureSet, disabled map[string]struct{}) resiliencecontract.GovernanceFeatureSet {
	if len(disabled) == 0 {
		return features
	}

	if _, ok := disabled["request_identity"]; ok {
		features.RequestIdentity = false
	}
	if _, ok := disabled["logging"]; ok {
		features.Logging = false
	}
	if _, ok := disabled["recovery"]; ok {
		features.Recovery = false
	}
	if _, ok := disabled["timeout"]; ok {
		features.Timeout = false
	}
	if _, ok := disabled["metrics"]; ok {
		features.Metrics = false
	}
	if _, ok := disabled["metadata"]; ok {
		features.MetadataPropagation = false
	}
	if _, ok := disabled["tracing"]; ok {
		features.Tracing = false
	}
	if _, ok := disabled["selector"]; ok {
		features.Selector = false
	}
	if _, ok := disabled["serviceauth"]; ok {
		features.ServiceAuth = false
	}
	if _, ok := disabled["circuitbreaker"]; ok {
		features.CircuitBreaker = false
	}
	if _, ok := disabled["retry"]; ok {
		features.Retry = false
	}
	if _, ok := disabled["loadshedding"]; ok {
		features.LoadShedding = false
	}
	if _, ok := disabled["discovery"]; ok {
		features.Discovery = false
	}

	return features
}

func applyGovernanceProviderOverrides(defaults GovernanceProviderDefaults, overrides map[string]string) GovernanceProviderDefaults {
	if len(overrides) == 0 {
		return defaults
	}

	if value := overrides["configsource"]; value != "" {
		defaults.ConfigSource = value
	}
	if value := overrides["discovery"]; value != "" {
		defaults.Discovery = value
	}
	if value := overrides["selector"]; value != "" {
		defaults.Selector = value
	}
	if value := overrides["rpc"]; value != "" {
		defaults.RPC = value
	}
	if value := overrides["tracing"]; value != "" {
		defaults.Tracing = value
	}
	if value := overrides["metadata"]; value != "" {
		defaults.Metadata = value
	}
	if value := overrides["serviceauth"]; value != "" {
		defaults.ServiceAuth = value
	}
	if value := overrides["circuitbreaker"]; value != "" {
		defaults.CircuitBreaker = value
	}
	if value := overrides["dtm"]; value != "" {
		defaults.DTM = value
	}
	if value := overrides["message_queue"]; value != "" {
		defaults.MessageQueue = value
	}
	if value := overrides["distributed_lock"]; value != "" {
		defaults.DistributedLock = value
	}
	return defaults
}

func normalizeGovernanceKey(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	replacer := strings.NewReplacer("-", "", "_", "", " ", "")
	value = replacer.Replace(value)
	switch value {
	case "requestidentity":
		return "request_identity"
	case "serviceauth":
		return "serviceauth"
	case "circuitbreaker":
		return "circuitbreaker"
	case "messagequeue":
		return "message_queue"
	case "distributedlock":
		return "distributed_lock"
	case "loadshedding":
		return "loadshedding"
	case "metadata":
		return "metadata"
	default:
		return value
	}
}

func mergeGovernanceDisabled(base map[string]struct{}, overlay map[string]struct{}) map[string]struct{} {
	if len(base) == 0 && len(overlay) == 0 {
		return nil
	}
	merged := make(map[string]struct{}, len(base)+len(overlay))
	for key := range base {
		merged[key] = struct{}{}
	}
	for key := range overlay {
		merged[key] = struct{}{}
	}
	return merged
}

func mergeGovernanceProviderBackends(base map[string]string, overlay map[string]string) map[string]string {
	if len(base) == 0 && len(overlay) == 0 {
		return nil
	}
	merged := cloneGovernanceProviderMap(base)
	if merged == nil {
		merged = make(map[string]string, len(overlay))
	}
	for key, value := range overlay {
		merged[key] = value
	}
	return merged
}
