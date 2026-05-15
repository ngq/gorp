// Package wrr provides weighted round-robin load balancing selector.
// The selector picks instances based on their weight values in metadata.
// Supported weight values in metadata["weight"]: 100, 80, 50, 20, 10, default 1.
// When the instance count exceeds wrrP2CFallbackThreshold (100), it automatically
// falls back to P2C to avoid O(n) per-selection overhead becoming a CPU hotspot.
//
// 加权轮询负载均衡选择器包，提供基于权重的轮询选择算法。
// 选择器根据实例 metadata 中的权重值进行选择。
// 支持的 metadata["weight"] 权重值：100, 80, 50, 20, 10，默认为 1。
// 当实例数量超过 wrrP2CFallbackThreshold (100) 时，自动降级到 P2C，
// 避免 O(n) 的每次选择开销成为 CPU 热点。
//
// Eg:
//
//	// 注册 Provider
//	app.Register(wrr.NewProvider())
//
//	// 使用选择器
//	selector := c.MustMake(discoverycontract.SelectorKey).(discoverycontract.Selector)
//	instance, doneFunc, _ := selector.Select(ctx, instances)
package wrr

import (
	"context"
	"sync"

	discoverycontract "github.com/ngq/gorp/framework/contract/discovery"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	"github.com/ngq/gorp/framework/provider/selector/p2c"
)

// wrrP2CFallbackThreshold is the instance count above which WRR falls back to P2C.
// WRR has O(n) weight computation per selection; at large scale P2C is faster
// and provides comparable adaptive load balancing.
//
// wrrP2CFallbackThreshold 是实例数量阈值，超过此值后 WRR 自动降级到 P2C。
// WRR 每次选择需要 O(n) 权重计算；大规模场景下 P2C 更快且提供类似的自适应负载均衡。
const wrrP2CFallbackThreshold = 100

// Provider registers the WRR selector contract.
//
// Provider 注册加权轮询选择器契约。
type Provider struct{}

// NewProvider creates a new WRR selector provider instance.
//
// NewProvider 创建新的加权轮询选择器 Provider 实例。
func NewProvider() *Provider { return &Provider{} }

// Name returns the provider name "selector.wrr".
//
// Name 返回 Provider 名称 "selector.wrr"。
func (p *Provider) Name() string { return "selector.wrr" }

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

// DependsOn returns the keys this provider depends on.
// WRR selector has no dependencies.
//
// DependsOn 返回该 provider 依赖的 key。
// WRR selector 无依赖。
func (p *Provider) DependsOn() []string { return nil }

// Register binds the WRR selector factory to the container.
//
// Register 将加权轮询选择器工厂绑定到容器。
func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(discoverycontract.SelectorBuilderKey, func(c runtimecontract.Container) (any, error) {
		return &wrrBuilder{}, nil
	}, true)
	c.Bind(discoverycontract.SelectorKey, func(c runtimecontract.Container) (any, error) {
		return (&wrrBuilder{}).Build(), nil
	}, true)
	return nil
}

// Boot is a no-op for WRR selector provider.
//
// Boot 加权轮询选择器 Provider 无启动逻辑。
func (p *Provider) Boot(c runtimecontract.Container) error {
	return nil
}

// wrrBuilder builds WRR selector instances.
//
// wrrBuilder 构建加权轮询选择器实例。
type wrrBuilder struct{}

// Build creates a new WRR selector.
//
// Build 创建新的加权轮询选择器。
func (b *wrrBuilder) Build() discoverycontract.Selector {
	return NewWRRSelector()
}

// WRRSelector implements discoverycontract.Selector with weighted round-robin algorithm.
// When the instance count exceeds wrrP2CFallbackThreshold, it falls back to P2C
// to avoid the O(n) per-selection overhead becoming a CPU hotspot.
//
// WRRSelector 使用加权轮询算法实现 discoverycontract.Selector 接口。
// 当实例数量超过 wrrP2CFallbackThreshold 时，自动降级到 P2C，
// 避免 O(n) 的每次选择开销成为 CPU 热点。
type WRRSelector struct {
	mu            sync.Mutex
	currentWeight map[string]float64
	lastInstances []transportcontract.ServiceInstance
	fallback      discoverycontract.Selector // 大实例集的降级选择器。
}

// NewWRRSelector creates a new WRR selector instance.
//
// NewWRRSelector 创建新的加权轮询选择器实例。
func NewWRRSelector() *WRRSelector {
	return &WRRSelector{
		fallback:      p2c.NewP2CSelector(),
		currentWeight: make(map[string]float64),
	}
}

// Select picks an instance using weighted round-robin algorithm.
// Core logic: Calculate weights, add to current, pick max, subtract total.
// Falls back to P2C when instance count exceeds wrrP2CFallbackThreshold.
//
// Select 使用加权轮询算法选择实例。
// 核心逻辑：计算权重、累加到当前权重、选择最大值、减去总权重。
// 实例数量超过 wrrP2CFallbackThreshold 时自动降级到 P2C。
func (s *WRRSelector) Select(ctx context.Context, instances []transportcontract.ServiceInstance, opts ...discoverycontract.SelectOption) (
	selected transportcontract.ServiceInstance, done discoverycontract.DoneFunc, err error,
) {
	options := &discoverycontract.SelectOptions{}
	for _, opt := range opts {
		opt(options)
	}

	if options.ForceInstance != nil {
		return *options.ForceInstance, noopDone, nil
	}

	filtered := s.filterHealthy(instances, options.Filters)
	if len(filtered) == 0 {
		return transportcontract.ServiceInstance{}, noopDone, discoverycontract.ErrNoAvailable
	}
	// 大实例集降级到 P2C，避免 O(n) 权重计算成为 CPU 热点。
	if len(filtered) > wrrP2CFallbackThreshold {
		return s.fallback.Select(ctx, filtered, opts...)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.cleanupStaleWeights(filtered)

	var totalWeight float64
	var maxWeight float64
	var chosenInstance transportcontract.ServiceInstance

	for _, instance := range filtered {
		weight := s.getWeight(instance)
		totalWeight += weight
		current := s.currentWeight[instance.Address] + weight
		s.currentWeight[instance.Address] = current

		if current > maxWeight {
			maxWeight = current
			chosenInstance = instance
		}
	}

	if chosenInstance.Address != "" {
		s.currentWeight[chosenInstance.Address] -= totalWeight
	}

	s.lastInstances = filtered

	return chosenInstance, noopDone, nil
}

// filterHealthy filters instances by health status and custom filters.
//
// filterHealthy 根据健康状态和自定义过滤器过滤实例。
func (s *WRRSelector) filterHealthy(instances []transportcontract.ServiceInstance, filters []discoverycontract.NodeFilter) []transportcontract.ServiceInstance {
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

// getWeight retrieves weight from instance metadata.
// Supported values: 100, 80, 50, 20, 10. Default is 1.
//
// getWeight 从实例 metadata 获取权重值。
// 支持的值：100, 80, 50, 20, 10。默认为 1。
func (s *WRRSelector) getWeight(instance transportcontract.ServiceInstance) float64 {
	if instance.Metadata == nil {
		return 1
	}
	if w, ok := instance.Metadata["weight"]; ok {
		switch w {
		case "100":
			return 100
		case "80":
			return 80
		case "50":
			return 50
		case "20":
			return 20
		case "10":
			return 10
		default:
			return 1
		}
	}
	return 1
}

// cleanupStaleWeights removes weight entries for instances no longer present.
//
// cleanupStaleWeights 删除不再存在的实例的权重记录。
func (s *WRRSelector) cleanupStaleWeights(currentInstances []transportcontract.ServiceInstance) {
	currentAddrSet := make(map[string]bool, len(currentInstances))
	for _, inst := range currentInstances {
		currentAddrSet[inst.Address] = true
	}

	for addr := range s.currentWeight {
		if !currentAddrSet[addr] {
			delete(s.currentWeight, addr)
		}
	}
}

// noopDone is a no-op DoneFunc for WRR selector.
//
// noopDone 是加权轮询选择器的空 DoneFunc。
func noopDone(ctx context.Context, info discoverycontract.DoneInfo) {}
