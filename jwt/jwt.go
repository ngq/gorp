package jwt

import (
	"context"
	"strings"

	"github.com/ngq/gorp/framework/container"
	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	securitycontract "github.com/ngq/gorp/framework/contract/security"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	frameworkjwt "github.com/ngq/gorp/framework/provider/auth/jwt"
)

const (
	ContextJWTClaimsKey   = frameworkjwt.ContextJWTClaimsKey
	ContextSubjectIDKey   = frameworkjwt.ContextSubjectIDKey
	ContextSubjectTypeKey = frameworkjwt.ContextSubjectTypeKey
)

type JWTClaims = securitycontract.JWTClaims
type JWTService = securitycontract.JWTService

// Get returns the business JWT service from the container.
// Get 从容器获取业务 JWT 服务。
func Get(c runtimecontract.Container) (securitycontract.JWTService, error) {
	return container.MakeWith[securitycontract.JWTService](c, securitycontract.AuthJWTKey)
}

// GetOrPanic returns the business JWT service from the container and panics on failure.
// GetOrPanic 从容器获取业务 JWT 服务，失败 panic。
func GetOrPanic(c runtimecontract.Container) securitycontract.JWTService {
	return frameworkjwt.MustMakeJWTService(c)
}

// NewService creates a business JWT service.
// NewService 创建业务 JWT 服务。
func NewService(secret, issuer, audience string) securitycontract.JWTService {
	return frameworkjwt.NewJWTService(secret, issuer, audience)
}

// SecretFromConfig resolves the business JWT secret from config.
// SecretFromConfig 统一从配置中解析业务 JWT secret。
func SecretFromConfig(cfg datacontract.Config) string {
	return frameworkjwt.JWTSecretFromConfig(cfg)
}

// AuthMiddleware creates a framework HTTP middleware based on business JWT verification.
// AuthMiddleware 创建基于业务 JWT 的 framework HTTP 中间件。
//
// Example:
//
//	router.Use(jwt.AuthMiddleware(jwtSvc, "user"))
func AuthMiddleware(jwtSvc securitycontract.JWTService, expectedSubjectType string) transportcontract.Middleware {
	return func(next transportcontract.Handler) transportcontract.Handler {
		return func(c transportcontract.Context) {
			if jwtSvc == nil {
				c.JSON(401, map[string]any{"error": "jwt service is not configured"})
				return
			}
			token := frameworkjwt.ClaimsFromRequestContext
			_ = token
			header := c.GetHeader("Authorization")
			claimsToken := ""
			const prefix = "Bearer "
			if strings.HasPrefix(strings.TrimSpace(header), prefix) {
				claimsToken = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(header), prefix))
			}
			if claimsToken == "" {
				c.JSON(401, map[string]any{"error": "missing bearer token"})
				return
			}
			claims, err := jwtSvc.Verify(claimsToken)
			if err != nil {
				c.JSON(401, map[string]any{"error": err.Error()})
				return
			}
			if expectedSubjectType != "" && claims.SubjectType != expectedSubjectType {
				c.JSON(403, map[string]any{"error": "unexpected subject type: " + claims.SubjectType})
				return
			}
			c.Set(ContextJWTClaimsKey, claims)
			c.Set(ContextSubjectIDKey, claims.SubjectID)
			c.Set(ContextSubjectTypeKey, claims.SubjectType)
			if next != nil {
				next(c)
			}
		}
	}
}

// SubjectIDFromContext extracts the business subject id from context.
// SubjectIDFromContext 从 request/framework context 中提取业务主体 ID。
func SubjectIDFromContext(ctx context.Context) (int64, bool) {
	return securitycontract.FromSubjectIDContext(ctx)
}

// ClaimsFromRequestContext extracts business JWT claims from request context.
// ClaimsFromRequestContext 从 request context 提取业务 JWT claims。
func ClaimsFromRequestContext(ctx context.Context) (*JWTClaims, bool) {
	return securitycontract.FromJWTClaimsContext(ctx)
}

// SubjectIDFromRequestContext extracts the business subject id from request context.
// SubjectIDFromRequestContext 从 request context 提取业务主体 ID。
func SubjectIDFromRequestContext(ctx context.Context) (int64, bool) {
	return securitycontract.FromSubjectIDContext(ctx)
}

// SubjectTypeFromRequestContext extracts the business subject type from request context.
// SubjectTypeFromRequestContext 从 request context 提取业务主体类型。
func SubjectTypeFromRequestContext(ctx context.Context) (string, bool) {
	return securitycontract.FromSubjectTypeContext(ctx)
}
