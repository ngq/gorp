package token

import (
	"context"
	"net/http"
	"strings"

	securitycontract "github.com/ngq/gorp/framework/contract/security"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// ServiceAuthMiddleware 把服务间认证相关 header 提取到 context，并在可用时执行强校验。
//
// 中文说明：
// - 负责把 Authorization / X-Service-Token 写入 request context；
// - 如果容器中存在 ServiceAuthenticator，则直接执行认证；
// - 认证失败时立刻返回 401，避免请求进入业务处理层。
func ServiceAuthMiddleware(authenticator securitycontract.ServiceAuthenticator) transportcontract.Middleware {
	return func(next transportcontract.Handler) transportcontract.Handler {
		return func(c transportcontract.Context) {
			auth := strings.TrimSpace(c.GetHeader("Authorization"))
			token := strings.TrimSpace(c.GetHeader("X-Service-Token"))

			if authenticator != nil && (auth != "" || token != "") {
				// Build context for authentication using standard context.Context
				authCtx := context.Context(c)
				if auth != "" {
					authCtx = context.WithValue(authCtx, "authorization", auth)
				}
				if token != "" {
					authCtx = context.WithValue(authCtx, "x-service-token", token)
				}

				identity, err := authenticator.Authenticate(authCtx)
				if err != nil {
					c.JSON(http.StatusUnauthorized, map[string]any{"error": "service authentication failed"})
					return
				}
				if identity != nil {
					c.Set("service_identity", identity)
				}
			} else {
				// Store headers in context for later use
				if auth != "" {
					c.Set("authorization", auth)
				}
				if token != "" {
					c.Set("x-service-token", token)
				}
			}

			if next != nil {
				next(c)
			}
		}
	}
}

// ServiceAuthHeaderInjector 把服务间认证 token 注入 HTTP 请求头。
//
// 中文说明：
// - 用于客户端发起 HTTP RPC 请求时；
// - 统一把生成出的 service token 写入 `X-Service-Token`；
// - 如果 token 为空，则不做注入。
func ServiceAuthHeaderInjector(req *http.Request, token string) {
	if req == nil || strings.TrimSpace(token) == "" {
		return
	}
	req.Header.Set("X-Service-Token", token)
}

// UnaryServerInterceptor 把 gRPC metadata 中的服务认证信息提取到 context，并在可用时执行强校验。
func UnaryServerInterceptor(authenticator securitycontract.ServiceAuthenticator) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if values := md.Get("authorization"); len(values) > 0 && strings.TrimSpace(values[0]) != "" {
				ctx = context.WithValue(ctx, "authorization", values[0])
			}
			if values := md.Get("x-service-token"); len(values) > 0 && strings.TrimSpace(values[0]) != "" {
				ctx = context.WithValue(ctx, "x-service-token", values[0])
			}
		}

		if authenticator != nil {
			identity, err := authenticator.Authenticate(ctx)
			if err != nil {
				return nil, status.Error(codes.Unauthenticated, "service authentication failed")
			}
			if identity != nil {
				ctx = securitycontract.NewServiceIdentityContext(ctx, identity)
			}
		}

		return handler(ctx, req)
	}
}

// StreamServerInterceptor 把 gRPC 流式请求中的服务认证信息提取到 context，并在可用时执行强校验。
func StreamServerInterceptor(authenticator securitycontract.ServiceAuthenticator) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := ss.Context()
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if values := md.Get("authorization"); len(values) > 0 && strings.TrimSpace(values[0]) != "" {
				ctx = context.WithValue(ctx, "authorization", values[0])
			}
			if values := md.Get("x-service-token"); len(values) > 0 && strings.TrimSpace(values[0]) != "" {
				ctx = context.WithValue(ctx, "x-service-token", values[0])
			}
		}

		if authenticator != nil {
			identity, err := authenticator.Authenticate(ctx)
			if err != nil {
				return status.Error(codes.Unauthenticated, "service authentication failed")
			}
			if identity != nil {
				ctx = securitycontract.NewServiceIdentityContext(ctx, identity)
			}
		}

		wrapped := &serviceAuthWrappedServerStream{ServerStream: ss, ctx: ctx}
		return handler(srv, wrapped)
	}
}

// UnaryClientInterceptor 在 gRPC 客户端调用前注入服务认证 token。
//
// 中文说明：
// - 基于目标服务名生成服务间认证 token；
// - 把 token 写入 outgoing metadata 的 `x-service-token`；
// - 这样 Proto-first 客户端无需手工拼接认证头。
func UnaryClientInterceptor(tokenIssuer securitycontract.ServiceTokenIssuer, targetService string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if tokenIssuer != nil {
			if token, err := tokenIssuer.GenerateToken(ctx, targetService); err == nil && strings.TrimSpace(token) != "" {
				md, ok := metadata.FromOutgoingContext(ctx)
				if !ok {
					md = metadata.New(nil)
				}
				md.Set("x-service-token", token)
				ctx = metadata.NewOutgoingContext(ctx, md)
			}
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// StreamClientInterceptor 在 gRPC 流式客户端调用前注入服务认证 token。
func StreamClientInterceptor(tokenIssuer securitycontract.ServiceTokenIssuer, targetService string) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		if tokenIssuer != nil {
			if token, err := tokenIssuer.GenerateToken(ctx, targetService); err == nil && strings.TrimSpace(token) != "" {
				md, ok := metadata.FromOutgoingContext(ctx)
				if !ok {
					md = metadata.New(nil)
				}
				md.Set("x-service-token", token)
				ctx = metadata.NewOutgoingContext(ctx, md)
			}
		}
		return streamer(ctx, desc, cc, method, opts...)
	}
}

type serviceAuthWrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *serviceAuthWrappedServerStream) Context() context.Context {
	return w.ctx
}
