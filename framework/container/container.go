// Package container provides runtime dependency injection container for gorp framework.
// Manages direct bindings, deferred providers, singleton lifecycle, and provider boot order.
// Serves as central assembly point for providers used by bootstrap and application startup.
//
// 容器包提供 gorp 框架的运行时依赖注入容器实现。
// 管理直接绑定、延迟 provider、单例生命周期和 provider 启动顺序。
// 作为 bootstrap 与 application 启动阶段装配 provider 的核心承载点。
package container

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

// singletonState tracks the lifecycle of a singleton binding.
//
// singletonState 跟踪单例绑定的生命周期状态。
type singletonState uint32

const (
	singletonUninit  singletonState = iota // not yet initialized
	singletonIniting                       // currently being initialized
	singletonInited                        // initialization complete (success or error)
)

type binding struct {
	factory   runtimecontract.Factory
	singleton bool

	// Singleton state machine (only used when singleton == true).
	// 单例状态机（仅当 singleton == true 时使用）。
	state atomic.Uint32 // stores singletonState
	mu    sync.Mutex    // protects state transitions and done channel creation
	done  chan struct{} // closed when initialization completes
	inst  any
	err   error
}

type providerState struct {
	p      runtimecontract.ServiceProvider
	loaded bool
	booted bool
}

type namedKey struct {
	name string
	key  string
}

type closerEntry struct {
	key    string
	closer io.Closer
}

// Container is the default runtime container implementation used by the framework.
//
// Container 是框架使用的默认运行时容器实现。
type Container struct {
	mu              sync.RWMutex
	bindings        map[string]*binding
	namedBindings   map[namedKey]*binding
	providersByName map[string]*providerState
	deferredByKey   map[string]string

	// Circular dependency detection: maps goroutine ID → resolution stack.
	// 循环依赖检测：goroutine ID → 解析栈。
	resolving sync.Map // uint64 → *[]string

	// Destroy lifecycle.
	// 销毁生命周期。
	closerMu  sync.Mutex
	closers   []closerEntry
	destroyed atomic.Bool
}

// New creates a new runtime container and self-binds the container contract.
//
// New 创建一个新的运行时容器，并将容器契约自身注册进去。
func New() *Container {
	c := &Container{
		bindings:        map[string]*binding{},
		namedBindings:   map[namedKey]*binding{},
		providersByName: map[string]*providerState{},
		deferredByKey:   map[string]string{},
	}
	c.Bind(runtimecontract.ContainerKey, func(runtimecontract.Container) (any, error) {
		return c, nil
	}, true)
	return c
}

// Bind registers a factory under the given key.
// If the key is already bound, a warning is logged and the previous binding is replaced.
//
// Bind 将指定 factory 注册到目标 key 之下。
// 如果 key 已绑定，打印警告日志并替换原有绑定。
func (c *Container) Bind(key string, factory runtimecontract.Factory, singleton bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, exists := c.bindings[key]; exists {
		slog.Warn("container: binding overridden", "key", key)
	}
	c.bindings[key] = &binding{factory: factory, singleton: singleton}
}

// NamedBind registers a named factory under the given key.
// Allows multiple implementations of the same key to coexist under different names.
// If the name+key combination is already bound, a warning is logged and the previous binding is replaced.
//
// NamedBind 将命名 factory 绑定到目标 key 之下。
// 允许同一 key 的多个实现以不同名称共存。
// 如果 name+key 组合已绑定，打印警告日志并替换原有绑定。
func (c *Container) NamedBind(name, key string, factory runtimecontract.Factory, singleton bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	nk := namedKey{name: name, key: key}
	if _, exists := c.namedBindings[nk]; exists {
		slog.Warn("container: named binding overridden", "name", name, "key", key)
	}
	c.namedBindings[nk] = &binding{factory: factory, singleton: singleton}
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

// IsBindNamed reports whether a named binding exists.
//
// IsBindNamed 返回指定命名绑定是否存在。
func (c *Container) IsBindNamed(name, key string) bool {
	c.mu.RLock()
	_, ok := c.namedBindings[namedKey{name: name, key: key}]
	c.mu.RUnlock()
	return ok
}

// RegisterProvider registers a provider into the container and loads it immediately when not deferred.
//
// RegisterProvider 将 provider 注册进容器；若不是延迟 provider，则立刻装载。
func (c *Container) RegisterProvider(p runtimecontract.ServiceProvider) error {
	name := p.Name()
	if name == "" {
		return errors.New("provider name is empty")
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
// Returns ErrContainerDestroyed if the container has been destroyed.
// Returns CircularDependencyError if a circular dependency is detected.
//
// Make 按 key 解析服务实例。
// 如果容器已销毁，返回 ErrContainerDestroyed。
// 如果检测到循环依赖，返回 CircularDependencyError。
func (c *Container) Make(key string) (any, error) {
	if c.destroyed.Load() {
		return nil, runtimecontract.ErrContainerDestroyed
	}

	if err := c.ensureProviderForKey(key); err != nil {
		return nil, err
	}

	c.mu.RLock()
	b, ok := c.bindings[key]
	c.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("service not bound: %s", key)
	}

	return c.resolveBinding(b, key)
}

// MakeNamed resolves a named service by name and key.
// Returns ErrContainerDestroyed if the container has been destroyed.
// Returns CircularDependencyError if a circular dependency is detected.
//
// MakeNamed 按名称和 key 解析命名服务实例。
// 如果容器已销毁，返回 ErrContainerDestroyed。
// 如果检测到循环依赖，返回 CircularDependencyError。
func (c *Container) MakeNamed(name, key string) (any, error) {
	if c.destroyed.Load() {
		return nil, runtimecontract.ErrContainerDestroyed
	}

	c.mu.RLock()
	b, ok := c.namedBindings[namedKey{name: name, key: key}]
	c.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("named service not bound: name=%s, key=%s", name, key)
	}

	return c.resolveBinding(b, fmt.Sprintf("%s/%s", name, key))
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

// MustMakeNamed resolves a named service by name and key and panics on failure.
//
// MustMakeNamed 按名称和 key 解析命名服务，失败时直接 panic。
func (c *Container) MustMakeNamed(name, key string) any {
	v, err := c.MakeNamed(name, key)
	if err != nil {
		panic(err)
	}
	return v
}

// RegisterCloser registers an io.Closer to be called during Destroy.
// Closers are called in reverse registration order.
//
// RegisterCloser 注册一个 io.Closer，在 Destroy 时调用。
// Closer 按注册逆序调用。
func (c *Container) RegisterCloser(key string, closer io.Closer) {
	c.closerMu.Lock()
	defer c.closerMu.Unlock()
	c.closers = append(c.closers, closerEntry{key: key, closer: closer})
}

// Destroy calls all registered closers in reverse order and marks the container as destroyed.
// After Destroy, Make/MakeNamed return ErrContainerDestroyed.
//
// Destroy 按注册逆序调用所有 Closer，并将容器标记为已销毁。
// 销毁后 Make/MakeNamed 返回 ErrContainerDestroyed。
func (c *Container) Destroy() error {
	if !c.destroyed.CompareAndSwap(false, true) {
		return runtimecontract.ErrContainerDestroyed
	}

	c.closerMu.Lock()
	closers := c.closers
	c.closers = nil
	c.closerMu.Unlock()

	var errs []error
	// Close in reverse registration order.
	// 按注册逆序关闭。
	for i := len(closers) - 1; i >= 0; i-- {
		if err := closers[i].closer.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close %s: %w", closers[i].key, err))
		}
	}
	return errors.Join(errs...)
}

// RegisteredProviders returns information about all registered providers.
//
// RegisteredProviders 返回所有已注册 provider 的信息。
func (c *Container) RegisteredProviders() []runtimecontract.ProviderInfo {
	c.mu.RLock()
	defer c.mu.RUnlock()
	infos := make([]runtimecontract.ProviderInfo, 0, len(c.providersByName))
	for _, st := range c.providersByName {
		infos = append(infos, runtimecontract.ProviderInfo{
			Name:    st.p.Name(),
			Loaded:  st.loaded,
			Booted:  st.booted,
			IsDefer: st.p.IsDefer(),
		})
	}
	sort.Slice(infos, func(i, j int) bool { return infos[i].Name < infos[j].Name })
	return infos
}

// DebugPrint returns a human-readable snapshot of the container state.
//
// DebugPrint 返回容器状态的人类可读快照。
func (c *Container) DebugPrint() string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "=== Container Debug ===\n")

	c.mu.RLock()
	fmt.Fprintf(&buf, "Bindings (%d):\n", len(c.bindings))
	keys := make([]string, 0, len(c.bindings))
	for k := range c.bindings {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		b := c.bindings[k]
		kind := "transient"
		if b.singleton {
			kind = "singleton"
		}
		fmt.Fprintf(&buf, "  %s [%s]\n", k, kind)
	}

	fmt.Fprintf(&buf, "Named Bindings (%d):\n", len(c.namedBindings))
	nks := make([]namedKey, 0, len(c.namedBindings))
	for nk := range c.namedBindings {
		nks = append(nks, nk)
	}
	sort.Slice(nks, func(i, j int) bool {
		if nks[i].name != nks[j].name {
			return nks[i].name < nks[j].name
		}
		return nks[i].key < nks[j].key
	})
	for _, nk := range nks {
		b := c.namedBindings[nk]
		kind := "transient"
		if b.singleton {
			kind = "singleton"
		}
		fmt.Fprintf(&buf, "  %s/%s [%s]\n", nk.name, nk.key, kind)
	}

	fmt.Fprintf(&buf, "Providers (%d):\n", len(c.providersByName))
	pNames := make([]string, 0, len(c.providersByName))
	for name := range c.providersByName {
		pNames = append(pNames, name)
	}
	sort.Strings(pNames)
	for _, name := range pNames {
		st := c.providersByName[name]
		fmt.Fprintf(&buf, "  %s (loaded=%v,booted=%v,defer=%v)\n", name, st.loaded, st.booted, st.p.IsDefer())
	}

	fmt.Fprintf(&buf, "Deferred Keys (%d):\n", len(c.deferredByKey))
	dKeys := make([]string, 0, len(c.deferredByKey))
	for k := range c.deferredByKey {
		dKeys = append(dKeys, k)
	}
	sort.Strings(dKeys)
	for _, k := range dKeys {
		fmt.Fprintf(&buf, "  %s → %s\n", k, c.deferredByKey[k])
	}

	fmt.Fprintf(&buf, "Closers (%d):\n", len(c.closers))
	for _, ce := range c.closers {
		fmt.Fprintf(&buf, "  %s\n", ce.key)
	}

	fmt.Fprintf(&buf, "Destroyed: %v\n", c.destroyed.Load())

	c.mu.RUnlock()
	return buf.String()
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

// resolveBinding resolves a binding with circular dependency detection for singletons.
//
// resolveBinding 解析绑定，对单例执行循环依赖检测。
func (c *Container) resolveBinding(b *binding, key string) (any, error) {
	if !b.singleton {
		return b.factory(c)
	}

	// Fast path: already initialized.
	// 快速路径：已初始化。
	if singletonState(b.state.Load()) == singletonInited {
		return b.inst, b.err
	}

	// Try to become the initializer.
	// 尝试成为初始化者。
	b.mu.Lock()
	switch singletonState(b.state.Load()) {
	case singletonInited:
		b.mu.Unlock()
		return b.inst, b.err

	case singletonIniting:
		// Another goroutine (or this one via circular dependency) is initializing.
		// 另一个 goroutine（或本 goroutine 通过循环依赖）正在初始化。
		done := b.done
		b.mu.Unlock()

		// Check for circular dependency within this goroutine.
		// 检查本 goroutine 的循环依赖。
		gid := goroutineID()
		if stackVal, ok := c.resolving.Load(gid); ok {
			stack := stackVal.(*[]string)
			for _, k := range *stack {
				if k == key {
					chain := append(append([]string(nil), *stack...), key)
					return nil, &runtimecontract.CircularDependencyError{Key: key, Chain: chain}
				}
			}
		}

		// Wait for initialization to complete.
		// 等待初始化完成。
		<-done
		return b.inst, b.err

	default:
		// singletonUninit → start initialization.
		// singletonUninit → 开始初始化。
		b.state.Store(uint32(singletonIniting))
		b.done = make(chan struct{})
		b.mu.Unlock()

		// Track this key in the per-goroutine resolution stack.
		// 在 per-goroutine 解析栈中跟踪此 key。
		gid := goroutineID()
		stack := c.getOrCreateStack(gid)
		*stack = append(*stack, key)

		inst, err := b.factory(c)

		// Pop the key from the resolution stack.
		// 从解析栈中弹出此 key。
		*stack = (*stack)[:len(*stack)-1]
		if len(*stack) == 0 {
			c.resolving.Delete(gid)
		}

		b.mu.Lock()
		b.inst = inst
		b.err = err
		b.state.Store(uint32(singletonInited))
		close(b.done)
		b.mu.Unlock()

		return inst, err
	}
}

// getOrCreateStack returns the per-goroutine resolution stack, creating it if needed.
//
// getOrCreateStack 返回 per-goroutine 解析栈，不存在则创建。
func (c *Container) getOrCreateStack(gid uint64) *[]string {
	if v, ok := c.resolving.Load(gid); ok {
		return v.(*[]string)
	}
	s := make([]string, 0, 8)
	actual, _ := c.resolving.LoadOrStore(gid, &s)
	return actual.(*[]string)
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

// goroutineID returns the current goroutine ID for circular dependency tracking.
// Uses runtime.Stack to extract the goroutine ID.
//
// goroutineID 返回当前 goroutine ID，用于循环依赖跟踪。
// 使用 runtime.Stack 提取 goroutine ID。
func goroutineID() uint64 {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	// buf[:n] starts with "goroutine XXX "
	s := string(buf[:n])
	const prefix = "goroutine "
	if !strings.HasPrefix(s, prefix) {
		return 0
	}
	s = s[len(prefix):]
	idx := strings.IndexByte(s, ' ')
	if idx < 0 {
		return 0
	}
	id, _ := strconv.ParseUint(s[:idx], 10, 64)
	return id
}
