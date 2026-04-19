package jwt

import (
	"strings"

	"github.com/ngq/gorp/framework/container"
	"github.com/ngq/gorp/framework/contract"
)

// MustMakeJWTService 从容器中获取 framework 级业务 JWT 服务。
//
// 中文说明：
// - 这是 auth/jwt 边界下的推荐 helper；
// - 内部继续复用 container 层统一入口；
// - 供业务 service、handler、middleware 在启动期直接获取 JWT 服务。
func MustMakeJWTService(c contract.Container) contract.JWTService {
	return container.MustMakeJWTService(c)
}

// JWTSecretFromConfig 统一从配置中解析业务 JWT secret。
//
// 中文说明：
// - 优先读取 `auth.jwt.secret`；
// - 兼容读取旧键 `auth.jwt_secret`；
// - 这里属于业务 JWT 配置读取，不再继续挂在 serviceauth/token 边界下。
func JWTSecretFromConfig(cfg contract.Config) string {
	if cfg == nil {
		return ""
	}
	secret := strings.TrimSpace(cfg.GetString("auth.jwt.secret"))
	if secret != "" {
		return secret
	}
	return strings.TrimSpace(cfg.GetString("auth.jwt_secret"))
}
