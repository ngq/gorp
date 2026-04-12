package grpc

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	grpc_health_v1 "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

// NewServer 创建并注册当前母仓默认的 gRPC Server。
//
// 中文说明：
// - 当前母仓不再内置业务级 gRPC demo 服务；
// - 这里只保留最小的 health + reflection，作为框架运行时能力入口；
// - 真正的业务 gRPC 示例转移到外部案例中展示。
func NewServer() *grpc.Server {
	srv := grpc.NewServer(
		grpc.UnaryInterceptor(UnaryServerInterceptor()),
		grpc.StreamInterceptor(StreamServerInterceptor()),
	)

	hs := health.NewServer()
	hs.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(srv, hs)
	reflection.Register(srv)

	return srv
}
