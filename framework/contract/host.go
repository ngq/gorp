package contract

import "context"

// Root/Host 相关的 key 定义。
//
// 中文说明：
// - RootKey: 应用根目录服务
// - HostKey: 应用宿主生命周期服务
const (
	RootKey = "framework.root"
	HostKey = "framework.host"
)

// Root 定义应用根目录相关的能力。
//
// 中文说明：
// - 这是"dedicated root contract 正式化"的核心抽象；
// - 统一管理应用的根目录、存储路径、运行时路径、日志路径等；
// - 所有路径相关的能力都通过这个接口访问，避免散落在各个 provider 中。
type Root interface {
	// BasePath 返回应用根目录。
	// 中文说明：这是应用的基础路径，通常包含 go.mod 的目录。
	BasePath() string

	// StoragePath 返回存储目录。
	// 中文说明：用于存放文件、上传内容等持久化数据。
	StoragePath() string

	// RuntimePath 返回运行时目录。
	// 中文说明：用于存放 PID 文件、socket 文件等运行时数据。
	RuntimePath() string

	// LogPath 返回日志目录。
	// 中文说明：用于存放应用日志文件。
	LogPath() string

	// ConfigPath 返回配置目录。
	// 中文说明：用于存放配置文件，默认为 BasePath/config。
	ConfigPath() string

	// TempPath 返回临时文件目录。
	// 中文说明：用于存放临时文件，默认为 RuntimePath/tmp。
	TempPath() string
}

// Host 定义应用宿主的生命周期管理能力。
//
// 中文说明：
// - 这是"dedicated host contract 正式化"的核心抽象；
// - 统一管理应用的启动、停止、优雅关闭等生命周期；
// - 支持 HTTP、gRPC、Cron 等多种运行模式的统一管理。
type Host interface {
	// Start 启动应用宿主。
	// 中文说明：
	// - 根据配置启动相应的服务（HTTP/gRPC/Cron）；
	// - 阻塞直到收到停止信号。
	Start(ctx context.Context) error

	// Stop 停止应用宿主。
	// 中文说明：
	// - 触发优雅关闭流程；
	// - 等待所有服务完成清理工作。
	Stop(ctx context.Context) error

	// Shutdown 触发优雅关闭。
	// 中文说明：
	// - 发送关闭信号给所有注册的服务；
	// - 按注册顺序逆序关闭。
	Shutdown(ctx context.Context) error

	// RegisterService 注册服务。
	// 中文说明：
	// - 服务需要实现 Hostable 接口；
	// - 服务会在 Host 启动时按注册顺序启动；
	// - 服务会在 Host 关闭时按逆序关闭。
	RegisterService(name string, service Hostable) error

	// Services 返回所有已注册的服务名称。
	Services() []string
}

// Hostable 定义可被 Host 管理的服务接口。
//
// 中文说明：
// - 所有需要被 Host 统一管理的服务都需要实现这个接口；
// - 例如 HTTP Server、gRPC Server、Cron Scheduler 等；
// - 框架通过这个接口实现统一的生命周期管理。
type Hostable interface {
	// Name 返回服务名称。
	Name() string

	// Start 启动服务。
	// 中文说明：
	// - 非阻塞启动，服务应在后台运行；
	// - 启动失败应返回 error。
	Start(ctx context.Context) error

	// Stop 停止服务。
	// 中文说明：
	// - 阻塞直到服务完全停止；
	// - 应实现优雅关闭，等待进行中的请求完成。
	Stop(ctx context.Context) error
}

// Lifecycle 定义服务生命周期的钩子接口。
//
// 中文说明：
// - 可选接口，服务可以实现它来接收生命周期事件；
// - OnStarting 在服务启动前调用；
// - OnStarted 在服务启动后调用；
// - OnStopping 在服务停止前调用；
// - OnStopped 在服务停止后调用。
type Lifecycle interface {
	OnStarting(ctx context.Context) error
	OnStarted(ctx context.Context) error
	OnStopping(ctx context.Context) error
	OnStopped(ctx context.Context) error
}