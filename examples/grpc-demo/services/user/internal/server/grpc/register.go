package grpc

import (
	userv1 "grpc-demo/proto/user/v1"
	"grpc-demo/services/user/internal/service"
	"google.golang.org/grpc"
)

func RegisterUserService(server *grpc.Server, services *service.Services) {
	userv1.RegisterUserServiceServer(server, services.User)
}
