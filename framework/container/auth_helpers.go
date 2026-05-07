// Application scenarios:
// - Expose focused security-related convenience helpers on top of the runtime container.
// - Keep common JWT capability access compact for handlers, middleware, and bootstrap hooks.
// - Avoid scattering repeated auth service lookup code across packages.
//
// 适用场景：
// - 在运行时容器之上暴露聚焦安全能力的便捷 helper。
// - 让 handler、middleware 和 bootstrap hook 获取 JWT 能力时保持简洁。
// - 避免认证服务查找代码散落在各个包里重复出现。
package container

import (
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	securitycontract "github.com/ngq/gorp/framework/contract/security"
)

// MustMakeJWTService resolves the JWT service and panics on failure.
//
// MustMakeJWTService 解析 JWT 服务，失败时 panic。
func MustMakeJWTService(c runtimecontract.Container) securitycontract.JWTService {
	v := c.MustMake(securitycontract.AuthJWTKey)
	return v.(securitycontract.JWTService)
}
