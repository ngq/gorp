package security

import "context"

type jwtClaimsContextKey struct{}
type subjectIDContextKey struct{}
type subjectTypeContextKey struct{}

func NewJWTClaimsContext(ctx context.Context, claims *JWTClaims) context.Context {
	return context.WithValue(ctx, jwtClaimsContextKey{}, claims)
}

func FromJWTClaimsContext(ctx context.Context) (*JWTClaims, bool) {
	if ctx == nil {
		return nil, false
	}
	claims, ok := ctx.Value(jwtClaimsContextKey{}).(*JWTClaims)
	return claims, ok && claims != nil
}

func NewSubjectIDContext(ctx context.Context, subjectID int64) context.Context {
	return context.WithValue(ctx, subjectIDContextKey{}, subjectID)
}

func FromSubjectIDContext(ctx context.Context) (int64, bool) {
	if ctx == nil {
		return 0, false
	}
	subjectID, ok := ctx.Value(subjectIDContextKey{}).(int64)
	return subjectID, ok
}

func NewSubjectTypeContext(ctx context.Context, subjectType string) context.Context {
	return context.WithValue(ctx, subjectTypeContextKey{}, subjectType)
}

func FromSubjectTypeContext(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}
	subjectType, ok := ctx.Value(subjectTypeContextKey{}).(string)
	return subjectType, ok
}
