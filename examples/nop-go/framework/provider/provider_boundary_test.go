// Package provider_test provides boundary tests for provider registration and container integration.
//
// 适用场景：
// - 验证多个 provider 的注册顺序和 container 绑定行为。
// - 确保 app、jwt、cache、host 等 provider 的集成正确。
package provider

import (
	"testing"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	securitycontract "github.com/ngq/gorp/framework/contract/security"
	appprovider "github.com/ngq/gorp/framework/provider/app"
	jwtprovider "github.com/ngq/gorp/framework/provider/auth/jwt"
	cacheprovider "github.com/ngq/gorp/framework/provider/cache"
	hostprovider "github.com/ngq/gorp/framework/provider/host"
	"github.com/stretchr/testify/require"
)

// TestCoreProvidersFollowProviderFirstBoundary verifies that core providers (app, cache, auth.jwt, host)
// follow the provider-first pattern with correct deferred status and provided keys.
//
// TestCoreProvidersFollowProviderFirstBoundary 验证核心 provider（app、cache、auth.jwt、host）
// 遵循 provider-first 模式，具有正确的延迟状态和提供的键。
func TestCoreProvidersFollowProviderFirstBoundary(t *testing.T) {
	tests := []struct {
		name        string
		provider    runtimecontract.ServiceProvider
		deferred    bool
		providesKey []string
	}{
		{name: "app", provider: appprovider.NewProvider(), deferred: false, providesKey: []string{appprovider.AppKey}},
		{name: "cache", provider: cacheprovider.NewProvider(), deferred: false, providesKey: []string{datacontract.CacheKey}},
		{name: "auth.jwt", provider: jwtprovider.NewProvider(), deferred: true, providesKey: []string{securitycontract.AuthJWTKey}},
		{name: "host", provider: hostprovider.NewProvider(), deferred: false, providesKey: []string{runtimecontract.HostKey}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NotEmpty(t, tt.provider.Name())
			require.Equal(t, tt.deferred, tt.provider.IsDefer())
			require.ElementsMatch(t, tt.providesKey, tt.provider.Provides())
		})
	}
}
