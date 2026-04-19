package contract

import (
	"context"

	"google.golang.org/grpc"
)

const (
	// GRPCConnFactoryKey 是 gRPC 连接工厂在容器中的绑定 key。
	//
	// 中文说明：
	// - 这是 Proto-first gRPC 主线的客户端连接入口；
	// - 业务侧通过它按服务名获取由 framework 托管的 `*grpc.ClientConn`；
	// - 它不再要求业务方手工维护连接池、服务发现、selector 或 metadata/tracing/serviceauth 注入。
	GRPCConnFactoryKey = "framework.grpc.conn_factory"

	// GRPCServerRegistrarKey 是 gRPC 服务端注册器在容器中的绑定 key。
	//
	// 中文说明：
	// - 这是 Proto-first gRPC 主线的服务端注册入口；
	// - 业务侧通过它把 `pb.RegisterXxxServer(...)` 挂到 framework 托管的 `grpc.Server`；
	// - 它不再要求业务方通过弱类型 `RPCServer.Register(service string, handler any)` 表达 gRPC 注册。
	GRPCServerRegistrarKey = "framework.grpc.server_registrar"
)

// GRPCConnFactory 提供按服务名获取 gRPC 连接的能力。
//
// 中文说明：
// - Proto-first 客户端主线应围绕 `pb.NewXxxClient(conn)` 工作；
// - framework 负责服务发现、selector、metadata、tracing、serviceauth 与连接复用；
// - 业务侧只消费标准 `*grpc.ClientConn`。
type GRPCConnFactory interface {
	Conn(ctx context.Context, service string) (*grpc.ClientConn, error)
}

// GRPCServerRegistrar 提供标准 gRPC service 注册能力。
//
// 中文说明：
// - 业务侧通过 `RegisterProto(func(server *grpc.Server) error { ... })` 挂接 protobuf service；
// - framework 负责 server 生命周期、公共 interceptor、health/reflection 等通用能力；
// - 这样业务实现可以保持标准 `pb.RegisterXxxServer(...)` 心智。
type GRPCServerRegistrar interface {
	RegisterProto(func(server *grpc.Server) error) error

	// Server 返回底层 `grpc.Server`，供需要直接访问底层 server 的场景使用。
	//
	// 中文说明：
	// - 常规业务接入优先用 RegisterProto；
	// - 暴露该方法主要是为了保留最小逃生舱和底层扩展能力。
	Server() *grpc.Server
}
