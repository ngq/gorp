package jwt

import (
	"context"
	"strings"

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

// AuthMiddleware 创建基于业务 JWT 的 framework HTTP 中间件。
func AuthMiddleware(jwtSvc contract.JWTService, expectedSubjectType string) contract.HTTPMiddleware {
	return func(c contract.HTTPContext, next contract.HTTPNext) {
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
		ctx := contract.NewJWTClaimsContext(c.Context(), claims)
		ctx = contract.NewSubjectIDContext(ctx, claims.SubjectID)
		ctx = contract.NewSubjectTypeContext(ctx, claims.SubjectType)
		c.SetContext(ctx)
		if next != nil {
			next()
		}
	}
}

// SubjectIDFromContext 从 request/framework context 中提取业务主体 ID。
func SubjectIDFromContext(ctx context.Context) (int64, bool) {
	return contract.FromSubjectIDContext(ctx)
}

// ClaimsFromRequestContext 从 request context 提取业务 JWT claims。
func ClaimsFromRequestContext(ctx context.Context) (*JWTClaims, bool) {
	return contract.FromJWTClaimsContext(ctx)
}

// SubjectIDFromRequestContext 从 request context 提取业务主体 ID。
func SubjectIDFromRequestContext(ctx context.Context) (int64, bool) {
	return contract.FromSubjectIDContext(ctx)
}

// SubjectTypeFromRequestContext 从 request context 提取业务主体类型。
func SubjectTypeFromRequestContext(ctx context.Context) (string, bool) {
	return contract.FromSubjectTypeContext(ctx)
}
