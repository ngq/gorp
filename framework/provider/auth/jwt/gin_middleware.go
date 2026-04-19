package jwt

import (
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

// AuthMiddleware 基于 framework JWTService 的最小 gin middleware。
//
// 中文说明：
// - 这是面向业务主体的 JWT 鉴权中间件；
// - 负责读取 Authorization Bearer Token、校验 claims、写入主体上下文；
// - expectedSubjectType 非空时，会额外限制主体类型（如 admin/customer）。
func AuthMiddleware(jwtSvc contract.JWTService, expectedSubjectType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if jwtSvc == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "jwt service is not configured"})
			return
		}
		token := extractBearerToken(c.GetHeader("Authorization"))
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}
		claims, err := jwtSvc.Verify(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		if expectedSubjectType != "" && claims.SubjectType != expectedSubjectType {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": fmt.Sprintf("unexpected subject type: %s", claims.SubjectType)})
			return
		}
		c.Set(ContextJWTClaimsKey, claims)
		c.Set(ContextSubjectIDKey, claims.SubjectID)
		c.Set(ContextSubjectTypeKey, claims.SubjectType)
		c.Next()
	}
}

// SubjectIDFromContext 从 gin context 中提取当前业务主体 ID。
func SubjectIDFromContext(c *gin.Context) (int64, bool) {
	v, ok := c.Get(ContextSubjectIDKey)
	if !ok {
		return 0, false
	}
	id, ok := v.(int64)
	return id, ok
}
