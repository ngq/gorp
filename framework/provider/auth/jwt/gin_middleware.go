// Package jwt provides Gin HTTP middleware for JWT authentication.
// The middleware extracts Bearer token from Authorization header, verifies it,
// and injects claims into request context.
// Context keys:
//
// 本文件提供 Gin HTTP 中间件，用于 JWT 认证。
// 中间件从 Authorization header 提取 Bearer token，验证后注入 context。
// Context 键：
//   - ContextJWTClaimsKey: JWT claims
//   - ContextSubjectIDKey: Subject ID
//   - ContextSubjectTypeKey: Subject type
//
// Eg:
//
//	// 使用中间件
//	router.Use(jwt.AuthMiddleware(jwtSvc, "user"))
//
//	// 从 context 获取 claims
//	claims, ok := jwt.ClaimsFromRequestContext(ctx)
//	subjectID, ok := jwt.SubjectIDFromRequestContext(ctx)
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

// Context keys for storing JWT information in Gin context.
//
// Gin context 中存储 JWT 信息的键。
const (
	// ContextJWTClaimsKey is the key for JWT claims in Gin context.
	//
	// ContextJWTClaimsKey 是 Gin context 中 JWT claims 的键。
	ContextJWTClaimsKey = "framework.jwt.claims"
	// ContextSubjectIDKey is the key for subject ID in Gin context.
	//
	// ContextSubjectIDKey 是 Gin context 中主体 ID 的键。
	ContextSubjectIDKey = "framework.subject.id"
	// ContextSubjectTypeKey is the key for subject type in Gin context.
	//
	// ContextSubjectTypeKey 是 Gin context 中主体类型的键。
	ContextSubjectTypeKey = "framework.subject.type"
)

// extractBearerToken extracts the Bearer token from Authorization header value.
// Returns empty string if header is not a valid Bearer token.
//
// extractBearerToken 从 Authorization header 值中提取 Bearer token。
// 如果 header 不是有效的 Bearer token，返回空字符串。
func extractBearerToken(header string) string {
	const prefix = "Bearer "
	if !strings.HasPrefix(strings.TrimSpace(header), prefix) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(header), prefix))
}

// AuthMiddleware creates an HTTP middleware for JWT authentication.
// Core logic: Extract Bearer token, verify with JWT service, inject claims to context.
// Parameters:
//
// AuthMiddleware 创建用于 JWT 认证的 HTTP 中间件。
// 核心逻辑：提取 Bearer token，用 JWT 服务验证，将 claims 注入 context。
// 参数：
//   - jwtSvc: JWT service for token verification
//   - expectedSubjectType: optional subject type validation (empty means no check)
//
// Returns 401 if token is missing or invalid, 403 if subject type mismatch.
//
// token 缺失或无效返回 401，主体类型不匹配返回 403。
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

// SubjectIDFromContext extracts subject ID from Gin context.
//
// SubjectIDFromContext 从 Gin context 提取主体 ID。
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

// ClaimsFromRequestContext extracts JWT claims from request context.
//
// ClaimsFromRequestContext 从请求 context 提取 JWT claims。
func ClaimsFromRequestContext(ctx context.Context) (*securitycontract.JWTClaims, bool) {
	return securitycontract.FromJWTClaimsContext(ctx)
}

// SubjectIDFromRequestContext extracts subject ID from request context.
//
// SubjectIDFromRequestContext 从请求 context 提取主体 ID。
func SubjectIDFromRequestContext(ctx context.Context) (int64, bool) {
	return securitycontract.FromSubjectIDContext(ctx)
}

// SubjectTypeFromRequestContext extracts subject type from request context.
//
// SubjectTypeFromRequestContext 从请求 context 提取主体类型。
func SubjectTypeFromRequestContext(ctx context.Context) (string, bool) {
	return securitycontract.FromSubjectTypeContext(ctx)
}