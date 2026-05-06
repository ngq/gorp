package support

import "context"

type tenantContextKey struct{}

// NewTenantContext stores tenant identity in context.
func NewTenantContext(ctx context.Context, tenant string) context.Context {
	return context.WithValue(ctx, tenantContextKey{}, tenant)
}

// FromTenantContext reads tenant identity from context.
func FromTenantContext(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}
	tenant, ok := ctx.Value(tenantContextKey{}).(string)
	return tenant, ok && tenant != ""
}
