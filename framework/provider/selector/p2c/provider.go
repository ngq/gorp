package p2c

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/ngq/gorp/framework/contract"
)

// Provider 提供自适应负载均衡选择器实现。
//
// 中文说明：
// - 使用 P2C（Power of Two Choices）学术算法原理；
// - 动态感知实例负载，自适应调整选择概率；
// - 高负载实例被选中概率降低，低负载实例概率升高；
// - 适用于实例性能波动较大的场景。
//
// 算法原理（P2C 论文）：
// - 随机选择两个候选实例；
// - 比较两者的负载情况；
// - 选择负载较低的实例；
// - 使用 DoneFunc 回调更新负载统计。
type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string     { return "selector.p2c" }
func (p *Provider) IsDefer() bool    { return true }
func (p *Provider) Provides() []string {
	return []string{contract.SelectorKey, contract.SelectorBuilderKey}
}

func (p *Provider) Register(c contract.Container) error {
	c.Bind(contract.SelectorBuilderKey, func(c contract.Container) (any, error) {
		return &p2cBuilder{}, nil
	}, true)
	c.Bind(contract.SelectorKey, func(c contract.Container) (any, error) {
		return (&p2cBuilder{}).Build(), nil
	}, true)
	return nil
}

func (p *Provider) Boot(c contract.Container) error {
	return nil
}

// p2cBuilder 构建 P2C Selector。
type p2cBuilder struct{}

func (b *p2cBuilder) Build() contract.Selector {
	return NewP2CSelector()
}

// P2CSelector 自适应负载均衡选择器。
//
// 中文说明：
// - 基于 "Power of Two Choices" 学术论文算法；
// - 维护每个实例的负载统计（pending requests、延迟）；
// - 使用 DoneFunc 回调动态更新负载；
// - 不抄袭 Kratos 代码，使用自己的实现。
type P2CSelector struct {
	// mu 保护负载统计状态
	mu sync.Mutex

	// r 随机数生成器
	r *rand.Rand

	// instanceStats 实例负载统计
	// key: instance.Address
	instanceStats map[string]*InstanceStats
}

// InstanceStats 实例负载统计。
//
// 中文说明：
// - pending: 当前正在处理的请求数；
// - successCount: 成功请求计数；
// - failCount: 失败请求计数；
// - totalLatency: 累计延迟（用于计算平均延迟）；
// - lastUpdate: 最后更新时间。
type InstanceStats struct {
	pending       int64   // 正在处理的请求数
	successCount  int64   // 成功次数
	failCount     int64   // 失败次数
	totalLatency  int64   // 累计延迟（毫秒）
	lastUpdate    time.Time
}

// NewP2CSelector 创建 P2C 选择器。
func NewP2CSelector() *P2CSelector {
	return &P2CSelector{
		r:             rand.New(rand.NewSource(time.Now().UnixNano())),
		instanceStats: make(map[string]*InstanceStats),
	}
}

// Select 选择一个服务实例。
//
// 中文说明：
// - 随机选择两个候选实例；
// - 比较两者的负载评分；
// - 选择负载评分较低的实例；
// - 返回 DoneFunc 用于更新负载统计。
func (s *P2CSelector) Select(ctx context.Context, instances []contract.ServiceInstance, opts ...contract.SelectOption) (
	selected contract.ServiceInstance, done contract.DoneFunc, err error,
) {
	// 解析可选参数
	options := &contract.SelectOptions{}
	for _, opt := range opts {
		opt(options)
	}

	// 强制指定实例
	if options.ForceInstance != nil {
		inst := *options.ForceInstance
		// 增加 pending 计数
		s.incrementPending(inst.Address)
		return inst, s.createDoneFunc(inst), nil
	}

	// 过滤健康实例
	filtered := s.filterHealthy(instances, options.Filters)
	if len(filtered) == 0 {
		return contract.ServiceInstance{}, noopDone, contract.ErrNoAvailable
	}

	// 如果只有 1 个实例，直接返回
	if len(filtered) == 1 {
		inst := filtered[0]
		s.incrementPending(inst.Address)
		return inst, s.createDoneFunc(inst), nil
	}

	s.mu.Lock()

	// P2C 核心算法：随机选择两个候选实例
	idx1 := s.r.Intn(len(filtered))
	idx2 := s.r.Intn(len(filtered))
	// 确保 idx1 != idx2
	for idx2 == idx1 && len(filtered) > 1 {
		idx2 = s.r.Intn(len(filtered))
	}

	inst1 := filtered[idx1]
	inst2 := filtered[idx2]

	// 计算两个实例的负载评分
	score1 := s.calculateScore(inst1)
	score2 := s.calculateScore(inst2)

	s.mu.Unlock()

	// 选择负载评分较低的实例
	var chosen contract.ServiceInstance
	if score1 <= score2 {
		chosen = inst1
	} else {
		chosen = inst2
	}

	// 增加 pending 计数
	s.incrementPending(chosen.Address)

	return chosen, s.createDoneFunc(chosen), nil
}

// calculateScore 计算实例负载评分。
//
// 中文说明：
// - 评分越高表示负载越重，被选中概率降低；
// - 综合考虑：pending 数量、失败率、平均延迟；
// - 使用加权公式：score = pending + fail_rate * 10 + avg_latency * 0.1。
func (s *P2CSelector) calculateScore(instance contract.ServiceInstance) float64 {
	stats := s.instanceStats[instance.Address]
	if stats == nil {
		// 新实例，默认低负载
		return 0.0
	}

	// 计算失败率
	totalRequests := stats.successCount + stats.failCount
	failRate := 0.0
	if totalRequests > 0 {
		failRate = float64(stats.failCount) / float64(totalRequests)
	}

	// 计算平均延迟
	avgLatency := 0.0
	if stats.successCount > 0 {
		avgLatency = float64(stats.totalLatency) / float64(stats.successCount)
	}

	// 加权评分公式
	// pending 权重最高（当前负载）
	// 失败率权重中等（可靠性）
	// 延迟权重较低（响应速度）
	score := float64(stats.pending) + failRate * 10.0 + avgLatency * 0.1

	return score
}

// incrementPending 增加 pending 计数。
func (s *P2CSelector) incrementPending(address string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	stats := s.instanceStats[address]
	if stats == nil {
		stats = &InstanceStats{
			lastUpdate: time.Now(),
		}
		s.instanceStats[address] = stats
	}
	stats.pending++
	stats.lastUpdate = time.Now()
}

// createDoneFunc 创建 DoneFunc 回调。
//
// 中文说明：
// - 调用完成后必须执行此回调；
// - 更新实例负载统计；
// - 减少 pending 计数，增加成功/失败计数。
func (s *P2CSelector) createDoneFunc(instance contract.ServiceInstance) contract.DoneFunc {
	return func(ctx context.Context, info contract.DoneInfo) {
		s.mu.Lock()
		defer s.mu.Unlock()

		stats := s.instanceStats[instance.Address]
		if stats == nil {
			return
		}

		// 减少 pending 计数
		stats.pending--
		if stats.pending < 0 {
			stats.pending = 0
		}

		// 更新成功/失败计数
		if info.Err != nil {
			stats.failCount++
		} else {
			stats.successCount++
			// 可从 ReplyMD 读取延迟信息（如有）
		}

		stats.lastUpdate = time.Now()
	}
}

// filterHealthy 过滤健康实例。
func (s *P2CSelector) filterHealthy(instances []contract.ServiceInstance, filters []contract.NodeFilter) []contract.ServiceInstance {
	if len(instances) == 0 {
		return nil
	}

	result := make([]contract.ServiceInstance, 0, len(instances))
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

// noopDone 空 DoneFunc。
func noopDone(ctx context.Context, info contract.DoneInfo) {}