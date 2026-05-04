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

type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string  { return "selector.random" }
func (p *Provider) IsDefer() bool { return true }
func (p *Provider) Provides() []string {
	return []string{discoverycontract.SelectorKey, discoverycontract.SelectorBuilderKey}
}

func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(discoverycontract.SelectorBuilderKey, func(c runtimecontract.Container) (any, error) {
		return &randomBuilder{}, nil
	}, true)

	c.Bind(discoverycontract.SelectorKey, func(c runtimecontract.Container) (any, error) {
		return (&randomBuilder{}).Build(), nil
	}, true)

	return nil
}

func (p *Provider) Boot(c runtimecontract.Container) error {
	return nil
}

type randomBuilder struct{}

func (b *randomBuilder) Build() discoverycontract.Selector {
	return NewRandomSelector()
}

type RandomSelector struct {
	r  *rand.Rand
	mu sync.Mutex
}

func NewRandomSelector() *RandomSelector {
	return &RandomSelector{
		r: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

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

func noopDone(ctx context.Context, info discoverycontract.DoneInfo) {}
