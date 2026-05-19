// Application scenarios:
// - Protect authenticated routes with a unified authorization layer.
// - Reuse JWT subject and role information in route-level access control.
// - Build higher-level permission checks on top of a simple transport middleware contract.
//
// 适用场景：
// - 用统一鉴权层保护需要登录态的路由。
// - 在路由级访问控制中复用 JWT 的主体和角色信息。
// - 基于简洁的传输层中间件契约构建更高层的权限判断。
package middleware

import (
	"net/http"
	"strings"

	securitycontract "github.com/ngq/gorp/framework/contract/security"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

// AuthorizationChecker performs a request-level authorization decision.
//
// AuthorizationChecker 用于执行请求级鉴权判断。
type AuthorizationChecker func(c transportcontract.Context, claims *securitycontract.JWTClaims) error

// Authorize enforces JWT claim presence first, then applies the custom checker.
//
// Authorize 先校验 JWT claims 是否存在，再执行自定义鉴权逻辑。
func Authorize(checker AuthorizationChecker) transportcontract.Middleware {
	return func(next transportcontract.Handler) transportcontract.Handler {
		return func(c transportcontract.Context) {
			claims, ok := claimsFromContext(c)
			if !ok {
				respondUnauthorized(c, "authentication required")
				return
			}
			if checker != nil {
				if err := checker(c, claims); err != nil {
					respondForbidden(c, err.Error())
					return
				}
			}
			if next != nil {
				next(c)
			}
		}
	}
}

// RequireAuthorization requires authenticated JWT claims to exist in the request context.
//
// RequireAuthorization 要求请求上下文中必须存在已认证的 JWT claims。
func RequireAuthorization() transportcontract.Middleware {
	return Authorize(nil)
}

// RequireSubjectType requires the JWT subject type to match one of the expected values.
//
// RequireSubjectType 要求 JWT subject type 命中给定候选值之一。
func RequireSubjectType(subjectTypes ...string) transportcontract.Middleware {
	expected := normalizeRequiredValues(subjectTypes)
	return Authorize(func(c transportcontract.Context, claims *securitycontract.JWTClaims) error {
		if len(expected) == 0 {
			return nil
		}
		for _, item := range expected {
			if strings.EqualFold(claims.SubjectType, item) {
				return nil
			}
		}
		return ErrForbidden("subject type is not allowed")
	})
}

// RequireAnyRole requires the caller to own at least one of the expected roles.
//
// RequireAnyRole 要求调用方至少拥有一个指定角色。
func RequireAnyRole(roles ...string) transportcontract.Middleware {
	required := normalizeRequiredValues(roles)
	return Authorize(func(c transportcontract.Context, claims *securitycontract.JWTClaims) error {
		if len(required) == 0 {
			return nil
		}
		roleSet := claimsRoleSet(claims)
		for _, role := range required {
			if _, ok := roleSet[strings.ToLower(role)]; ok {
				return nil
			}
		}
		return ErrForbidden("required role is missing")
	})
}

// RequireAllRoles requires the caller to own all expected roles.
//
// RequireAllRoles 要求调用方拥有全部指定角色。
func RequireAllRoles(roles ...string) transportcontract.Middleware {
	required := normalizeRequiredValues(roles)
	return Authorize(func(c transportcontract.Context, claims *securitycontract.JWTClaims) error {
		if len(required) == 0 {
			return nil
		}
		roleSet := claimsRoleSet(claims)
		for _, role := range required {
			if _, ok := roleSet[strings.ToLower(role)]; !ok {
				return ErrForbidden("required role is missing")
			}
		}
		return nil
	})
}

// claimsFromContext extracts JWT claims from the current request context.
//
// claimsFromContext 从当前请求上下文中提取 JWT claims。
func claimsFromContext(c transportcontract.Context) (*securitycontract.JWTClaims, bool) {
	if c == nil {
		return nil, false
	}
	return securitycontract.FromJWTClaimsContext(c)
}

// claimsRoleSet normalizes role values into a set for quick membership checks.
// Uses strings.ToLower to ensure consistency with EqualFold comparisons.
//
// claimsRoleSet 把角色列表归一化为集合，便于快速判断是否命中。
// 使用 strings.ToLower 确保与 EqualFold 比较一致。
func claimsRoleSet(claims *securitycontract.JWTClaims) map[string]struct{} {
	set := make(map[string]struct{})
	if claims == nil {
		return set
	}
	for _, role := range claims.Roles {
		role = strings.TrimSpace(role)
		if role == "" {
			continue
		}
		set[strings.ToLower(role)] = struct{}{}
	}
	return set
}

// normalizeRequiredValues normalizes required subject or role values for comparisons.
// Uses strings.ToLower to ensure consistency with EqualFold comparisons.
//
// normalizeRequiredValues 统一规范 subject 或 role 的目标值，便于比较。
// 使用 strings.ToLower 确保与 EqualFold 比较一致。
func normalizeRequiredValues(values []string) []string {
	normalized := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		normalized = append(normalized, strings.ToLower(value))
	}
	return normalized
}

// respondUnauthorized writes the unified unauthorized response.
//
// respondUnauthorized 输出统一的未认证响应。
func respondUnauthorized(c transportcontract.Context, message string) {
	if gc, ok := unwrapGinContext(c); ok {
		writeGinResponseHeaders(gc)
		resp := Response{
			Code:    CodeUnauthorized,
			Message: message,
		}
		gc.JSON(http.StatusUnauthorized, resp)
		gc.Abort()
		return
	}

	c.JSON(http.StatusUnauthorized, map[string]any{
		"code":    CodeUnauthorized,
		"message": message,
	})
}

// respondForbidden writes the unified forbidden response.
//
// respondForbidden 输出统一的无权限响应。
func respondForbidden(c transportcontract.Context, message string) {
	if gc, ok := unwrapGinContext(c); ok {
		writeGinResponseHeaders(gc)
		resp := Response{
			Code:    CodeForbidden,
			Message: message,
		}
		gc.JSON(http.StatusForbidden, resp)
		gc.Abort()
		return
	}

	c.JSON(http.StatusForbidden, map[string]any{
		"code":    CodeForbidden,
		"message": message,
	})
}
