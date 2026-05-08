// Application scenarios:
// - Define one mode-aware default governance feature set shared by bootstrap, summaries, and docs.
// - Separate "which governance capabilities are on by default" from concrete provider-backend selection.
// - Keep monolith, microservice, and gin-first default behavior explicit and testable.
//
// 适用场景：
// - 定义一套按模式生效的默认治理能力集合，供 bootstrap、生效摘要和文档共用。
// - 将“默认启用哪些治理能力”与“选择哪个 provider backend”解耦。
// - 让 monolith、microservice、gin-first 三种模式的默认行为显式且可测试。
package resilience

// GovernanceFeatureSet describes which governance capabilities are enabled by default.
//
// GovernanceFeatureSet 描述某个治理模式下默认启用的治理能力集合。
type GovernanceFeatureSet struct {
	RequestIdentity bool
	Logging         bool
	Recovery        bool
	Timeout         bool
	Metrics         bool

	MetadataPropagation bool
	Tracing             bool
	Selector            bool
	ServiceAuth         bool
	CircuitBreaker      bool
	Retry               bool
	LoadShedding        bool
	Discovery           bool
}

// DefaultGovernanceFeatureSet returns the mode-aware default governance feature set.
//
// DefaultGovernanceFeatureSet 返回按治理模式生效的默认治理能力集合。
func DefaultGovernanceFeatureSet(mode GovernanceMode) GovernanceFeatureSet {
	base := GovernanceFeatureSet{
		RequestIdentity: true,
		Logging:         true,
		Recovery:        true,
		Timeout:         true,
		Metrics:         true,
	}

	switch mode {
	case GovernanceModeMicroservice:
		base.MetadataPropagation = true
		base.Tracing = true
		base.Selector = true
		base.ServiceAuth = true
		base.CircuitBreaker = true
	case GovernanceModeGinFirst:
		fallthrough
	case GovernanceModeMonolith:
		return base
	default:
		return DefaultGovernanceFeatureSet(GovernanceModeMonolith)
	}

	return base
}
