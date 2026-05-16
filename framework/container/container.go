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
	loadMu sync.Mutex
	loaded bool
	bootMu sync.Mutex
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

	if p.IsDefer() {
		// Deferred providers advertise their keys first and only load when one of those keys is requested.
		// Register deferred keys within the same lock that registered the provider to prevent TOCTOU race.
		// 延迟 provider 先声明自己能提供的 key，等这些 key 真被请求时再装载。
		// 在注册 provider 的同一把锁内注册延迟 key，防止 TOCTOU 竞态。
		for _, key := range p.Provides() {
			if _, exists := c.deferredByKey[key]; !exists {
				c.deferredByKey[key] = name
			}
		}
		c.mu.Unlock()
		return nil
	}
	c.mu.Unlock()

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
// The panic message includes the key name for easier debugging.
//
// MustMake 按 key 解析服务，失败时直接 panic。
// panic 信息包含 key 名称以便调试。
func (c *Container) MustMake(key string) any {
	v, err := c.Make(key)
	if err != nil {
		panic(fmt.Sprintf("container: MustMake(%q): %v", key, err))
	}
	return v
}

// MustMakeNamed resolves a named service by name and key and panics on failure.
// The panic message includes the name and key for easier debugging.
//
// MustMakeNamed 按名称和 key 解析命名服务，失败时直接 panic。
// panic 信息包含 name 和 key 以便调试。
func (c *Container) MustMakeNamed(name, key string) any {
	v, err := c.MakeNamed(name, key)
	if err != nil {
		panic(fmt.Sprintf("container: MustMakeNamed(name=%q, key=%q): %v", name, key, err))
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
		if err := c.checkCircularDependency(key); err != nil {
			return nil, err
		}
		gid := goroutineID()
		c.pushResolvingKey(gid, key)
		defer c.popResolvingKey(gid)
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
		if err := c.checkCircularDependency(key); err != nil {
			return nil, err
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
		c.pushResolvingKey(gid, key)

		// Use recover to protect against factory panics.
		// Without this, a panic would leave b.done unclosed forever,
		// causing all subsequent goroutines to block indefinitely on <-done.
		// 使用 recover 防止工厂 panic。
		// 否则 panic 会导致 b.done 永远不关闭，
		// 所有后续 goroutine 将在 <-done 处永久阻塞。
		var inst any
		var err error
		func() {
			defer func() {
				if r := recover(); r != nil {
					err = fmt.Errorf("container: singleton factory panic for key %q: %v", key, r)
				}
			}()
			inst, err = b.factory(c)
		}()
		c.popResolvingKey(gid)

		b.mu.Lock()
		if err != nil {
			// Do not cache error instances; reset state so the next call can retry.
			// The done channel is still closed so waiters are unblocked,
			// but they will see state=singletonUninit and re-attempt initialization.
			// 不缓存错误实例；重置状态使下次调用可重试。
			// done channel 仍然被关闭以解除等待者阻塞，
			// 但等待者会看到 state=singletonUninit 并重新尝试初始化。
			b.inst = nil
			b.err = err
			b.state.Store(uint32(singletonUninit))
			close(b.done)
			b.mu.Unlock()
			return nil, err
		}
		b.inst = inst
		b.err = nil
		b.state.Store(uint32(singletonInited))
		close(b.done)
		b.mu.Unlock()
		return inst, nil
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

// checkCircularDependency checks whether the target key already exists in the current goroutine resolution stack.
//
// checkCircularDependency 检查目标 key 是否已存在于当前 goroutine 的解析栈中。
func (c *Container) checkCircularDependency(key string) error {
	gid := goroutineID()
	if stackVal, ok := c.resolving.Load(gid); ok {
		stack := stackVal.(*[]string)
		for _, resolvingKey := range *stack {
			if resolvingKey == key {
				chain := append(append([]string(nil), *stack...), key)
				return &runtimecontract.CircularDependencyError{Key: key, Chain: chain}
			}
		}
	}
	return nil
}

// pushResolvingKey pushes one key into the current goroutine resolution stack.
//
// pushResolvingKey 将一个 key 压入当前 goroutine 的解析栈。
func (c *Container) pushResolvingKey(gid uint64, key string) {
	stack := c.getOrCreateStack(gid)
	*stack = append(*stack, key)
}

// popResolvingKey pops one key from the current goroutine resolution stack.
//
// popResolvingKey 从当前 goroutine 的解析栈弹出一个 key。
func (c *Container) popResolvingKey(gid uint64) {
	stackVal, ok := c.resolving.Load(gid)
	if !ok {
		return
	}
	stack := stackVal.(*[]string)
	*stack = (*stack)[:len(*stack)-1]
	if len(*stack) == 0 {
		c.resolving.Delete(gid)
	}
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

	st.loadMu.Lock()
	defer st.loadMu.Unlock()
	if st.loaded {
		return nil
	}

	if err := st.p.Register(c); err != nil {
		return err
	}

	st.loaded = true
	return nil
}

func (c *Container) bootProvider(name string) error {
	c.mu.RLock()
	st, ok := c.providersByName[name]
	c.mu.RUnlock()
	if !ok {
		return fmt.Errorf("provider not registered: %s", name)
	}

	st.bootMu.Lock()
	defer st.bootMu.Unlock()
	if st.booted {
		return nil
	}

	if err := st.p.Boot(c); err != nil {
		return err
	}

	st.booted = true
	return nil
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

// ProviderDAG builds and returns the provider dependency graph.
// It analyzes all registered providers, their provides/depends declarations,
// and computes the recommended load order with cycle detection.
//
// ProviderDAG 构建并返回 provider 依赖图。
// 分析所有已注册的 provider，其 provides/depends 声明，
// 计算推荐加载顺序并检测循环依赖。
func (c *Container) ProviderDAG() runtimecontract.ProviderDAG {
	c.mu.RLock()
	defer c.mu.RUnlock()

	dag := runtimecontract.ProviderDAG{
		Nodes:     make([]runtimecontract.ProviderDAGNode, 0, len(c.providersByName)),
		Edges:     make([]runtimecontract.DAGEdge, 0),
		Cycles:    make([][]string, 0),
		LoadOrder: make([]string, 0, len(c.providersByName)),
	}

	// Build node map: provider name -> node.
	// 构建节点映射：provider 名称 -> 节点。
	nodeMap := make(map[string]*runtimecontract.ProviderDAGNode)
	for name, st := range c.providersByName {
		node := &runtimecontract.ProviderDAGNode{
			Name:      name,
			Provides:  st.p.Provides(),
			DependsOn: st.p.DependsOn(),
			IsDefer:   st.p.IsDefer(),
			Loaded:    st.loaded,
			Booted:    st.booted,
		}
		nodeMap[name] = node
		dag.Nodes = append(dag.Nodes, *node)
	}

	// Build key -> provider name reverse map.
	// 构建 key -> provider 名称的反向映射。
	keyToProvider := make(map[string]string)
	for name, st := range c.providersByName {
		for _, key := range st.p.Provides() {
			keyToProvider[key] = name
		}
	}

	// Build edges: for each provider's DependsOn, find which provider provides that key.
	// 构建边：对每个 provider 的 DependsOn，找到提供该 key 的 provider。
	for name, node := range nodeMap {
		for _, depKey := range node.DependsOn {
			edge := runtimecontract.DAGEdge{
				From: name,
				Key:  depKey,
			}
			// Find which provider provides this key.
			// 查找哪个 provider 提供这个 key。
			if providerName, ok := keyToProvider[depKey]; ok {
				edge.To = providerName
			} else if _, bound := c.bindings[depKey]; bound {
				// Key is directly bound, no provider dependency.
				// key 已直接绑定，无 provider 依赖。
				edge.To = ""
			} else if _, deferred := c.deferredByKey[depKey]; deferred {
				// Key is promised by a deferred provider.
				// key 由延迟 provider 承诺提供。
				edge.To = c.deferredByKey[depKey]
			}
			if edge.To != "" || edge.To == "" {
				// Always add edge to show dependency, even if external.
				// 始终添加边以显示依赖关系，即使是外部依赖。
				dag.Edges = append(dag.Edges, edge)
			}
		}
	}

	// Detect cycles using DFS.
	// 使用 DFS 检测循环依赖。
	visited := make(map[string]bool)
	inStack := make(map[string]bool)
	var cyclePath []string

	var dfs func(name string) bool
	dfs = func(name string) bool {
		visited[name] = true
		inStack[name] = true
		cyclePath = append(cyclePath, name)

		// Find all dependencies of this provider.
		// 查找此 provider 的所有依赖。
		for _, edge := range dag.Edges {
			if edge.From != name || edge.To == "" {
				continue
			}
			if !visited[edge.To] {
				if dfs(edge.To) {
					return true
				}
			} else if inStack[edge.To] {
				// Found cycle.
				// 发现循环。
				cycleStart := -1
				for i, n := range cyclePath {
					if n == edge.To {
						cycleStart = i
						break
					}
				}
				if cycleStart >= 0 {
					cycle := make([]string, len(cyclePath)-cycleStart)
					copy(cycle, cyclePath[cycleStart:])
					dag.Cycles = append(dag.Cycles, cycle)
				}
				return true
			}
		}

		cyclePath = cyclePath[:len(cyclePath)-1]
		inStack[name] = false
		return false
	}

	for name := range nodeMap {
		if !visited[name] {
			dfs(name)
		}
	}

	// Compute load order using topological sort (Kahn's algorithm).
	// 使用拓扑排序（Kahn 算法）计算加载顺序。
	inDegree := make(map[string]int)
	for name := range nodeMap {
		inDegree[name] = 0
	}
	for _, edge := range dag.Edges {
		if edge.To != "" {
			inDegree[edge.To]++ // edge.To is depended by edge.From
		}
	}

	// Reverse: edge.From depends on edge.To, so edge.To should be loaded first.
	// Actually we need to reverse the direction for load order.
	// 实际上需要反转方向来计算加载顺序。
	adj := make(map[string][]string) // provider -> providers that depend on it
	for name := range nodeMap {
		adj[name] = nil
	}
	for _, edge := range dag.Edges {
		if edge.To != "" {
			adj[edge.To] = append(adj[edge.To], edge.From)
		}
	}

	// Recompute in-degree: count how many providers each provider depends on.
	// 重新计算入度：统计每个 provider 依赖多少个其他 provider。
	inDegree = make(map[string]int)
	for name := range nodeMap {
		inDegree[name] = 0
	}
	for _, edge := range dag.Edges {
		if edge.To != "" {
			inDegree[edge.From]++ // edge.From depends on edge.To
		}
	}

	// Queue of providers with no dependencies.
	// 无依赖的 provider 队列。
	queue := make([]string, 0)
	for name, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, name)
		}
	}

	// Sort queue for deterministic output.
	// 对队列排序以保证输出确定性。
	sort.Strings(queue)

	for len(queue) > 0 {
		name := queue[0]
		queue = queue[1:]
		dag.LoadOrder = append(dag.LoadOrder, name)

		// Find providers that depend on this one.
		// 查找依赖此 provider 的其他 provider。
		for _, edge := range dag.Edges {
			if edge.To == name {
				inDegree[edge.From]--
				if inDegree[edge.From] == 0 {
					queue = append(queue, edge.From)
					sort.Strings(queue) // Keep sorted for determinism.
				}
			}
		}
	}

	// If there are cycles, some providers won't be in LoadOrder.
	// 如果存在循环，某些 provider 不会出现在 LoadOrder 中。
	// Add them at the end (they will fail to load anyway).
	// 将它们追加到末尾（它们本来就无法加载）。
	if len(dag.LoadOrder) < len(nodeMap) {
		visited := make(map[string]bool)
		for _, name := range dag.LoadOrder {
			visited[name] = true
		}
		var remaining []string
		for name := range nodeMap {
			if !visited[name] {
				remaining = append(remaining, name)
			}
		}
		sort.Strings(remaining)
		dag.LoadOrder = append(dag.LoadOrder, remaining...)
	}

	return dag
}
