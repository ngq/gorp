// Application scenarios:
// - Propagate authenticated service identity through request-scoped contexts.
// - Let service-auth middleware, handlers, and downstream services share one consistent identity access path.
// - Keep service-identity context keys private and collision-free.
//
// 适用场景：
// - 在请求级 context 中透传已认证的服务身份。
// - 让服务鉴权 middleware、handler 和下游 service 共享统一的身份读取路径。
// - 保持服务身份相关 context key 私有且避免冲突。
package security

import "context"

type serviceIdentityContextKey struct{}

// NewServiceIdentityContext stores service identity in context.
//
// NewServiceIdentityContext 将服务身份写入 context。
func NewServiceIdentityContext(ctx context.Context, identity *ServiceIdentity) context.Context {
	return context.WithValue(ctx, serviceIdentityContextKey{}, identity)
}

// FromServiceIdentityContext retrieves service identity from context.
//
// FromServiceIdentityContext 从 context 中读取服务身份。
func FromServiceIdentityContext(ctx context.Context) (*ServiceIdentity, bool) {
	if ctx == nil {
		return nil, false
	}
	identity, ok := ctx.Value(serviceIdentityContextKey{}).(*ServiceIdentity)
	return identity, ok && identity != nil
}
