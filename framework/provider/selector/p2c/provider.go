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

type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string  { return "selector.p2c" }
func (p *Provider) IsDefer() bool { return true }
func (p *Provider) Provides() []string {
	return []string{discoverycontract.SelectorKey, discoverycontract.SelectorBuilderKey}
}

func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(discoverycontract.SelectorBuilderKey, func(c runtimecontract.Container) (any, error) {
		return &p2cBuilder{}, nil
	}, true)
	c.Bind(discoverycontract.SelectorKey, func(c runtimecontract.Container) (any, error) {
		return (&p2cBuilder{}).Build(), nil
	}, true)
	return nil
}

func (p *Provider) Boot(c runtimecontract.Container) error {
	return nil
}

type p2cBuilder struct{}

func (b *p2cBuilder) Build() discoverycontract.Selector {
	return NewP2CSelector()
}

type P2CSelector struct {
	mu            sync.Mutex
	r             *rand.Rand
	instanceStats map[string]*InstanceStats
}

type InstanceStats struct {
	pending      int64
	successCount int64
	failCount    int64
	totalLatency int64
	lastUpdate   time.Time
}

func NewP2CSelector() *P2CSelector {
	return &P2CSelector{
		r:             rand.New(rand.NewSource(time.Now().UnixNano())),
		instanceStats: make(map[string]*InstanceStats),
	}
}

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

	avgLatency := 0.0
	if stats.successCount > 0 {
		avgLatency = float64(stats.totalLatency) / float64(stats.successCount)
	}

	return float64(stats.pending) + failRate*10 + avgLatency*0.1
}

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

		stats.lastUpdate = time.Now()
	}
}

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

func noopDone(ctx context.Context, info discoverycontract.DoneInfo) {}
