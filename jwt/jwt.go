package jwt

import (
	"github.com/gin-gonic/gin"
	frameworkjwt "github.com/ngq/gorp/framework/provider/auth/jwt"
	"github.com/ngq/gorp/framework/contract"
)

const (
	ContextJWTClaimsKey  = frameworkjwt.ContextJWTClaimsKey
	ContextSubjectIDKey  = frameworkjwt.ContextSubjectIDKey
	ContextSubjectTypeKey = frameworkjwt.ContextSubjectTypeKey
)

type JWTClaims = contract.JWTClaims
type JWTService = contract.JWTService

// Make 从容器获取业务 JWT 服务。
func Make(c contract.Container) (contract.JWTService, error) {
	v, err := c.Make(contract.AuthJWTKey)
	if err != nil {
		return nil, err
	}
	return v.(contract.JWTService), nil
}

// MustMake 从容器获取业务 JWT 服务，失败 panic。
func MustMake(c contract.Container) contract.JWTService {
	return frameworkjwt.MustMakeJWTService(c)
}

// NewService 创建业务 JWT 服务。
func NewService(secret, issuer, audience string) contract.JWTService {
	return frameworkjwt.NewJWTService(secret, issuer, audience)
}

// SecretFromConfig 统一从配置中解析业务 JWT secret。
func SecretFromConfig(cfg contract.Config) string {
	return frameworkjwt.JWTSecretFromConfig(cfg)
}

// AuthMiddleware 创建基于业务 JWT 的 gin 中间件。
func AuthMiddleware(jwtSvc contract.JWTService, expectedSubjectType string) gin.HandlerFunc {
	return frameworkjwt.AuthMiddleware(jwtSvc, expectedSubjectType)
}

// SubjectIDFromContext 从 gin context 中提取业务主体 ID。
func SubjectIDFromContext(c *gin.Context) (int64, bool) {
	return frameworkjwt.SubjectIDFromContext(c)
}
