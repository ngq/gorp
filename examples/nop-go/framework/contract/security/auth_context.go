// Application scenarios:
// - Propagate JWT-derived identity information through request-scoped contexts.
// - Let middleware, handlers, and services share one consistent access path for claims and subject identity.
// - Keep security-related context keys private to avoid cross-package collisions.
//
// 适用场景：
// - 在请求级 context 中透传由 JWT 解析出的身份信息。
// - 让 middleware、handler 和 service 共享统一的 claims 与主体身份读取路径。
// - 将安全相关 context key 保持为私有，避免跨包冲突。
package security

import "context"

type jwtClaimsContextKey struct{}
type subjectIDContextKey struct{}
type subjectTypeContextKey struct{}

// NewJWTClaimsContext writes JWT claims into the context.
//
// NewJWTClaimsContext 将 JWT claims 写入 context。
func NewJWTClaimsContext(ctx context.Context, claims *JWTClaims) context.Context {
	return context.WithValue(ctx, jwtClaimsContextKey{}, claims)
}

// FromJWTClaimsContext reads JWT claims from the context.
//
// FromJWTClaimsContext 从 context 中读取 JWT claims。
func FromJWTClaimsContext(ctx context.Context) (*JWTClaims, bool) {
	if ctx == nil {
		return nil, false
	}
	claims, ok := ctx.Value(jwtClaimsContextKey{}).(*JWTClaims)
	return claims, ok && claims != nil
}

// NewSubjectIDContext writes the authenticated subject ID into the context.
//
// NewSubjectIDContext 将认证主体 ID 写入 context。
func NewSubjectIDContext(ctx context.Context, subjectID int64) context.Context {
	return context.WithValue(ctx, subjectIDContextKey{}, subjectID)
}

// FromSubjectIDContext reads the authenticated subject ID from the context.
//
// FromSubjectIDContext 从 context 中读取认证主体 ID。
func FromSubjectIDContext(ctx context.Context) (int64, bool) {
	if ctx == nil {
		return 0, false
	}
	subjectID, ok := ctx.Value(subjectIDContextKey{}).(int64)
	return subjectID, ok
}

// NewSubjectTypeContext writes the authenticated subject type into the context.
//
// NewSubjectTypeContext 将认证主体类型写入 context。
func NewSubjectTypeContext(ctx context.Context, subjectType string) context.Context {
	return context.WithValue(ctx, subjectTypeContextKey{}, subjectType)
}

// FromSubjectTypeContext reads the authenticated subject type from the context.
//
// FromSubjectTypeContext 从 context 中读取认证主体类型。
func FromSubjectTypeContext(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}
	subjectType, ok := ctx.Value(subjectTypeContextKey{}).(string)
	return subjectType, ok
}
