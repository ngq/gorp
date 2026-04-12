package framework

import (
	"github.com/ngq/gorp/framework/container"
	"github.com/ngq/gorp/framework/contract"
)

// Application 是框架运行时的最小持有者。
//
// 中文说明：
// - Application 的职责非常克制：只负责持有一个 Container。
// - 这样做的目的，是把”框架启动时需要共享的根对象”收敛到一个地方，
//   避免后续各类 provider、命令入口、业务模块直接散落依赖具体实现。
// - 后续如果要扩展全局生命周期（例如 shutdown hook、全局 event bus、统一 runtime 状态），
//   可以继续挂在 Application 上，而不需要改变外部调用方式。
//
// 设计上，Application 本身不直接负责：
// - 注册 provider
// - 启动 HTTP/gRPC/Cron
// - 读取配置
// 这些工作都在上层 bootstrap / cmd 层完成。
type Application struct {
	// container 是整个框架运行时的依赖注入容器。
	//
	// 说明：
	// - 所有 provider 注册、服务解析，最终都围绕这个 container 展开。
	// - 这里用接口类型而不是具体实现，是为了让 framework 包只依赖 contract，
	//   保持上层调用面稳定。
	container contract.Container
}

// NewApplication 创建一个新的框架运行时实例。
//
// 中文说明：
// - 该函数会同时创建默认容器实现 `container.New()`。
// - 容器在创建时会把“容器自己”绑定到 `contract.ContainerKey`，
//   因此后续 factory/provider 内部也能反向取回 container 本身。
// - 当前函数不做任何 provider 注册，返回的 Application 只是一个“空运行时骨架”。
func NewApplication() *Application {
	c := container.New()
	return &Application{container: c}
}

// Container 返回当前运行时持有的依赖注入容器。
//
// 中文说明：
// - 上层通常通过这个入口继续注册 provider 或解析服务。
// - 这里不返回具体实现 `*container.Container`，是为了尽量让外部只依赖 contract。
func (a *Application) Container() contract.Container {
	return a.container
}
