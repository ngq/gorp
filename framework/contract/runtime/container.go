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

// ContainerKey is the container self-binding key.
//
// ContainerKey 是容器自绑定时使用的 key。
const ContainerKey = "framework.container"

// Container defines the runtime dependency injection surface.
//
// Container 定义运行时依赖注入接口面。
type Container interface {
	// Bind registers a factory under a key.
	//
	// Bind 将 factory 绑定到目标 key。
	Bind(key string, factory Factory, singleton bool)

	// IsBind reports whether a key is bound.
	//
	// IsBind 返回指定 key 是否已绑定。
	IsBind(key string) bool

	// Make resolves a service by key.
	//
	// Make 按 key 解析服务。
	Make(key string) (any, error)

	// MustMake resolves a service by key and panics on error.
	//
	// MustMake 按 key 解析服务，失败时 panic。
	MustMake(key string) any

	// RegisterProvider registers one provider into the container.
	//
	// RegisterProvider 将单个 provider 注册进容器。
	RegisterProvider(p ServiceProvider) error

	// RegisterProviders registers multiple providers into the container.
	//
	// RegisterProviders 将多个 provider 注册进容器。
	RegisterProviders(providers ...ServiceProvider) error
}

// Factory creates one service value from the container.
//
// Factory 定义从容器创建服务值的工厂函数。
type Factory func(Container) (any, error)
