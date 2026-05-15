// Package noop provides a no-op service authenticator for monolith scenarios.
// This authenticator returns a local identity for all authentication requests.
// Note: Service-to-service authentication is not available in monolith mode.
//
// 服务认证实现包，用于单体应用场景。
// 此认证器对所有认证请求返回本地身份。
// 注意：服务间认证在单体模式下不可用。
package noop

import (
	"context"
	"crypto/x509"
	"errors"

	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	securitycontract "github.com/ngq/gorp/framework/contract/security"
)

// Provider registers no-op service authenticator contracts.
//
// Provider 注册空服务认证契约。
type Provider struct{}

// NewProvider creates a new no-op service auth provider instance.
//
// NewProvider 创建新的空服务认证 Provider 实例。
func NewProvider() *Provider { return &Provider{} }

// Name returns the provider name "serviceauth.noop".
//
// Name 返回 Provider 名称 "serviceauth.noop"。
func (p *Provider) Name() string { return "serviceauth.noop" }

// IsDefer returns true, service auth can be deferred until first use.
//
// IsDefer 返回 true，服务认证可延迟初始化直到首次使用。
func (p *Provider) IsDefer() bool { return true }

// Provides returns the service auth contract keys.
//
// Provides 返回服务认证契约键列表。
func (p *Provider) Provides() []string {
	return []string{securitycontract.ServiceAuthKey, securitycontract.ServiceIdentityKey}
}

// DependsOn returns the keys this provider depends on.
// Noop service auth has no dependencies.
//
// DependsOn 返回该 provider 依赖的 key。
// Noop service auth 无依赖。
func (p *Provider) DependsOn() []string { return nil }

// Register binds the no-op service authenticator to the container.
//
// Register 将空服务认证绑定到容器。
func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(securitycontract.ServiceAuthKey, func(c runtimecontract.Container) (any, error) {
		return &noopAuthenticator{}, nil
	}, true)

	c.Bind(securitycontract.ServiceIdentityKey, func(c runtimecontract.Container) (any, error) {
		return &securitycontract.ServiceIdentity{
			ServiceName: "local",
		}, nil
	}, true)

	return nil
}

// Boot is a no-op for this provider.
//
// Boot 此 Provider 无启动逻辑。
func (p *Provider) Boot(runtimecontract.Container) error { return nil }

// ErrNoopAuth indicates service authentication is not available in monolith mode.
//
// ErrNoopAuth 表示服务认证在单体模式下不可用。
var ErrNoopAuth = errors.New("serviceauth: noop mode, authentication not available in monolith")

// noopAuthenticator implements ServiceAuthenticator with no-op behavior.
//
// noopAuthenticator 使用空行为实现 ServiceAuthenticator 接口。
type noopAuthenticator struct{}

// Authenticate returns a local service identity.
//
// Authenticate 返回本地服务身份。
func (a *noopAuthenticator) Authenticate(ctx context.Context) (*securitycontract.ServiceIdentity, error) {
	_ = ctx
	return &securitycontract.ServiceIdentity{
		ServiceID:   "local",
		ServiceName: "local",
		Namespace:   "default",
	}, nil
}

// GenerateToken returns a placeholder token.
//
// GenerateToken 返回占位符 token。
func (a *noopAuthenticator) GenerateToken(ctx context.Context, targetService string) (string, error) {
	_, _ = ctx, targetService
	return "noop-token", nil
}

// VerifyToken returns a local service identity.
//
// VerifyToken 返回本地服务身份。
func (a *noopAuthenticator) VerifyToken(ctx context.Context, token string) (*securitycontract.ServiceIdentity, error) {
	_, _ = ctx, token
	return a.Authenticate(ctx)
}

// AuthenticatePeerCertificate returns a local service identity.
//
// AuthenticatePeerCertificate 返回本地服务身份。
func (a *noopAuthenticator) AuthenticatePeerCertificate(ctx context.Context, cert *x509.Certificate) (*securitycontract.ServiceIdentity, error) {
	_, _ = ctx, cert
	return a.Authenticate(ctx)
}