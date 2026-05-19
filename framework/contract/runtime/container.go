// Application scenarios:
// - Define the minimal runtime container contract used across framework packages.
// - Keep dependency resolution, provider registration, and singleton binding semantics stable.
// - Let different container implementations conform to one shared runtime abstraction.
//
// 适用场景：
// - 定义框架各包共享的最小运行时容器契约。
// - 稳定维护依赖解析、provider 注册和单例绑定语义。
// - 让不同容器实现都能遵循同一套运行时抽象。
package runtime

import "io"

// ContainerKey is the container self-binding key.
//
// ContainerKey 是容器自绑定时使用的 key。
const ContainerKey = "framework.container"

// Container defines the runtime dependency injection surface.
//
// Container 定义运行时依赖注入接口面。
type Container interface {
	// Bind registers a factory under a key.
	// If the key is already bound, a warning is logged and the previous binding is replaced.
	//
	// Bind 将 factory 绑定到目标 key。
	// 如果 key 已绑定，打印警告日志并替换原有绑定。
	Bind(key string, factory Factory, singleton bool)

	// NamedBind registers a named factory under a key.
	// Allows multiple implementations of the same key to coexist under different names.
	//
	// NamedBind 将命名 factory 绑定到目标 key。
	// 允许同一 key 的多个实现以不同名称共存。
	NamedBind(name, key string, factory Factory, singleton bool)

	// IsBind reports whether a key is bound.
	//
	// IsBind 返回指定 key 是否已绑定。
	IsBind(key string) bool

	// IsBindNamed reports whether a named binding exists.
	//
	// IsBindNamed 返回指定命名绑定是否存在。
	IsBindNamed(name, key string) bool

	// Make resolves a service by key.
	//
	// Make 按 key 解析服务。
	Make(key string) (any, error)

	// MakeNamed resolves a named service by name and key.
	//
	// MakeNamed 按名称和 key 解析命名服务。
	MakeNamed(name, key string) (any, error)

	// MustMake resolves a service by key and panics on error.
	//
	// MustMake 按 key 解析服务，失败时 panic。
	MustMake(key string) any

	// MustMakeNamed resolves a named service by name and key and panics on error.
	//
	// MustMakeNamed 按名称和 key 解析命名服务，失败时 panic。
	MustMakeNamed(name, key string) any

	// RegisterCloser registers an io.Closer to be called during Destroy.
	// Closers are called in reverse registration order.
	//
	// RegisterCloser 注册一个 io.Closer，在 Destroy 时调用。
	// Closer 按注册逆序调用。
	RegisterCloser(key string, closer io.Closer)

	// Destroy calls all registered closers in reverse order and marks the container as destroyed.
	// After Destroy, Make/MakeNamed return ErrContainerDestroyed.
	//
	// Destroy 按注册逆序调用所有 Closer，并将容器标记为已销毁。
	// 销毁后 Make/MakeNamed 返回 ErrContainerDestroyed。
	Destroy() error

	// RegisterProvider registers one provider into the container.
	//
	// RegisterProvider 将单个 provider 注册进容器。
	RegisterProvider(p ServiceProvider) error

	// RegisterProviders registers multiple providers into the container.
	//
	// RegisterProviders 将多个 provider 注册进容器。
	RegisterProviders(providers ...ServiceProvider) error

	// RegisteredProviders returns the names of all registered providers with their load/boot status.
	//
	// RegisteredProviders 返回所有已注册 provider 的名称及其加载/引导状态。
	RegisteredProviders() []ProviderInfo

	// DebugPrint returns a human-readable snapshot of the container state for diagnostics.
	//
	// DebugPrint 返回容器状态的人类可读快照，用于诊断。
	DebugPrint() string

	// ProviderDAG returns the provider dependency graph for visualization and analysis.
	// Useful for debugging provider load order and detecting circular dependencies.
	//
	// ProviderDAG 返回 provider 依赖图，用于可视化和分析。
	// 用于调试 provider 加载顺序和检测循环依赖。
	ProviderDAG() ProviderDAG
}

// ProviderInfo describes a registered provider's state.
//
// ProviderInfo 描述已注册 provider 的状态。
type ProviderInfo struct {
	Name    string
	Loaded  bool
	Booted  bool
	IsDefer bool
}

// ProviderDAGNode represents a node in the provider dependency graph.
//
// ProviderDAGNode 表示 provider 依赖图中的一个节点。
type ProviderDAGNode struct {
	Name      string   // Provider 名称
	Provides  []string // 提供的契约键
	DependsOn []string // 依赖的契约键
	IsDefer   bool     // 是否延迟加载
	Loaded    bool     // 是否已加载
	Booted    bool     // 是否已启动
}

// ProviderDAG represents the provider dependency graph.
//
// ProviderDAG 表示 provider 依赖图。
type ProviderDAG struct {
	Nodes     []ProviderDAGNode // 所有节点
	Edges     []DAGEdge         // 依赖边
	Cycles    [][]string        // 检测到的循环依赖
	LoadOrder []string          // 推荐加载顺序
}

// DAGEdge represents a dependency edge in the DAG.
//
// DAGEdge 表示 DAG 中的依赖边。
type DAGEdge struct {
	From string // 依赖方 provider
	To   string // 被依赖方 provider（可能为空，表示依赖外部绑定）
	Key  string // 依赖的契约键
}

// Factory creates one service value from the container.
//
// Factory 定义从容器创建服务值的工厂函数。
type Factory func(Container) (any, error)
