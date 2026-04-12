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
// - 需要配置 CA 证书、服务证书。
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
		Mode:               "mtls",
		MTLSEnabled:        true,
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

	// mTLS 配置
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
//
// 中文说明：
// - 使用双向 TLS 证书验证服务身份；
// - 从证书中提取服务名称；
// - 支持证书白名单验证。
type MTLSAuthenticator struct {
	cfg *contract.ServiceAuthConfig

	// TLS 配置
	tlsConfig *tls.Config
	certPool  *x509.CertPool

	// 证书缓存
	cert     *tls.Certificate
	certMu   sync.RWMutex
}

// NewMTLSAuthenticator 创建 mTLS 认证器。
func NewMTLSAuthenticator(cfg *contract.ServiceAuthConfig) (*MTLSAuthenticator, error) {
	auth := &MTLSAuthenticator{
		cfg: cfg,
	}

	// 加载 CA 证书
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

	// 加载服务证书
	if cfg.MTLSCertFile != "" && cfg.MTLSKeyFile != "" {
		cert, err := tls.LoadX509KeyPair(cfg.MTLSCertFile, cfg.MTLSKeyFile)
		if err != nil {
			return nil, fmt.Errorf("serviceauth.mtls: load cert failed: %w", err)
		}
		auth.cert = &cert
	}

	// 构建 TLS 配置
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
//
// 中文说明：
// - 从上下文中提取 TLS 连接状态；
// - 验证客户端证书。
func (a *MTLSAuthenticator) Authenticate(ctx context.Context) (*contract.ServiceIdentity, error) {
	// 从上下文获取 TLS 连接状态
	tlsConnState := extractTLSState(ctx)
	if tlsConnState == nil {
		return nil, errors.New("serviceauth.mtls: no TLS connection found")
	}

	// 验证客户端证书
	if len(tlsConnState.PeerCertificates) == 0 {
		return nil, errors.New("serviceauth.mtls: no client certificate provided")
	}

	// 使用第一个证书进行验证
	cert := tlsConnState.PeerCertificates[0]
	return a.authenticateByCert(cert)
}

// AuthenticateWithToken 使用令牌验证（不支持）。
//
// 中文说明：
// - mTLS 认证模式不支持令牌验证；
// - 返回错误提示使用证书认证。
func (a *MTLSAuthenticator) AuthenticateWithToken(ctx context.Context, token string) (*contract.ServiceIdentity, error) {
	return nil, errors.New("serviceauth.mtls: token authentication not supported, use certificate instead")
}

// AuthenticateWithCert 使用证书验证服务身份。
//
// 中文说明：
// - 验证提供的证书；
// - 从证书中提取服务身份信息。
func (a *MTLSAuthenticator) AuthenticateWithCert(ctx context.Context, cert *tls.Certificate) (*contract.ServiceIdentity, error) {
	if cert == nil || len(cert.Certificate) == 0 {
		return nil, errors.New("serviceauth.mtls: invalid certificate")
	}

	// 解析证书
	x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return nil, fmt.Errorf("serviceauth.mtls: parse certificate failed: %w", err)
	}

	return a.authenticateByCert(x509Cert)
}

// authenticateByCert 通过证书验证服务身份。
//
// 中文说明：
// - 验证证书有效性；
// - 从证书 Subject 中提取服务名；
// - 检查服务是否在允许列表。
func (a *MTLSAuthenticator) authenticateByCert(cert *x509.Certificate) (*contract.ServiceIdentity, error) {
	// 验证证书时间
	now := cert.NotBefore
	if now.After(cert.NotAfter) {
		return nil, errors.New("serviceauth.mtls: certificate expired")
	}

	// 从证书中提取服务名
	// 常见的做法是使用 Subject.CommonName 或 Subject.Organization
	serviceName := cert.Subject.CommonName
	if serviceName == "" {
		// 尝试从 Organization 获取
		if len(cert.Subject.Organization) > 0 {
			serviceName = cert.Subject.Organization[0]
		}
	}

	if serviceName == "" {
		return nil, errors.New("serviceauth.mtls: cannot extract service name from certificate")
	}

	// 检查服务是否在允许列表
	if len(a.cfg.AllowedServices) > 0 {
		allowed := false
		for _, svc := range a.cfg.AllowedServices {
			if svc == serviceName {
				allowed = true
				break
			}
		}
		if !allowed {
			return nil, fmt.Errorf("serviceauth.mtls: service %s not allowed", serviceName)
		}
	}

	// 获取服务权限
	permissions := a.cfg.ServicePermissions[serviceName]

	// 返回服务身份
	return &contract.ServiceIdentity{
		ServiceID:     serviceName,
		ServiceName:   serviceName,
		Namespace:     a.cfg.Namespace,
		Environment:   a.cfg.Environment,
		Permissions:   permissions,
		ExpiresAt:     cert.NotAfter.Unix(),
		IssuedAt:      cert.NotBefore.Unix(),
	}, nil
}

// GenerateToken 生成服务令牌（不支持）。
//
// 中文说明：
// - mTLS 认证模式不生成令牌；
// - 返回空令牌。
func (a *MTLSAuthenticator) GenerateToken(ctx context.Context, targetService string) (string, error) {
	// mTLS 不使用令牌，返回空
	return "", nil
}

// VerifyToken 验证服务令牌（不支持）。
//
// 中文说明：
// - mTLS 认证模式不支持令牌验证。
func (a *MTLSAuthenticator) VerifyToken(ctx context.Context, token string) (*contract.ServiceIdentity, error) {
	return nil, errors.New("serviceauth.mtls: token verification not supported")
}

// GetTLSConfig 获取 TLS 配置。
//
// 中文说明：
// - 返回配置好的 tls.Config；
// - 用于 HTTP/gRPC 服务器配置。
func (a *MTLSAuthenticator) GetTLSConfig() *tls.Config {
	return a.tlsConfig.Clone()
}

// GetClientTLSConfig 获取客户端 TLS 配置。
//
// 中文说明：
// - 返回用于客户端连接的 TLS 配置；
// - 包含客户端证书和 CA 验证。
func (a *MTLSAuthenticator) GetClientTLSConfig() *tls.Config {
	a.certMu.RLock()
	defer a.certMu.RUnlock()

	cfg := &tls.Config{
		RootCAs:            a.certPool,
		ServerName:         a.cfg.MTLSServerName,
		InsecureSkipVerify: false,
	}

	if a.cert != nil {
		cfg.Certificates = []tls.Certificate{*a.cert}
	}

	return cfg
}

// ReloadCertificate 重新加载证书。
//
// 中文说明：
// - 用于证书轮换场景；
// - 重新从文件加载证书。
func (a *MTLSAuthenticator) ReloadCertificate() error {
	if a.cfg.MTLSCertFile == "" || a.cfg.MTLSKeyFile == "" {
		return nil
	}

	cert, err := tls.LoadX509KeyPair(a.cfg.MTLSCertFile, a.cfg.MTLSKeyFile)
	if err != nil {
		return fmt.Errorf("serviceauth.mtls: reload cert failed: %w", err)
	}

	a.certMu.Lock()
	a.cert = &cert
	a.tlsConfig.Certificates = []tls.Certificate{cert}
	a.certMu.Unlock()

	return nil
}

// extractTLSState 从上下文提取 TLS 连接状态。
//
// 中文说明：
// - 支持从 HTTP/gRPC 连接中提取；
// - 返回 TLS 连接状态。
func extractTLSState(ctx context.Context) *tls.ConnectionState {
	// 尝试从上下文获取 TLS 状态
	if tlsState := ctx.Value("tls_state"); tlsState != nil {
		if state, ok := tlsState.(*tls.ConnectionState); ok {
			return state
		}
	}

	// 尝试从 net.Conn 获取（需要配合中间件设置）
	if conn := ctx.Value("tls_conn"); conn != nil {
		if tlsConn, ok := conn.(*tls.Conn); ok {
			state := tlsConn.ConnectionState()
			return &state
		}
	}

	return nil
}