package mtls

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	securitycontract "github.com/ngq/gorp/framework/contract/security"
	"github.com/stretchr/testify/require"
)

func TestMTLSAuthenticator_AuthenticateRejectsMissingTLSState(t *testing.T) {
	auth, err := NewMTLSAuthenticator(&securitycontract.ServiceAuthConfig{})
	require.NoError(t, err)

	_, err = auth.Authenticate(context.Background())
	require.Error(t, err)
	require.Contains(t, err.Error(), "no TLS connection found")
}

func TestMTLSAuthenticator_AuthenticateRejectsMissingPeerCertificate(t *testing.T) {
	auth, err := NewMTLSAuthenticator(&securitycontract.ServiceAuthConfig{})
	require.NoError(t, err)

	ctx := context.WithValue(context.Background(), "tls_state", &tls.ConnectionState{})
	_, err = auth.Authenticate(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "no client certificate provided")
}

func TestMTLSAuthenticator_AuthenticatePeerCertificateExtractsIdentity(t *testing.T) {
	cert := testX509Certificate(t, "user-service")
	auth, err := NewMTLSAuthenticator(&securitycontract.ServiceAuthConfig{
		Namespace:   "prod",
		Environment: "online",
	})
	require.NoError(t, err)

	identity, err := auth.AuthenticatePeerCertificate(context.Background(), cert)
	require.NoError(t, err)
	require.Equal(t, "user-service", identity.ServiceID)
	require.Equal(t, "user-service", identity.ServiceName)
	require.Equal(t, "prod", identity.Namespace)
	require.Equal(t, "online", identity.Environment)
}

func TestMTLSAuthenticator_AuthenticatePeerCertificateRejectsInvalidCertificate(t *testing.T) {
	auth, err := NewMTLSAuthenticator(&securitycontract.ServiceAuthConfig{})
	require.NoError(t, err)

	_, err = auth.AuthenticatePeerCertificate(context.Background(), nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "nil certificate")
}

func TestMTLSAuthenticator_GetCurrentIdentity(t *testing.T) {
	auth, err := NewMTLSAuthenticator(&securitycontract.ServiceAuthConfig{
		ServiceName: "order-service",
		Namespace:   "prod",
		Environment: "online",
	})
	require.NoError(t, err)

	identity, err := auth.GetCurrentIdentity(context.Background())
	require.NoError(t, err)
	require.Equal(t, "order-service", identity.ServiceName)
	require.Equal(t, "prod", identity.Namespace)
}

func TestNewMTLSAuthenticator_LoadsCAAndKeyPair(t *testing.T) {
	dir := t.TempDir()
	caPath, certPath, keyPath := writeTestMTLSFiles(t, dir, "svc-a")

	auth, err := NewMTLSAuthenticator(&securitycontract.ServiceAuthConfig{
		MTLSCAFile:   caPath,
		MTLSCertFile: certPath,
		MTLSKeyFile:  keyPath,
	})
	require.NoError(t, err)
	require.NotNil(t, auth.certPool)
	require.NotNil(t, auth.cert)
	require.Len(t, auth.tlsConfig.Certificates, 1)
}

// TestMTLSAuthenticator_AuthenticatePeerCertificateRejectsMissingCommonName verifies that a certificate without CN is rejected.
//
// TestMTLSAuthenticator_AuthenticatePeerCertificateRejectsMissingCommonName 验证缺少 CommonName 的证书会被拒绝。
func TestMTLSAuthenticator_AuthenticatePeerCertificateRejectsMissingCommonName(t *testing.T) {
	auth, err := NewMTLSAuthenticator(&securitycontract.ServiceAuthConfig{
		Namespace:   "prod",
		Environment: "online",
	})
	require.NoError(t, err)

	// 构建一个没有 CommonName 的证书
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject:      pkix.Name{}, // CommonName 为空
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &priv.PublicKey, priv)
	require.NoError(t, err)
	parsedCert, err := x509.ParseCertificate(der)
	require.NoError(t, err)

	_, err = auth.AuthenticatePeerCertificate(context.Background(), parsedCert)
	require.Error(t, err)
	require.Contains(t, err.Error(), "certificate missing common name")
}

// TestMTLSAuthenticator_AuthenticateExtractsIdentityFromTLSPeerCert verifies that Authenticate extracts identity from a real TLS peer certificate.
//
// TestMTLSAuthenticator_AuthenticateExtractsIdentityFromTLSPeerCert 验证 Authenticate 能从真实 TLS 对端证书中提取身份。
func TestMTLSAuthenticator_AuthenticateExtractsIdentityFromTLSPeerCert(t *testing.T) {
	auth, err := NewMTLSAuthenticator(&securitycontract.ServiceAuthConfig{
		Namespace:   "prod",
		Environment: "online",
	})
	require.NoError(t, err)

	// 构建一个带 CommonName 的证书，放入 TLS ConnectionState 的 PeerCertificates
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(3),
		Subject:      pkix.Name{CommonName: "payment-service"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &priv.PublicKey, priv)
	require.NoError(t, err)
	parsedCert, err := x509.ParseCertificate(der)
	require.NoError(t, err)

	connState := &tls.ConnectionState{
		PeerCertificates: []*x509.Certificate{parsedCert},
	}
	ctx := context.WithValue(context.Background(), "tls_state", connState)

	identity, err := auth.Authenticate(ctx)
	require.NoError(t, err)
	require.Equal(t, "payment-service", identity.ServiceID)
	require.Equal(t, "payment-service", identity.ServiceName)
	require.Equal(t, "prod", identity.Namespace)
	require.Equal(t, "online", identity.Environment)
}

// TestNewMTLSAuthenticator_RejectsInvalidCAFile verifies that loading a nonexistent CA file returns an error.
//
// TestNewMTLSAuthenticator_RejectsInvalidCAFile 验证加载不存在的 CA 文件会返回错误。
func TestNewMTLSAuthenticator_RejectsInvalidCAFile(t *testing.T) {
	_, err := NewMTLSAuthenticator(&securitycontract.ServiceAuthConfig{
		MTLSCAFile: "/nonexistent/path/ca.pem",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "load CA cert failed")
}

// TestNewMTLSAuthenticator_RejectsInvalidKeyPair verifies that loading an invalid cert/key pair returns an error.
//
// TestNewMTLSAuthenticator_RejectsInvalidKeyPair 验证加载无效的证书/密钥对会返回错误。
func TestNewMTLSAuthenticator_RejectsInvalidKeyPair(t *testing.T) {
	dir := t.TempDir()
	// 写入无效的 cert 和 key 文件
	certPath := filepath.Join(dir, "bad_cert.pem")
	keyPath := filepath.Join(dir, "bad_key.pem")
	require.NoError(t, os.WriteFile(certPath, []byte("not a valid cert"), 0o600))
	require.NoError(t, os.WriteFile(keyPath, []byte("not a valid key"), 0o600))

	_, err := NewMTLSAuthenticator(&securitycontract.ServiceAuthConfig{
		MTLSCertFile: certPath,
		MTLSKeyFile:  keyPath,
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "load cert failed")
}

func testX509Certificate(t *testing.T, cn string) *x509.Certificate {
	t.Helper()
	_, cert, _ := generateCertificatePEM(t, cn)
	parsed, err := tls.X509KeyPair(cert, cert)
	require.Error(t, err)
	_ = parsed

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: cn},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &priv.PublicKey, priv)
	require.NoError(t, err)
	parsedCert, err := x509.ParseCertificate(der)
	require.NoError(t, err)
	return parsedCert
}

func writeTestMTLSFiles(t *testing.T, dir, cn string) (string, string, string) {
	t.Helper()
	keyPEM, certPEM, caPEM := generateCertificatePEM(t, cn)
	caPath := filepath.Join(dir, "ca.pem")
	certPath := filepath.Join(dir, "cert.pem")
	keyPath := filepath.Join(dir, "key.pem")
	require.NoError(t, os.WriteFile(caPath, caPEM, 0o600))
	require.NoError(t, os.WriteFile(certPath, certPEM, 0o600))
	require.NoError(t, os.WriteFile(keyPath, keyPEM, 0o600))
	return caPath, certPath, keyPath
}

func generateCertificatePEM(t *testing.T, cn string) ([]byte, []byte, []byte) {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: cn},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment | x509.KeyUsageCertSign,
		IsCA:                  true,
		BasicConstraintsValid: true,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &priv.PublicKey, priv)
	require.NoError(t, err)
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	return keyPEM, certPEM, certPEM
}
