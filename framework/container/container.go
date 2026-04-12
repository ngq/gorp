package container

import (
	"fmt"
	"sync"

	"github.com/ngq/gorp/framework/contract"
)

// binding 描述了“一个 key 对应如何被实例化”。
//
// 中文说明：
// - factory 定义了实例的创建逻辑；
// - singleton 决定该实例是否只创建一次；
// - once/inst/err 是 singleton 模式下的运行时缓存。
//
// 这里有一个非常关键的语义：
// - 对 singleton 来说，factory 只会被 `sync.Once` 执行一次；
// - 即使第一次执行返回 error，该 error 也会被缓存下来；
// - 后续再次 `Make(key)` 不会重试，而是直接返回第一次的结果。
//
// 这意味着：
// - singleton factory 一旦失败，通常代表初始化阶段存在真实问题；
// - 如果希望“失败后允许重试”，就不应该使用当前这种 singleton 绑定方式。
type binding struct {
	// factory 是创建服务实例的工厂函数。
	factory contract.Factory
	// singleton 表示该服务是否按单例模式缓存。
	singleton bool

	// once 保证 singleton factory 最多执行一次。
	once sync.Once
	// inst 保存 singleton 模式下首次创建成功的实例。
	inst any
	// err 保存 singleton 模式下首次创建的错误（如果有）。
	err error
}

// providerState 保存一个 provider 在容器中的生命周期状态。
//
// 中文说明：
// - loaded 表示 `Register` 是否已经执行过；
// - booted 表示 `Boot` 是否已经执行过。
//
// 注意：
//   - 当前实现里，一旦某阶段被标记为已执行，就不会再重试；
//   - 即使 `Register/Boot` 返回 error，状态位也已经写入。
//   - 因此 provider 的失败通常应当被视为“需要修配置/修代码”的启动错误，
//     而不是运行时自动恢复的场景。
type providerState struct {
	// p 是实际注册进来的 provider 对象。
	p contract.ServiceProvider
	// loaded 表示 Register 阶段是否已执行。
	loaded bool
	// booted 表示 Boot 阶段是否已执行。
	booted bool
}

// Container 是默认的依赖注入容器实现。
//
// 中文说明：
// - 它负责维护所有服务绑定（bindings）、provider 注册状态，以及 deferred provider 的触发映射；
// - 整个框架运行时的服务解析，最终都会落到这里；
// - 当前实现是“字符串 key -> factory/instance”的设计，而不是按类型自动注入。
//
// 内部三张核心表：
// - bindings：已绑定的服务 key -> binding
// - providersByName：provider name -> providerState
// - deferredByKey：某个 key 首次被解析时，应触发哪个 deferred provider
//
// 并发说明：
// - 容器自身使用 RWMutex 保护内部映射；
// - singleton 实例化则由 binding 内部的 sync.Once 保证只执行一次。
type Container struct {
	// mu 保护容器内部的映射结构。
	mu sync.RWMutex

	// bindings 保存已经注册完成的服务绑定关系。
	bindings map[string]*binding

	// providersByName 保存所有 provider 的生命周期状态。
	providersByName map[string]*providerState
	// deferredByKey 记录：某个 key 第一次被 Make 时，需要触发哪个 deferred provider 加载。
	deferredByKey map[string]string // key -> providerName
}

// New 创建一个默认容器实例。
//
// 中文说明：
// - 创建时会初始化三张核心映射表；
// - 同时会把“容器自己”绑定到 `contract.ContainerKey`；
// - 这样一来，后续任何 factory/provider 在执行过程中，都能再把 container 解析出来继续做依赖装配。
func New() *Container {
	c := &Container{
		bindings:        map[string]*binding{},
		providersByName: map[string]*providerState{},
		deferredByKey:   map[string]string{},
	}

	// 把容器自己绑定进去，便于后续 factory / provider 嵌套解析依赖。
	c.Bind(contract.ContainerKey, func(contract.Container) (any, error) {
		return c, nil
	}, true)
	return c
}

// Bind 把一个 key 绑定到 factory。
//
// 中文说明：
// - 这个方法只负责登记绑定关系，不会立刻执行 factory；
// - factory 会在第一次 `Make(key)` 时才真正运行；
// - 如果 singleton=true，则实例会被缓存；否则每次 Make 都重新构造。
func (c *Container) Bind(key string, factory contract.Factory, singleton bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.bindings[key] = &binding{factory: factory, singleton: singleton}
}

// IsBind 判断某个 key 当前是否可被容器解析。
//
// 中文说明：
// - 它不仅检查已经显式 `Bind` 的服务；
// - 还会检查 deferred provider 是否声明“未来可提供”这个 key；
// - 因此 IsBind=true 并不总是意味着绑定已经完成，也可能只是“首次 Make 时会触发加载”。
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

// RegisterProvider 把一个 provider 注册进容器。
//
// 中文说明：
// - provider name 必须唯一；
// - 非延迟 provider：注册时立刻执行 `Register -> Boot`；
// - 延迟 provider：先只登记到 `providersByName/deferredByKey`，等第一次 `Make()` 命中相关 key 时再真正加载。
//
// 对 deferred provider 来说：
// - `Provides()` 决定它能响应哪些 key；
// - 当前实现采用“先注册者优先”的策略；
// - 如果两个 deferred provider 都声明了相同 key，后注册者不会覆盖前者，避免出现“后来的 provider 抢走解析路径”这种难以理解的行为。
func (c *Container) RegisterProvider(p contract.ServiceProvider) error {
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
			// later provider wins is usually surprising; keep first.
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

// Make 解析指定 key 对应的服务实例。
//
// 中文说明：
// - 解析前会先检查该 key 是否属于某个 deferred provider；
// - 如果是，则先触发该 provider 的 `Register + Boot`；
// - 之后再查 bindings，最后按 singleton / transient 的方式返回实例。
//
// singleton 语义：
// - 第一次调用时执行 factory；
// - 后续直接返回缓存的 inst / err；
// - 不会自动重试失败的初始化。
func (c *Container) Make(key string) (any, error) {
	// 如果该 key 属于 deferred provider，则在真正解析前先把 provider 加载进来。
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

// MustMake 解析指定 key；如果失败则直接 panic。
//
// 中文说明：
// - 适用于“缺少该依赖就说明程序装配有误”的内部场景；
// - 不适合直接用于面向用户输入的外层逻辑。
func (c *Container) MustMake(key string) any {
	v, err := c.Make(key)
	if err != nil {
		panic(err)
	}
	return v
}

// ensureProviderForKey 确保某个 key 对应的 deferred provider 已被加载。
//
// 中文说明：
// - 如果该 key 不属于任何 deferred provider，直接返回 nil；
// - 如果属于某个 deferred provider，则按顺序执行 `loadProvider -> bootProvider`；
// - 这也是 deferred provider 真正“懒加载”的触发点。
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

// loadProvider 执行 provider 的 Register 阶段。
//
// 中文说明：
// - 这里只负责“把服务绑定进容器”，不负责 Boot；
// - 同一个 provider 只会 Register 一次；
// - 当前实现会先把 `loaded=true` 写入状态，再执行 `Register`。
//
// 易踩坑点：
// - 如果 `Register` 返回 error，当前 provider 仍然会保持 `loaded=true`；
// - 后续再次触发不会自动重试；
// - 因此 Register 失败应被视为需要修复的启动错误，而不是临时可恢复错误。
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

// bootProvider 执行 provider 的 Boot 阶段。
//
// 中文说明：
// - Boot 总是在 Register 之后执行；
// - 同一个 provider 只会 Boot 一次；
// - 与 loadProvider 一样，当前实现会先写 `booted=true`，再执行 `Boot`。
//
// 易踩坑点：
// - 如果 Boot 返回 error，后续也不会自动重试；
// - 因此 Boot 中的失败同样应按”初始化失败”来处理。
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

// RegisterProviders 批量注册多个服务提供者。
//
// 中文说明：
// - 按传入顺序依次注册，遇到错误立即停止；
// - 错误信息会包含失败的 provider 名称，便于定位；
// - 这是简化 bootstrap 代码的推荐方式。
//
// 使用示例：
//
//	if err := c.RegisterProviders(
//	    app.NewProvider(),
//	    config.NewProvider(),
//	    log.NewProvider(),
//	    ginprovider.NewProvider(),
//	); err != nil {
//	    return err
//	}
func (c *Container) RegisterProviders(providers ...contract.ServiceProvider) error {
	for _, p := range providers {
		if err := c.RegisterProvider(p); err != nil {
			return fmt.Errorf("register provider %s: %w", p.Name(), err)
		}
	}
	return nil
}
