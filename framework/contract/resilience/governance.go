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
	// GovernanceModeMonolith keeps the runtime lightweight and local-first.
	//
	// GovernanceModeMonolith 表示继续走轻量、本地优先的单体主线。
	GovernanceModeMonolith GovernanceMode = "monolith"
	// GovernanceModeMicroservice enables the default microservice governance mainline.
	//
	// GovernanceModeMicroservice 表示启用默认微服务治理主线。
	GovernanceModeMicroservice GovernanceMode = "microservice"
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
