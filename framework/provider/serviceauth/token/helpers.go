package token

import (
	"strings"

	"github.com/ngq/gorp/framework/container"
	"github.com/ngq/gorp/framework/contract"
)

// MustMakeJWTService 从容器中获取 framework 级 JWT 服务。
func MustMakeJWTService(c contract.Container) contract.JWTService {
	return container.MustMakeJWTService(c)
}

// JWTSecretFromConfig 统一从配置中解析业务 JWT secret。
//
// 中文说明：
// - 优先读取 `auth.jwt.secret`；
// - 兼容读取旧键 `auth.jwt_secret`；
// - 统一放在 framework/provider/serviceauth/token 边界，避免业务层散写配置读取逻辑。
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
