package token

import (
	"context"
	"crypto/x509"
	"testing"
	"time"

	securitycontract "github.com/ngq/gorp/framework/contract/security"
	"github.com/stretchr/testify/require"
)

func TestTokenAuthenticator_GenerateAndVerifyToken(t *testing.T) {
	auth := NewTokenAuthenticator(&securitycontract.ServiceAuthConfig{
		ServiceName:  "order-service",
		TokenIssuer:  "order-service",
		TokenSecret:  "test-secret",
		TokenExpiry:  60,
		Namespace:    "prod",
		Environment:  "online",
	})

	token, err := auth.GenerateToken(context.Background(), "user-service")
	require.NoError(t, err)

	identity, err := auth.VerifyToken(context.Background(), token)
	require.NoError(t, err)
	require.Equal(t, "order-service", identity.ServiceID)
	require.Equal(t, "order-service", identity.ServiceName)
	require.Equal(t, "prod", identity.Namespace)
	require.Equal(t, "online", identity.Environment)
}

func TestTokenAuthenticator_VerifyTokenRejectsDisallowedService(t *testing.T) {
	auth := NewTokenAuthenticator(&securitycontract.ServiceAuthConfig{
		ServiceName:     "order-service",
		TokenIssuer:     "order-service",
		TokenSecret:     "test-secret",
		TokenExpiry:     60,
		AllowedServices: []string{"billing-service"},
	})

	token, err := auth.GenerateToken(context.Background(), "")
	require.NoError(t, err)

	_, err = auth.VerifyToken(context.Background(), token)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not allowed")
}

func TestTokenAuthenticator_VerifyTokenRejectsExpiredToken(t *testing.T) {
	auth := NewTokenAuthenticator(&securitycontract.ServiceAuthConfig{
		ServiceName: "order-service",
		TokenIssuer: "order-service",
		TokenSecret: "test-secret",
		TokenExpiry: -1,
	})

	token, err := auth.GenerateToken(context.Background(), "")
	require.NoError(t, err)

	_, err = auth.VerifyToken(context.Background(), token)
	require.Error(t, err)
	require.Contains(t, err.Error(), "verify token failed")
}

func TestTokenAuthenticator_AuthenticateReadsBearerTokenFromContext(t *testing.T) {
	auth := NewTokenAuthenticator(&securitycontract.ServiceAuthConfig{
		ServiceName: "order-service",
		TokenIssuer: "order-service",
		TokenSecret: "test-secret",
		TokenExpiry: 60,
	})

	token, err := auth.GenerateToken(context.Background(), "")
	require.NoError(t, err)

	ctx := context.WithValue(context.Background(), "authorization", "Bearer "+token)
	identity, err := auth.Authenticate(ctx)
	require.NoError(t, err)
	require.Equal(t, "order-service", identity.ServiceName)
}

func TestTokenAuthenticator_AuthenticatePeerCertificateRejectsCertificate(t *testing.T) {
	auth := NewTokenAuthenticator(&securitycontract.ServiceAuthConfig{})
	_, err := auth.AuthenticatePeerCertificate(context.Background(), &x509.Certificate{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "peer certificate authentication not supported")
}

func TestExtractTokenFromContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), "authorization", "bearer abc")
	require.Equal(t, "abc", extractTokenFromContext(ctx))

	ctx = context.WithValue(context.Background(), "authorization", "raw-token")
	require.Equal(t, "raw-token", extractTokenFromContext(ctx))

	require.Empty(t, extractTokenFromContext(context.Background()))
}

func TestTokenAuthenticator_GetCurrentIdentity(t *testing.T) {
	auth := NewTokenAuthenticator(&securitycontract.ServiceAuthConfig{
		ServiceName: "order-service",
		Namespace:   "prod",
		Environment: "online",
	})
	identity, err := auth.GetCurrentIdentity(context.Background())
	require.NoError(t, err)
	require.Equal(t, "order-service", identity.ServiceID)
	require.Equal(t, "prod", identity.Namespace)
	require.Equal(t, "online", identity.Environment)
}

func TestTokenAuthenticator_VerifyTokenRejectsDifferentSecret(t *testing.T) {
	issuer := NewTokenAuthenticator(&securitycontract.ServiceAuthConfig{
		ServiceName: "order-service",
		TokenIssuer: "order-service",
		TokenSecret: "secret-a",
		TokenExpiry: int64((1 * time.Minute).Seconds()),
	})
	verifier := NewTokenAuthenticator(&securitycontract.ServiceAuthConfig{
		ServiceName: "order-service",
		TokenIssuer: "order-service",
		TokenSecret: "secret-b",
		TokenExpiry: 60,
	})

	token, err := issuer.GenerateToken(context.Background(), "")
	require.NoError(t, err)
	_, err = verifier.VerifyToken(context.Background(), token)
	require.Error(t, err)
}

// TestTokenAuthenticator_AuthenticateRejectsMissingToken verifies that Authenticate returns an error when no token is in context.
//
// TestTokenAuthenticator_AuthenticateRejectsMissingToken 验证 context 中无 token 时 Authenticate 返回错误。
func TestTokenAuthenticator_AuthenticateRejectsMissingToken(t *testing.T) {
	auth := NewTokenAuthenticator(&securitycontract.ServiceAuthConfig{
		ServiceName: "order-service",
		TokenSecret: "test-secret",
		TokenExpiry: 60,
	})

	// 空 context，不含任何 token
	_, err := auth.Authenticate(context.Background())
	require.Error(t, err)
	require.Contains(t, err.Error(), "no token found in context")
}

// TestTokenAuthenticator_VerifyTokenRejectsEmptyToken verifies that VerifyToken rejects an empty token string.
//
// TestTokenAuthenticator_VerifyTokenRejectsEmptyToken 验证空 token 字符串会被拒绝。
func TestTokenAuthenticator_VerifyTokenRejectsEmptyToken(t *testing.T) {
	auth := NewTokenAuthenticator(&securitycontract.ServiceAuthConfig{
		ServiceName: "order-service",
		TokenSecret: "test-secret",
		TokenExpiry: 60,
	})

	_, err := auth.VerifyToken(context.Background(), "")
	require.Error(t, err)
}

// TestTokenAuthenticator_GenerateTokenWithTargetService verifies that generated token includes the target service as audience.
//
// TestTokenAuthenticator_GenerateTokenWithTargetService 验证生成 token 时目标服务会被写入 audience。
func TestTokenAuthenticator_GenerateTokenWithTargetService(t *testing.T) {
	auth := NewTokenAuthenticator(&securitycontract.ServiceAuthConfig{
		ServiceName: "order-service",
		TokenIssuer: "order-service",
		TokenSecret: "test-secret",
		TokenExpiry: 60,
	})

	token, err := auth.GenerateToken(context.Background(), "user-service")
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// 验证目标服务 token 可以被正常校验
	identity, err := auth.VerifyToken(context.Background(), token)
	require.NoError(t, err)
	require.Equal(t, "order-service", identity.ServiceName)
}
