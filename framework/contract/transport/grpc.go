package transport

import (
	"context"

	"google.golang.org/grpc"
)

const (
	GRPCConnFactoryKey     = "framework.grpc.conn_factory"
	GRPCServerRegistrarKey = "framework.grpc.server_registrar"
)

type GRPCConnFactory interface {
	Conn(ctx context.Context, service string) (*grpc.ClientConn, error)
}

type GRPCServerRegistrar interface {
	RegisterProto(func(server *grpc.Server) error) error
	Server() *grpc.Server
}
