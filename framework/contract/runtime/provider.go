// Application scenarios:
// - Define the provider contract used by the runtime container and bootstrap flows.
// - Keep provider registration, boot order, and deferred loading semantics explicit.
// - Offer one stable provider abstraction across all framework capabilities.
//
// 适用场景：
// - 定义运行时容器与 bootstrap 流程使用的 provider 契约。
// - 显式表达 provider 的注册、启动顺序和延迟加载语义。
// - 为所有框架能力提供统一稳定的 provider 抽象。
package runtime

// ServiceProvider defines one runtime service provider.
//
// ServiceProvider 定义一个运行时服务 provider。
type ServiceProvider interface {
	// Name returns the unique provider name.
	//
	// Name 返回唯一的 provider 名称。
	Name() string

	// Register binds provider services into the container.
	//
	// Register 将 provider 服务绑定到容器中。
	Register(c Container) error

	// Boot performs boot-time initialization after registration.
	//
	// Boot 在注册完成后执行启动期初始化。
	Boot(c Container) error

	// IsDefer reports whether registration should be deferred until needed.
	//
	// IsDefer 表示该 provider 是否应延迟到真正需要时再装载。
	IsDefer() bool

	// Provides returns the keys that can be resolved from this provider.
	//
	// Provides 返回该 provider 可以提供的 key 列表。
	Provides() []string

	// DependsOn returns the keys that this provider depends on.
	// Used for dependency graph construction and load order computation.
	//
	// DependsOn 返回该 provider 依赖的 key 列表。
	// 用于构建依赖图和计算加载顺序。
	DependsOn() []string
}
