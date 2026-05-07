// Application scenarios:
// - Expose root-package security context helpers and shared security aliases.
// - Keep JWT subject and service-identity access convenient for handlers and middleware.
// - Re-export common security primitives without forcing direct dependency on lower-level packages.
//
// 适用场景：
// - 暴露根包层的安全上下文 helper 和共享安全别名。
// - 让 handler 和 middleware 更方便地访问 JWT 主体信息与服务身份信息。
// - 在不强迫业务直接依赖更底层包的前提下重导出常用安全原语。
package gorp

import (
	"context"

	"github.com/ngq/gorp/framework/application"
	"github.com/ngq/gorp/framework/contract/security"
)

type ServiceIdentity = security.ServiceIdentity

func WithServiceIdentity(ctx context.Context, identity *ServiceIdentity) context.Context {
	return application.WithServiceIdentity(ctx, identity)
}

func FromServiceIdentity(ctx context.Context) (*ServiceIdentity, bool) {
	return application.FromServiceIdentity(ctx)
}

func FromJWTClaimsContext(ctx context.Context) (*security.JWTClaims, bool) {
	return security.FromJWTClaimsContext(ctx)
}

func FromSubjectIDContext(ctx context.Context) (int64, bool) {
	return security.FromSubjectIDContext(ctx)
}

func FromSubjectTypeContext(ctx context.Context) (string, bool) {
	return security.FromSubjectTypeContext(ctx)
}
