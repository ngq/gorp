// Package p2c provides Power of Two Choices load balancing selector.
// The P2C algorithm picks two random instances and selects the one with lower load.
// Load calculation considers: pending requests, failure rate, latency EWMA.
// Eg:
//
// P2C（二选一）负载均衡选择器包，提供基于负载感知的选择算法。
// P2C 算法随机选择两个实例，然后选择负载较低的那个。
// 负载计算考虑：待处理请求、失败率、延迟 EWMA。
// Eg:
//
//	// 注册 Provider
//	app.Register(p2c.NewProvider())
//
//	// 使用选择器
//	selector := c.MustMake(discoverycontract.SelectorKey).(discoverycontract.Selector)
//	instance, doneFunc, _ := selector.Select(ctx, instances)
package p2c

import (
	"context"
	"math/rand"
	"sync"
	"time"

	discoverycontract "github.com/ngq/gorp/framework/contract/discovery"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

// Provider registers the P2C selector contract.
//
// Provider 注册 P2C 选择器契约。
type Provider struct{}

// NewProvider creates a new P2C selector provider instance.
//
// NewProvider 创建新的 P2C 选择器 Provider 实例。
func NewProvider() *Provider { return &Provider{} }

// Name returns the provider name "selector.p2c".
//
// Name 返回 Provider 名称 "selector.p2c"。
func (p *Provider) Name() string { return "selector.p2c" }

// IsDefer returns true, selector can be deferred until first use.
//
// IsDefer 返回 true，选择器可延迟初始化直到首次使用。
func (p *Provider) IsDefer() bool { return true }

// Provides returns the selector contract keys.
//
// Provides 返回选择器契约键列表。
func (p *Provider) Provides() []string {
	return []string{discoverycontract.SelectorKey, discoverycontract.SelectorBuilderKey}
}

// Register binds the P2C selector factory to the container.
//
// Register 将 P2C 选择器工厂绑定到容器。
func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(discoverycontract.SelectorBuilderKey, func(c runtimecontract.Container) (any, error) {
		return &p2cBuilder{}, nil
	}, true)
	c.Bind(discoverycontract.SelectorKey, func(c runtimecontract.Container) (any, error) {
		return (&p2cBuilder{}).Build(), nil
	}, true)
	return nil
}

// Boot is a no-op for P2C selector provider.
//
// Boot P2C 选择器 Provider 无启动逻辑。
func (p *Provider) Boot(c runtimecontract.Container) error {
	return nil
}

// p2cBuilder builds P2C selector instances.
//
// p2cBuilder 构建 P2C 选择器实例。
type p2cBuilder struct{}

// Build creates a new P2C selector.
//
// Build 创建新的 P2C 选择器。
func (b *p2cBuilder) Build() discoverycontract.Selector {
	return NewP2CSelector()
}

// P2CSelector implements discoverycontract.Selector with Power of Two Choices algorithm.
//
// P2CSelector 使用 P2C 算法实现 discoverycontract.Selector 接口。
type P2CSelector struct {
	mu            sync.Mutex                // mu protects instance stats.
	                                        //
	                                         // mu 保护实例统计。
	r             *rand.Rand                // r is the random generator.
	                                        //
	                                         // r 随机数生成器。
	instanceStats map[string]*InstanceStats // instanceStats tracks per-instance metrics.
	                                        //
	                                         // instanceStats 跟踪每个实例的指标。
}

// InstanceStats tracks load metrics for a single instance.
//
// InstanceStats 跟踪单个实例的负载指标。
type InstanceStats struct {
	pending        int64     // pending is the count of active requests.
	                          //
	                           // pending 活跃请求计数。
	successCount   int64     // successCount is the count of successful requests.
	                          //
	                           // successCount 成功请求计数。
	failCount      int64     // failCount is the count of failed requests.
	                          //
	                           // failCount 失败请求计数。
	latencyEWMA    float64   // latencyEWMA is the exponentially weighted moving average of latency.
	                          //
	                           // latencyEWMA 延迟的指数加权移动平均值。
	latencySamples int64     // latencySamples is the count of latency samples collected.
	                          //
	                           // latencySamples 收集的延迟样本计数。
	lastUpdate     time.Time // lastUpdate is the timestamp of last update.
	                          //
	                           // lastUpdate 最后更新时间戳。
}

// NewP2CSelector creates a new P2C selector instance.
//
// NewP2CSelector 创建新的 P2C 选择器实例。
func NewP2CSelector() *P2CSelector {
	return &P2CSelector{
		r:             rand.New(rand.NewSource(time.Now().UnixNano())),
		instanceStats: make(map[string]*InstanceStats),
	}
}

// Select picks an instance using P2C algorithm with load awareness.
// Core logic: Pick two random instances, compare their scores, pick lower load.
//
// Select 使用 P2C 算法选择实例，考虑负载感知。
// 核心逻辑：随机选择两个实例，比较它们的分数，选择负载较低的那个。
func (s *P2CSelector) Select(ctx context.Context, instances []transportcontract.ServiceInstance, opts ...discoverycontract.SelectOption) (
	selected transportcontract.ServiceInstance, done discoverycontract.DoneFunc, err error,
) {
	options := &discoverycontract.SelectOptions{}
	for _, opt := range opts {
		opt(options)
	}

	if options.ForceInstance != nil {
		inst := *options.ForceInstance
		s.incrementPending(inst.Address)
		return inst, s.createDoneFunc(inst), nil
	}

	filtered := s.filterHealthy(instances, options.Filters)
	if len(filtered) == 0 {
		return transportcontract.ServiceInstance{}, noopDone, discoverycontract.ErrNoAvailable
	}

	if len(filtered) == 1 {
		inst := filtered[0]
		s.incrementPending(inst.Address)
		return inst, s.createDoneFunc(inst), nil
	}

	s.mu.Lock()
	idx1 := s.r.Intn(len(filtered))
	idx2 := s.r.Intn(len(filtered))
	for idx2 == idx1 && len(filtered) > 1 {
		idx2 = s.r.Intn(len(filtered))
	}

	inst1 := filtered[idx1]
	inst2 := filtered[idx2]
	score1 := s.calculateScore(inst1)
	score2 := s.calculateScore(inst2)
	s.mu.Unlock()

	var chosen transportcontract.ServiceInstance
	if score1 <= score2 {
		chosen = inst1
	} else {
		chosen = inst2
	}

	s.incrementPending(chosen.Address)

	return chosen, s.createDoneFunc(chosen), nil
}

// calculateScore calculates load score for an instance.
// Formula: pending + failRate*10 + latencyEWMA*0.001.
//
// calculateScore 计算实例的负载分数。
// 计算公式：pending + failRate*10 + latencyEWMA*0.001。
func (s *P2CSelector) calculateScore(instance transportcontract.ServiceInstance) float64 {
	stats := s.instanceStats[instance.Address]
	if stats == nil {
		return 0
	}

	totalRequests := stats.successCount + stats.failCount
	failRate := 0.0
	if totalRequests > 0 {
		failRate = float64(stats.failCount) / float64(totalRequests)
	}

	return float64(stats.pending) + failRate*10 + stats.latencyEWMA*0.001
}

// incrementPending increases pending request count for an instance.
//
// incrementPending 增加实例的待处理请求计数。
func (s *P2CSelector) incrementPending(address string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	stats := s.instanceStats[address]
	if stats == nil {
		stats = &InstanceStats{lastUpdate: time.Now()}
		s.instanceStats[address] = stats
	}
	stats.pending++
	stats.lastUpdate = time.Now()
}

// createDoneFunc creates a DoneFunc that updates instance stats after request completion.
//
// createDoneFunc 创建 DoneFunc，在请求完成后更新实例统计。
func (s *P2CSelector) createDoneFunc(instance transportcontract.ServiceInstance) discoverycontract.DoneFunc {
	return func(ctx context.Context, info discoverycontract.DoneInfo) {
		s.mu.Lock()
		defer s.mu.Unlock()

		stats := s.instanceStats[instance.Address]
		if stats == nil {
			return
		}

		stats.pending--
		if stats.pending < 0 {
			stats.pending = 0
		}

		if info.Err != nil {
			stats.failCount++
		} else {
			stats.successCount++
		}
		if info.Latency > 0 {
			sample := float64(info.Latency.Milliseconds())
			if stats.latencySamples == 0 {
				stats.latencyEWMA = sample
			} else {
				const alpha = 0.2
				stats.latencyEWMA = alpha*sample + (1-alpha)*stats.latencyEWMA
			}
			stats.latencySamples++
		}

		stats.lastUpdate = time.Now()
	}
}

// filterHealthy filters instances by health status and custom filters.
//
// filterHealthy 根据健康状态和自定义过滤器过滤实例。
func (s *P2CSelector) filterHealthy(instances []transportcontract.ServiceInstance, filters []discoverycontract.NodeFilter) []transportcontract.ServiceInstance {
	if len(instances) == 0 {
		return nil
	}

	result := make([]transportcontract.ServiceInstance, 0, len(instances))
	for _, inst := range instances {
		if !inst.Healthy {
			continue
		}
		ok := true
		for _, f := range filters {
			if !f(inst) {
				ok = false
				break
			}
		}
		if ok {
			result = append(result, inst)
		}
	}
	return result
}

// noopDone is a no-op DoneFunc for edge cases.
//
// noopDone 是边缘情况的空 DoneFunc。
func noopDone(ctx context.Context, info discoverycontract.DoneInfo) {}