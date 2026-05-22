// Package application provides application startup entrypoints for gorp framework.
// Exposes container-backed capability accessors and transport convenience APIs.
// Offers stable helpers for service identity propagation and gRPC metadata access.
//
// 应用启动包提供 gorp 框架的应用启动入口。
// 暴露基于容器的能力获取入口和 transport 便捷 API。
// 提供稳定的服务身份透传助手和 gRPC 元数据读取助手。
package application

import (
	"context"

	frameworkcontainer "github.com/ngq/gorp/framework/container"
	datacontract "github.com/ngq/gorp/framework/contract/data"
	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	securitycontract "github.com/ngq/gorp/framework/contract/security"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	frameworkgrpc "github.com/ngq/gorp/framework/provider/grpc"
)

// MakeGRPCConnFactory returns the proto-first gRPC connection factory from the container.
//
// MakeGRPCConnFactory 获取 Proto-first gRPC 连接工厂。
func MakeGRPCConnFactory(c runtimecontract.Container) (transportcontract.GRPCConnFactory, error) {
	return frameworkcontainer.MakeGRPCConnFactory(c)
}

// MakeGRPCServerRegistrar returns the proto-first gRPC server registrar from the container.
//
// MakeGRPCServerRegistrar 获取 Proto-first gRPC 服务端注册器。
func MakeGRPCServerRegistrar(c runtimecontract.Container) (transportcontract.GRPCServerRegistrar, error) {
	return frameworkcontainer.MakeGRPCServerRegistrar(c)
}

// MakeDistributedLock returns the distributed lock capability from the container.
//
// MakeDistributedLock 获取分布式锁能力。
func MakeDistributedLock(c runtimecontract.Container) (datacontract.DistributedLock, error) {
	return frameworkcontainer.MakeDistributedLock(c)
}

// MakeMessagePublisher returns the message publishing capability from the container.
//
// MakeMessagePublisher 获取消息发布能力。
func MakeMessagePublisher(c runtimecontract.Container) (integrationcontract.MessagePublisher, error) {
	return frameworkcontainer.MakeMessagePublisher(c)
}

// MakeMessageSubscriber returns the message subscription capability from the container.
//
// MakeMessageSubscriber 获取消息订阅能力。
func MakeMessageSubscriber(c runtimecontract.Container) (integrationcontract.MessageSubscriber, error) {
	return frameworkcontainer.MakeMessageSubscriber(c)
}

// WithServiceIdentity writes service identity into the context.
//
// WithServiceIdentity 把服务身份写入上下文。
//
// Example:
//
//	ctx = application.WithServiceIdentity(ctx, identity)
func WithServiceIdentity(ctx context.Context, identity *securitycontract.ServiceIdentity) context.Context {
	return securitycontract.NewServiceIdentityContext(ctx, identity)
}

// FromServiceIdentity reads service identity from the context.
//
// FromServiceIdentity 读取上下文中的服务身份。
func FromServiceIdentity(ctx context.Context) (*securitycontract.ServiceIdentity, bool) {
	return securitycontract.FromServiceIdentityContext(ctx)
}

// GetGRPCTraceID reads the trace id from a gRPC context.
//
// GetGRPCTraceID 从 gRPC context 读取 trace id。
func GetGRPCTraceID(ctx context.Context) string {
	return frameworkgrpc.GetTraceID(ctx)
}

// GetGRPCRequestID reads the request id from a gRPC context.
//
// GetGRPCRequestID 从 gRPC context 读取 request id。
func GetGRPCRequestID(ctx context.Context) string {
	return frameworkgrpc.GetRequestID(ctx)
}
