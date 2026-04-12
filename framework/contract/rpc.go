package contract

import (
	"context"
)

const (
	// RPCClientKey 是 RPC 客户端在容器中的绑定 key。
	//
	// 中文说明：
	// - 用于服务间调用，支持 HTTP/gRPC 双实现；
	// - 单体项目使用 noop 实现，零依赖。
	RPCClientKey = "framework.rpc.client"

	// RPCServerKey 是 RPC 服务端在容器中的绑定 key。
	//
	// 中文说明：
	// - 用于暴露服务接口，支持 HTTP/gRPC 双实现；
	// - 单体项目可不注册，由 HTTP 服务替代。
	RPCServerKey = "framework.rpc.server"

	// RPCRegistryKey 是服务注册中心在容器中的绑定 key。
	//
	// 中文说明：
	// - 用于服务发现与注册；
	// - noop 实现返回空注册中心，单体项目不依赖 Consul/etcd。
	RPCRegistryKey = "framework.rpc.registry"
)

// RPCConfig 定义 RPC 配置。
//
// 中文说明：
// - mode 控制使用哪种 RPC 实现：noop/http/grpc；
// - noop 模式下单体项目零依赖，直接返回错误或本地调用。
type RPCConfig struct {
	// Mode: noop, http, grpc
	Mode string `mapstructure:"mode"`

	// 服务发现配置（grpc 模式使用）
	Registry string `mapstructure:"registry"`
	Address  string `mapstructure:"address"`

	// HTTP 模式配置
	BaseURL string `mapstructure:"base_url"`

	// gRPC 模式配置
	Target    string `mapstructure:"target"`
	Insecure  bool   `mapstructure:"insecure"`
	TimeoutMS int    `mapstructure:"timeout_ms"`
}

// RPCClient 定义 RPC 客户端抽象。
//
// 中文说明：
// - 这是服务间调用的统一入口；
// - HTTP 实现通过 REST API 调用；
// - gRPC 实现通过 protobuf 调用；
// - noop 实现返回错误，单体项目无需服务间调用。
type RPCClient interface {
	// Call 执行 RPC 调用。
	//
	// 中文说明：
	// - service: 目标服务名称（如 "user-service"）；
	// - method: 方法名称（如 "GetUser"）；
	// - req: 请求对象；
	// - resp: 响应对象（指针）；
	// - HTTP 模式下 service 映射为 baseURL，method 映射为 path；
	// - gRPC 模式下 service 映射为服务定义，method 映射为 RPC 方法。
	Call(ctx context.Context, service, method string, req, resp any) error

	// CallRaw 执行原始数据 RPC 调用。
	//
	// 中文说明：
	// - 用于不依赖 protobuf 的场景；
	// - 返回原始字节，由调用方自行反序列化。
	CallRaw(ctx context.Context, service, method string, data []byte) ([]byte, error)

	// Close 关闭客户端连接。
	Close() error
}

// RPCServer 定义 RPC 服务端抽象。
//
// 中文说明：
// - 用于暴露服务接口供其他服务调用；
// - HTTP 实现复用 Gin Engine；
// - gRPC 实现使用 grpc.Server；
// - noop 实现空操作，单体项目无需暴露 RPC。
type RPCServer interface {
	// Register 注册服务处理器。
	//
	// 中文说明：
	// - service: 服务名称；
	// - handler: 服务处理器（具体类型由实现决定）；
	// - HTTP 模式下 handler 为 gin.HandlerFunc；
	// - gRPC 模式下 handler 为 protobuf service implementation。
	Register(service string, handler any) error

	// Start 启动 RPC 服务。
	Start(ctx context.Context) error

	// Stop 停止 RPC 服务。
	Stop(ctx context.Context) error

	// Addr 返回服务监听地址。
	Addr() string
}

// ServiceRegistry 定义服务注册中心抽象。
//
// 中文说明：
// - 用于服务发现与注册；
// - Consul/etcd/Nacos 等实现此接口；
// - noop 实现空操作，单体项目无需服务发现。
type ServiceRegistry interface {
	// Register 注册服务实例。
	//
	// 中文说明：
	// - name: 服务名称；
	// - addr: 服务地址（如 "192.168.1.1:8080"）；
	// - meta: 元数据（如版本、权重等）。
	Register(ctx context.Context, name, addr string, meta map[string]string) error

	// Deregister 注销服务实例。
	Deregister(ctx context.Context, name, addr string) error

	// Discover 发现服务实例。
	//
	// 中文说明：
	// - 返回可用的服务地址列表；
	// - noop 实现返回空列表。
	Discover(ctx context.Context, name string) ([]ServiceInstance, error)

	// Close 关闭注册中心连接。
	Close() error
}

// ServiceInstance 表示一个服务实例。
//
// 中文说明：
// - 包含服务地址和元数据；
// - 用于负载均衡选择目标实例。
type ServiceInstance struct {
	// ID 实例唯一标识
	ID string

	// Name 服务名称
	Name string

	// Address 服务地址（如 "192.168.1.1:8080"）
	Address string

	// Metadata 服务元数据
	Metadata map[string]string

	// Healthy 是否健康
	Healthy bool
}