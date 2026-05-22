package governance

import (
	"time"

	discoverycontract "github.com/ngq/gorp/framework/contract/discovery"
	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	securitycontract "github.com/ngq/gorp/framework/contract/security"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

// DefaultClientPresetOptions controls the default outbound RPC governance preset.
//
// DefaultClientPresetOptions 用于控制默认出站 RPC 治理预设。
type DefaultClientPresetOptions struct {
	Timeout time.Duration
}

// DefaultClientPresetDependencies contains all dependencies needed to build the full 7-stage RPC governance chain.
// Each field is optional; nil dependencies result in the corresponding middleware being skipped.
//
// DefaultClientPresetDependencies 包含构建完整 7 阶段 RPC 治理链所需的所有依赖。
// 每个字段都是可选的；nil 依赖会导致对应中间件被跳过。
type DefaultClientPresetDependencies struct {
	Selector           discoverycontract.Selector
	Registry           transportcontract.ServiceRegistry
	Tracer             observabilitycontract.Tracer
	ServiceName        string
	MetadataPropagator transportcontract.MetadataPropagator
	ServiceAuthIssuer  securitycontract.ServiceTokenIssuer
	CircuitBreaker     resiliencecontract.CircuitBreaker
	Retry              resiliencecontract.Retry
	LoadShedder        resiliencecontract.LoadShedder
	ResourceNamer      func(service, method string) string
}

// DefaultClientPresetOrder returns the stable logical order of the outbound governance chain.
//
// DefaultClientPresetOrder 返回出站治理链的稳定逻辑顺序。
func DefaultClientPresetOrder() []string {
	return []string{
		"selector",
		"timeout",
		"tracing",
		"metadata",
		"serviceauth",
		"loadshedding",
		"breaker",
		"retry",
	}
}

// DefaultClientPresetSet builds the complete 8-stage RPC governance middleware chain from options and dependencies.
// Returns middleware in execution order: selector → timeout → tracing → metadata → serviceauth → loadshedding → breaker → retry.
// Nil dependencies cause corresponding middleware to be skipped.
//
// DefaultClientPresetSet 从 options 和 dependencies 构建完整 8 阶段 RPC 治理中间件链。
// 返回的中间件按执行顺序排列：selector → timeout → tracing → metadata → serviceauth → loadshedding → breaker → retry。
// nil 依赖会导致对应中间件被跳过。
func DefaultClientPresetSet(opts DefaultClientPresetOptions, deps DefaultClientPresetDependencies) []transportcontract.RPCClientMiddleware {
	chain := make([]transportcontract.RPCClientMiddleware, 0, 8)

	// 1. selector - instance selection before routing
	// selector - 路由前的实例选择
	if deps.Selector != nil && deps.Registry != nil {
		chain = append(chain, SelectorMiddleware(deps.Selector, deps.Registry))
	}

	// 2. timeout - request timeout control
	// timeout - 请求超时控制
	if opts.Timeout > 0 {
		chain = append(chain, TimeoutMiddleware(opts.Timeout))
	}

	// 3. tracing - distributed tracing span creation
	// tracing - 分布式追踪 span 创建
	if deps.Tracer != nil {
		chain = append(chain, TracingMiddleware(deps.Tracer, deps.ServiceName))
	}

	// 4. metadata - context metadata propagation
	// metadata - context metadata 传播
	if deps.MetadataPropagator != nil {
		chain = append(chain, MetadataMiddleware(deps.MetadataPropagator))
	}

	// 5. serviceauth - service-to-service authentication token injection
	// serviceauth - 服务间认证令牌注入
	if deps.ServiceAuthIssuer != nil {
		chain = append(chain, ServiceAuthMiddleware(deps.ServiceAuthIssuer))
	}

	// 6. loadshedding - overload protection
	// loadshedding - 过载保护
	if deps.LoadShedder != nil {
		chain = append(chain, LoadSheddingMiddleware(deps.LoadShedder, deps.ResourceNamer))
	}

	// 7. breaker - circuit breaker protection
	// breaker - 熔断器保护
	if deps.CircuitBreaker != nil {
		chain = append(chain, CircuitBreakerMiddleware(deps.CircuitBreaker, deps.ResourceNamer))
	}

	// 8. retry - automatic retry on failure
	// retry - 失败自动重试
	if deps.Retry != nil {
		chain = append(chain, RetryMiddlewareWithResource(deps.Retry, deps.ResourceNamer))
	}

	return chain
}
