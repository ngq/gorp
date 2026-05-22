// Application scenarios:
// - Propagate tenant identity through request-scoped or task-scoped context.
// - Support multi-tenant business flows without coupling tenant access to one transport layer.
// - Keep tenant context keys private and collision-free.
//
// 适用场景：
// - 在请求级或任务级 context 中透传租户身份。
// - 支持多租户业务流程，而不把租户读取绑定到某一种 transport 层。
// - 保持 tenant context key 私有且避免冲突。
package support

import "context"

type tenantContextKey struct{}

// NewTenantContext stores tenant identity in context.
//
// NewTenantContext 将租户身份写入 context。
func NewTenantContext(ctx context.Context, tenant string) context.Context {
	return context.WithValue(ctx, tenantContextKey{}, tenant)
}

// FromTenantContext reads tenant identity from context.
//
// FromTenantContext 从 context 中读取租户身份。
func FromTenantContext(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}
	tenant, ok := ctx.Value(tenantContextKey{}).(string)
	return tenant, ok && tenant != ""
}
