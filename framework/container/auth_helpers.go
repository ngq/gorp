// Package container provides runtime dependency injection container for gorp framework.
// This file exposes security-related convenience helpers for JWT capability access.
// Keeps common auth service lookup compact for handlers, middleware, bootstrap hooks.
//
// 容器包提供 gorp 框架的运行时依赖注入容器实现。
// 本文件暴露聚焦安全能力的便捷 helper。
// 让 handler、middleware 和 bootstrap hook 获取 JWT 能力时保持简洁。
package container

import (
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	securitycontract "github.com/ngq/gorp/framework/contract/security"
)

// MustMakeJWTService resolves the JWT service and panics on failure.
//
// MustMakeJWTService 解析 JWT 服务，失败时 panic。
func MustMakeJWTService(c runtimecontract.Container) securitycontract.JWTService {
	return MustMakeWith[securitycontract.JWTService](c, securitycontract.AuthJWTKey)
}
