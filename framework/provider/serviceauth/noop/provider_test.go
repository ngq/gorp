// Package noop_test provides unit tests for the service auth noop provider.
//
// 适用场景：
// - 验证服务鉴权 noop provider 的注册与空操作行为。
package noop

import (
	"testing"

	securitycontract "github.com/ngq/gorp/framework/contract/security"
	"github.com/stretchr/testify/require"
)

// TestProviderContract verifies that the provider registers with correct name and provided keys.
//
// TestProviderContract 验证 provider 以正确的名称和提供的键注册。
func TestProviderContract(t *testing.T) {
	p := NewProvider()
	require.Equal(t, "serviceauth.noop", p.Name())
	require.True(t, p.IsDefer())
	require.ElementsMatch(t, []string{securitycontract.ServiceAuthKey, securitycontract.ServiceIdentityKey}, p.Provides())
}
