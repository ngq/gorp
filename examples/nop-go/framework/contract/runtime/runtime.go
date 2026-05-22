// Application scenarios:
// - Define runtime configurator contracts used during bootstrap assembly.
// - Let features inject HTTP, cron, and gRPC runtime behavior without hard-coding one implementation.
// - Keep configurator interfaces minimal and composition-friendly.
//
// 适用场景：
// - 定义 bootstrap 装配阶段使用的运行时配置器契约。
// - 让特性模块在不写死实现的前提下注入 HTTP、cron 和 gRPC 运行时行为。
// - 保持配置器接口最小化且易于组合。
package runtime

import "google.golang.org/grpc"

const (
	HTTPRuntimeConfiguratorKey = "app.runtime.http_configurator"
	CronRuntimeConfiguratorKey = "app.runtime.cron_configurator"
	GRPCRuntimeBuilderKey      = "app.runtime.grpc_builder"
)

// HTTPRuntimeConfigurator configures the HTTP runtime inside a container.
//
// HTTPRuntimeConfigurator 定义容器内的 HTTP 运行时配置器。
type HTTPRuntimeConfigurator interface {
	// ConfigureHTTPRuntime configures the HTTP runtime.
	//
	// ConfigureHTTPRuntime 配置 HTTP 运行时。
	ConfigureHTTPRuntime(Container) error
}

// CronRuntimeConfigurator configures the cron runtime inside a container.
//
// CronRuntimeConfigurator 定义容器内的 cron 运行时配置器。
type CronRuntimeConfigurator interface {
	// ConfigureCronRuntime configures the cron runtime and returns the number of registered jobs.
	//
	// ConfigureCronRuntime 配置 cron 运行时，并返回注册任务数量。
	ConfigureCronRuntime(Container) (int, error)
}

// GRPCRuntimeBuilder builds a gRPC server runtime.
//
// GRPCRuntimeBuilder 定义 gRPC 服务端运行时构建器。
type GRPCRuntimeBuilder interface {
	// BuildGRPCServer builds and returns a gRPC server.
	//
	// BuildGRPCServer 构建并返回一个 gRPC 服务端。
	BuildGRPCServer() *grpc.Server
}
