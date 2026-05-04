package gorp

import (
	"context"

	"github.com/ngq/gorp/framework/contract/runtime"
	"github.com/ngq/gorp/framework/contract/transport"
	"github.com/ngq/gorp/framework/facade"
)

type Container = runtime.Container
type HTTPRouter = transport.HTTPRouter
type HTTPContext = transport.HTTPContext
type HTTPHandler = transport.HTTPHandler
type HTTPMiddleware = transport.HTTPMiddleware
type Metadata = transport.Metadata
type GRPCConnFactory = transport.GRPCConnFactory
type GRPCServerRegistrar = transport.GRPCServerRegistrar

func MakeGRPCConnFactory(c Container) (GRPCConnFactory, error) {
	return facade.MakeGRPCConnFactory(c)
}

func MakeGRPCServerRegistrar(c Container) (GRPCServerRegistrar, error) {
	return facade.MakeGRPCServerRegistrar(c)
}

func NewMetadata() Metadata {
	return transport.NewMetadata()
}

func NewServerContext(ctx context.Context, md Metadata) context.Context {
	return transport.NewServerContext(ctx, md)
}

func FromServerContext(ctx context.Context) (Metadata, bool) {
	return transport.FromServerContext(ctx)
}

func NewClientContext(ctx context.Context, md Metadata) context.Context {
	return transport.NewClientContext(ctx, md)
}

func FromClientContext(ctx context.Context) (Metadata, bool) {
	return transport.FromClientContext(ctx)
}

func AppendToClientContext(ctx context.Context, kv ...string) context.Context {
	return transport.AppendToClientContext(ctx, kv...)
}

func GetGRPCTraceID(ctx context.Context) string {
	return facade.GetGRPCTraceID(ctx)
}

func GetGRPCRequestID(ctx context.Context) string {
	return facade.GetGRPCRequestID(ctx)
}
