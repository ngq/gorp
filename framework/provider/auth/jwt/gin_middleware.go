package jwt

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ngq/gorp/framework/contract"
)

const (
	// ContextJWTClaimsKey 是业务 JWT claims 在 gin context 中的键。
	ContextJWTClaimsKey = "framework.jwt.claims"
	// ContextSubjectIDKey 是当前业务主体 ID 在 gin context 中的键。
	ContextSubjectIDKey = "framework.subject.id"
	// ContextSubjectTypeKey 是当前业务主体类型在 gin context 中的键。
	ContextSubjectTypeKey = "framework.subject.type"
)

func extractBearerToken(header string) string {
	const prefix = "Bearer "
	if !strings.HasPrefix(strings.TrimSpace(header), prefix) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(header), prefix))
}

// AuthMiddleware 基于 framework JWTService 的默认 HTTP 鉴权中间件。
//
// 中文说明：
// - 这是面向业务主体的 JWT 鉴权中间件；
// - 负责读取 Authorization Bearer Token、校验 claims、写入主体上下文；
// - expectedSubjectType 非空时，会额外限制主体类型（如 admin/customer）；
// - 默认主线返回 framework `contract.HTTPMiddleware`，Gin 只作为 provider 内部承接层。
func AuthMiddleware(jwtSvc contract.JWTService, expectedSubjectType string) contract.HTTPMiddleware {
	return func(c contract.HTTPContext, next contract.HTTPNext) {
		if jwtSvc == nil {
			c.JSON(http.StatusUnauthorized, map[string]any{"error": "jwt service is not configured"})
			return
		}
		token := extractBearerToken(c.GetHeader("Authorization"))
		if token == "" {
			c.JSON(http.StatusUnauthorized, map[string]any{"error": "missing bearer token"})
			return
		}
		claims, err := jwtSvc.Verify(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, map[string]any{"error": err.Error()})
			return
		}
		if expectedSubjectType != "" && claims.SubjectType != expectedSubjectType {
			c.JSON(http.StatusForbidden, map[string]any{"error": fmt.Sprintf("unexpected subject type: %s", claims.SubjectType)})
			return
		}
		ctx := c.Context()
		ctx = contract.NewJWTClaimsContext(ctx, claims)
		ctx = contract.NewSubjectIDContext(ctx, claims.SubjectID)
		ctx = contract.NewSubjectTypeContext(ctx, claims.SubjectType)
		c.SetContext(ctx)
		if next != nil {
			next()
		}
	}
}

// SubjectIDFromContext 从 gin context 中提取当前业务主体 ID。
func SubjectIDFromContext(c *gin.Context) (int64, bool) {
	if c == nil {
		return 0, false
	}
	v, ok := c.Get(ContextSubjectIDKey)
	if ok {
		id, ok := v.(int64)
		return id, ok
	}
	if c.Request == nil {
		return 0, false
	}
	return contract.FromSubjectIDContext(c.Request.Context())
}

// ClaimsFromRequestContext 从 request context 中提取当前业务 JWT claims。
func ClaimsFromRequestContext(ctx context.Context) (*contract.JWTClaims, bool) {
	return contract.FromJWTClaimsContext(ctx)
}

// SubjectIDFromRequestContext 从 request context 中提取当前业务主体 ID。
func SubjectIDFromRequestContext(ctx context.Context) (int64, bool) {
	return contract.FromSubjectIDContext(ctx)
}

// SubjectTypeFromRequestContext 从 request context 中提取当前业务主体类型。
func SubjectTypeFromRequestContext(ctx context.Context) (string, bool) {
	return contract.FromSubjectTypeContext(ctx)
}
