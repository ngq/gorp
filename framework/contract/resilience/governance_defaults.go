// Application scenarios:
// - Define one mode-aware default governance feature set shared by bootstrap, summaries, and docs.
// - Separate “which governance capabilities are on by default” from concrete provider-backend selection.
// - Keep monolith and microservice default behavior explicit and testable.
//
// 适用场景：
// - 定义一套按模式生效的默认治理能力集合，供 bootstrap、生效摘要和文档共用。
// - 将”默认启用哪些治理能力”与”选择哪个 provider backend”解耦。
// - 让 monolith 与 microservice 两种模式的默认行为显式且可测试。
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
	case GovernanceModeMicro:
		base.MetadataPropagation = true
		base.Tracing = true
		base.Selector = true
		base.ServiceAuth = true
		base.CircuitBreaker = true
		// LoadShedding 在微服务模式下默认启用，提供过载保护。
		// 默认值：MaxConcurrent = runtime.GOMAXPROCS(0) * 100（约 100-800 并发）
		base.LoadShedding = true
	case GovernanceModeMono:
		return base
	default:
		return DefaultGovernanceFeatureSet(GovernanceModeMono)
	}

	return base
}

// GovernanceDefaultsTable captures all governance default values for one mode,
// projected into a serializable, inspection-friendly format.
// This struct is populated lazily (only when view=defaults is requested)
// to avoid bloating the default GovernanceSummary JSON output.
//
// GovernanceDefaultsTable 捕获某个治理模式下所有默认值，
// 以可序列化、可检查的格式呈现。
// 此结构体按需填充（仅在 view=defaults 时请求），
// 避免撑大默认的 GovernanceSummary JSON 输出。
type GovernanceDefaultsTable struct {
	Mode               GovernanceMode     `json:"mode"`
	FeatureDefaults    map[string]bool    `json:"feature_defaults"`
	ProviderDefaults   map[string]string  `json:"provider_defaults"`
	MiddlewareDefaults MiddlewareDefaults `json:"middleware_defaults"`
	RPCClientDefaults  RPCClientDefaults  `json:"rpc_client_defaults"`
}

// MiddlewareDefaults captures the default middleware option values.
//
// MiddlewareDefaults 捕获中间件的默认选项值。
type MiddlewareDefaults struct {
	Timeout           string                 `json:"timeout"`
	BodyLimit         string                 `json:"body_limit"`
	MaxConcurrent     int                    `json:"max_concurrent"`
	EnableMetrics     bool                   `json:"enable_metrics"`
	EnableCompression bool                   `json:"enable_compression"`
	CORS              CORSDefaults           `json:"cors"`
	SecurityHeaders   SecurityHeaderDefaults `json:"security_headers"`
	Locale            LocaleDefaults         `json:"locale"`
}

// CORSDefaults captures CORS-specific default values (applied when CORS is explicitly enabled).
//
// CORSDefaults 捕获 CORS 相关默认值（显式启用 CORS 时生效）。
type CORSDefaults struct {
	AllowOrigins  []string `json:"allow_origins"`
	MaxAgeSeconds int      `json:"max_age_seconds"`
}

// SecurityHeaderDefaults captures security-header-specific default values.
//
// SecurityHeaderDefaults 捕获安全头相关默认值。
type SecurityHeaderDefaults struct {
	XFrameOptions       string `json:"x_frame_options"`
	XContentTypeOptions string `json:"x_content_type_options"`
	ReferrerPolicy      string `json:"referrer_policy"`
}

// LocaleDefaults captures locale-specific default values.
//
// LocaleDefaults 捕获本地化相关默认值。
type LocaleDefaults struct {
	Supported []string `json:"supported"`
	Default   string   `json:"default"`
	QueryKeys []string `json:"query_keys"`
}

// RPCClientDefaults captures the default RPC client option values.
//
// RPCClientDefaults 捕获 RPC 客户端默认选项值。
type RPCClientDefaults struct {
	Timeout string `json:"timeout"`
}
