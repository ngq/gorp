// Package middleware provides gRPC metadata propagation middleware.
// Implements MetadataCarrier interface for gRPC metadata.
// Supports unary and stream interceptors for both client and server.
//
// 中间件包提供 gRPC 元数据传播中间件。
// 为 gRPC metadata 实现 MetadataCarrier 接口。
// 支持客户端和服务端的一元和流拦截器。
package middleware

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

// GRPCCarrier wraps gRPC metadata.MD to implement MetadataCarrier interface.
// Core logic: Delegate Get/Set/Add operations to underlying metadata.MD.
//
// GRPCCarrier 包装 gRPC metadata.MD，实现 MetadataCarrier 接口。
// 核心逻辑：将 Get/Set/Add 操作委托给底层 metadata.MD。
type GRPCCarrier struct {
	md metadata.MD
}

// NewGRPCCarrier creates a new GRPCCarrier with optional metadata.
// Core logic: Initialize metadata.MD if nil, wrap in carrier.
//
// NewGRPCCarrier 创建新的 GRPCCarrier（可选 metadata）。
// 核心逻辑：若 metadata 为 nil 则初始化、包装为 carrier。
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

// UnaryServerInterceptor creates gRPC unary server interceptor for metadata extraction.
// Core logic: Extract metadata from incoming context, inject into handler context.
//
// UnaryServerInterceptor 创建 gRPC 一元服务端拦截器，用于提取元数据。
// 核心逻辑：从 incoming context 提取元数据、注入到 handler context。
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

// StreamServerInterceptor creates gRPC stream server interceptor for metadata extraction.
// Core logic: Extract metadata, wrap stream with updated context.
//
// StreamServerInterceptor 创建 gRPC 流服务端拦截器，用于提取元数据。
// 核心逻辑：提取元数据、用更新后的 context 包装 stream。
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

// UnaryClientInterceptor creates gRPC unary client interceptor for metadata injection.
// Core logic: Inject metadata into outgoing context before invoking.
//
// UnaryClientInterceptor 创建 gRPC 一元客户端拦截器，用于注入元数据。
// 核心逻辑：在调用前将元数据注入到 outgoing context。
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

// StreamClientInterceptor creates gRPC stream client interceptor for metadata injection.
// Core logic: Inject metadata into outgoing context before streaming.
//
// StreamClientInterceptor 创建 gRPC 流客户端拦截器，用于注入元数据。
// 核心逻辑：在流调用前将元数据注入到 outgoing context。
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

// GetGRPCMetadata retrieves metadata from gRPC context (incoming or outgoing).
// Core logic: Try incoming context first, fallback to outgoing.
//
// GetGRPCMetadata 从 gRPC context 获取元数据（incoming 或 outgoing）。
// 核心逻辑：先尝试 incoming context，回退到 outgoing。
func GetGRPCMetadata(ctx context.Context) metadata.MD {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		return md
	}
	md, _ := metadata.FromOutgoingContext(ctx)
	return md
}
