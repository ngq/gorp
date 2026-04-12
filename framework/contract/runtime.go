package contract

import "google.golang.org/grpc"

const (
	HTTPRuntimeConfiguratorKey = "app.runtime.http_configurator"
	CronRuntimeConfiguratorKey = "app.runtime.cron_configurator"
	GRPCRuntimeBuilderKey      = "app.runtime.grpc_builder"
)

// HTTPRuntimeConfigurator 收口 HTTP 启动前的 app 装配动作。
type HTTPRuntimeConfigurator interface {
	ConfigureHTTPRuntime(Container) error
}

// CronRuntimeConfigurator 收口 cron worker 启动前的任务装配动作。
type CronRuntimeConfigurator interface {
	ConfigureCronRuntime(Container) (int, error)
}

// GRPCRuntimeBuilder 收口 gRPC server 的 app 装配动作。
type GRPCRuntimeBuilder interface {
	BuildGRPCServer() *grpc.Server
}
