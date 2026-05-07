package interceptor

import (
	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	providergrpc "github.com/ngq/gorp/framework/provider/grpc"
	metadatamw "github.com/ngq/gorp/framework/provider/metadata/middleware"
	tracingmw "github.com/ngq/gorp/framework/provider/tracing/middleware"
	"google.golang.org/grpc"
)

// DefaultServerPresetOptions controls the default gRPC server governance preset.
//
// DefaultServerPresetOptions 用于控制默认 gRPC 服务端治理预设。
type DefaultServerPresetOptions struct {
	Tracer             observabilitycontract.Tracer
	ServiceName        string
	MetadataPropagator transportcontract.MetadataPropagator
}

// DefaultUnaryServerInterceptors returns the default unary server governance interceptors.
//
// DefaultUnaryServerInterceptors 返回默认 unary 服务端治理拦截器。
func DefaultUnaryServerInterceptors(opts DefaultServerPresetOptions) []grpc.UnaryServerInterceptor {
	interceptors := make([]grpc.UnaryServerInterceptor, 0, 3)
	interceptors = append(interceptors, providergrpc.UnaryServerInterceptor())
	if opts.Tracer != nil {
		interceptors = append(interceptors, tracingmw.UnaryServerInterceptor(opts.Tracer, opts.ServiceName))
	}
	if opts.MetadataPropagator != nil {
		interceptors = append(interceptors, metadatamw.UnaryServerInterceptor(opts.MetadataPropagator))
	}
	return ChainUnary(interceptors...)
}

// DefaultStreamServerInterceptors returns the default stream server governance interceptors.
//
// DefaultStreamServerInterceptors 返回默认 stream 服务端治理拦截器。
func DefaultStreamServerInterceptors(opts DefaultServerPresetOptions) []grpc.StreamServerInterceptor {
	interceptors := make([]grpc.StreamServerInterceptor, 0, 2)
	interceptors = append(interceptors, providergrpc.StreamServerInterceptor())
	if opts.MetadataPropagator != nil {
		interceptors = append(interceptors, metadatamw.StreamServerInterceptor(opts.MetadataPropagator))
	}
	return ChainStream(interceptors...)
}
