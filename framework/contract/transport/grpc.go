// Application scenarios:
// - Define transport-layer gRPC connection and server registration contracts.
// - Keep gRPC client/server assembly provider-neutral at the transport contract layer.
// - Let application and container helpers fetch gRPC capabilities through stable keys.
//
// 本契约刻意不 import google.golang.org/grpc：契约层不应绑定具体传输实现
// （依赖倒置）。实现方（framework/provider/rpc/grpc）用 *grpc.ClientConn /
// *grpc.Server 满足这里声明的 any 类型，调用方（proto-first 应用本就引入
// grpc）用类型断言或辅助函数转换。这让纯 HTTP 应用不必为 gRPC 的体积买单。
package transport

import "context"

const (
	GRPCConnFactoryKey     = "framework.grpc.conn_factory"
	GRPCServerRegistrarKey = "framework.grpc.server_registrar"
)

// GRPCConnFactory resolves outbound gRPC connections by service name.
//
// GRPCConnFactory 按服务名解析出站 gRPC 连接。
type GRPCConnFactory interface {
	// Conn returns the client connection of the target service.
	// The returned value is a *grpc.ClientConn; assert it with a type
	// assertion (or a helper from framework/provider/rpc/grpc).
	//
	// Conn 返回目标服务的客户端连接。返回值实际为 *grpc.ClientConn，
	// 用类型断言（或 rpc/grpc 的辅助函数）转换。
	Conn(ctx context.Context, service string) (any, error)
}

// GRPCServerRegistrar registers proto services onto a gRPC server.
//
// GRPCServerRegistrar 定义将 proto 服务注册到 gRPC server 的能力。
type GRPCServerRegistrar interface {
	// RegisterProto registers a proto service against the underlying gRPC server.
	// The register callback receives a *grpc.Server as any.
	//
	// RegisterProto 将 proto 服务注册到底层 gRPC server。
	// register 回调收到的 any 实际为 *grpc.Server。
	RegisterProto(register func(server any) error) error

	// Server returns the underlying gRPC server (*grpc.Server as any).
	//
	// Server 返回底层 gRPC server（*grpc.Server，以 any 形式返回）。
	Server() any
}
