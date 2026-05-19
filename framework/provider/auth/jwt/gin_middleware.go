// Package jwt provides Gin HTTP middleware for JWT authentication.
// The middleware extracts Bearer token from Authorization header, verifies it,
// and injects claims into request context.
// Supports SkipPaths option to skip authentication for specific paths.
// Context keys:
//
// 本文件提供 Gin HTTP 中间件，用于 JWT 认证。
// 中间件从 Authorization header 提取 Bearer token，验证后注入 context。
// 支持 SkipPaths 选项跳过指定路径的认证。
// Context 键：
//   - ContextJWTClaimsKey: JWT claims
//   - ContextSubjectIDKey: Subject ID
//   - ContextSubjectTypeKey: Subject type
//
// Eg:
//
//	// 基本用法
//	router.Use(jwt.AuthMiddleware(jwtSvc, "user"))
//
//	// 全局注册 + 白名单跳过
//	router.Use(jwt.AuthMiddleware(jwtSvc, "",
//	    jwt.WithSkipPaths("/api/v1/auth/login", "/api/v1/auth/register"),
//	))
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

// authMiddlewareConfig 包含 JWT 认证中间件的配置选项。
// 支持白名单路径跳过、角色校验等扩展能力。
type authMiddlewareConfig struct {
	// SkipPaths 是不需要认证的路径白名单。
	// 匹配时使用精确匹配，不支持的通配符。
	// 适用于全局注册中间件后，跳过登录/注册等公开接口。
	SkipPaths []string

	// RequiredRoles 是访问所需的角色列表。
	// 为空时不做角色校验；非空时要求 JWT claims 中的 Roles
	// 至少包含其中一个角色。
	RequiredRoles []string
}

// AuthMiddlewareOption 是 JWT 认证中间件的配置选项函数。
//
// AuthMiddlewareOption is the config option function for JWT auth middleware.
type AuthMiddlewareOption func(*authMiddlewareConfig)

// WithSkipPaths 设置不需要认证的路径白名单。
// 匹配规则：精确匹配请求路径（不含 query string）。
// 适用于全局注册中间件后，跳过 /api/v1/auth/login 等公开接口。
//
// WithSkipPaths sets the path whitelist that skips authentication.
// Matching rule: exact match on request path (excluding query string).
// Use when registering middleware globally but need to skip public endpoints.
func WithSkipPaths(paths ...string) AuthMiddlewareOption {
	return func(cfg *authMiddlewareConfig) {
		cfg.SkipPaths = append(cfg.SkipPaths, paths...)
	}
}

// WithRequiredRoles 设置访问所需的角色列表。
// JWT claims 中的 Roles 至少包含其中一个角色才能通过。
// 为空时不做角色校验。
//
// WithRequiredRoles sets the required roles for access.
// JWT claims must contain at least one of the specified roles.
// Empty means no role check.
func WithRequiredRoles(roles ...string) AuthMiddlewareOption {
	return func(cfg *authMiddlewareConfig) {
		cfg.RequiredRoles = append(cfg.RequiredRoles, roles...)
	}
}

// isSkipPath 判断请求路径是否在白名单中。
// 使用精确匹配，确保 O(1) 查找性能。
//
// isSkipPath checks if the request path is in the skip list.
// Uses exact matching for O(1) lookup performance.
func isSkipPath(path string, skipPaths []string) bool {
	for _, p := range skipPaths {
		if p == path {
			return true
		}
	}
	return false
}

// hasRequiredRole 检查 JWT claims 是否包含所需角色中的至少一个。
//
// hasRequiredRole checks if JWT claims contain at least one of the required roles.
func hasRequiredRole(claimsRoles, requiredRoles []string) bool {
	if len(requiredRoles) == 0 {
		return true
	}
	for _, required := range requiredRoles {
		for _, got := range claimsRoles {
			if required == got {
				return true
			}
		}
	}
	return false
}

// AuthMiddleware creates an HTTP middleware for JWT authentication.
// Core logic: Extract Bearer token, verify with JWT service, inject claims to context.
// Supports SkipPaths to skip authentication for specific paths (e.g., login, register).
//
// AuthMiddleware 创建用于 JWT 认证的 HTTP 中间件。
// 核心逻辑：提取 Bearer token，用 JWT 服务验证，将 claims 注入 context。
// 支持 SkipPaths 跳过指定路径的认证（如登录、注册接口）。
//
// Parameters:
//   - jwtSvc: JWT service for token verification
//   - expectedSubjectType: optional subject type validation (empty means no check)
//   - opts: optional configuration (WithSkipPaths, WithRequiredRoles)
//
// Returns 401 if token is missing or invalid, 403 if subject type or role mismatch.
//
// token 缺失或无效返回 401，主体类型或角色不匹配返回 403。
func AuthMiddleware(jwtSvc securitycontract.JWTService, expectedSubjectType string, opts ...AuthMiddlewareOption) transportcontract.Middleware {
	// 应用配置选项。
	cfg := &authMiddlewareConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	// 预构建 skipPaths 集合，避免每次请求都分配。
	skipPaths := cfg.SkipPaths
	requiredRoles := cfg.RequiredRoles

	return func(next transportcontract.Handler) transportcontract.Handler {
		return func(c transportcontract.Context) {
			// 白名单路径跳过认证，直接放行。
			if len(skipPaths) > 0 && c.Request() != nil && isSkipPath(c.Request().URL.Path, skipPaths) {
				if next != nil {
					next(c)
				}
				return
			}

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
			// 角色校验：如果配置了 RequiredRoles，检查 claims 中是否包含。
			if len(requiredRoles) > 0 && !hasRequiredRole(claims.Roles, requiredRoles) {
				c.JSON(http.StatusForbidden, map[string]any{"error": "insufficient role"})
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
