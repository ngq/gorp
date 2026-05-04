package host

import (
	"context"
	"sync"

	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	"github.com/ngq/gorp/framework/lifecycle"
)

type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string { return "host" }

func (p *Provider) IsDefer() bool { return false }

func (p *Provider) Provides() []string { return []string{runtimecontract.HostKey} }

func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(runtimecontract.HostKey, func(c runtimecontract.Container) (any, error) {
		return NewDefaultHost(c), nil
	}, true)
	return nil
}

func (p *Provider) Boot(runtimecontract.Container) error { return nil }

type DefaultHost struct {
	container runtimecontract.Container
	manager   *lifecycle.Manager
	mu        sync.RWMutex
	running   bool
}

func NewDefaultHost(container runtimecontract.Container) *DefaultHost {
	return &DefaultHost{
		container: container,
		manager:   lifecycle.NewManager(),
	}
}

func (h *DefaultHost) RegisterService(name string, service runtimecontract.Hostable) error {
	return h.RegisterServiceWithPriority(name, service, nil, 100)
}

func (h *DefaultHost) RegisterServiceWithPriority(name string, service runtimecontract.Hostable, hooks runtimecontract.Lifecycle, priority int) error {
	h.manager.Register(name, service, hooks, priority)
	return nil
}

func (h *DefaultHost) Services() []string {
	return h.manager.Services()
}

func (h *DefaultHost) State() lifecycle.State {
	return h.manager.State()
}

func (h *DefaultHost) Start(ctx context.Context) error {
	h.mu.Lock()
	if h.running {
		h.mu.Unlock()
		return nil
	}
	h.running = true
	h.mu.Unlock()

	return h.manager.Start(ctx)
}

func (h *DefaultHost) Stop(ctx context.Context) error {
	return h.Shutdown(ctx)
}

func (h *DefaultHost) Shutdown(ctx context.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if !h.running {
		return nil
	}

	err := h.manager.Stop(ctx)
	h.running = false
	return err
}
