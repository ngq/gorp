// Package gorp provides the root-package application startup surface for gorp framework.
// This file exposes transport contracts, middleware composition helpers, HTTP middleware presets.
// Re-exports stable transport surface without forcing callers into lower-level packages.
//
// Gorp 包提供 gorp 框架的根包层应用启动入口。
// 本文件暴露根包层的 transport 契约、中间件组合 helper 和 HTTP 中间件预设入口。
// 在不强迫调用方下沉到更底层 framework 包的前提下重导出稳定 transport 能力。
package gorp

import (
	"context"

	"github.com/ngq/gorp/framework/application"
	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	"github.com/ngq/gorp/framework/contract/runtime"
	"github.com/ngq/gorp/framework/contract/transport"
	httpmiddleware "github.com/ngq/gorp/framework/http/middleware"
)

// Container is the top-level alias of the runtime container contract.
//
// Container 是 runtime 容器契约的顶层别名。
type Container = runtime.Container

// Router is the top-level alias of the transport router contract.
//
// Router 是 transport 层 Router 契约的顶层别名。
type Router = transport.Router

// Context is the top-level alias of the transport context contract.
//
// Context 是 transport 层 Context 契约的顶层别名。
type Context = transport.Context

// Handler is the top-level alias of the transport handler contract.
//
// Handler 是 transport 层 Handler 契约的顶层别名。
type Handler = transport.Handler

// Middleware is the top-level alias of the transport middleware contract.
//
// Middleware 是 transport 层 Middleware 契约的顶层别名。
type Middleware = transport.Middleware

// MiddlewareFunc is the business-friendly helper signature used by Middleware.
//
// MiddlewareFunc 是 Middleware 使用的业务友好辅助签名。
type MiddlewareFunc func(ctx Context, next Handler)

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

// RecommendedMiddlewareOptions is the top-level alias of the recommended HTTP preset options.
//
// RecommendedMiddlewareOptions 是推荐 HTTP 预设选项的顶层别名。
type RecommendedMiddlewareOptions = httpmiddleware.RecommendedMiddlewareOptions

// InternalMiddlewareOptions is the top-level alias of the internal HTTP preset options.
//
// InternalMiddlewareOptions 是内网 HTTP 预设选项的顶层别名。
type InternalMiddlewareOptions = httpmiddleware.InternalMiddlewareOptions

// AdminMiddlewareOptions is the top-level alias of the admin HTTP preset options.
//
// AdminMiddlewareOptions 是管理 HTTP 预设选项的顶层别名。
type AdminMiddlewareOptions = httpmiddleware.AdminMiddlewareOptions

// DefaultServiceGovernanceOptions is the top-level alias of the HTTP service governance preset options.
//
// DefaultServiceGovernanceOptions 是 HTTP 服务治理预设选项的顶层别名。
type DefaultServiceGovernanceOptions = httpmiddleware.DefaultHTTPServiceGovernanceOptions

// TenantOptions is the top-level alias of tenant middleware options.
//
// TenantOptions 是租户中间件选项的顶层别名。
type TenantOptions = httpmiddleware.TenantOptions

// BodyDumpOptions is the top-level alias of request/response capture middleware options.
//
// BodyDumpOptions 是请求响应抓取中间件选项的顶层别名。
type BodyDumpOptions = httpmiddleware.BodyDumpOptions

// ExchangeCapture is the top-level alias of the request/response capture result.
//
// ExchangeCapture 是请求响应抓取结果的顶层别名。
type ExchangeCapture = httpmiddleware.HTTPExchangeCapture

// Chain combines multiple middlewares into one middleware.
//
// Chain 将多个中间件组合成一个中间件。
//
// Example:
//
//	apiMiddleware := gorp.Chain(
//	    gorp.RequestIdentity(),
//	    gorp.LoggingMiddleware(logger),
//	    gorp.RecoveryMiddleware(),
//	)
//
//	router.Use(apiMiddleware)
func Chain(middleware ...Middleware) Middleware {
	return func(next Handler) Handler {
		for i := len(middleware) - 1; i >= 0; i-- {
			if middleware[i] == nil {
				continue
			}
			next = middleware[i](next)
		}
		return next
	}
}

// AdaptMiddleware adapts a business-friendly middleware callback to Middleware.
//
// AdaptMiddleware 将业务友好的回调签名适配为 Middleware。
//
// Example:
//
//	func AccessLog() gorp.Middleware {
//	    return gorp.AdaptMiddleware(func(ctx gorp.Context, next gorp.Handler) {
//	        if next != nil {
//	            next(ctx)
//	        }
//	    })
//	}
func AdaptMiddleware(fn MiddlewareFunc) Middleware {
	return func(next Handler) Handler {
		return func(ctx Context) {
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

// DefaultRecommendedMiddlewareOptions returns the default public API middleware preset options.
//
// DefaultRecommendedMiddlewareOptions 返回默认的对外 API 中间件预设选项。
func DefaultRecommendedMiddlewareOptions() RecommendedMiddlewareOptions {
	return httpmiddleware.DefaultRecommendedMiddlewareOptions()
}

// DefaultInternalMiddlewareOptions returns the default internal API middleware preset options.
//
// DefaultInternalMiddlewareOptions 返回默认的内网 API 中间件预设选项。
func DefaultInternalMiddlewareOptions() InternalMiddlewareOptions {
	return httpmiddleware.DefaultInternalMiddlewareOptions()
}

// DefaultAdminMiddlewareOptions returns the default admin API middleware preset options.
//
// DefaultAdminMiddlewareOptions 返回默认的管理 API 中间件预设选项。
func DefaultAdminMiddlewareOptions() AdminMiddlewareOptions {
	return httpmiddleware.DefaultAdminMiddlewareOptions()
}

// DefaultServiceGovernanceDefaults returns the default HTTP service governance preset options.
//
// DefaultServiceGovernanceDefaults 返回默认 HTTP 服务治理预设选项。
func DefaultServiceGovernanceDefaults() DefaultServiceGovernanceOptions {
	return httpmiddleware.DefaultHTTPServiceGovernanceDefaults()
}

// DefaultTenantOptions returns the default tenant middleware options.
//
// DefaultTenantOptions 返回默认租户中间件选项。
func DefaultTenantOptions() TenantOptions {
	return httpmiddleware.DefaultTenantOptions()
}

// DefaultMiddleware returns the stable default HTTP middleware baseline.
//
// DefaultMiddleware 返回稳定的默认 HTTP 中间件基线。
func DefaultMiddleware(base observabilitycontract.Logger) Middleware {
	return httpmiddleware.DefaultMiddleware(base)
}

// DefaultMiddlewareSet returns the stable default HTTP middleware baseline as an ordered slice.
//
// DefaultMiddlewareSet 以有序切片形式返回稳定的默认 HTTP 中间件基线。
func DefaultMiddlewareSet(base observabilitycontract.Logger) []Middleware {
	return httpmiddleware.DefaultMiddlewareSet(base)
}

// RecommendedAPIMiddleware returns the recommended public API middleware preset.
//
// RecommendedAPIMiddleware 返回推荐的对外 API 中间件预设。
func RecommendedAPIMiddleware(base observabilitycontract.Logger, opts RecommendedMiddlewareOptions) Middleware {
	return httpmiddleware.RecommendedAPIMiddleware(base, opts)
}

// RecommendedAPIMiddlewareSet returns the recommended public API middleware preset as an ordered slice.
//
// RecommendedAPIMiddlewareSet 以有序切片形式返回推荐的对外 API 中间件预设。
func RecommendedAPIMiddlewareSet(base observabilitycontract.Logger, opts RecommendedMiddlewareOptions) []Middleware {
	return httpmiddleware.RecommendedAPIMiddlewareSet(base, opts)
}

// InternalAPIMiddleware returns the recommended internal API middleware preset.
//
// InternalAPIMiddleware 返回推荐的内网 API 中间件预设。
func InternalAPIMiddleware(base observabilitycontract.Logger, opts InternalMiddlewareOptions) Middleware {
	return httpmiddleware.InternalAPIMiddleware(base, opts)
}

// InternalAPIMiddlewareSet returns the recommended internal API middleware preset as an ordered slice.
//
// InternalAPIMiddlewareSet 以有序切片形式返回推荐的内网 API 中间件预设。
func InternalAPIMiddlewareSet(base observabilitycontract.Logger, opts InternalMiddlewareOptions) []Middleware {
	return httpmiddleware.InternalAPIMiddlewareSet(base, opts)
}

// AdminAPIMiddleware returns the recommended admin API middleware preset.
//
// AdminAPIMiddleware 返回推荐的管理 API 中间件预设。
func AdminAPIMiddleware(base observabilitycontract.Logger, opts AdminMiddlewareOptions) Middleware {
	return httpmiddleware.AdminAPIMiddleware(base, opts)
}

// AdminAPIMiddlewareSet returns the recommended admin API middleware preset as an ordered slice.
//
// AdminAPIMiddlewareSet 以有序切片形式返回推荐的管理 API 中间件预设。
func AdminAPIMiddlewareSet(base observabilitycontract.Logger, opts AdminMiddlewareOptions) []Middleware {
	return httpmiddleware.AdminAPIMiddlewareSet(base, opts)
}

// DefaultServiceGovernancePreset returns the default HTTP service governance middleware preset.
//
// DefaultServiceGovernancePreset 返回默认 HTTP 服务治理中间件预设。
func DefaultServiceGovernancePreset(base observabilitycontract.Logger, opts DefaultServiceGovernanceOptions) Middleware {
	return httpmiddleware.DefaultHTTPServiceGovernancePreset(base, opts)
}

// DefaultServiceGovernanceSet returns the default HTTP service governance preset as an ordered slice.
//
// DefaultServiceGovernanceSet 以有序切片形式返回默认 HTTP 服务治理预设。
func DefaultServiceGovernanceSet(base observabilitycontract.Logger, opts DefaultServiceGovernanceOptions) []Middleware {
	return httpmiddleware.DefaultHTTPServiceGovernanceSet(base, opts)
}

// UseDefaultMiddleware applies the default HTTP middleware baseline to the router.
//
// UseDefaultMiddleware 将默认 HTTP 中间件基线装配到路由器。
func UseDefaultMiddleware(router Router, base observabilitycontract.Logger) {
	httpmiddleware.UseDefaultMiddleware(router, base)
}

// UseRecommendedAPIMiddleware applies the recommended public API middleware preset to the router.
//
// UseRecommendedAPIMiddleware 将推荐的对外 API 中间件预设装配到路由器。
func UseRecommendedAPIMiddleware(router Router, base observabilitycontract.Logger, opts RecommendedMiddlewareOptions) {
	httpmiddleware.UseRecommendedAPIMiddleware(router, base, opts)
}

// UseInternalAPIMiddleware applies the recommended internal API middleware preset to the router.
//
// UseInternalAPIMiddleware 将推荐的内网 API 中间件预设装配到路由器。
func UseInternalAPIMiddleware(router Router, base observabilitycontract.Logger, opts InternalMiddlewareOptions) {
	httpmiddleware.UseInternalAPIMiddleware(router, base, opts)
}

// UseAdminAPIMiddleware applies the recommended admin API middleware preset to the router.
//
// UseAdminAPIMiddleware 将推荐的管理 API 中间件预设装配到路由器。
func UseAdminAPIMiddleware(router Router, base observabilitycontract.Logger, opts AdminMiddlewareOptions) {
	httpmiddleware.UseAdminAPIMiddleware(router, base, opts)
}

// UseDefaultServiceGovernance applies the default HTTP service governance preset to the router.
//
// UseDefaultServiceGovernance 将默认 HTTP 服务治理预设装配到路由器。
func UseDefaultServiceGovernance(router Router, base observabilitycontract.Logger, opts DefaultServiceGovernanceOptions) {
	httpmiddleware.UseDefaultHTTPServiceGovernance(router, base, opts)
}

// Tenant returns the tenant middleware from the mainline HTTP middleware package.
//
// Tenant 返回主线 HTTP 中间件中的租户中间件。
func Tenant(opts TenantOptions) Middleware {
	return httpmiddleware.Tenant(opts)
}

// BodyDump returns the request/response capture middleware from the mainline HTTP middleware package.
//
// BodyDump 返回主线 HTTP 中间件中的请求响应抓取中间件。
func BodyDump(opts BodyDumpOptions) Middleware {
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
