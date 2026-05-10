// Package wrr provides weighted round-robin load balancing selector.
// The selector picks instances based on their weight values in metadata.
// Supported weight values in metadata["weight"]: 100, 80, 50, 20, 10, default 1.
// Eg:
//
// 加权轮询负载均衡选择器包，提供基于权重的轮询选择算法。
// 选择器根据实例 metadata 中的权重值进行选择。
// 支持的 metadata["weight"] 权重值：100, 80, 50, 20, 10，默认为 1。
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
)

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
//
// WRRSelector 使用加权轮询算法实现 discoverycontract.Selector 接口。
type WRRSelector struct {
	mu            sync.Mutex                     // mu protects weight state.
	                                             //
	                                              // mu 保护权重状态。
	currentWeight map[string]float64             // currentWeight tracks dynamic weights per instance.
	                                             //
	                                              // currentWeight 跟踪每个实例的动态权重。
	lastInstances []transportcontract.ServiceInstance // lastInstances caches last seen instances.
	                                             //
	                                              // lastInstances 缓存上次看到的实例。
}

// NewWRRSelector creates a new WRR selector instance.
//
// NewWRRSelector 创建新的加权轮询选择器实例。
func NewWRRSelector() *WRRSelector {
	return &WRRSelector{
		currentWeight: make(map[string]float64),
	}
}

// Select picks an instance using weighted round-robin algorithm.
// Core logic: Calculate weights, add to current, pick max, subtract total.
//
// Select 使用加权轮询算法选择实例。
// 核心逻辑：计算权重、累加到当前权重、选择最大值、减去总权重。
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