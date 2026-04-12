package token

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ngq/gorp/framework/contract"
	configprovider "github.com/ngq/gorp/framework/provider/config"
)

// Provider 提供服务令牌认证实现。
//
// 中文说明：
// - 基于 JWT 令牌实现服务间认证；
// - 支持令牌签名和验证；
// - 支持令牌过期时间配置；
// - 支持服务权限映射；
// - 需要项目引入 github.com/golang-jwt/jwt/v5 依赖。
type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string  { return "serviceauth.token" }
func (p *Provider) IsDefer() bool { return true }
func (p *Provider) Provides() []string {
	return []string{contract.ServiceAuthKey, contract.ServiceIdentityKey}
}

func (p *Provider) Register(c contract.Container) error {
	c.Bind(contract.ServiceAuthKey, func(c contract.Container) (any, error) {
		cfg, err := getServiceAuthConfig(c)
		if err != nil {
			return nil, err
		}
		return NewTokenAuthenticator(cfg), nil
	}, true)

	c.Bind(contract.ServiceIdentityKey, func(c contract.Container) (any, error) {
		cfg, err := getServiceAuthConfig(c)
		if err != nil {
			return nil, err
		}
		return &contract.ServiceIdentity{
			ServiceID:   cfg.ServiceName,
			ServiceName: cfg.ServiceName,
			Namespace:   cfg.Namespace,
			Environment: cfg.Environment,
		}, nil
	}, true)

	return nil
}

func (p *Provider) Boot(c contract.Container) error {
	return nil
}

// getServiceAuthConfig 从容器获取服务认证配置。
func getServiceAuthConfig(c contract.Container) (*contract.ServiceAuthConfig, error) {
	cfgAny, err := c.Make(contract.ConfigKey)
	if err != nil {
		return nil, err
	}

	cfg, ok := cfgAny.(contract.Config)
	if !ok {
		return nil, errors.New("serviceauth: invalid config service")
	}

	authCfg := &contract.ServiceAuthConfig{
		Mode:               "token",
		TokenExpiry:        3600, // 默认 1 小时
		ServicePermissions: make(map[string][]string),
	}

	if mode := configprovider.GetStringAny(cfg,
		"serviceauth.mode",
		"service_auth.mode",
	); mode != "" {
		authCfg.Mode = mode
	}

	// 服务信息
	if name := configprovider.GetStringAny(cfg,
		"service.name",
	); name != "" {
		authCfg.ServiceName = name
	}
	if ns := configprovider.GetStringAny(cfg,
		"service.namespace",
	); ns != "" {
		authCfg.Namespace = ns
	}
	if env := configprovider.GetStringAny(cfg,
		"service.environment",
	); env != "" {
		authCfg.Environment = env
	}

	// 令牌配置
	if secret := configprovider.GetStringAny(cfg,
		"serviceauth.token.secret",
		"service_auth.token.secret",
		"service_auth.token_secret",
	); secret != "" {
		authCfg.TokenSecret = secret
	} else {
		// 如果没有配置密钥，生成一个默认值（生产环境必须配置）
		authCfg.TokenSecret = "default-secret-change-in-production"
	}
	if issuer := configprovider.GetStringAny(cfg,
		"serviceauth.token.issuer",
		"service_auth.token.issuer",
		"service_auth.token_issuer",
	); issuer != "" {
		authCfg.TokenIssuer = issuer
	} else {
		authCfg.TokenIssuer = authCfg.ServiceName
	}
	if audience := configprovider.GetStringAny(cfg,
		"serviceauth.token.audience",
		"service_auth.token.audience",
		"service_auth.token_audience",
	); audience != "" {
		authCfg.TokenAudience = audience
	}
	if expiry := configprovider.GetIntAny(cfg,
		"serviceauth.token.expiry",
		"service_auth.token.expiry",
		"service_auth.token_expiry",
	); expiry > 0 {
		authCfg.TokenExpiry = int64(expiry)
	}

	if allowed := configprovider.GetStringSliceAny(cfg,
		"serviceauth.allowed_services",
		"service_auth.allowed_services",
	); len(allowed) > 0 {
		authCfg.AllowedServices = allowed
	}
	if perms := configprovider.GetStringMapAny(cfg,
		"serviceauth.service_permissions",
		"service_auth.service_permissions",
	); len(perms) > 0 {
		authCfg.ServicePermissions = make(map[string][]string, len(perms))
		for serviceName, permList := range perms {
			parts := strings.Split(permList, ",")
			arr := make([]string, 0, len(parts))
			for _, part := range parts {
				part = strings.TrimSpace(part)
				if part != "" {
					arr = append(arr, part)
				}
			}
			authCfg.ServicePermissions[serviceName] = arr
		}
	}

	return authCfg, nil
}

// TokenAuthenticator 是服务令牌认证器实现。
//
// 中文说明：
// - 使用 JWT 令牌实现服务身份验证；
// - 支持 HS256 签名算法；
// - 支持令牌过期验证；
// - 支持服务权限映射。
type TokenAuthenticator struct {
	cfg *contract.ServiceAuthConfig

	// 令牌缓存
	tokenCache sync.Map // map[string]*cachedToken
}

// cachedToken 缓存的令牌。
type cachedToken struct {
	token     string
	expiresAt time.Time
}

// NewTokenAuthenticator 创建令牌认证器。
func NewTokenAuthenticator(cfg *contract.ServiceAuthConfig) *TokenAuthenticator {
	return &TokenAuthenticator{
		cfg: cfg,
	}
}

// Authenticate 验证服务身份（从上下文提取令牌）。
//
// 中文说明：
// - 从上下文中提取 Authorization 头；
// - 验证 Bearer 令牌。
func (a *TokenAuthenticator) Authenticate(ctx context.Context) (*contract.ServiceIdentity, error) {
	// 从上下文获取令牌
	token := extractTokenFromContext(ctx)
	if token == "" {
		return nil, errors.New("serviceauth: no token found in context")
	}

	return a.VerifyToken(ctx, token)
}

// AuthenticateWithToken 使用令牌验证服务身份。
//
// 中文说明：
// - 直接验证提供的令牌。
func (a *TokenAuthenticator) AuthenticateWithToken(ctx context.Context, token string) (*contract.ServiceIdentity, error) {
	return a.VerifyToken(ctx, token)
}

// AuthenticateWithCert 使用证书验证（不支持）。
//
// 中文说明：
// - 令牌认证模式不支持证书验证；
// - 返回错误提示使用令牌认证。
func (a *TokenAuthenticator) AuthenticateWithCert(ctx context.Context, cert *tls.Certificate) (*contract.ServiceIdentity, error) {
	return nil, errors.New("serviceauth.token: certificate authentication not supported, use token instead")
}

// GenerateToken 生成服务令牌。
//
// 中文说明：
// - 为当前服务生成访问令牌；
// - targetService: 目标服务名称，用于设置 audience；
// - 令牌包含服务名、权限等信息。
func (a *TokenAuthenticator) GenerateToken(ctx context.Context, targetService string) (string, error) {
	now := time.Now()
	expiresAt := now.Add(time.Duration(a.cfg.TokenExpiry) * time.Second)

	// 获取服务权限
	permissions := a.cfg.ServicePermissions[a.cfg.ServiceName]

	// 构建令牌声明
	claims := &serviceTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   a.cfg.ServiceName,
			Issuer:    a.cfg.TokenIssuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
		ServiceID:   a.cfg.ServiceName,
		ServiceName: a.cfg.ServiceName,
		Namespace:   a.cfg.Namespace,
		Environment: a.cfg.Environment,
		Permissions: permissions,
	}

	// 设置受众
	if targetService != "" {
		claims.Audience = []string{targetService}
	}

	// 创建令牌
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名
	tokenString, err := token.SignedString([]byte(a.cfg.TokenSecret))
	if err != nil {
		return "", fmt.Errorf("serviceauth: sign token failed: %w", err)
	}

	return tokenString, nil
}

// VerifyToken 验证服务令牌。
//
// 中文说明：
// - 解析并验证 JWT 令牌；
// - 检查签名、过期时间；
// - 返回服务身份信息。
func (a *TokenAuthenticator) VerifyToken(ctx context.Context, tokenString string) (*contract.ServiceIdentity, error) {
	// 解析令牌
	token, err := jwt.ParseWithClaims(tokenString, &serviceTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名算法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("serviceauth: unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(a.cfg.TokenSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("serviceauth: parse token failed: %w", err)
	}

	// 验证令牌有效性
	if !token.Valid {
		return nil, errors.New("serviceauth: invalid token")
	}

	// 提取声明
	claims, ok := token.Claims.(*serviceTokenClaims)
	if !ok {
		return nil, errors.New("serviceauth: invalid token claims")
	}

	// 检查过期时间
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("serviceauth: token expired")
	}

	// 检查服务是否在允许列表
	if len(a.cfg.AllowedServices) > 0 {
		allowed := false
		for _, svc := range a.cfg.AllowedServices {
			if svc == claims.ServiceName {
				allowed = true
				break
			}
		}
		if !allowed {
			return nil, fmt.Errorf("serviceauth: service %s not allowed", claims.ServiceName)
		}
	}

	// 返回服务身份
	return &contract.ServiceIdentity{
		ServiceID:   claims.ServiceID,
		ServiceName: claims.ServiceName,
		Namespace:   claims.Namespace,
		Environment: claims.Environment,
		Permissions: claims.Permissions,
		ExpiresAt:   claims.ExpiresAt.Unix(),
		IssuedAt:    claims.IssuedAt.Unix(),
		Issuer:      claims.Issuer,
	}, nil
}

// serviceTokenClaims 服务令牌声明。
type serviceTokenClaims struct {
	jwt.RegisteredClaims

	ServiceID   string   `json:"service_id"`
	ServiceName string   `json:"service_name"`
	Namespace   string   `json:"namespace,omitempty"`
	Environment string   `json:"environment,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
}

// extractTokenFromContext 从上下文提取令牌。
//
// 中文说明：
// - 支持从 HTTP Authorization 头提取；
// - 支持从 gRPC metadata 提取。
func extractTokenFromContext(ctx context.Context) string {
	// 尝试从 HTTP Authorization 头提取
	if authHeader := ctx.Value("authorization"); authHeader != nil {
		if authStr, ok := authHeader.(string); ok {
			if strings.HasPrefix(authStr, "Bearer ") {
				return strings.TrimPrefix(authStr, "Bearer ")
			}
			return authStr
		}
	}

	// 尝试从 x-service-token 头提取
	if tokenHeader := ctx.Value("x-service-token"); tokenHeader != nil {
		if tokenStr, ok := tokenHeader.(string); ok {
			return tokenStr
		}
	}

	return ""
}
