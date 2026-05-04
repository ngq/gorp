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

// ServiceAuthenticator 服务身份认证接口。
type ServiceAuthenticator interface {
	Authenticate(ctx context.Context) (*ServiceIdentity, error)
}

// ServiceTokenIssuer 服务令牌签发接口。
type ServiceTokenIssuer interface {
	GenerateToken(ctx context.Context, targetService string) (string, error)
}

// ServiceTokenVerifier 服务令牌校验接口。
type ServiceTokenVerifier interface {
	VerifyToken(ctx context.Context, token string) (*ServiceIdentity, error)
}

// ServicePeerCertificateAuthenticator 对等端证书认证接口。
type ServicePeerCertificateAuthenticator interface {
	AuthenticatePeerCertificate(ctx context.Context, cert *x509.Certificate) (*ServiceIdentity, error)
}

// ServiceIdentity 服务身份信息。
type ServiceIdentity struct {
	ServiceID   string
	ServiceName string
	Namespace   string
	Environment string
	Metadata    map[string]string
}

// ServiceAuthConfig 服务认证配置。
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
type ServiceProvider interface {
	Name() string
	Register(Container) error
	Boot(Container) error
	IsDefer() bool
	Provides() []string
}

// Container defines the minimal container surface needed by service auth providers.
type Container interface{}

// ServiceAuthProvider 服务认证 Provider 接口。
type ServiceAuthProvider interface {
	ServiceProvider
	CreateAuthenticator(cfg *ServiceAuthConfig) (ServiceAuthenticator, error)
}

// TLSCertificateLoader TLS 证书加载接口。
type TLSCertificateLoader interface {
	LoadCertificate(certFile, keyFile string) (*tls.Certificate, error)
	LoadCA(caFile string) (*x509.CertPool, error)
}
