package contract

// ContainerKey 是“容器自己”在容器中的默认绑定 key。
//
// 中文说明：
// - 创建默认容器时，框架会把 container 自己绑定到这个 key。
// - 这样做的好处是：在 factory / provider / service 初始化过程中，
//   如果需要继续做依赖解析，可以通过 container 再把自己取出来。
const ContainerKey = "framework.container"

// Container 定义了框架中依赖注入容器的最小能力。
//
// 中文说明：
// - 这是整个框架的“服务解析中枢”。
// - 所有服务都通过字符串 key 标识，而不是直接按类型解析；这是本仓库延续教程设计的约定。
// - Provider 负责把一组相关服务注册进 Container，并在需要时执行生命周期钩子。
// - Container 接口只暴露最核心的动作：绑定、判断是否已绑定、解析、强制解析、注册 Provider。
//
// 使用建议：
// - 在 provider 内部优先使用 `Bind` 注册服务；
// - 在业务/命令入口里优先使用 `Make`；
// - 只有在“缺失依赖就属于程序错误”的场景下，才使用 `MustMake`。
type Container interface {
	// Bind 把一个字符串 key 绑定到一个 factory。
	//
	// 中文说明：
	// - factory 只有在真正 `Make(key)` 时才会执行；
	// - 如果 singleton=true，则第一次创建出的实例会被缓存，后续重复解析返回同一实例；
	// - 如果 singleton=false，则每次解析都会重新执行 factory。
	Bind(key string, factory Factory, singleton bool)

	// IsBind 判断一个 key 当前是否“可被容器解析”。
	//
	// 中文说明：
	// - 这里不仅检查已经显式 Bind 的 key；
	// - 对于 deferred provider 提供的 key，只要 provider 已登记，也会返回 true。
	IsBind(key string) bool

	// Make 解析一个 key，并返回对应服务实例。
	//
	// 中文说明：
	// - 如果 key 属于某个 deferred provider，Make 会先触发该 provider 的 Register/Boot；
	// - 如果 key 未绑定，则返回 error；
	// - 推荐业务代码优先使用这个方法，而不是直接 panic。
	Make(key string) (any, error)

	// MustMake 解析一个 key；如果失败则直接 panic。
	//
	// 中文说明：
	// - 适用于“缺少该依赖属于程序启动/装配错误”的场景；
	// - 不适合直接用于处理外部输入相关逻辑。
	MustMake(key string) any

	// RegisterProvider 注册一个服务提供者。
	//
	// 中文说明：
	// - 非延迟 provider：注册时立即执行 Register，然后 Boot；
	// - 延迟 provider：先登记，等第一次 Make 命中其 Provides() 中任一 key 时再真正加载。
	RegisterProvider(p ServiceProvider) error

	// RegisterProviders 批量注册多个服务提供者。
	//
	// 中文说明：
	// - 按顺序依次注册，遇到错误立即停止并返回；
	// - 错误信息会包含失败的 provider 名称，便于定位问题；
	// - 相比多次调用 RegisterProvider，这种方式更简洁且错误信息更友好。
	RegisterProviders(providers ...ServiceProvider) error
}

// Factory 定义了“如何创建某个服务实例”。
//
// 中文说明：
// - Factory 会接收当前 Container，因此它内部可以继续解析其他依赖；
// - 这也是 provider/service 初始化链能逐层展开的基础。
type Factory func(Container) (any, error)
