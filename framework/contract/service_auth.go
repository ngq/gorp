package contract

import (
	"context"
	"crypto/tls"
	"crypto/x509"
)

const (
	// ServiceAuthKey 是服务认证器在容器中的绑定 key。
	//
	// 中文说明：
	// - 用于服务间身份验证；
	// - 支持 mTLS、服务令牌等多种认证方式；
	// - noop 实现不验证身份，单体项目零依赖。
	ServiceAuthKey = "framework.service.auth"

	// ServiceIdentityKey 是服务身份在容器中的绑定 key。
	//
	// 中文说明：
	// - 当前服务的身份标识；
	// - 包含服务名、证书、令牌等信息。
	ServiceIdentityKey = "framework.service.identity"
)

// ServiceAuthenticator 服务认证器接口。
//
// 中文说明：
// - 验证服务间调用的身份；
// - 支持 mTLS 双向认证；
// - 支持服务令牌认证；
// - 与 HTTP/gRPC 拦截器集成。
type ServiceAuthenticator interface {
	// Authenticate 验证服务身份。
	//
	// 中文说明：
	// - ctx: 请求上下文；
	// - 返回验证后的服务身份信息；
	// - 认证失败返回错误。
	Authenticate(ctx context.Context) (*ServiceIdentity, error)

	// AuthenticateWithToken 使用令牌验证服务身份。
	//
	// 中文说明：
	// - 用于服务令牌认证方式；
	// - token: 服务令牌（如 JWT）。
	AuthenticateWithToken(ctx context.Context, token string) (*ServiceIdentity, error)

	// AuthenticateWithCert 使用证书验证服务身份。
	//
	// 中文说明：
	// - 用于 mTLS 认证方式；
	// - cert: 客户端证书。
	AuthenticateWithCert(ctx context.Context, cert *tls.Certificate) (*ServiceIdentity, error)

	// GenerateToken 生成服务令牌。
	//
	// 中文说明：
	// - 为当前服务生成访问令牌；
	// - 用于服务间调用时携带身份凭证。
	GenerateToken(ctx context.Context, targetService string) (string, error)

	// VerifyToken 验证服务令牌。
	//
	// 中文说明：
	// - 验证令牌的有效性和签名；
	// - 返回令牌对应的服务身份。
	VerifyToken(ctx context.Context, token string) (*ServiceIdentity, error)
}

// ServiceIdentity 服务身份信息。
//
// 中文说明：
// - 表示一个服务的身份标识；
// - 只保留服务认证与身份识别所需的最小字段；
// - 不在 framework 合同层内解释服务权限语义。
type ServiceIdentity struct {
	// ServiceID 服务唯一标识
	ServiceID string

	// ServiceName 服务名称
	ServiceName string

	// Namespace 命名空间（可选）
	Namespace string

	// Environment 环境标识
	Environment string

	// Metadata 服务元数据
	Metadata map[string]string

	// ExpiresAt 令牌过期时间（可选）
	ExpiresAt int64

	// IssuedAt 令牌签发时间
	IssuedAt int64

	// Issuer 令牌签发者
	Issuer string
}

// ServiceAuthConfig 服务认证配置。
type ServiceAuthConfig struct {
	// Mode 认证模式：noop/mtls/token
	Mode string

	// ServiceName 当前服务名称
	ServiceName string

	// Namespace 命名空间
	Namespace string

	// Environment 环境标识
	Environment string

	// mTLS 配置
	MTLSEnabled    bool
	MTLSCertFile   string
	MTLSKeyFile    string
	MTLSCAFile     string
	MTLSServerName string

	// 服务令牌配置
	TokenSecret   string // 令牌签名密钥
	TokenExpiry   int64  // 令牌有效期（秒）
	TokenIssuer   string // 令牌签发者
	TokenAudience string // 令牌受众

	// 允许的服务列表
	AllowedServices []string
}

// ServiceAuthProvider 服务认证 Provider 接口。
//
// 中文说明：
// - 扩展 ServiceProvider 接口；
// - 提供创建 Authenticator 的工厂方法。
type ServiceAuthProvider interface {
	ServiceProvider

	// CreateAuthenticator 创建认证器实例。
	CreateAuthenticator(cfg *ServiceAuthConfig) (ServiceAuthenticator, error)
}

// TLSCertificateLoader TLS 证书加载接口。
//
// 中文说明：
// - 用于加载 TLS 证书；
// - 支持从文件、Secret 等来源加载。
type TLSCertificateLoader interface {
	LoadCertificate(certFile, keyFile string) (*tls.Certificate, error)
	LoadCA(caFile string) (*x509.CertPool, error)
}