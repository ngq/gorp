package mtls

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"os"
	"sync"

	configprovider "github.com/ngq/gorp/framework/provider/config"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	securitycontract "github.com/ngq/gorp/framework/contract/security"
)

type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string  { return "serviceauth.mtls" }
func (p *Provider) IsDefer() bool { return true }
func (p *Provider) Provides() []string {
	return []string{securitycontract.ServiceAuthKey, securitycontract.ServiceIdentityKey}
}

func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(securitycontract.ServiceAuthKey, func(c runtimecontract.Container) (any, error) {
		cfg, err := getServiceAuthConfig(c)
		if err != nil {
			return nil, err
		}
		return NewMTLSAuthenticator(cfg)
	}, true)

	c.Bind(securitycontract.ServiceIdentityKey, func(c runtimecontract.Container) (any, error) {
		cfg, err := getServiceAuthConfig(c)
		if err != nil {
			return nil, err
		}
		return &securitycontract.ServiceIdentity{
			ServiceID:   cfg.ServiceName,
			ServiceName: cfg.ServiceName,
			Namespace:   cfg.Namespace,
			Environment: cfg.Environment,
		}, nil
	}, true)

	return nil
}

func (p *Provider) Boot(c runtimecontract.Container) error { return nil }

func getServiceAuthConfig(c runtimecontract.Container) (*securitycontract.ServiceAuthConfig, error) {
	cfgAny, err := c.Make(datacontract.ConfigKey)
	if err != nil {
		return nil, err
	}

	cfg, ok := cfgAny.(datacontract.Config)
	if !ok {
		return nil, errors.New("serviceauth: invalid config service")
	}

	authCfg := &securitycontract.ServiceAuthConfig{
		Mode:        "mtls",
		MTLSEnabled: true,
	}

	if mode := configprovider.GetStringAny(cfg, "serviceauth.mode", "service_auth.mode"); mode != "" {
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

	if certFile := configprovider.GetStringAny(cfg, "serviceauth.mtls.cert_file", "service_auth.mtls.cert_file", "service_auth.mtls_cert_file"); certFile != "" {
		authCfg.MTLSCertFile = certFile
	}
	if keyFile := configprovider.GetStringAny(cfg, "serviceauth.mtls.key_file", "service_auth.mtls.key_file", "service_auth.mtls_key_file"); keyFile != "" {
		authCfg.MTLSKeyFile = keyFile
	}
	if caFile := configprovider.GetStringAny(cfg, "serviceauth.mtls.ca_file", "service_auth.mtls.ca_file", "service_auth.mtls_ca_file"); caFile != "" {
		authCfg.MTLSCAFile = caFile
	}
	if serverName := configprovider.GetStringAny(cfg, "serviceauth.mtls.server_name", "service_auth.mtls.server_name", "service_auth.mtls_server_name"); serverName != "" {
		authCfg.MTLSServerName = serverName
	}

	return authCfg, nil
}

type MTLSAuthenticator struct {
	cfg *securitycontract.ServiceAuthConfig

	tlsConfig *tls.Config
	certPool  *x509.CertPool
	cert      *tls.Certificate
	certMu    sync.RWMutex
}

func NewMTLSAuthenticator(cfg *securitycontract.ServiceAuthConfig) (*MTLSAuthenticator, error) {
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

func (a *MTLSAuthenticator) Authenticate(ctx context.Context) (*securitycontract.ServiceIdentity, error) {
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

func (a *MTLSAuthenticator) AuthenticatePeerCertificate(ctx context.Context, cert *x509.Certificate) (*securitycontract.ServiceIdentity, error) {
	_ = ctx
	return a.authenticateByCert(cert)
}

func (a *MTLSAuthenticator) GetCurrentIdentity(ctx context.Context) (*securitycontract.ServiceIdentity, error) {
	_ = ctx
	return &securitycontract.ServiceIdentity{
		ServiceID:   a.cfg.ServiceName,
		ServiceName: a.cfg.ServiceName,
		Namespace:   a.cfg.Namespace,
		Environment: a.cfg.Environment,
	}, nil
}

func (a *MTLSAuthenticator) authenticateByCert(cert *x509.Certificate) (*securitycontract.ServiceIdentity, error) {
	if cert == nil {
		return nil, errors.New("serviceauth.mtls: nil certificate")
	}
	serviceName := cert.Subject.CommonName
	if serviceName == "" {
		return nil, errors.New("serviceauth.mtls: certificate missing common name")
	}
	return &securitycontract.ServiceIdentity{
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
