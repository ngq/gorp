package provider

import (
	"testing"

	appprovider "github.com/ngq/gorp/framework/provider/app"
	jwtprovider "github.com/ngq/gorp/framework/provider/auth/jwt"
	cacheprovider "github.com/ngq/gorp/framework/provider/cache"
	hostprovider "github.com/ngq/gorp/framework/provider/host"
	"github.com/ngq/gorp/framework/contract"
	"github.com/stretchr/testify/require"
)

func TestCoreProvidersFollowProviderFirstBoundary(t *testing.T) {
	tests := []struct {
		name        string
		provider    contract.ServiceProvider
		deferred    bool
		providesKey []string
	}{
		{name: "app", provider: appprovider.NewProvider(), deferred: false, providesKey: []string{appprovider.AppKey}},
		{name: "cache", provider: cacheprovider.NewProvider(), deferred: false, providesKey: []string{contract.CacheKey}},
		{name: "auth.jwt", provider: jwtprovider.NewProvider(), deferred: true, providesKey: []string{contract.AuthJWTKey}},
		{name: "host", provider: hostprovider.NewProvider(), deferred: false, providesKey: []string{contract.HostKey}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NotEmpty(t, tt.provider.Name())
			require.Equal(t, tt.deferred, tt.provider.IsDefer())
			require.ElementsMatch(t, tt.providesKey, tt.provider.Provides())
		})
	}
}
