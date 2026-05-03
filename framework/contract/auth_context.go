package contract

import "context"

type jwtClaimsContextKey struct{}
type subjectIDContextKey struct{}
type subjectTypeContextKey struct{}

// NewJWTClaimsContext stores JWT claims in request context.
func NewJWTClaimsContext(ctx context.Context, claims *JWTClaims) context.Context {
	return context.WithValue(ctx, jwtClaimsContextKey{}, claims)
}

// FromJWTClaimsContext retrieves JWT claims from request context.
func FromJWTClaimsContext(ctx context.Context) (*JWTClaims, bool) {
	if ctx == nil {
		return nil, false
	}
	claims, ok := ctx.Value(jwtClaimsContextKey{}).(*JWTClaims)
	return claims, ok && claims != nil
}

// NewSubjectIDContext stores the current subject id in request context.
func NewSubjectIDContext(ctx context.Context, subjectID int64) context.Context {
	return context.WithValue(ctx, subjectIDContextKey{}, subjectID)
}

// FromSubjectIDContext retrieves the current subject id from request context.
func FromSubjectIDContext(ctx context.Context) (int64, bool) {
	if ctx == nil {
		return 0, false
	}
	subjectID, ok := ctx.Value(subjectIDContextKey{}).(int64)
	return subjectID, ok
}

// NewSubjectTypeContext stores the current subject type in request context.
func NewSubjectTypeContext(ctx context.Context, subjectType string) context.Context {
	return context.WithValue(ctx, subjectTypeContextKey{}, subjectType)
}

// FromSubjectTypeContext retrieves the current subject type from request context.
func FromSubjectTypeContext(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}
	subjectType, ok := ctx.Value(subjectTypeContextKey{}).(string)
	return subjectType, ok
}
