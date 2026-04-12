package noop

import (
	"context"
	"crypto/tls"
	"errors"

	"github.com/ngq/gorp/framework/contract"
)

// Provider 提供 noop 服务认证实现。
//
// 中文说明：
// - 单体项目默认使用此 provider；
// - 不引入任何外部依赖；
// - 所有认证操作返回默认身份，不进行实际验证。
type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string     { return "serviceauth.noop" }
func (p *Provider) IsDefer() bool    { return true }
func (p *Provider) Provides() []string {
	return []string{contract.ServiceAuthKey, contract.ServiceIdentityKey}
}

func (p *Provider) Register(c contract.Container) error {
	c.Bind(contract.ServiceAuthKey, func(c contract.Container) (any, error) {
		return &noopAuthenticator{}, nil
	}, true)

	c.Bind(contract.ServiceIdentityKey, func(c contract.Container) (any, error) {
		return &contract.ServiceIdentity{
			ServiceName: "local",
		}, nil
	}, true)

	return nil
}

func (p *Provider) Boot(c contract.Container) error {
	return nil
}

// ErrNoopAuth 表示 noop 认证模式不支持实际验证。
var ErrNoopAuth = errors.New("serviceauth: noop mode, authentication not available in monolith")

// noopAuthenticator 是 ServiceAuthenticator 的空实现。
//
// 中文说明：
// - 单体项目所有服务在同一进程内，无需服务间认证；
// - 所有认证方法返回默认身份。
type noopAuthenticator struct{}

// Authenticate 验证服务身份（返回默认身份）。
//
// 中文说明：
// - 单体项目返回本地服务身份；
// - 不进行实际验证。
func (a *noopAuthenticator) Authenticate(ctx context.Context) (*contract.ServiceIdentity, error) {
	return &contract.ServiceIdentity{
		ServiceID:   "local",
		ServiceName: "local",
		Namespace:   "default",
	}, nil
}

// AuthenticateWithToken 使用令牌验证（返回默认身份）。
//
// 中文说明：
// - noop 模式忽略令牌，返回默认身份。
func (a *noopAuthenticator) AuthenticateWithToken(ctx context.Context, token string) (*contract.ServiceIdentity, error) {
	return a.Authenticate(ctx)
}

// AuthenticateWithCert 使用证书验证（返回默认身份）。
//
// 中文说明：
// - noop 模式忽略证书，返回默认身份。
func (a *noopAuthenticator) AuthenticateWithCert(ctx context.Context, cert *tls.Certificate) (*contract.ServiceIdentity, error) {
	return a.Authenticate(ctx)
}

// GenerateToken 生成服务令牌（返回空令牌）。
//
// 中文说明：
// - noop 模式不生成实际令牌。
func (a *noopAuthenticator) GenerateToken(ctx context.Context, targetService string) (string, error) {
	return "noop-token", nil
}

// VerifyToken 验证服务令牌（返回默认身份）。
//
// 中文说明：
// - noop 模式不验证令牌，直接返回默认身份。
func (a *noopAuthenticator) VerifyToken(ctx context.Context, token string) (*contract.ServiceIdentity, error) {
	return a.Authenticate(ctx)
}