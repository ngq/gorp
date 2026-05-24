// Package gorp provides the root-package application startup surface for gorp framework.
// This file exposes security context helpers and shared security aliases.
// Keeps JWT subject and service-identity access convenient for handlers.
//
// Gorp 包提供 gorp 框架的根包层应用启动入口。
// 本文件暴露根包层的安全上下文 helper 和共享安全别名。
// 让 handler 和 middleware 更方便地访问 JWT 主体与服务身份信息。
package gorp

import (
	"context"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	securitycontract "github.com/ngq/gorp/framework/contract/security"
	"github.com/ngq/gorp/framework/application"
	frameworkcontainer "github.com/ngq/gorp/framework/container"
	frameworkjwt "github.com/ngq/gorp/framework/provider/auth/jwt"
)

type ServiceIdentity = securitycontract.ServiceIdentity

// JWTClaims 是共享的业务 JWT claims 模型的顶层别名。
// 业务 handler 通过 gorp.ClaimsFromContext(c) 获取 claims 后可直接使用此类型。
type JWTClaims = securitycontract.JWTClaims

func WithServiceIdentity(ctx context.Context, identity *ServiceIdentity) context.Context {
	return application.WithServiceIdentity(ctx, identity)
}

func FromServiceIdentity(ctx context.Context) (*ServiceIdentity, bool) {
	return application.FromServiceIdentity(ctx)
}

func FromJWTClaimsContext(ctx context.Context) (*securitycontract.JWTClaims, bool) {
	return securitycontract.FromJWTClaimsContext(ctx)
}

func FromSubjectIDContext(ctx context.Context) (int64, bool) {
	return securitycontract.FromSubjectIDContext(ctx)
}

func FromSubjectTypeContext(ctx context.Context) (string, bool) {
	return securitycontract.FromSubjectTypeContext(ctx)
}

// ClaimsFromContext 从 transport.Context 提取 JWT claims。
// 优先从 c.Get() 读取（中间件主路径），回退到 c.Context() 标准上下文读取。
// 这是业务 handler 提取 claims 的推荐方式。
//
// 示例：
//
//	claims, ok := gorp.ClaimsFromContext(c)
func ClaimsFromContext(c Context) (*JWTClaims, bool) {
	if c == nil {
		return nil, false
	}
	// 优先从 Get() 读取，利用 (any, bool) 返回值区分 key 不存在 vs 值为 nil
	val, ok := c.Get(securitycontract.ContextJWTClaimsKey)
	if ok {
		if claims, valid := val.(*securitycontract.JWTClaims); valid && claims != nil {
			return claims, true
		}
	}
	// 回退到标准 context 读取
	return securitycontract.FromJWTClaimsContext(c.Context())
}

// JWTService 从 ctx 或全局默认容器中获取 JWT 服务实例。
// 优先从 ctx 提取 Container，提取不到使用全局默认。
//
// 示例：
//
//	jwtSvc, err := gorp.JWTService(ctx)
//	token, err := jwtSvc.Sign(claims)
func JWTService(ctx context.Context) (securitycontract.JWTService, error) {
	cont := frameworkcontainer.Resolve(ctx)
	return frameworkcontainer.GetJWT(cont)
}

// MustJWTService 从 ctx 或全局默认容器获取 JWT 服务实例，失败时 panic。
//
// MustJWTService 获取 JWT 服务，失败 panic。
func MustJWTService(ctx context.Context) securitycontract.JWTService {
	svc, err := JWTService(ctx)
	if err != nil {
		panic(err)
	}
	return svc
}

// NewJWTService 从配置创建 JWT 服务实例。
// 不依赖 Container，直接传入 secret/issuer/audience 构造。
//
// 示例：
//
//	jwtSvc := gorp.NewJWTService("my-secret", "gorp-app", "api-users")
func NewJWTService(secret, issuer, audience string) securitycontract.JWTService {
	return frameworkjwt.NewJWTService(secret, issuer, audience)
}

// JWTSecretFromConfig 从容器配置中解析 JWT 密钥。
// 读取 `jwt.secret` 配置项。
//
// JWTSecretFromConfig 统一从配置中解析 JWT secret。
func JWTSecretFromConfig(cfg datacontract.Config) string {
	return frameworkjwt.JWTSecretFromConfig(cfg)
}