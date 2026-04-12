package random

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/ngq/gorp/framework/contract"
)

// Provider 提供随机负载均衡选择器实现。
//
// 中文说明：
// - 随机选择一个服务实例；
// - 简单高效，无状态维护；
// - 适用于实例性能相近的场景；
// - 微服务项目可启用此算法。
type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string     { return "selector.random" }
func (p *Provider) IsDefer() bool    { return true }
func (p *Provider) Provides() []string {
	return []string{contract.SelectorKey, contract.SelectorBuilderKey}
}

func (p *Provider) Register(c contract.Container) error {
	// 注册 Selector Builder
	c.Bind(contract.SelectorBuilderKey, func(c contract.Container) (any, error) {
		return &randomBuilder{}, nil
	}, true)

	// 注册默认 Selector 实例
	c.Bind(contract.SelectorKey, func(c contract.Container) (any, error) {
		return (&randomBuilder{}).Build(), nil
	}, true)

	return nil
}

func (p *Provider) Boot(c contract.Container) error {
	return nil
}

// randomBuilder 构建随机 Selector。
type randomBuilder struct{}

func (b *randomBuilder) Build() contract.Selector {
	return NewRandomSelector()
}

// RandomSelector 随机负载均衡选择器。
//
// 中文说明：
// - 随机选择一个健康实例；
// - 支持实例过滤；
// - 支持强制指定实例；
// - 无状态维护，每次选择独立随机。
type RandomSelector struct {
	// r 随机数生成器
	r *rand.Rand

	// mu 保护随机数生成器
	mu sync.Mutex
}

// NewRandomSelector 创建随机选择器。
func NewRandomSelector() *RandomSelector {
	return &RandomSelector{
		r: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Select 随机选择一个服务实例。
//
// 中文说明：
// - 从 instances 中随机选择一个健康实例；
// - 如果 opts.Filters 指定了过滤器，先过滤实例列表；
// - 如果 opts.ForceInstance 指定了实例，直接返回该实例；
// - 如果无健康实例，返回 ErrNoAvailable。
func (s *RandomSelector) Select(ctx context.Context, instances []contract.ServiceInstance, opts ...contract.SelectOption) (
	selected contract.ServiceInstance, done contract.DoneFunc, err error,
) {
	// 解析可选参数
	options := &contract.SelectOptions{}
	for _, opt := range opts {
		opt(options)
	}

	// 如果强制指定实例，则返回该实例
	if options.ForceInstance != nil {
		return *options.ForceInstance, noopDone, nil
	}

	// 过滤实例
	filtered := s.filterInstances(instances, options.Filters)

	// 如果无可用实例，返回错误
	if len(filtered) == 0 {
		return contract.ServiceInstance{}, noopDone, contract.ErrNoAvailable
	}

	// 随机选择一个实例
	s.mu.Lock()
	idx := s.r.Intn(len(filtered))
	s.mu.Unlock()

	return filtered[idx], noopDone, nil
}

// filterInstances 过滤服务实例。
//
// 中文说明：
// - 应用所有过滤器；
// - 只保留健康实例；
// - 返回过滤后的实例列表。
func (s *RandomSelector) filterInstances(instances []contract.ServiceInstance, filters []contract.NodeFilter) []contract.ServiceInstance {
	if len(instances) == 0 {
		return nil
	}

	result := make([]contract.ServiceInstance, 0, len(instances))
	for _, instance := range instances {
		// 检查健康状态
		if !instance.Healthy {
			continue
		}

		// 应用过滤器
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

// noopDone 是空的 DoneFunc 实现。
//
// 中文说明：
// - 随机算法无需权重调整；
// - 无性能统计。
func noopDone(ctx context.Context, info contract.DoneInfo) {
	// 空操作
}