// Application scenarios:
// - Define the shared governance-mode and policy-provider contracts used by bootstrap and runtime assembly.
// - Centralize timeout, retry, breaker, and load-shedding policy lookup behind one provider-neutral interface.
// - Keep monolith and microservice mode selection explicit and testable.
//
// 适用场景：
// - 定义 bootstrap 与运行时装配共用的治理模式与策略提供器契约。
// - 将 timeout、retry、breaker、load shedding 的策略读取统一收口到一个 provider-neutral 接口。
// - 让 monolith 与 microservice 两种模式的选择显式、可测试。
package resilience

// GovernanceMode identifies the runtime governance mode.
//
// GovernanceMode 标识运行时治理模式。
type GovernanceMode string

const (
	// GovernanceModeMono keeps the runtime lightweight and local-first.
	//
	// GovernanceModeMono 表示继续走轻量、本地优先的单体主线。
	GovernanceModeMono GovernanceMode = "mono"
	// GovernanceModeMicro enables the default microservice governance mainline.
	//
	// GovernanceModeMicro 表示启用默认微服务治理主线。
	GovernanceModeMicro GovernanceMode = "micro"
)

// HTTPMode identifies the HTTP handling abstraction mode.
// This is an orthogonal dimension to GovernanceMode: HTTP mode controls
// handler signature style (gorp.HTTPContext vs gin.Context), while
// GovernanceMode controls governance capability set.
//
// HTTPMode 标识 HTTP 处理抽象模式。
// 这是与 GovernanceMode 正交的维度：HTTP 模式控制 handler 签名风格
// （gorp.HTTPContext vs gin.Context），GovernanceMode 控制治理能力集。
type HTTPMode string

const (
	// HTTPModeContract uses gorp.HTTPContext abstraction.
	//
	// HTTPModeContract 使用 gorp.HTTPContext 契约抽象。
	HTTPModeContract HTTPMode = "contract"
	// HTTPModeGin uses native gin.Context directly.
	//
	// HTTPModeGin 使用原生 gin.Context。
	HTTPModeGin HTTPMode = "gin"
)

// GovernancePolicyProvider exposes unified policy lookups for runtime governance.
//
// GovernancePolicyProvider 暴露统一的运行时治理策略读取入口。
type GovernancePolicyProvider interface {
	Mode() GovernanceMode
	TimeoutPolicy(resource string) TimeoutPolicy
	RetryPolicy(resource string) RetryPolicy
	BreakerPolicy(resource string) BreakerPolicy
	LoadSheddingPolicy(resource string) LoadSheddingPolicy
}

// TimeoutPolicy describes one timeout strategy.
//
// TimeoutPolicy 描述一条超时策略。
type TimeoutPolicy struct {
	Enabled   bool
	TimeoutMS int
}

// BreakerPolicy describes one circuit-breaker strategy.
//
// BreakerPolicy 描述一条熔断策略。
type BreakerPolicy struct {
	Enabled               bool
	Strategy              string
	Threshold             float64
	MinRequestCount       int64
	MaxConcurrentRequests int64
	RetryTimeoutMs        int64
}
