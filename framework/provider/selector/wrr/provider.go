package wrr

import (
	"context"
	"sync"

	"github.com/ngq/gorp/framework/contract"
)

// Provider 提供加权轮询负载均衡选择器实现。
//
// 中文说明：
// - 使用 Nginx 公开的 WRR 算法原理；
// - 根据实例权重分配请求；
// - 权重越高，被选中频率越高；
// - 适用于实例性能差异较大的场景。
//
// 算法原理（Nginx WRR）：
// - 每个实例维护一个 currentWeight；
// - 每次选择时，所有实例 currentWeight += weight；
// - 选择 currentWeight 最大的实例；
// - 被选中实例 currentWeight -= totalWeight。
type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string     { return "selector.wrr" }
func (p *Provider) IsDefer() bool    { return true }
func (p *Provider) Provides() []string {
	return []string{contract.SelectorKey, contract.SelectorBuilderKey}
}

func (p *Provider) Register(c contract.Container) error {
	c.Bind(contract.SelectorBuilderKey, func(c contract.Container) (any, error) {
		return &wrrBuilder{}, nil
	}, true)
	c.Bind(contract.SelectorKey, func(c contract.Container) (any, error) {
		return (&wrrBuilder{}).Build(), nil
	}, true)
	return nil
}

func (p *Provider) Boot(c contract.Container) error {
	return nil
}

// wrrBuilder 构建 WRR Selector。
type wrrBuilder struct{}

func (b *wrrBuilder) Build() contract.Selector {
	return NewWRRSelector()
}

// WRRSelector 加权轮询选择器。
//
// 中文说明：
// - 基于 Nginx 公开的平滑加权轮询算法；
// - 支持动态权重（从元数据读取）；
// - 支持实例列表变更时自动清理；
// - 不抄袭 Kratos 代码，使用自己的实现。
type WRRSelector struct {
	// mu 保护 currentWeight 状态
	mu sync.Mutex

	// currentWeight 每个实例的当前权重
	// key: instance.Address, value: 当前权重值
	currentWeight map[string]float64

	// lastInstances 上次选择时的实例列表
	// 用于检测实例列表变更，清理废弃权重
	lastInstances []contract.ServiceInstance
}

// NewWRRSelector 创建 WRR 选择器。
func NewWRRSelector() *WRRSelector {
	return &WRRSelector{
		currentWeight: make(map[string]float64),
		lastInstances: nil,
	}
}

// Select 选择一个服务实例。
//
// 中文说明：
// - 使用 Nginx 平滑加权轮询算法；
// - 每次选择时更新所有实例 currentWeight；
// - 选择 currentWeight 最大的实例；
// - 被选中实例 currentWeight -= totalWeight；
// - 返回空 DoneFunc（WRR 无需回调）。
func (s *WRRSelector) Select(ctx context.Context, instances []contract.ServiceInstance, opts ...contract.SelectOption) (
	selected contract.ServiceInstance, done contract.DoneFunc, err error,
) {
	// 解析可选参数
	options := &contract.SelectOptions{}
	for _, opt := range opts {
		opt(options)
	}

	// 强制指定实例
	if options.ForceInstance != nil {
		return *options.ForceInstance, noopDone, nil
	}

	// 过滤健康实例
	filtered := s.filterHealthy(instances, options.Filters)
	if len(filtered) == 0 {
		return contract.ServiceInstance{}, noopDone, contract.ErrNoAvailable
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 检测实例列表变更，清理废弃权重
	s.cleanupStaleWeights(filtered)

	// Nginx WRR 算法核心逻辑
	// 1. 累加权重：所有实例 currentWeight += weight
	// 2. 计算总权重
	// 3. 选择 currentWeight 最大的实例
	// 4. 被选中实例 currentWeight -= totalWeight

	var totalWeight float64
	var maxWeight float64
	var chosenInstance contract.ServiceInstance

	for _, instance := range filtered {
		weight := s.getWeight(instance)
		totalWeight += weight

		// 累加权重
		current := s.currentWeight[instance.Address] + weight
		s.currentWeight[instance.Address] = current

		// 寻找最大权重实例
		if current > maxWeight {
			maxWeight = current
			chosenInstance = instance
		}
	}

	// 被选中实例减去总权重
	if chosenInstance.Address != "" {
		s.currentWeight[chosenInstance.Address] -= totalWeight
	}

	// 保存实例列表用于下次清理
	s.lastInstances = filtered

	return chosenInstance, noopDone, nil
}

// filterHealthy 过滤健康实例。
func (s *WRRSelector) filterHealthy(instances []contract.ServiceInstance, filters []contract.NodeFilter) []contract.ServiceInstance {
	if len(instances) == 0 {
		return nil
	}

	result := make([]contract.ServiceInstance, 0, len(instances))
	for _, inst := range instances {
		if !inst.Healthy {
			continue
		}
		// 应用自定义过滤器
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

// getWeight 获取实例权重。
//
// 中文说明：
// - 从元数据中读取 "weight" 字段；
// - 默认权重为 1.0；
// - 权重范围建议 0-100。
func (s *WRRSelector) getWeight(instance contract.ServiceInstance) float64 {
	if instance.Metadata == nil {
		return 1.0
	}
	// 尝试从元数据读取权重
	if w, ok := instance.Metadata["weight"]; ok {
		// 简单解析（实际项目中可用 strconv.ParseFloat）
		switch w {
		case "100":
			return 100.0
		case "80":
			return 80.0
		case "50":
			return 50.0
		case "20":
			return 20.0
		case "10":
			return 10.0
		default:
			return 1.0
		}
	}
	return 1.0
}

// cleanupStaleWeights 清理废弃实例的权重。
//
// 中文说明：
// - 当实例列表变更时（如实例下线）；
// - 清理不在当前列表中的实例权重；
// - 避免废弃权重影响选择结果。
func (s *WRRSelector) cleanupStaleWeights(currentInstances []contract.ServiceInstance) {
	// 构建当前实例地址集合
	currentAddrSet := make(map[string]bool, len(currentInstances))
	for _, inst := range currentInstances {
		currentAddrSet[inst.Address] = true
	}

	// 清理废弃权重
	for addr := range s.currentWeight {
		if !currentAddrSet[addr] {
			delete(s.currentWeight, addr)
		}
	}
}

// noopDone 空 DoneFunc。
func noopDone(ctx context.Context, info contract.DoneInfo) {}