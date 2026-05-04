package jwt

import (
	"strings"

	"github.com/ngq/gorp/framework/container"
	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	securitycontract "github.com/ngq/gorp/framework/contract/security"
)

func MustMakeJWTService(c runtimecontract.Container) securitycontract.JWTService {
	return container.MustMakeJWTService(c)
}

func JWTSecretFromConfig(cfg datacontract.Config) string {
	if cfg == nil {
		return ""
	}
	secret := strings.TrimSpace(cfg.GetString("auth.jwt.secret"))
	if secret != "" {
		return secret
	}
	return strings.TrimSpace(cfg.GetString("auth.jwt_secret"))
}
