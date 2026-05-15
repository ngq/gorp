// Package jwt provides JWT authentication service for gorp framework.
// The provider reads JWT configuration from config service.
// Configuration via config.yaml:
//
// JWT 认证包，提供基于 JWT 的身份认证能力。
// Provider 从配置服务读取 JWT 配置。
// 通过 config.yaml 配置：
//
//	auth:
//	  jwt:
//	    secret: "your-secret-key"
//	    issuer: "your-service-name"
//	    audience: "your-audience"
//	  jwt_secret: "fallback-secret"  # legacy config key
//
// Eg:
//
//	// 注册 Provider
//	app.Register(jwt.NewProvider())
//
//	// 使用 JWT 服务
//	jwtSvc := c.MustMake(securitycontract.AuthJWTKey).(securitycontract.JWTService)
//	token, _ := jwtSvc.Sign(claims)
//	claims, _ := jwtSvc.Verify(token)
package jwt

import (
	"errors"
	"strings"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	securitycontract "github.com/ngq/gorp/framework/contract/security"
)

// Provider registers the JWT authentication service contract.
//
// Provider 注册 JWT 认证服务契约。
type Provider struct{}

// NewProvider creates a new JWT provider instance.
//
// NewProvider 创建新的 JWT Provider 实例。
func NewProvider() *Provider { return &Provider{} }

// Name returns the provider name "auth.jwt".
//
// Name 返回 Provider 名称 "auth.jwt"。
func (p *Provider) Name() string { return "auth.jwt" }

// IsDefer returns true, JWT service can be deferred until first use.
//
// IsDefer 返回 true，JWT 服务可延迟初始化直到首次使用。
func (p *Provider) IsDefer() bool { return true }

// Provides returns the JWT service contract key.
//
// Provides 返回 JWT 服务契约键。
func (p *Provider) Provides() []string {
	return []string{securitycontract.AuthJWTKey}
}

// DependsOn returns the keys this provider depends on.
// JWT provider depends on Config for JWT configuration.
//
// DependsOn 返回该 provider 依赖的 key。
// JWT provider 依赖 Config 获取 JWT 配置。
func (p *Provider) DependsOn() []string { return []string{datacontract.ConfigKey} }

// Register binds the JWT service factory to the container.
// Core logic: Read secret, issuer, audience from config, create JWTService.
// Note: Uses fallback secret if not configured (should change in production).
//
// Register 将 JWT 服务工厂绑定到容器。
// 核心逻辑：从配置读取 secret、issuer、audience，创建 JWTService。
// 注意：未配置时使用默认 secret（生产环境应修改）。
func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(securitycontract.AuthJWTKey, func(c runtimecontract.Container) (any, error) {
		cfgAny, err := c.Make(datacontract.ConfigKey)
		if err != nil {
			return NewJWTService("default-secret-change-in-production", "gorp", ""), nil
		}

		cfg, ok := cfgAny.(datacontract.Config)
		if !ok {
			return nil, errors.New("auth.jwt: invalid config service")
		}

		secret := JWTSecretFromConfig(cfg)
		if secret == "" {
			secret = "default-secret-change-in-production"
		}

		issuer := strings.TrimSpace(cfg.GetString("auth.jwt.issuer"))
		if issuer == "" {
			issuer = cfg.GetString("service.name")
		}

		audience := strings.TrimSpace(cfg.GetString("auth.jwt.audience"))
		return NewJWTService(secret, issuer, audience), nil
	}, true)

	return nil
}

// Boot is a no-op for JWT provider.
//
// Boot JWT Provider 无启动逻辑。
func (p *Provider) Boot(c runtimecontract.Container) error {
	return nil
}