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

	"github.com/ngq/gorp/framework/contract"
	"github.com/stretchr/testify/require"
)

func TestMTLSAuthenticator_AuthenticateRejectsMissingTLSState(t *testing.T) {
	auth, err := NewMTLSAuthenticator(&contract.ServiceAuthConfig{})
	require.NoError(t, err)

	_, err = auth.Authenticate(context.Background())
	require.Error(t, err)
	require.Contains(t, err.Error(), "no TLS connection found")
}

func TestMTLSAuthenticator_AuthenticateRejectsMissingPeerCertificate(t *testing.T) {
	auth, err := NewMTLSAuthenticator(&contract.ServiceAuthConfig{})
	require.NoError(t, err)

	ctx := context.WithValue(context.Background(), "tls_state", &tls.ConnectionState{})
	_, err = auth.Authenticate(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "no client certificate provided")
}

func TestMTLSAuthenticator_AuthenticateWithCertExtractsIdentity(t *testing.T) {
	cert := testCertificate(t, "user-service")
	auth, err := NewMTLSAuthenticator(&contract.ServiceAuthConfig{
		Namespace:   "prod",
		Environment: "online",
	})
	require.NoError(t, err)

	identity, err := auth.AuthenticateWithCert(context.Background(), cert)
	require.NoError(t, err)
	require.Equal(t, "user-service", identity.ServiceID)
	require.Equal(t, "user-service", identity.ServiceName)
	require.Equal(t, "prod", identity.Namespace)
	require.Equal(t, "online", identity.Environment)
}

func TestMTLSAuthenticator_AuthenticateWithCertRejectsInvalidCertificate(t *testing.T) {
	auth, err := NewMTLSAuthenticator(&contract.ServiceAuthConfig{})
	require.NoError(t, err)

	_, err = auth.AuthenticateWithCert(context.Background(), &tls.Certificate{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid certificate")
}

func TestMTLSAuthenticator_GetCurrentIdentity(t *testing.T) {
	auth, err := NewMTLSAuthenticator(&contract.ServiceAuthConfig{
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

	auth, err := NewMTLSAuthenticator(&contract.ServiceAuthConfig{
		MTLSCAFile:   caPath,
		MTLSCertFile: certPath,
		MTLSKeyFile:  keyPath,
	})
	require.NoError(t, err)
	require.NotNil(t, auth.certPool)
	require.NotNil(t, auth.cert)
	require.Len(t, auth.tlsConfig.Certificates, 1)
}

func testCertificate(t *testing.T, cn string) *tls.Certificate {
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
	return &tls.Certificate{Certificate: [][]byte{der}, PrivateKey: priv}
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
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: cn},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment | x509.KeyUsageCertSign,
		IsCA:         true,
		BasicConstraintsValid: true,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &priv.PublicKey, priv)
	require.NoError(t, err)
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	return keyPEM, certPEM, certPEM
}
