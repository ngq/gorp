// Application scenarios:
// - Provide the framework's runtime dependency injection container implementation.
// - Manage direct bindings, deferred providers, singleton lifecycle, and provider boot order.
// - Serve as the central assembly point for providers used by bootstrap and application startup.
//
// 适用场景：
// - 提供框架运行时依赖注入容器实现。
// - 管理直接绑定、延迟 provider、单例生命周期和 provider 启动顺序。
// - 作为 bootstrap 与 application 启动阶段装配 provider 的核心承载点。
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

// Container is the default runtime container implementation used by the framework.
//
// Container 是框架使用的默认运行时容器实现。
type Container struct {
	mu              sync.RWMutex
	bindings        map[string]*binding
	providersByName map[string]*providerState
	deferredByKey   map[string]string
}

// New creates a new runtime container and self-binds the container contract.
//
// New 创建一个新的运行时容器，并将容器契约自身注册进去。
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

// Bind registers a factory under the given key.
//
// Bind 将指定 factory 注册到目标 key 之下。
func (c *Container) Bind(key string, factory runtimecontract.Factory, singleton bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.bindings[key] = &binding{factory: factory, singleton: singleton}
}

// IsBind reports whether the key is directly bound or promised by a deferred provider.
//
// IsBind 返回目标 key 是否已直接绑定，或是否由延迟 provider 承诺提供。
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

// RegisterProvider registers a provider into the container and loads it immediately when not deferred.
//
// RegisterProvider 将 provider 注册进容器；若不是延迟 provider，则立刻装载。
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
		// Deferred providers advertise their keys first and only load when one of those keys is requested.
		// 延迟 provider 先声明自己能提供的 key，等这些 key 真被请求时再装载。
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

// Make resolves a service by key.
//
// Make 按 key 解析服务实例。
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
		// Singleton factories are guarded by sync.Once so concurrent callers share one initialization path.
		// 单例 factory 通过 sync.Once 保护，保证并发调用共享同一条初始化路径。
		b.once.Do(func() {
			b.inst, b.err = b.factory(c)
		})
		return b.inst, b.err
	}
	return b.factory(c)
}

// MustMake resolves a service by key and panics on failure.
//
// MustMake 按 key 解析服务，失败时直接 panic。
func (c *Container) MustMake(key string) any {
	v, err := c.Make(key)
	if err != nil {
		panic(err)
	}
	return v
}

// RegisterProviders registers multiple providers in order.
//
// RegisterProviders 按顺序注册多个 provider。
func (c *Container) RegisterProviders(providers ...runtimecontract.ServiceProvider) error {
	for _, p := range providers {
		if err := c.RegisterProvider(p); err != nil {
			return fmt.Errorf("register provider %s: %w", p.Name(), err)
		}
	}
	return nil
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
	// Mark loaded before calling Register to make repeated load attempts idempotent.
	// 在调用 Register 之前先标记 loaded，保证重复装载请求具备幂等语义。
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
	// Boot is also guarded for idempotency because a deferred provider may be triggered multiple times.
	// Boot 同样需要幂等保护，因为延迟 provider 可能被多次触发。
	st.booted = true
	c.mu.Unlock()

	return st.p.Boot(c)
}
