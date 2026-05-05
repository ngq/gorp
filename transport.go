package gorp

import (
	"context"

	"github.com/ngq/gorp/framework/contract/runtime"
	"github.com/ngq/gorp/framework/contract/transport"
	"github.com/ngq/gorp/framework/application"
)

type Container = runtime.Container

// HTTPRouter is the top-level alias of the transport HTTP router contract.
// HTTPRouter 是 transport 层 HTTP 路由契约的顶层别名。
type HTTPRouter = transport.HTTPRouter

// HTTPContext is the top-level alias of the transport HTTP context contract.
// HTTPContext 是 transport 层 HTTP 上下文契约的顶层别名。
type HTTPContext = transport.HTTPContext

// HTTPHandler is the top-level alias of the transport HTTP handler contract.
// HTTPHandler 是 transport 层 HTTP 处理器契约的顶层别名。
type HTTPHandler = transport.HTTPHandler

// HTTPMiddleware is the top-level alias of the transport HTTP middleware contract.
// HTTPMiddleware 是 transport 层 HTTP 中间件契约的顶层别名。
type HTTPMiddleware = transport.HTTPMiddleware

// HTTPMiddlewareFunc is the business-friendly helper signature used by Middleware.
// HTTPMiddlewareFunc 是 Middleware 使用的业务友好 helper 签名。
type HTTPMiddlewareFunc func(ctx HTTPContext, next HTTPHandler)

// Metadata is the top-level alias of the transport metadata contract.
// Metadata 是 transport 层元数据契约的顶层别名。
type Metadata = transport.Metadata

// GRPCConnFactory is the top-level alias of the proto-first gRPC connection factory.
// GRPCConnFactory 是 Proto-first gRPC 连接工厂的顶层别名。
type GRPCConnFactory = transport.GRPCConnFactory

// GRPCServerRegistrar is the top-level alias of the proto-first gRPC server registrar.
// GRPCServerRegistrar 是 Proto-first gRPC 服务端注册器的顶层别名。
type GRPCServerRegistrar = transport.GRPCServerRegistrar

// Chain combines multiple HTTP middlewares into one middleware.
// Chain 把多个 HTTP 中间件预组合成一个中间件。
//
// Example:
//
//	apiMiddleware := gorp.Chain(
//	    RequestIdentity(),
//	    LoggingMiddleware(logger),
//	    RecoveryMiddleware(),
//	)
//
//	router.Use(apiMiddleware)
func Chain(middleware ...HTTPMiddleware) HTTPMiddleware {
	return func(next HTTPHandler) HTTPHandler {
		for i := len(middleware) - 1; i >= 0; i-- {
			if middleware[i] == nil {
				continue
			}
			next = middleware[i](next)
		}
		return next
	}
}

// Middleware adapts a business-friendly middleware callback to HTTPMiddleware.
// Middleware 把业务友好的中间件回调适配成 HTTPMiddleware。
//
// Example:
//
//	func AccessLog() gorp.HTTPMiddleware {
//	    return gorp.Middleware(func(ctx gorp.HTTPContext, next gorp.HTTPHandler) {
//	        start := time.Now()
//	        if next != nil {
//	            next(ctx)
//	        }
//	        _ = start
//	    })
//	}
func Middleware(fn HTTPMiddlewareFunc) HTTPMiddleware {
	return func(next HTTPHandler) HTTPHandler {
		return func(ctx HTTPContext) {
			if fn == nil {
				if next != nil {
					next(ctx)
				}
				return
			}
			fn(ctx, next)
		}
	}
}

// MakeGRPCConnFactory returns the proto-first gRPC connection factory from the container.
// MakeGRPCConnFactory 从容器中获取 Proto-first gRPC 连接工厂。
func MakeGRPCConnFactory(c Container) (GRPCConnFactory, error) {
	return application.MakeGRPCConnFactory(c)
}

// MakeGRPCServerRegistrar returns the proto-first gRPC server registrar from the container.
// MakeGRPCServerRegistrar 从容器中获取 Proto-first gRPC 服务端注册器。
func MakeGRPCServerRegistrar(c Container) (GRPCServerRegistrar, error) {
	return application.MakeGRPCServerRegistrar(c)
}

// NewMetadata creates a transport metadata container.
// NewMetadata 创建 transport 元数据容器。
func NewMetadata() Metadata {
	return transport.NewMetadata()
}

// NewServerContext attaches metadata to a server-side context.
// NewServerContext 把元数据写入服务端 context。
func NewServerContext(ctx context.Context, md Metadata) context.Context {
	return transport.NewServerContext(ctx, md)
}

// FromServerContext reads transport metadata from a server-side context.
// FromServerContext 从服务端 context 读取 transport 元数据。
func FromServerContext(ctx context.Context) (Metadata, bool) {
	return transport.FromServerContext(ctx)
}

// NewClientContext attaches metadata to a client-side context.
// NewClientContext 把元数据写入客户端 context。
func NewClientContext(ctx context.Context, md Metadata) context.Context {
	return transport.NewClientContext(ctx, md)
}

// FromClientContext reads transport metadata from a client-side context.
// FromClientContext 从客户端 context 读取 transport 元数据。
func FromClientContext(ctx context.Context) (Metadata, bool) {
	return transport.FromClientContext(ctx)
}

// AppendToClientContext appends key-value pairs to client transport metadata.
// AppendToClientContext 向客户端 transport 元数据追加键值对。
func AppendToClientContext(ctx context.Context, kv ...string) context.Context {
	return transport.AppendToClientContext(ctx, kv...)
}

// GetGRPCTraceID reads the trace id from a gRPC context.
// GetGRPCTraceID 从 gRPC context 读取 trace id。
func GetGRPCTraceID(ctx context.Context) string {
	return application.GetGRPCTraceID(ctx)
}

// GetGRPCRequestID reads the request id from a gRPC context.
// GetGRPCRequestID 从 gRPC context 读取 request id。
func GetGRPCRequestID(ctx context.Context) string {
	return application.GetGRPCRequestID(ctx)
}
