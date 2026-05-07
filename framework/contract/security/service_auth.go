// Application scenarios:
// - Define service-to-service authentication contracts used by microservice runtime flows.
// - Support token-based, mTLS-based, and peer-certificate-based identity verification.
// - Keep service-auth provider creation and config models independent from concrete implementations.
//
// 适用场景：
// - 定义微服务运行时流程使用的服务间身份认证契约。
// - 支持基于 token、mTLS 和对端证书的身份校验方式。
// - 让服务鉴权 provider 的创建和配置模型不依赖具体实现。
package security

import (
	"context"
	"crypto/tls"
	"crypto/x509"
)

const (
	ServiceAuthKey     = "framework.service.auth"
	ServiceIdentityKey = "framework.service.identity"
)

// ServiceAuthenticator defines the service-identity authentication contract.
//
// ServiceAuthenticator 定义服务身份认证契约。
type ServiceAuthenticator interface {
	// Authenticate authenticates the current request and returns the resolved service identity.
	//
	// Authenticate 认证当前请求并返回解析出的服务身份。
	Authenticate(ctx context.Context) (*ServiceIdentity, error)
}

// ServiceTokenIssuer defines the service token issuing contract.
//
// ServiceTokenIssuer 定义服务令牌签发契约。
type ServiceTokenIssuer interface {
	// GenerateToken generates a token for the target service.
	//
	// GenerateToken 为目标服务生成令牌。
	GenerateToken(ctx context.Context, targetService string) (string, error)
}

// ServiceTokenVerifier defines the service token verification contract.
//
// ServiceTokenVerifier 定义服务令牌校验契约。
type ServiceTokenVerifier interface {
	// VerifyToken verifies a token and returns the resolved service identity.
	//
	// VerifyToken 校验令牌并返回解析出的服务身份。
	VerifyToken(ctx context.Context, token string) (*ServiceIdentity, error)
}

// ServicePeerCertificateAuthenticator defines peer-certificate authentication for service identity.
//
// ServicePeerCertificateAuthenticator 定义基于对端证书的服务身份认证契约。
type ServicePeerCertificateAuthenticator interface {
	// AuthenticatePeerCertificate authenticates the peer certificate and returns the resolved service identity.
	//
	// AuthenticatePeerCertificate 校验对端证书并返回解析出的服务身份。
	AuthenticatePeerCertificate(ctx context.Context, cert *x509.Certificate) (*ServiceIdentity, error)
}

// ServiceIdentity describes one authenticated service identity.
//
// ServiceIdentity 描述一个已认证的服务身份。
type ServiceIdentity struct {
	ServiceID   string
	ServiceName string
	Namespace   string
	Environment string
	Metadata    map[string]string
}

// ServiceAuthConfig describes service-auth configuration.
//
// ServiceAuthConfig 描述服务鉴权配置。
type ServiceAuthConfig struct {
	Mode        string
	ServiceName string
	Namespace   string
	Environment string

	MTLSEnabled    bool
	MTLSCertFile   string
	MTLSKeyFile    string
	MTLSCAFile     string
	MTLSServerName string

	TokenSecret   string
	TokenExpiry   int64
	TokenIssuer   string
	TokenAudience string

	AllowedServices []string
}

// ServiceProvider defines the provider registration surface needed by service auth providers.
//
// ServiceProvider 定义服务鉴权 provider 所需的 provider 注册接口面。
type ServiceProvider interface {
	Name() string
	Register(Container) error
	Boot(Container) error
	IsDefer() bool
	Provides() []string
}

// Container defines the minimal container surface needed by service auth providers.
//
// Container 定义服务鉴权 provider 所需的最小容器接口面。
type Container interface{}

// ServiceAuthProvider defines the provider contract for creating service authenticators.
//
// ServiceAuthProvider 定义创建服务认证器的 provider 契约。
type ServiceAuthProvider interface {
	ServiceProvider

	// CreateAuthenticator creates a service authenticator from config.
	//
	// CreateAuthenticator 根据配置创建服务认证器。
	CreateAuthenticator(cfg *ServiceAuthConfig) (ServiceAuthenticator, error)
}

// TLSCertificateLoader defines the certificate-loading contract used by mTLS auth providers.
//
// TLSCertificateLoader 定义 mTLS 鉴权 provider 使用的证书加载契约。
type TLSCertificateLoader interface {
	LoadCertificate(certFile, keyFile string) (*tls.Certificate, error)
	LoadCA(caFile string) (*x509.CertPool, error)
}
