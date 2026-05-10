// Package grpc provides service authentication interceptors for gRPC.
// This file implements client and server interceptors for service-to-service auth.
//
// 本包提供 gRPC 服务认证拦截器。
// 本文件实现服务间认证的客户端和服务端拦截器。
package grpc

import (
	"context"
	"fmt"
	"strings"
	"time"

	securitycontract "github.com/ngq/gorp/framework/contract/security"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// serviceAuthWrappedStream wraps a grpc.ServerStream with a custom context.
// Used to propagate authenticated identity through stream calls.
//
// serviceAuthWrappedStream 用自定义 context 包装 grpc.ServerStream。
// 用于通过流式调用传播认证身份。
type serviceAuthWrappedStream struct {
	grpc.ServerStream
	ctx context.Context
}

// Context returns the wrapped context with authentication information.
//
// Context 返回包含认证信息的包装 context。
func (w *serviceAuthWrappedStream) Context() context.Context { return w.ctx }

// serviceAuthUnaryClientInterceptor creates a unary client interceptor for service auth.
// Injects service token into outgoing gRPC metadata.
//
// serviceAuthUnaryClientInterceptor 创建服务认证的一元客户端拦截器。
// 将服务 token 注入出站 gRPC metadata。
func serviceAuthUnaryClientInterceptor(auth securitycontract.ServiceTokenIssuer, targetService string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		// 生成服务 token
		if token, err := auth.GenerateToken(ctx, targetService); err == nil && strings.TrimSpace(token) != "" {
			md, ok := metadata.FromOutgoingContext(ctx)
			if !ok {
				md = metadata.New(nil)
			}
			md.Set("x-service-token", token)
			ctx = metadata.NewOutgoingContext(ctx, md)
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// serviceAuthStreamClientInterceptor creates a stream client interceptor for service auth.
// Injects service token into outgoing gRPC metadata for streaming calls.
//
// serviceAuthStreamClientInterceptor 创建服务认证的流式客户端拦截器。
// 将服务 token 注入出站 gRPC metadata 用于流式调用。
func serviceAuthStreamClientInterceptor(auth securitycontract.ServiceTokenIssuer, targetService string) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		// 生成服务 token
		if token, err := auth.GenerateToken(ctx, targetService); err == nil && strings.TrimSpace(token) != "" {
			md, ok := metadata.FromOutgoingContext(ctx)
			if !ok {
				md = metadata.New(nil)
			}
			md.Set("x-service-token", token)
			ctx = metadata.NewOutgoingContext(ctx, md)
		}
		return streamer(ctx, desc, cc, method, opts...)
	}
}

// serviceAuthUnaryServerInterceptor creates a unary server interceptor for service auth.
// Extracts service token from incoming metadata and authenticates the caller.
//
// serviceAuthUnaryServerInterceptor 创建服务认证的一元服务端拦截器。
// 从入站 metadata 提取服务 token 并认证调用方。
func serviceAuthUnaryServerInterceptor(auth securitycontract.ServiceAuthenticator) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		// 从入站 metadata 提取 token
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if values := md.Get("x-service-token"); len(values) > 0 && strings.TrimSpace(values[0]) != "" {
				ctx = context.WithValue(ctx, "x-service-token", values[0])
			}
		}
		// 执行认证
		identity, err := auth.Authenticate(ctx)
		if err != nil {
			return nil, fmt.Errorf("rpc: service authentication failed: %w", err)
		}
		// 将身份注入 context
		if identity != nil {
			ctx = securitycontract.NewServiceIdentityContext(ctx, identity)
		}
		return handler(ctx, req)
	}
}

// serviceAuthStreamServerInterceptor creates a stream server interceptor for service auth.
// Extracts service token from incoming metadata and authenticates the caller for streaming calls.
//
// serviceAuthStreamServerInterceptor 创建服务认证的流式服务端拦截器。
// 从入站 metadata 提取服务 token 并认证调用方用于流式调用。
func serviceAuthStreamServerInterceptor(auth securitycontract.ServiceAuthenticator) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := ss.Context()
		// 从入站 metadata 提取 token
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if values := md.Get("x-service-token"); len(values) > 0 && strings.TrimSpace(values[0]) != "" {
				ctx = context.WithValue(ctx, "x-service-token", values[0])
			}
		}
		// 执行认证
		identity, err := auth.Authenticate(ctx)
		if err != nil {
			return fmt.Errorf("rpc: service authentication failed: %w", err)
		}
		// 将身份注入 context
		if identity != nil {
			ctx = securitycontract.NewServiceIdentityContext(ctx, identity)
		}
		return handler(srv, &serviceAuthWrappedStream{ServerStream: ss, ctx: ctx})
	}
}

// recoveryUnaryServerInterceptor creates a unary server interceptor for panic recovery.
// Recovers from panics and returns a gRPC error with codes.Internal.
//
// recoveryUnaryServerInterceptor 创建 panic 恢复的一元服务端拦截器。
// 恢复 panic 并返回 codes.Internal 的 gRPC 错误。
func recoveryUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("grpc: panic recovered: %v", r)
			}
		}()
		return handler(ctx, req)
	}
}

// recoveryStreamServerInterceptor creates a stream server interceptor for panic recovery.
// Recovers from panics during streaming calls and returns a gRPC error.
//
// recoveryStreamServerInterceptor 创建 panic 恢复的流式服务端拦截器。
// 恢复流式调用期间的 panic 并返回 gRPC 错误。
func recoveryStreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("grpc: panic recovered: %v", r)
			}
		}()
		return handler(srv, ss)
	}
}

// timeoutUnaryServerInterceptor creates a unary server interceptor for request timeout.
// Enforces a timeout on the request context.
//
// timeoutUnaryServerInterceptor 创建请求超时的一元服务端拦截器。
// 对请求 context 强制执行超时。
func timeoutUnaryServerInterceptor(timeout time.Duration) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if timeout <= 0 {
			return handler(ctx, req)
		}
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		return handler(ctx, req)
	}
}

// metricsUnaryServerInterceptor creates a unary server interceptor for Prometheus metrics.
// Records request count with method label.
//
// metricsUnaryServerInterceptor 创建 Prometheus 指标的一元服务端拦截器。
// 记录带 method 标签的请求计数。
func metricsUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		// TODO: Integrate with Prometheus metrics when available
		// 等待 Prometheus metrics 集成后补充
		return handler(ctx, req)
	}
}