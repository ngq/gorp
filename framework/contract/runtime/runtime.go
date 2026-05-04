package runtime

import "google.golang.org/grpc"

const (
	HTTPRuntimeConfiguratorKey = "app.runtime.http_configurator"
	CronRuntimeConfiguratorKey = "app.runtime.cron_configurator"
	GRPCRuntimeBuilderKey      = "app.runtime.grpc_builder"
)

type HTTPRuntimeConfigurator interface {
	ConfigureHTTPRuntime(Container) error
}

type CronRuntimeConfigurator interface {
	ConfigureCronRuntime(Container) (int, error)
}

type GRPCRuntimeBuilder interface {
	BuildGRPCServer() *grpc.Server
}
