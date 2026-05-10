// Package jwt provides helper functions for JWT service integration.
// These helpers simplify JWT service creation and configuration retrieval.
//
// 本文件提供 JWT 服务集成的辅助函数。
// 这些辅助函数简化 JWT 服务创建和配置获取。
package jwt

import (
	"strings"

	"github.com/ngq/gorp/framework/container"
	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	securitycontract "github.com/ngq/gorp/framework/contract/security"
)

// MustMakeJWTService retrieves JWT service from container, panics on failure.
// Use this when JWT service is guaranteed to be available.
//
// MustMakeJWTService 从容器获取 JWT 服务，失败时 panic。
// 当 JWT 服务保证可用时使用此函数。
func MustMakeJWTService(c runtimecontract.Container) securitycontract.JWTService {
	return container.MustMakeJWTService(c)
}

// JWTSecretFromConfig retrieves JWT secret from config service.
// Checks both "auth.jwt.secret" and legacy "auth.jwt_secret" keys.
// Returns empty string if not found.
//
// JWTSecretFromConfig 从配置服务获取 JWT secret。
// 同时检查 "auth.jwt.secret" 和遗留的 "auth.jwt_secret" 配置键。
// 未找到时返回空字符串。
func JWTSecretFromConfig(cfg datacontract.Config) string {
	if cfg == nil {
		return ""
	}
	secret := strings.TrimSpace(cfg.GetString("auth.jwt.secret"))
	if secret != "" {
		return secret
	}
	return strings.TrimSpace(cfg.GetString("auth.jwt_secret"))
}