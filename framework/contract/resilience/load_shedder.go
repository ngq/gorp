// Application scenarios:
// - Define the provider-neutral load-shedding contract used by HTTP, gRPC, and RPC client governance.
// - Separate "should reject now" semantics from concrete adaptive or static implementations.
// - Provide one shared policy/config model for overload protection.
//
// 适用场景：
// - 定义 HTTP、gRPC 与 RPC client 治理共用的 provider-neutral 过载保护契约。
// - 将“当前是否应立即拒绝”语义与具体静态或自适应实现分离。
// - 为过载保护提供统一的策略与配置模型。
package resilience

import "context"

// LoadShedderKey is the container key for the load-shedding capability.
//
// LoadShedderKey 是过载保护能力的容器键。
const LoadShedderKey = "framework.load_shedder"

// LoadShedder decides whether a request should be shed immediately.
//
// LoadShedder 用于判断一个请求是否应被立即丢弃。
type LoadShedder interface {
	Allow(ctx context.Context, resource string) error
	Done(ctx context.Context, resource string, err error)
}

// LoadSheddingConfig describes runtime overload-protection settings.
//
// LoadSheddingConfig 描述运行时过载保护配置。
type LoadSheddingConfig struct {
	Enabled          bool
	Strategy         string
	MaxConcurrency   int
	ResourcePolicies map[string]LoadSheddingPolicy
	DefaultPolicy    LoadSheddingPolicy
}

// LoadSheddingPolicy describes one overload-protection policy.
//
// LoadSheddingPolicy 描述一条过载保护策略。
type LoadSheddingPolicy struct {
	Enabled        bool
	Strategy       string
	MaxConcurrency int
}

// GetPolicy returns the resource-specific policy or falls back to the default one.
//
// GetPolicy 返回资源级策略；若未命中，则回退到默认策略。
func (c *LoadSheddingConfig) GetPolicy(resource string) LoadSheddingPolicy {
	if c == nil {
		return LoadSheddingPolicy{}
	}
	if policy, ok := c.ResourcePolicies[resource]; ok {
		return policy
	}
	return c.DefaultPolicy
}
