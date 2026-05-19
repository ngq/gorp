package grpc

import (
	"google.golang.org/grpc"
	userv1 "grpc-demo/proto/user/v1"
	"grpc-demo/services/user/internal/service"
)

func RegisterUserService(server *grpc.Server, services *service.Services) {
	userv1.RegisterUserServiceServer(server, services.User)
}
