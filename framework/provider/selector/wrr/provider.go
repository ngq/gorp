package wrr

import (
	"context"
	"sync"

	discoverycontract "github.com/ngq/gorp/framework/contract/discovery"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string  { return "selector.wrr" }
func (p *Provider) IsDefer() bool { return true }
func (p *Provider) Provides() []string {
	return []string{discoverycontract.SelectorKey, discoverycontract.SelectorBuilderKey}
}

func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(discoverycontract.SelectorBuilderKey, func(c runtimecontract.Container) (any, error) {
		return &wrrBuilder{}, nil
	}, true)
	c.Bind(discoverycontract.SelectorKey, func(c runtimecontract.Container) (any, error) {
		return (&wrrBuilder{}).Build(), nil
	}, true)
	return nil
}

func (p *Provider) Boot(c runtimecontract.Container) error {
	return nil
}

type wrrBuilder struct{}

func (b *wrrBuilder) Build() discoverycontract.Selector {
	return NewWRRSelector()
}

type WRRSelector struct {
	mu            sync.Mutex
	currentWeight map[string]float64
	lastInstances []transportcontract.ServiceInstance
}

func NewWRRSelector() *WRRSelector {
	return &WRRSelector{
		currentWeight: make(map[string]float64),
	}
}

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

func noopDone(ctx context.Context, info discoverycontract.DoneInfo) {}
