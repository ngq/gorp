// Application scenarios:
// - Define transport-layer RPC contracts shared by providers and higher-level runtime code.
// - Keep client, server, and registry semantics provider-neutral across HTTP and gRPC style transports.
// - Provide a shared config model for RPC mode, target, registry, and timeout settings.
//
// 适用场景：
// - 定义 provider 与上层运行时代码共享的 transport 层 RPC 契约。
// - 在 HTTP、gRPC 等不同传输模式下保持客户端、服务端和注册中心语义的 provider 中立。
// - 为 RPC 模式、目标地址、注册中心和超时设置提供共享配置模型。
package transport

import "context"

const (
	RPCClientKey   = "framework.rpc.client"
	RPCServerKey   = "framework.rpc.server"
	RPCRegistryKey = "framework.rpc.registry"
)

// RPCConfig describes RPC-related runtime configuration.
//
// RPCConfig 描述 RPC 相关运行时配置。
type RPCConfig struct {
	Mode string `mapstructure:"mode"`

	Registry string `mapstructure:"registry"`
	Address  string `mapstructure:"address"`

	BaseURL string `mapstructure:"base_url"`

	Target    string `mapstructure:"target"`
	Insecure  bool   `mapstructure:"insecure"`
	TimeoutMS int    `mapstructure:"timeout_ms"`
}

// RPCClient defines the outbound RPC client contract.
//
// RPCClient 定义出站 RPC 客户端契约。
type RPCClient interface {
	Call(ctx context.Context, service, method string, req, resp any) error
	CallRaw(ctx context.Context, service, method string, data []byte) ([]byte, error)
	Close() error
}

// RPCInvoker defines one outbound RPC invocation function.
//
// RPCInvoker 定义一次出站 RPC 调用函数。
type RPCInvoker func(ctx context.Context, service, method string, req, resp any) error

// RPCClientMiddleware defines one outbound RPC governance middleware.
//
// RPCClientMiddleware 定义一个出站 RPC 治理中间件。
type RPCClientMiddleware func(next RPCInvoker) RPCInvoker

// RPCServer defines the inbound RPC server contract.
//
// RPCServer 定义入站 RPC 服务端契约。
type RPCServer interface {
	Register(service string, handler any) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Addr() string
}

// ServiceRegistry defines the RPC service discovery registry contract.
//
// ServiceRegistry 定义 RPC 服务发现注册中心契约。
type ServiceRegistry interface {
	Register(ctx context.Context, name, addr string, meta map[string]string) error
	Deregister(ctx context.Context, name, addr string) error
	Discover(ctx context.Context, name string) ([]ServiceInstance, error)
	Close() error
}

// ServiceInstance describes one discovered RPC service instance.
//
// ServiceInstance 描述一个被发现的 RPC 服务实例。
type ServiceInstance struct {
	ID       string
	Name     string
	Address  string
	Metadata map[string]string
	Healthy  bool
}
