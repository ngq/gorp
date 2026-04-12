package jwt

import (
	"errors"
	"strings"

	"github.com/ngq/gorp/framework/contract"
	tokensvc "github.com/ngq/gorp/framework/provider/serviceauth/token"
)

// Provider 提供 framework 级业务 JWT 服务实现。
//
// 中文说明：
// - 仅负责业务 JWT（AuthJWTKey）绑定；
// - 与 ServiceAuthKey（服务间认证）完全解耦；
// - 统一复用 token.JWTSecretFromConfig 的配置读取策略，避免重复配置解析逻辑。
type Provider struct{}

// NewProvider 创建业务 JWT Provider。
func NewProvider() *Provider { return &Provider{} }

// Name 返回 Provider 名称。
func (p *Provider) Name() string { return "auth.jwt" }

// IsDefer 返回是否延迟注册。
func (p *Provider) IsDefer() bool { return true }

// Provides 返回当前 Provider 提供的容器 key。
func (p *Provider) Provides() []string {
	return []string{contract.AuthJWTKey}
}

// Register 注册业务 JWT 服务。
//
// 中文说明：
// - 若配置服务不可用，则回退到默认 secret/issuer；
// - secret 优先读取 auth.jwt.secret，兼容 auth.jwt_secret；
// - issuer 未配置时回退到 service.name。
func (p *Provider) Register(c contract.Container) error {
	c.Bind(contract.AuthJWTKey, func(c contract.Container) (any, error) {
		cfgAny, err := c.Make(contract.ConfigKey)
		if err != nil {
			return tokensvc.NewJWTService("default-secret-change-in-production", "gorp", ""), nil
		}

		cfg, ok := cfgAny.(contract.Config)
		if !ok {
			return nil, errors.New("auth.jwt: invalid config service")
		}

		secret := tokensvc.JWTSecretFromConfig(cfg)
		if secret == "" {
			secret = "default-secret-change-in-production"
		}

		issuer := strings.TrimSpace(cfg.GetString("auth.jwt.issuer"))
		if issuer == "" {
			issuer = cfg.GetString("service.name")
		}

		audience := strings.TrimSpace(cfg.GetString("auth.jwt.audience"))
		return tokensvc.NewJWTService(secret, issuer, audience), nil
	}, true)

	return nil
}

// Boot 启动 Provider。
func (p *Provider) Boot(c contract.Container) error {
	return nil
}
