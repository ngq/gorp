package jwt

import (
	"context"
	"strings"

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

// Make 从容器获取业务 JWT 服务。
func Make(c runtimecontract.Container) (securitycontract.JWTService, error) {
	v, err := c.Make(securitycontract.AuthJWTKey)
	if err != nil {
		return nil, err
	}
	return v.(securitycontract.JWTService), nil
}

// MustMake 从容器获取业务 JWT 服务，失败 panic。
func MustMake(c runtimecontract.Container) securitycontract.JWTService {
	return frameworkjwt.MustMakeJWTService(c)
}

// NewService 创建业务 JWT 服务。
func NewService(secret, issuer, audience string) securitycontract.JWTService {
	return frameworkjwt.NewJWTService(secret, issuer, audience)
}

// SecretFromConfig 统一从配置中解析业务 JWT secret。
func SecretFromConfig(cfg datacontract.Config) string {
	return frameworkjwt.JWTSecretFromConfig(cfg)
}

// AuthMiddleware 创建基于业务 JWT 的 framework HTTP 中间件。
func AuthMiddleware(jwtSvc securitycontract.JWTService, expectedSubjectType string) transportcontract.HTTPMiddleware {
	return func(next transportcontract.HTTPHandler) transportcontract.HTTPHandler {
		return func(c transportcontract.HTTPContext) {
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
			ctx := securitycontract.NewJWTClaimsContext(c.Context(), claims)
			ctx = securitycontract.NewSubjectIDContext(ctx, claims.SubjectID)
			ctx = securitycontract.NewSubjectTypeContext(ctx, claims.SubjectType)
			c.SetContext(ctx)
			if next != nil {
				next(c)
			}
		}
	}
}

// SubjectIDFromContext 从 request/framework context 中提取业务主体 ID。
func SubjectIDFromContext(ctx context.Context) (int64, bool) {
	return securitycontract.FromSubjectIDContext(ctx)
}

// ClaimsFromRequestContext 从 request context 提取业务 JWT claims。
func ClaimsFromRequestContext(ctx context.Context) (*JWTClaims, bool) {
	return securitycontract.FromJWTClaimsContext(ctx)
}

// SubjectIDFromRequestContext 从 request context 提取业务主体 ID。
func SubjectIDFromRequestContext(ctx context.Context) (int64, bool) {
	return securitycontract.FromSubjectIDContext(ctx)
}

// SubjectTypeFromRequestContext 从 request context 提取业务主体类型。
func SubjectTypeFromRequestContext(ctx context.Context) (string, bool) {
	return securitycontract.FromSubjectTypeContext(ctx)
}
