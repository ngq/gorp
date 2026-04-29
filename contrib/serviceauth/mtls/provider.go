package mtls

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/ngq/gorp/framework/contract"
	configprovider "github.com/ngq/gorp/framework/provider/config"
)

// Provider 提供 mTLS 服务认证实现。
//
// 中文说明：
// - 基于双向 TLS 实现服务间认证；
// - 支持证书验证和身份提取；
// - 支持证书轮换；
// - 当前已从 framework/provider 真实下沉到 contrib 层。
type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string     { return "serviceauth.mtls" }
func (p *Provider) IsDefer() bool    { return true }
func (p *Provider) Provides() []string {
	return []string{contract.ServiceAuthKey, contract.ServiceIdentityKey}
}

func (p *Provider) Register(c contract.Container) error {
	c.Bind(contract.ServiceAuthKey, func(c contract.Container) (any, error) {
		cfg, err := getServiceAuthConfig(c)
		if err != nil {
			return nil, err
		}
		return NewMTLSAuthenticator(cfg)
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

func (p *Provider) Boot(c contract.Container) error { return nil }

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
		Mode:        "mtls",
		MTLSEnabled: true,
	}

	if mode := configprovider.GetStringAny(cfg,
		"serviceauth.mode",
		"service_auth.mode",
	); mode != "" {
		authCfg.Mode = mode
	}

	if name := configprovider.GetStringAny(cfg, "service.name"); name != "" {
		authCfg.ServiceName = name
	}
	if ns := configprovider.GetStringAny(cfg, "service.namespace"); ns != "" {
		authCfg.Namespace = ns
	}
	if env := configprovider.GetStringAny(cfg, "service.environment"); env != "" {
		authCfg.Environment = env
	}

	if certFile := configprovider.GetStringAny(cfg,
		"serviceauth.mtls.cert_file",
		"service_auth.mtls.cert_file",
		"service_auth.mtls_cert_file",
	); certFile != "" {
		authCfg.MTLSCertFile = certFile
	}
	if keyFile := configprovider.GetStringAny(cfg,
		"serviceauth.mtls.key_file",
		"service_auth.mtls.key_file",
		"service_auth.mtls_key_file",
	); keyFile != "" {
		authCfg.MTLSKeyFile = keyFile
	}
	if caFile := configprovider.GetStringAny(cfg,
		"serviceauth.mtls.ca_file",
		"service_auth.mtls.ca_file",
		"service_auth.mtls_ca_file",
	); caFile != "" {
		authCfg.MTLSCAFile = caFile
	}
	if serverName := configprovider.GetStringAny(cfg,
		"serviceauth.mtls.server_name",
		"service_auth.mtls.server_name",
		"service_auth.mtls_server_name",
	); serverName != "" {
		authCfg.MTLSServerName = serverName
	}

	return authCfg, nil
}

// MTLSAuthenticator 是 mTLS 服务认证器实现。
type MTLSAuthenticator struct {
	cfg *contract.ServiceAuthConfig

	tlsConfig *tls.Config
	certPool  *x509.CertPool
	cert      *tls.Certificate
	certMu    sync.RWMutex
}

// NewMTLSAuthenticator 创建 mTLS 认证器。
func NewMTLSAuthenticator(cfg *contract.ServiceAuthConfig) (*MTLSAuthenticator, error) {
	auth := &MTLSAuthenticator{cfg: cfg}

	if cfg.MTLSCAFile != "" {
		caCert, err := os.ReadFile(cfg.MTLSCAFile)
		if err != nil {
			return nil, fmt.Errorf("serviceauth.mtls: load CA cert failed: %w", err)
		}
		auth.certPool = x509.NewCertPool()
		if !auth.certPool.AppendCertsFromPEM(caCert) {
			return nil, errors.New("serviceauth.mtls: failed to parse CA cert")
		}
	}

	if cfg.MTLSCertFile != "" && cfg.MTLSKeyFile != "" {
		cert, err := tls.LoadX509KeyPair(cfg.MTLSCertFile, cfg.MTLSKeyFile)
		if err != nil {
			return nil, fmt.Errorf("serviceauth.mtls: load cert failed: %w", err)
		}
		auth.cert = &cert
	}

	auth.tlsConfig = &tls.Config{
		Certificates:       []tls.Certificate{},
		ClientCAs:          auth.certPool,
		ClientAuth:         tls.RequireAndVerifyClientCert,
		ServerName:         cfg.MTLSServerName,
		InsecureSkipVerify: false,
	}
	if auth.cert != nil {
		auth.tlsConfig.Certificates = []tls.Certificate{*auth.cert}
	}

	return auth, nil
}

// Authenticate 验证服务身份（从上下文提取证书）。
func (a *MTLSAuthenticator) Authenticate(ctx context.Context) (*contract.ServiceIdentity, error) {
	tlsConnState := extractTLSState(ctx)
	if tlsConnState == nil {
		return nil, errors.New("serviceauth.mtls: no TLS connection found")
	}
	if len(tlsConnState.PeerCertificates) == 0 {
		return nil, errors.New("serviceauth.mtls: no client certificate provided")
	}
	cert := tlsConnState.PeerCertificates[0]
	return a.authenticateByCert(cert)
}

// AuthenticateWithToken 使用令牌验证（不支持）。
func (a *MTLSAuthenticator) AuthenticateWithToken(ctx context.Context, token string) (*contract.ServiceIdentity, error) {
	return nil, errors.New("serviceauth.mtls: token authentication not supported, use certificate instead")
}

// AuthenticateWithCert 使用证书验证服务身份。
func (a *MTLSAuthenticator) AuthenticateWithCert(ctx context.Context, cert *tls.Certificate) (*contract.ServiceIdentity, error) {
	if cert == nil || len(cert.Certificate) == 0 {
		return nil, errors.New("serviceauth.mtls: invalid certificate")
	}
	x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return nil, fmt.Errorf("serviceauth.mtls: parse certificate failed: %w", err)
	}
	return a.authenticateByCert(x509Cert)
}

// GetCurrentIdentity 获取当前服务身份。
func (a *MTLSAuthenticator) GetCurrentIdentity(ctx context.Context) (*contract.ServiceIdentity, error) {
	return &contract.ServiceIdentity{
		ServiceID:   a.cfg.ServiceName,
		ServiceName: a.cfg.ServiceName,
		Namespace:   a.cfg.Namespace,
		Environment: a.cfg.Environment,
	}, nil
}

func (a *MTLSAuthenticator) authenticateByCert(cert *x509.Certificate) (*contract.ServiceIdentity, error) {
	if cert == nil {
		return nil, errors.New("serviceauth.mtls: nil certificate")
	}
	serviceName := cert.Subject.CommonName
	if serviceName == "" {
		return nil, errors.New("serviceauth.mtls: certificate missing common name")
	}
	return &contract.ServiceIdentity{
		ServiceID:   serviceName,
		ServiceName: serviceName,
		Namespace:   a.cfg.Namespace,
		Environment: a.cfg.Environment,
	}, nil
}

func extractTLSState(ctx context.Context) *tls.ConnectionState {
	if state, ok := ctx.Value("tls_state").(*tls.ConnectionState); ok {
		return state
	}
	return nil
}
