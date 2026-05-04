package middleware

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

type GRPCCarrier struct {
	md metadata.MD
}

func NewGRPCCarrier(md metadata.MD) *GRPCCarrier {
	if md == nil {
		md = make(metadata.MD)
	}
	return &GRPCCarrier{md: md}
}

func (c *GRPCCarrier) Get(key string) string {
	values := c.md.Get(key)
	if len(values) > 0 {
		return values[0]
	}
	return ""
}

func (c *GRPCCarrier) Set(key, value string) {
	c.md.Set(key, value)
}

func (c *GRPCCarrier) Add(key, value string) {
	c.md.Append(key, value)
}

func (c *GRPCCarrier) Keys() []string {
	keys := make([]string, 0, len(c.md))
	for k := range c.md {
		keys = append(keys, k)
	}
	return keys
}

func (c *GRPCCarrier) Values(key string) []string {
	return c.md.Get(key)
}

func (c *GRPCCarrier) MD() metadata.MD {
	return c.md
}

func UnaryServerInterceptor(propagator transportcontract.MetadataPropagator) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			carrier := NewGRPCCarrier(md)
			ctx = propagator.Extract(ctx, carrier)
		}
		return handler(ctx, req)
	}
}

func StreamServerInterceptor(propagator transportcontract.MetadataPropagator) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := ss.Context()
		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			carrier := NewGRPCCarrier(md)
			ctx = propagator.Extract(ctx, carrier)
		}
		wrapped := &wrappedServerStream{
			ServerStream: ss,
			ctx:          ctx,
		}
		return handler(srv, wrapped)
	}
}

type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}

func UnaryClientInterceptor(propagator transportcontract.MetadataPropagator) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok || md == nil {
			md = make(metadata.MD)
		} else {
			md = md.Copy()
		}
		carrier := NewGRPCCarrier(md)

		propagator.Inject(ctx, carrier)

		if len(md) > 0 {
			ctx = metadata.NewOutgoingContext(ctx, md)
		}

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func StreamClientInterceptor(propagator transportcontract.MetadataPropagator) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok || md == nil {
			md = make(metadata.MD)
		} else {
			md = md.Copy()
		}
		carrier := NewGRPCCarrier(md)

		propagator.Inject(ctx, carrier)

		if len(md) > 0 {
			ctx = metadata.NewOutgoingContext(ctx, md)
		}

		return streamer(ctx, desc, cc, method, opts...)
	}
}

func GetGRPCMetadata(ctx context.Context) metadata.MD {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		return md
	}
	md, _ := metadata.FromOutgoingContext(ctx)
	return md
}
