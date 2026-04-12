package middleware

import (
	"context"

	"github.com/ngq/gorp/framework/contract"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// GRPCCarrier 实现 MetadataCarrier 接口。
//
// 中文说明：
// - 包装 gRPC metadata.MD；
// - 用于 gRPC 请求的 metadata 提取/注入。
type GRPCCarrier struct {
	md metadata.MD
}

// NewGRPCCarrier 创建 GRPCCarrier。
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

// MD 返回 gRPC metadata.MD。
func (c *GRPCCarrier) MD() metadata.MD {
	return c.md
}

// UnaryServerInterceptor gRPC 服务端一元拦截器。
//
// 中文说明：
// - 从 gRPC Metadata 提取 metadata 存入 context；
// - 支持前缀过滤（默认 x-md- 前缀）；
// - 自己实现，不抄袭 Kratos。
func UnaryServerInterceptor(propagator contract.MetadataPropagator) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		// 从 gRPC Metadata 提取
		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			carrier := NewGRPCCarrier(md)
			ctx = propagator.Extract(ctx, carrier)
		}
		return handler(ctx, req)
	}
}

// StreamServerInterceptor gRPC 服务端流拦截器。
func StreamServerInterceptor(propagator contract.MetadataPropagator) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := ss.Context()
		// 从 gRPC Metadata 提取
		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			carrier := NewGRPCCarrier(md)
			ctx = propagator.Extract(ctx, carrier)
		}
		// 包装 ServerStream 以使用新的 context
		wrapped := &wrappedServerStream{
			ServerStream: ss,
			ctx:          ctx,
		}
		return handler(srv, wrapped)
	}
}

// wrappedServerStream 包装 grpc.ServerStream 以替换 context。
type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}

// UnaryClientInterceptor gRPC 客户端一元拦截器。
//
// 中文说明：
// - 从 context 读取 metadata 注入 gRPC Metadata；
// - 用于客户端发起请求时。
func UnaryClientInterceptor(propagator contract.MetadataPropagator) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		// 创建 gRPC Metadata
		md := make(metadata.MD)
		carrier := NewGRPCCarrier(md)

		// 注入 metadata
		propagator.Inject(ctx, carrier)

		// 如果有 metadata，添加到 outgoing context
		if len(md) > 0 {
			ctx = metadata.NewOutgoingContext(ctx, md)
		}

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// StreamClientInterceptor gRPC 客户端流拦截器。
func StreamClientInterceptor(propagator contract.MetadataPropagator) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		// 创建 gRPC Metadata
		md := make(metadata.MD)
		carrier := NewGRPCCarrier(md)

		// 注入 metadata
		propagator.Inject(ctx, carrier)

		// 如果有 metadata，添加到 outgoing context
		if len(md) > 0 {
			ctx = metadata.NewOutgoingContext(ctx, md)
		}

		return streamer(ctx, desc, cc, method, opts...)
	}
}

// GetGRPCMetadata 从 gRPC context 获取 metadata。
//
// 中文说明：
// - 优先从 incoming context 获取（服务端）；
// - 其次从 outgoing context 获取（客户端）。
func GetGRPCMetadata(ctx context.Context) metadata.MD {
	// 先尝试 incoming
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		return md
	}
	// 再尝试 outgoing
	md, _ := metadata.FromOutgoingContext(ctx)
	return md
}