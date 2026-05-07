// Application scenarios:
// - Expose the root-package transport contracts, middleware composition helpers, and HTTP middleware presets.
// - Keep business-facing HTTP/gRPC metadata helpers available from a short public entrypoint.
// - Re-export the stable transport surface without forcing callers into lower-level framework packages.
//
// 适用场景：
// - 暴露根包层的 transport 契约、中间件组合 helper 和 HTTP 中间件预设入口。
// - 通过简短的公共入口提供业务侧常用的 HTTP/gRPC metadata helper。
// - 在不强迫调用方下沉到更底层 framework 包的前提下，重导出稳定 transport 能力。
package gorp

import (
	"context"

	"github.com/ngq/gorp/framework/application"
	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	"github.com/ngq/gorp/framework/contract/runtime"
	"github.com/ngq/gorp/framework/contract/transport"
	httpmiddleware "github.com/ngq/gorp/framework/http/middleware"
)

type Container = runtime.Container

// HTTPRouter is the top-level alias of the transport HTTP router contract.
//
// HTTPRouter 是 transport 层 HTTPRouter 契约的顶层别名。
type HTTPRouter = transport.HTTPRouter

// HTTPContext is the top-level alias of the transport HTTP context contract.
//
// HTTPContext 是 transport 层 HTTPContext 契约的顶层别名。
type HTTPContext = transport.HTTPContext

// HTTPHandler is the top-level alias of the transport HTTP handler contract.
//
// HTTPHandler 是 transport 层 HTTPHandler 契约的顶层别名。
type HTTPHandler = transport.HTTPHandler

// HTTPMiddleware is the top-level alias of the transport HTTP middleware contract.
//
// HTTPMiddleware 是 transport 层 HTTPMiddleware 契约的顶层别名。
type HTTPMiddleware = transport.HTTPMiddleware

// HTTPMiddlewareFunc is the business-friendly helper signature used by Middleware.
//
// HTTPMiddlewareFunc 是 Middleware 使用的业务友好辅助签名。
type HTTPMiddlewareFunc func(ctx HTTPContext, next HTTPHandler)

// Metadata is the top-level alias of the transport metadata contract.
//
// Metadata 是 transport 层 metadata 契约的顶层别名。
type Metadata = transport.Metadata

// GRPCConnFactory is the top-level alias of the proto-first gRPC connection factory.
//
// GRPCConnFactory 是 proto-first gRPC 连接工厂的顶层别名。
type GRPCConnFactory = transport.GRPCConnFactory

// GRPCServerRegistrar is the top-level alias of the proto-first gRPC server registrar.
//
// GRPCServerRegistrar 是 proto-first gRPC 服务注册器的顶层别名。
type GRPCServerRegistrar = transport.GRPCServerRegistrar

// RecommendedHTTPMiddlewareOptions is the top-level alias of the recommended HTTP preset options.
//
// RecommendedHTTPMiddlewareOptions 是推荐 HTTP 预设选项的顶层别名。
type RecommendedHTTPMiddlewareOptions = httpmiddleware.RecommendedMiddlewareOptions

// InternalHTTPMiddlewareOptions is the top-level alias of the internal HTTP preset options.
//
// InternalHTTPMiddlewareOptions 是内网 HTTP 预设选项的顶层别名。
type InternalHTTPMiddlewareOptions = httpmiddleware.InternalMiddlewareOptions

// AdminHTTPMiddlewareOptions is the top-level alias of the admin HTTP preset options.
//
// AdminHTTPMiddlewareOptions 是管理 HTTP 预设选项的顶层别名。
type AdminHTTPMiddlewareOptions = httpmiddleware.AdminMiddlewareOptions
type TenantOptions = httpmiddleware.TenantOptions
type BodyDumpOptions = httpmiddleware.BodyDumpOptions
type HTTPExchangeCapture = httpmiddleware.HTTPExchangeCapture

// Chain combines multiple HTTP middlewares into one middleware.
//
// Chain 将多个 HTTP 中间件组合成一个中间件。
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
//
// Middleware 将业务友好的回调签名适配为 HTTPMiddleware。
//
// Example:
//
//	func AccessLog() gorp.HTTPMiddleware {
//	    return gorp.Middleware(func(ctx gorp.HTTPContext, next gorp.HTTPHandler) {
//	        if next != nil {
//	            next(ctx)
//	        }
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

// DefaultRecommendedHTTPMiddlewareOptions returns the default public API middleware preset options.
//
// DefaultRecommendedHTTPMiddlewareOptions 返回默认的对外 API 中间件预设选项。
func DefaultRecommendedHTTPMiddlewareOptions() RecommendedHTTPMiddlewareOptions {
	return httpmiddleware.DefaultRecommendedMiddlewareOptions()
}

// DefaultInternalHTTPMiddlewareOptions returns the default internal API middleware preset options.
//
// DefaultInternalHTTPMiddlewareOptions 返回默认的内网 API 中间件预设选项。
func DefaultInternalHTTPMiddlewareOptions() InternalHTTPMiddlewareOptions {
	return httpmiddleware.DefaultInternalMiddlewareOptions()
}

// DefaultAdminHTTPMiddlewareOptions returns the default admin API middleware preset options.
//
// DefaultAdminHTTPMiddlewareOptions 返回默认的管理 API 中间件预设选项。
func DefaultAdminHTTPMiddlewareOptions() AdminHTTPMiddlewareOptions {
	return httpmiddleware.DefaultAdminMiddlewareOptions()
}

func DefaultTenantOptions() TenantOptions {
	return httpmiddleware.DefaultTenantOptions()
}

// DefaultHTTPMiddleware returns the stable default HTTP middleware baseline.
//
// DefaultHTTPMiddleware 返回稳定的默认 HTTP 中间件基线。
func DefaultHTTPMiddleware(base observabilitycontract.Logger) HTTPMiddleware {
	return httpmiddleware.DefaultMiddleware(base)
}

// DefaultHTTPMiddlewareSet returns the stable default HTTP middleware baseline as an ordered slice.
//
// DefaultHTTPMiddlewareSet 以有序切片形式返回稳定的默认 HTTP 中间件基线。
func DefaultHTTPMiddlewareSet(base observabilitycontract.Logger) []HTTPMiddleware {
	return httpmiddleware.DefaultMiddlewareSet(base)
}

// RecommendedHTTPAPIMiddleware returns the recommended public API middleware preset.
//
// RecommendedHTTPAPIMiddleware 返回推荐的对外 API 中间件预设。
func RecommendedHTTPAPIMiddleware(base observabilitycontract.Logger, opts RecommendedHTTPMiddlewareOptions) HTTPMiddleware {
	return httpmiddleware.RecommendedAPIMiddleware(base, opts)
}

// RecommendedHTTPAPIMiddlewareSet returns the recommended public API middleware preset as an ordered slice.
//
// RecommendedHTTPAPIMiddlewareSet 以有序切片形式返回推荐的对外 API 中间件预设。
func RecommendedHTTPAPIMiddlewareSet(base observabilitycontract.Logger, opts RecommendedHTTPMiddlewareOptions) []HTTPMiddleware {
	return httpmiddleware.RecommendedAPIMiddlewareSet(base, opts)
}

// InternalHTTPAPIMiddleware returns the recommended internal API middleware preset.
//
// InternalHTTPAPIMiddleware 返回推荐的内网 API 中间件预设。
func InternalHTTPAPIMiddleware(base observabilitycontract.Logger, opts InternalHTTPMiddlewareOptions) HTTPMiddleware {
	return httpmiddleware.InternalAPIMiddleware(base, opts)
}

// InternalHTTPAPIMiddlewareSet returns the recommended internal API middleware preset as an ordered slice.
//
// InternalHTTPAPIMiddlewareSet 以有序切片形式返回推荐的内网 API 中间件预设。
func InternalHTTPAPIMiddlewareSet(base observabilitycontract.Logger, opts InternalHTTPMiddlewareOptions) []HTTPMiddleware {
	return httpmiddleware.InternalAPIMiddlewareSet(base, opts)
}

// AdminHTTPAPIMiddleware returns the recommended admin API middleware preset.
//
// AdminHTTPAPIMiddleware 返回推荐的管理 API 中间件预设。
func AdminHTTPAPIMiddleware(base observabilitycontract.Logger, opts AdminHTTPMiddlewareOptions) HTTPMiddleware {
	return httpmiddleware.AdminAPIMiddleware(base, opts)
}

// AdminHTTPAPIMiddlewareSet returns the recommended admin API middleware preset as an ordered slice.
//
// AdminHTTPAPIMiddlewareSet 以有序切片形式返回推荐的管理 API 中间件预设。
func AdminHTTPAPIMiddlewareSet(base observabilitycontract.Logger, opts AdminHTTPMiddlewareOptions) []HTTPMiddleware {
	return httpmiddleware.AdminAPIMiddlewareSet(base, opts)
}

// UseDefaultHTTPMiddleware applies the default HTTP middleware baseline to the router.
//
// UseDefaultHTTPMiddleware 将默认 HTTP 中间件基线装配到路由器。
func UseDefaultHTTPMiddleware(router HTTPRouter, base observabilitycontract.Logger) {
	httpmiddleware.UseDefaultMiddleware(router, base)
}

// UseRecommendedHTTPAPIMiddleware applies the recommended public API middleware preset to the router.
//
// UseRecommendedHTTPAPIMiddleware 将推荐的对外 API 中间件预设装配到路由器。
func UseRecommendedHTTPAPIMiddleware(router HTTPRouter, base observabilitycontract.Logger, opts RecommendedHTTPMiddlewareOptions) {
	httpmiddleware.UseRecommendedAPIMiddleware(router, base, opts)
}

// UseInternalHTTPAPIMiddleware applies the recommended internal API middleware preset to the router.
//
// UseInternalHTTPAPIMiddleware 将推荐的内网 API 中间件预设装配到路由器。
func UseInternalHTTPAPIMiddleware(router HTTPRouter, base observabilitycontract.Logger, opts InternalHTTPMiddlewareOptions) {
	httpmiddleware.UseInternalAPIMiddleware(router, base, opts)
}

// UseAdminHTTPAPIMiddleware applies the recommended admin API middleware preset to the router.
//
// UseAdminHTTPAPIMiddleware 将推荐的管理 API 中间件预设装配到路由器。
func UseAdminHTTPAPIMiddleware(router HTTPRouter, base observabilitycontract.Logger, opts AdminHTTPMiddlewareOptions) {
	httpmiddleware.UseAdminAPIMiddleware(router, base, opts)
}

func Tenant(opts TenantOptions) HTTPMiddleware {
	return httpmiddleware.Tenant(opts)
}

func BodyDump(opts BodyDumpOptions) HTTPMiddleware {
	return httpmiddleware.BodyDump(opts)
}

// MakeGRPCConnFactory returns the proto-first gRPC connection factory from the container.
//
// MakeGRPCConnFactory 从容器中返回 proto-first gRPC 连接工厂。
func MakeGRPCConnFactory(c Container) (GRPCConnFactory, error) {
	return application.MakeGRPCConnFactory(c)
}

// MakeGRPCServerRegistrar returns the proto-first gRPC server registrar from the container.
//
// MakeGRPCServerRegistrar 从容器中返回 proto-first gRPC 服务注册器。
func MakeGRPCServerRegistrar(c Container) (GRPCServerRegistrar, error) {
	return application.MakeGRPCServerRegistrar(c)
}

// NewMetadata creates a transport metadata container.
//
// NewMetadata 创建一个 transport metadata 容器。
func NewMetadata() Metadata {
	return transport.NewMetadata()
}

// NewServerContext attaches metadata to a server-side context.
//
// NewServerContext 向服务端 context 绑定 metadata。
func NewServerContext(ctx context.Context, md Metadata) context.Context {
	return transport.NewServerContext(ctx, md)
}

// FromServerContext reads transport metadata from a server-side context.
//
// FromServerContext 从服务端 context 读取 transport metadata。
func FromServerContext(ctx context.Context) (Metadata, bool) {
	return transport.FromServerContext(ctx)
}

// NewClientContext attaches metadata to a client-side context.
//
// NewClientContext 向客户端 context 绑定 metadata。
func NewClientContext(ctx context.Context, md Metadata) context.Context {
	return transport.NewClientContext(ctx, md)
}

// FromClientContext reads transport metadata from a client-side context.
//
// FromClientContext 从客户端 context 读取 transport metadata。
func FromClientContext(ctx context.Context) (Metadata, bool) {
	return transport.FromClientContext(ctx)
}

// AppendToClientContext appends key-value pairs to client transport metadata.
//
// AppendToClientContext 向客户端 transport metadata 追加键值对。
func AppendToClientContext(ctx context.Context, kv ...string) context.Context {
	return transport.AppendToClientContext(ctx, kv...)
}

// GetGRPCTraceID reads the trace id from a gRPC context.
//
// GetGRPCTraceID 从 gRPC context 读取 trace id。
func GetGRPCTraceID(ctx context.Context) string {
	return application.GetGRPCTraceID(ctx)
}

// GetGRPCRequestID reads the request id from a gRPC context.
//
// GetGRPCRequestID 从 gRPC context 读取 request id。
func GetGRPCRequestID(ctx context.Context) string {
	return application.GetGRPCRequestID(ctx)
}
