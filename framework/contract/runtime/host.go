// Application scenarios:
// - Define host and root-path contracts used by framework runtime assembly.
// - Provide a shared lifecycle model for long-running services managed by the host.
// - Keep service hosting abstractions independent from concrete server implementations.
//
// 适用场景：
// - 定义框架运行时装配使用的 host 与根路径契约。
// - 为 host 管理的长生命周期服务提供共享生命周期模型。
// - 让服务托管抽象不依赖具体服务器实现。
package runtime

import "context"

const (
	RootKey = "framework.root"
	HostKey = "framework.host"
)

// Root describes runtime filesystem layout access.
//
// Root 描述运行时文件系统布局访问能力。
type Root interface {
	BasePath() string
	StoragePath() string
	RuntimePath() string
	LogPath() string
	ConfigPath() string
	TempPath() string
}

// Host defines the host runtime that manages named services.
//
// Host 定义负责管理具名服务的宿主运行时。
type Host interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Shutdown(ctx context.Context) error
	RegisterService(name string, service Hostable) error
	Services() []string
}

// Hostable defines one host-managed service.
//
// Hostable 定义一个可被 host 管理的服务。
type Hostable interface {
	Name() string
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

// Lifecycle defines start/stop lifecycle hooks.
//
// Lifecycle 定义启动/停止生命周期钩子。
type Lifecycle interface {
	OnStarting(ctx context.Context) error
	OnStarted(ctx context.Context) error
	OnStopping(ctx context.Context) error
	OnStopped(ctx context.Context) error
}
