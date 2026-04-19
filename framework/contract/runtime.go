package contract

import "google.golang.org/grpc"

const (
	HTTPRuntimeConfiguratorKey = "app.runtime.http_configurator"
	CronRuntimeConfiguratorKey = "app.runtime.cron_configurator"
	GRPCRuntimeBuilderKey      = "app.runtime.grpc_builder"
)

// HTTPRuntimeConfigurator 收口 HTTP 启动前的项目侧装配动作。
//
// 中文说明：
// - 主要服务于 legacy/runtime CLI 命令组在启动前补齐项目自己的 migrate / route assembly；
// - generated starter 的默认公开启动路径仍应优先走项目自己的 `cmd/*/main.go`；
// - 因此这里是运行时扩展位，不是 starter 用户默认需要感知的入口。
type HTTPRuntimeConfigurator interface {
	ConfigureHTTPRuntime(Container) error
}

// CronRuntimeConfigurator 收口 cron worker 启动前的项目侧任务装配动作。
//
// 中文说明：
// - 主要服务于 legacy/runtime CLI 命令组在启动 worker 前补齐项目自己的任务注册；
// - 它属于运行时扩展位，而不是 generated starter 的默认公开启动心智。
type CronRuntimeConfigurator interface {
	ConfigureCronRuntime(Container) (int, error)
}

// GRPCRuntimeBuilder 收口 legacy gRPC runtime 命令组的 server 构造动作。
//
// 中文说明：
// - 当前只表达 CLI/runtime 维度的 gRPC server 构造扩展点；
// - 不代表 gRPC 的正式公开主线；
// - Proto-first 的 gRPC 主线路径会在后续专项中单独收口。
type GRPCRuntimeBuilder interface {
	BuildGRPCServer() *grpc.Server
}
