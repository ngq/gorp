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

import (
	"context"
	"crypto/tls"
)

// authorizationContextKey is the typed context key for the raw Authorization header value.
//
// authorizationContextKey 是原始 Authorization 头值的类型化 context key。
type authorizationContextKey struct{}

// serviceTokenContextKey is the typed context key for the raw X-Service-Token header value.
//
// serviceTokenContextKey 是原始 X-Service-Token 头值的类型化 context key。
type serviceTokenContextKey struct{}

// tlsStateContextKey is the typed context key for the peer TLS connection state.
// Used by mTLS service auth to obtain the verified client certificate.
//
// tlsStateContextKey 是 TLS 连接状态的类型化 context key。
// 供 mTLS 服务认证获取已验证的客户端证书。
type tlsStateContextKey struct{}

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

// WithTLSState stores the peer TLS connection state into the context.
// The transport layer (HTTP/gin or gRPC) injects this so mTLS authenticators
// can read the verified client certificate.
//
// WithTLSState 将对端 TLS 连接状态写入 context。
// 传输层（HTTP/gin 或 gRPC）注入该值，供 mTLS 认证器读取已验证的客户端证书。
func WithTLSState(ctx context.Context, state *tls.ConnectionState) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, tlsStateContextKey{}, state)
}

// TLSStateFrom retrieves the peer TLS connection state from the context.
// Returns nil when the connection is not TLS or the state was not injected.
//
// TLSStateFrom 从 context 中读取对端 TLS 连接状态。
// 连接非 TLS 或未注入时返回 nil。
func TLSStateFrom(ctx context.Context) *tls.ConnectionState {
	if ctx == nil {
		return nil
	}
	state, _ := ctx.Value(tlsStateContextKey{}).(*tls.ConnectionState)
	return state
}
