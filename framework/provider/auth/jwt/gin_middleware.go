package jwt

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	securitycontract "github.com/ngq/gorp/framework/contract/security"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
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

func AuthMiddleware(jwtSvc securitycontract.JWTService, expectedSubjectType string) transportcontract.HTTPMiddleware {
	return func(next transportcontract.HTTPHandler) transportcontract.HTTPHandler {
		return func(c transportcontract.HTTPContext) {
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
			ctx = securitycontract.NewJWTClaimsContext(ctx, claims)
			ctx = securitycontract.NewSubjectIDContext(ctx, claims.SubjectID)
			ctx = securitycontract.NewSubjectTypeContext(ctx, claims.SubjectType)
			c.SetContext(ctx)
			if next != nil {
				next(c)
			}
		}
	}
}

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
	return securitycontract.FromSubjectIDContext(c.Request.Context())
}

func ClaimsFromRequestContext(ctx context.Context) (*securitycontract.JWTClaims, bool) {
	return securitycontract.FromJWTClaimsContext(ctx)
}

func SubjectIDFromRequestContext(ctx context.Context) (int64, bool) {
	return securitycontract.FromSubjectIDContext(ctx)
}

func SubjectTypeFromRequestContext(ctx context.Context) (string, bool) {
	return securitycontract.FromSubjectTypeContext(ctx)
}
