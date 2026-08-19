package security_test

import (
	"testing"

	securityprovider "github.com/ngq/gorp/framework/provider/security"
	"github.com/stretchr/testify/require"
)

func TestMemoryAuthorizer_Enforce(t *testing.T) {
	authorizer := securityprovider.NewMemoryAuthorizer(map[string][]string{
		"admin": {"*"},
		"user":  {"GET:/api/v1/orders", "POST:/api/v1/orders"},
	})

	ctx := t.Context()

	// 1. admin -> wildcard allow
	allowed1, err := authorizer.Enforce(ctx, "admin", "/any/path", "DELETE")
	require.NoError(t, err)
	require.True(t, allowed1)

	// 2. user -> allowed GET /api/v1/orders
	allowed2, err := authorizer.Enforce(ctx, "user", "/api/v1/orders", "GET")
	require.NoError(t, err)
	require.True(t, allowed2)

	// 3. user -> denied DELETE /api/v1/orders
	allowed3, err := authorizer.Enforce(ctx, "user", "/api/v1/orders", "DELETE")
	require.NoError(t, err)
	require.False(t, allowed3)
}
