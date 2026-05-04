package container

import (
	"fmt"
	"sync"

	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

type binding struct {
	factory   runtimecontract.Factory
	singleton bool
	once      sync.Once
	inst      any
	err       error
}

type providerState struct {
	p      runtimecontract.ServiceProvider
	loaded bool
	booted bool
}

type Container struct {
	mu              sync.RWMutex
	bindings        map[string]*binding
	providersByName map[string]*providerState
	deferredByKey   map[string]string
}

func New() *Container {
	c := &Container{
		bindings:        map[string]*binding{},
		providersByName: map[string]*providerState{},
		deferredByKey:   map[string]string{},
	}
	c.Bind(runtimecontract.ContainerKey, func(runtimecontract.Container) (any, error) {
		return c, nil
	}, true)
	return c
}

func (c *Container) Bind(key string, factory runtimecontract.Factory, singleton bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.bindings[key] = &binding{factory: factory, singleton: singleton}
}

func (c *Container) IsBind(key string) bool {
	c.mu.RLock()
	_, ok := c.bindings[key]
	c.mu.RUnlock()
	if ok {
		return true
	}

	c.mu.RLock()
	_, ok = c.deferredByKey[key]
	c.mu.RUnlock()
	return ok
}

func (c *Container) RegisterProvider(p runtimecontract.ServiceProvider) error {
	name := p.Name()
	if name == "" {
		return fmt.Errorf("provider name is empty")
	}

	c.mu.Lock()
	if _, exists := c.providersByName[name]; exists {
		c.mu.Unlock()
		return fmt.Errorf("provider already registered: %s", name)
	}
	st := &providerState{p: p}
	c.providersByName[name] = st
	c.mu.Unlock()

	if p.IsDefer() {
		c.mu.Lock()
		for _, key := range p.Provides() {
			if _, exists := c.deferredByKey[key]; !exists {
				c.deferredByKey[key] = name
			}
		}
		c.mu.Unlock()
		return nil
	}

	if err := c.loadProvider(name); err != nil {
		return err
	}
	return c.bootProvider(name)
}

func (c *Container) Make(key string) (any, error) {
	if err := c.ensureProviderForKey(key); err != nil {
		return nil, err
	}

	c.mu.RLock()
	b, ok := c.bindings[key]
	c.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("service not bound: %s", key)
	}

	if b.singleton {
		b.once.Do(func() {
			b.inst, b.err = b.factory(c)
		})
		return b.inst, b.err
	}
	return b.factory(c)
}

func (c *Container) MustMake(key string) any {
	v, err := c.Make(key)
	if err != nil {
		panic(err)
	}
	return v
}

func (c *Container) ensureProviderForKey(key string) error {
	c.mu.RLock()
	providerName, ok := c.deferredByKey[key]
	c.mu.RUnlock()
	if !ok {
		return nil
	}
	if err := c.loadProvider(providerName); err != nil {
		return err
	}
	return c.bootProvider(providerName)
}

func (c *Container) loadProvider(name string) error {
	c.mu.RLock()
	st, ok := c.providersByName[name]
	c.mu.RUnlock()
	if !ok {
		return fmt.Errorf("provider not registered: %s", name)
	}

	c.mu.Lock()
	if st.loaded {
		c.mu.Unlock()
		return nil
	}
	st.loaded = true
	c.mu.Unlock()

	return st.p.Register(c)
}

func (c *Container) bootProvider(name string) error {
	c.mu.RLock()
	st, ok := c.providersByName[name]
	c.mu.RUnlock()
	if !ok {
		return fmt.Errorf("provider not registered: %s", name)
	}

	c.mu.Lock()
	if st.booted {
		c.mu.Unlock()
		return nil
	}
	st.booted = true
	c.mu.Unlock()

	return st.p.Boot(c)
}

func (c *Container) RegisterProviders(providers ...runtimecontract.ServiceProvider) error {
	for _, p := range providers {
		if err := c.RegisterProvider(p); err != nil {
			return fmt.Errorf("register provider %s: %w", p.Name(), err)
		}
	}
	return nil
}
