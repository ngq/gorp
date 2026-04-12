package contract

// ServiceProvider 定义了一组服务如何被注册到 Container 中。
//
// 中文说明：
// - Provider 是本仓库里“按能力分组注册服务”的核心抽象。
// - 例如 config/log/gin/gorm/cache 等能力，都会以 provider 的形式接入容器。
// - 这样做的好处是：
//   1. 可以把一组相关 binding 收拢在同一个模块里；
//   2. 可以为该组服务定义统一的初始化顺序；
//   3. 可以通过 `IsDefer + Provides` 支持延迟加载，降低启动时开销。
//
// 生命周期约定：
//  1. 外部调用 `RegisterProvider(p)`
//  2. 如果 `p.IsDefer() == false`：立即执行 `p.Register(c)`，随后执行 `p.Boot(c)`
//  3. 如果 `p.IsDefer() == true`：先记录 provider；等第一次 `Make()` 命中 `p.Provides()`
//     中任意 key 时，再触发 `Register + Boot`
//
// 这个设计直接对应教程中“服务基于协议 + provider 生命周期”的思路。
type ServiceProvider interface {
	// Name 返回 provider 的唯一名称。
	//
	// 中文说明：
	// - 这是 provider 在容器内部的唯一身份标识；
	// - 同名 provider 不允许重复注册。
	Name() string

	// Register 负责把当前 provider 提供的服务绑定进容器。
	//
	// 中文说明：
	// - 这个阶段通常只做 `Bind(...)` 等“注册动作”；
	// - 不建议在这里做过重的实际初始化（例如真正建立网络连接、启动 goroutine），
	//   这些更适合放到 Boot 阶段。
	Register(c Container) error

	// Boot 在 Register 之后执行，且只执行一次。
	//
	// 中文说明：
	// - 适合放“注册完成后才能安全执行”的初始化逻辑；
	// - 例如读取配置、做二次装配、启动内部组件等。
	Boot(c Container) error

	// IsDefer 表示该 provider 是否采用延迟加载。
	//
	// 中文说明：
	// - true：注册时只登记，首次解析相关 key 时再真正加载；
	// - false：注册阶段立即加载。
	IsDefer() bool

	// Provides 返回该 provider 能提供的 key 列表。
	//
	// 中文说明：
	// - 只有 deferred provider 会依赖这个列表来触发“首次命中加载”；
	// - 因此这个列表应当完整覆盖该 provider 对外暴露的服务 key。
	Provides() []string
}
