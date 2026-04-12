package token

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ngq/gorp/framework/contract"
)

const (
	ContextJWTClaimsKey   = "framework.jwt.claims"
	ContextSubjectIDKey   = "framework.subject.id"
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

func SubjectIDFromContext(c *gin.Context) (int64, bool) {
	v, ok := c.Get(ContextSubjectIDKey)
	if !ok {
		return 0, false
	}
	id, ok := v.(int64)
	return id, ok
}
