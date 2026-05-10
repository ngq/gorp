// Package random provides random load balancing selector for service discovery.
// The selector randomly picks an instance from available healthy instances.
// Eg:
//
// 随机负载均衡选择器包，提供服务发现的随机选择算法。
// 选择器从可用的健康实例中随机选择一个。
// Eg:
//
//	// 注册 Provider
//	app.Register(random.NewProvider())
//
//	// 使用选择器
//	selector := c.MustMake(discoverycontract.SelectorKey).(discoverycontract.Selector)
//	instance, doneFunc, _ := selector.Select(ctx, instances)
package random

import (
	"context"
	"math/rand"
	"sync"
	"time"

	discoverycontract "github.com/ngq/gorp/framework/contract/discovery"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

// Provider registers the random selector contract.
//
// Provider 注册随机选择器契约。
type Provider struct{}

// NewProvider creates a new random selector provider instance.
//
// NewProvider 创建新的随机选择器 Provider 实例。
func NewProvider() *Provider { return &Provider{} }

// Name returns the provider name "selector.random".
//
// Name 返回 Provider 名称 "selector.random"。
func (p *Provider) Name() string { return "selector.random" }

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

// Register binds the random selector factory to the container.
//
// Register 将随机选择器工厂绑定到容器。
func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(discoverycontract.SelectorBuilderKey, func(c runtimecontract.Container) (any, error) {
		return &randomBuilder{}, nil
	}, true)

	c.Bind(discoverycontract.SelectorKey, func(c runtimecontract.Container) (any, error) {
		return (&randomBuilder{}).Build(), nil
	}, true)

	return nil
}

// Boot is a no-op for random selector provider.
//
// Boot 随机选择器 Provider 无启动逻辑。
func (p *Provider) Boot(c runtimecontract.Container) error {
	return nil
}

// randomBuilder builds random selector instances.
//
// randomBuilder 构建随机选择器实例。
type randomBuilder struct{}

// Build creates a new random selector.
//
// Build 创建新的随机选择器。
func (b *randomBuilder) Build() discoverycontract.Selector {
	return NewRandomSelector()
}

// RandomSelector implements discoverycontract.Selector with random selection.
//
// RandomSelector 使用随机选择实现 discoverycontract.Selector 接口。
type RandomSelector struct {
	r  *rand.Rand // r is the random generator.
	                //
	                 // r 随机数生成器。
	mu sync.Mutex  // mu protects random generator access.
	                //
	                 // mu 保护随机数生成器访问。
}

// NewRandomSelector creates a new random selector instance.
//
// NewRandomSelector 创建新的随机选择器实例。
func NewRandomSelector() *RandomSelector {
	return &RandomSelector{
		r: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Select randomly picks an instance from available healthy instances.
// Core logic: Filter healthy instances, apply custom filters, then pick randomly.
//
// Select 从可用的健康实例中随机选择一个。
// 核心逻辑：过滤健康实例，应用自定义过滤器，然后随机选择。
func (s *RandomSelector) Select(ctx context.Context, instances []transportcontract.ServiceInstance, opts ...discoverycontract.SelectOption) (
	selected transportcontract.ServiceInstance, done discoverycontract.DoneFunc, err error,
) {
	options := &discoverycontract.SelectOptions{}
	for _, opt := range opts {
		opt(options)
	}

	if options.ForceInstance != nil {
		return *options.ForceInstance, noopDone, nil
	}

	filtered := s.filterInstances(instances, options.Filters)
	if len(filtered) == 0 {
		return transportcontract.ServiceInstance{}, noopDone, discoverycontract.ErrNoAvailable
	}

	s.mu.Lock()
	idx := s.r.Intn(len(filtered))
	s.mu.Unlock()

	return filtered[idx], noopDone, nil
}

// filterInstances filters instances by health status and custom filters.
// Core logic: Skip unhealthy instances, apply each filter function.
//
// filterInstances 根据健康状态和自定义过滤器过滤实例。
// 核心逻辑：跳过不健康实例，应用每个过滤器函数。
func (s *RandomSelector) filterInstances(instances []transportcontract.ServiceInstance, filters []discoverycontract.NodeFilter) []transportcontract.ServiceInstance {
	if len(instances) == 0 {
		return nil
	}

	result := make([]transportcontract.ServiceInstance, 0, len(instances))
	for _, instance := range instances {
		if !instance.Healthy {
			continue
		}

		accepted := true
		for _, filter := range filters {
			if !filter(instance) {
				accepted = false
				break
			}
		}

		if accepted {
			result = append(result, instance)
		}
	}

	return result
}

// noopDone is a no-op DoneFunc for simple selectors.
//
// noopDone 是简单选择器的空 DoneFunc。
func noopDone(ctx context.Context, info discoverycontract.DoneInfo) {}