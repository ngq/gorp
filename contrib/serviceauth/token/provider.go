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
// - 只在 framework 层处理服务身份认证，不解释服务权限语义；
// - 当前已从 framework/provider 真实下沉到 contrib 层。
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
		if err != nil { return nil, err }
		return NewTokenAuthenticator(cfg), nil
	}, true)
	c.Bind(contract.ServiceIdentityKey, func(c contract.Container) (any, error) {
		cfg, err := getServiceAuthConfig(c)
		if err != nil { return nil, err }
		return &contract.ServiceIdentity{ServiceID: cfg.ServiceName, ServiceName: cfg.ServiceName, Namespace: cfg.Namespace, Environment: cfg.Environment}, nil
	}, true)
	return nil
}
func (p *Provider) Boot(c contract.Container) error { return nil }

func getServiceAuthConfig(c contract.Container) (*contract.ServiceAuthConfig, error) {
	cfgAny, err := c.Make(contract.ConfigKey)
	if err != nil { return nil, err }
	cfg, ok := cfgAny.(contract.Config)
	if !ok { return nil, errors.New("serviceauth: invalid config service") }
	authCfg := &contract.ServiceAuthConfig{Mode: "token", TokenExpiry: 3600}
	if mode := configprovider.GetStringAny(cfg, "serviceauth.mode", "serviceauth.backend", "service_auth.mode", "service_auth.backend"); mode != "" { authCfg.Mode = mode }
	if name := configprovider.GetStringAny(cfg, "service.name"); name != "" { authCfg.ServiceName = name }
	if ns := configprovider.GetStringAny(cfg, "service.namespace"); ns != "" { authCfg.Namespace = ns }
	if env := configprovider.GetStringAny(cfg, "service.environment"); env != "" { authCfg.Environment = env }
	if secret := configprovider.GetStringAny(cfg, "serviceauth.token.secret", "service_auth.token.secret", "service_auth.token_secret"); secret != "" { authCfg.TokenSecret = secret } else { authCfg.TokenSecret = "default-secret-change-in-production" }
	if issuer := configprovider.GetStringAny(cfg, "serviceauth.token.issuer", "service_auth.token.issuer", "service_auth.token_issuer"); issuer != "" { authCfg.TokenIssuer = issuer } else { authCfg.TokenIssuer = authCfg.ServiceName }
	if audience := configprovider.GetStringAny(cfg, "serviceauth.token.audience", "service_auth.token.audience", "service_auth.token_audience"); audience != "" { authCfg.TokenAudience = audience }
	if expiry := configprovider.GetIntAny(cfg, "serviceauth.token.expiry", "service_auth.token.expiry", "service_auth.token_expiry"); expiry > 0 { authCfg.TokenExpiry = int64(expiry) }
	if allowed := configprovider.GetStringSliceAny(cfg, "serviceauth.allowed_services", "service_auth.allowed_services"); len(allowed) > 0 { authCfg.AllowedServices = allowed }
	return authCfg, nil
}

type TokenAuthenticator struct {
	cfg *contract.ServiceAuthConfig
	tokenCache sync.Map
}

type cachedToken struct {
	token string
	expiresAt time.Time
}

func NewTokenAuthenticator(cfg *contract.ServiceAuthConfig) *TokenAuthenticator { return &TokenAuthenticator{cfg: cfg} }

func (a *TokenAuthenticator) Authenticate(ctx context.Context) (*contract.ServiceIdentity, error) {
	token := extractTokenFromContext(ctx)
	if token == "" { return nil, errors.New("serviceauth: no token found in context") }
	return a.VerifyToken(ctx, token)
}
func (a *TokenAuthenticator) AuthenticateWithToken(ctx context.Context, token string) (*contract.ServiceIdentity, error) {
	return a.VerifyToken(ctx, token)
}
func (a *TokenAuthenticator) AuthenticateWithCert(ctx context.Context, cert *tls.Certificate) (*contract.ServiceIdentity, error) {
	return nil, errors.New("serviceauth.token: certificate authentication not supported, use token instead")
}

func (a *TokenAuthenticator) GenerateToken(ctx context.Context, targetService string) (string, error) {
	now := time.Now()
	expiresAt := now.Add(time.Duration(a.cfg.TokenExpiry) * time.Second)
	claims := &serviceTokenClaims{RegisteredClaims: jwt.RegisteredClaims{Subject: a.cfg.ServiceName, Issuer: a.cfg.TokenIssuer, IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(expiresAt)}, ServiceID: a.cfg.ServiceName, ServiceName: a.cfg.ServiceName, Namespace: a.cfg.Namespace, Environment: a.cfg.Environment}
	if targetService != "" { claims.Audience = []string{targetService} }
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(a.cfg.TokenSecret))
}

func (a *TokenAuthenticator) VerifyToken(ctx context.Context, tokenString string) (*contract.ServiceIdentity, error) {
	token, err := jwt.ParseWithClaims(tokenString, &serviceTokenClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(a.cfg.TokenSecret), nil
	})
	if err != nil { return nil, fmt.Errorf("serviceauth.token: verify token failed: %w", err) }
	claims, ok := token.Claims.(*serviceTokenClaims)
	if !ok || !token.Valid { return nil, errors.New("serviceauth.token: invalid token claims") }
	if len(a.cfg.AllowedServices) > 0 {
		allowed := false
		for _, service := range a.cfg.AllowedServices { if service == claims.ServiceName { allowed = true; break } }
		if !allowed { return nil, fmt.Errorf("serviceauth.token: service %s not allowed", claims.ServiceName) }
	}
	return &contract.ServiceIdentity{ServiceID: claims.ServiceID, ServiceName: claims.ServiceName, Namespace: claims.Namespace, Environment: claims.Environment}, nil
}

func (a *TokenAuthenticator) GetCurrentIdentity(ctx context.Context) (*contract.ServiceIdentity, error) {
	return &contract.ServiceIdentity{ServiceID: a.cfg.ServiceName, ServiceName: a.cfg.ServiceName, Namespace: a.cfg.Namespace, Environment: a.cfg.Environment}, nil
}

type serviceTokenClaims struct {
	jwt.RegisteredClaims
	ServiceID string `json:"service_id"`
	ServiceName string `json:"service_name"`
	Namespace string `json:"namespace,omitempty"`
	Environment string `json:"environment,omitempty"`
}

func extractTokenFromContext(ctx context.Context) string {
	if token, ok := ctx.Value("authorization").(string); ok {
		if strings.HasPrefix(strings.ToLower(token), "bearer ") { return token[7:] }
		return token
	}
	return ""
}
