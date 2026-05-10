// Package gorp provides the root-package application startup surface for gorp framework.
// This file exposes security context helpers and shared security aliases.
// Keeps JWT subject and service-identity access convenient for handlers.
//
// Gorp 包提供 gorp 框架的根包层应用启动入口。
// 本文件暴露根包层的安全上下文 helper 和共享安全别名。
// 让 handler 和 middleware 更方便地访问 JWT 主体与服务身份信息。
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
