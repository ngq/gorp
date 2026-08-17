// Application scenarios:
// - Propagate service-auth credentials (Authorization header, X-Service-Token) through request contexts.
// - Give service-auth middleware, interceptors, and authenticators one consistent, collision-free access path.
// - Replace fragile string context keys with typed keys so cross-package contracts are compile-checked.
//
// 适用场景：
// - 在请求 context 中透传服务认证凭据（Authorization 头、X-Service-Token）。
// - 让 service-auth middleware、interceptor 和 authenticator 共享统一的、无冲突的访问路径。
// - 用类型化 key 取代脆弱的字符串 context key，让跨包契约受编译期约束。
package security

import "context"

// authorizationContextKey is the typed context key for the raw Authorization header value.
//
// authorizationContextKey 是原始 Authorization 头值的类型化 context key。
type authorizationContextKey struct{}

// serviceTokenContextKey is the typed context key for the raw X-Service-Token header value.
//
// serviceTokenContextKey 是原始 X-Service-Token 头值的类型化 context key。
type serviceTokenContextKey struct{}

// WithAuthorization stores the raw Authorization header value into the context.
//
// WithAuthorization 将原始 Authorization 头值写入 context。
func WithAuthorization(ctx context.Context, value string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, authorizationContextKey{}, value)
}

// AuthorizationFrom retrieves the raw Authorization header value from the context.
//
// AuthorizationFrom 从 context 中读取原始 Authorization 头值。
func AuthorizationFrom(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	v, _ := ctx.Value(authorizationContextKey{}).(string)
	return v
}

// WithServiceToken stores the raw X-Service-Token header value into the context.
//
// WithServiceToken 将原始 X-Service-Token 头值写入 context。
func WithServiceToken(ctx context.Context, value string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, serviceTokenContextKey{}, value)
}

// ServiceTokenFrom retrieves the raw X-Service-Token header value from the context.
//
// ServiceTokenFrom 从 context 中读取原始 X-Service-Token 头值。
func ServiceTokenFrom(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	v, _ := ctx.Value(serviceTokenContextKey{}).(string)
	return v
}
