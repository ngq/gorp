// Application scenarios:
// - Define transport-layer gRPC connection and server registration contracts.
// - Keep gRPC client/server assembly provider-neutral at the transport contract layer.
// - Let application and container helpers fetch gRPC capabilities through stable keys.
//
// 适用场景：
// - 定义 transport 层 gRPC 连接与服务端注册契约。
// - 在 transport 契约层保持 gRPC 客户端/服务端装配与 provider 解耦。
// - 让 application 和 container helper 通过稳定 key 获取 gRPC 能力。
package transport

import (
	"context"

	"google.golang.org/grpc"
)

const (
	GRPCConnFactoryKey     = "framework.grpc.conn_factory"
	GRPCServerRegistrarKey = "framework.grpc.server_registrar"
)

// GRPCConnFactory resolves outbound gRPC connections by service name.
//
// GRPCConnFactory 按服务名解析出站 gRPC 连接。
type GRPCConnFactory interface {
	// Conn returns the client connection of the target service.
	//
	// Conn 返回目标服务的客户端连接。
	Conn(ctx context.Context, service string) (*grpc.ClientConn, error)
}

// GRPCServerRegistrar registers proto services onto a gRPC server.
//
// GRPCServerRegistrar 定义将 proto 服务注册到 gRPC server 的能力。
type GRPCServerRegistrar interface {
	// RegisterProto registers a proto service against the underlying gRPC server.
	//
	// RegisterProto 将 proto 服务注册到底层 gRPC server 上。
	RegisterProto(func(server *grpc.Server) error) error

	// Server returns the underlying gRPC server.
	//
	// Server 返回底层 gRPC server。
	Server() *grpc.Server
}
