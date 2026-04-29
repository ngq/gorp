package token

import (
	"context"
	"crypto/tls"
	"testing"
	"time"

	"github.com/ngq/gorp/framework/contract"
	"github.com/stretchr/testify/require"
)

func TestTokenAuthenticator_GenerateAndVerifyToken(t *testing.T) {
	auth := NewTokenAuthenticator(&contract.ServiceAuthConfig{
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
	auth := NewTokenAuthenticator(&contract.ServiceAuthConfig{
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
	auth := NewTokenAuthenticator(&contract.ServiceAuthConfig{
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
	auth := NewTokenAuthenticator(&contract.ServiceAuthConfig{
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

func TestTokenAuthenticator_AuthenticateWithCertRejectsCertificate(t *testing.T) {
	auth := NewTokenAuthenticator(&contract.ServiceAuthConfig{})
	_, err := auth.AuthenticateWithCert(context.Background(), &tls.Certificate{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "certificate authentication not supported")
}

func TestExtractTokenFromContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), "authorization", "bearer abc")
	require.Equal(t, "abc", extractTokenFromContext(ctx))

	ctx = context.WithValue(context.Background(), "authorization", "raw-token")
	require.Equal(t, "raw-token", extractTokenFromContext(ctx))

	require.Empty(t, extractTokenFromContext(context.Background()))
}

func TestTokenAuthenticator_GetCurrentIdentity(t *testing.T) {
	auth := NewTokenAuthenticator(&contract.ServiceAuthConfig{
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
	issuer := NewTokenAuthenticator(&contract.ServiceAuthConfig{
		ServiceName: "order-service",
		TokenIssuer: "order-service",
		TokenSecret: "secret-a",
		TokenExpiry: int64((1 * time.Minute).Seconds()),
	})
	verifier := NewTokenAuthenticator(&contract.ServiceAuthConfig{
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
