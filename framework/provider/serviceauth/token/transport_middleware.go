package token

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ngq/gorp/framework/contract"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// ServiceAuthHTTPMiddleware 把服务间认证相关 header 提取到 context，并在可用时执行强校验。
//
// 中文说明：
// - 负责把 Authorization / X-Service-Token 写入 request context；
// - 如果容器中存在 ServiceAuthenticator，则直接执行认证；
// - 认证失败时立刻返回 401，避免请求进入业务处理层。
func ServiceAuthHTTPMiddleware(authenticator contract.ServiceAuthenticator) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		if auth := strings.TrimSpace(c.GetHeader("Authorization")); auth != "" {
			ctx = context.WithValue(ctx, "authorization", auth)
		}
		if token := strings.TrimSpace(c.GetHeader("X-Service-Token")); token != "" {
			ctx = context.WithValue(ctx, "x-service-token", token)
		}

		if authenticator != nil {
			identity, err := authenticator.Authenticate(ctx)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "service authentication failed"})
				return
			}
			if identity != nil {
				ctx = context.WithValue(ctx, contract.ServiceIdentityKey, identity)
			}
		}

		c.Request = c.Request.WithContext(ctx)
		c.Next()
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
func UnaryServerInterceptor(authenticator contract.ServiceAuthenticator) grpc.UnaryServerInterceptor {
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
				ctx = context.WithValue(ctx, contract.ServiceIdentityKey, identity)
			}
		}

		return handler(ctx, req)
	}
}

// StreamServerInterceptor 把 gRPC 流式请求中的服务认证信息提取到 context，并在可用时执行强校验。
func StreamServerInterceptor(authenticator contract.ServiceAuthenticator) grpc.StreamServerInterceptor {
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
				ctx = context.WithValue(ctx, contract.ServiceIdentityKey, identity)
			}
		}

		wrapped := &serviceAuthWrappedServerStream{ServerStream: ss, ctx: ctx}
		return handler(srv, wrapped)
	}
}

type serviceAuthWrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *serviceAuthWrappedServerStream) Context() context.Context {
	return w.ctx
}
