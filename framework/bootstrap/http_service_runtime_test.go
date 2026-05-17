// Package bootstrap_test provides integration and boundary tests for HTTP service runtime bootstrapping.
//
// 适用场景：
// - 验证 HTTP Service runtime 的初始化、governance override 和 provider 注册行为。
// - 验证 pprof / governance inspect 端点的路由注册与响应格式。
// - 验证 governance summary 与 diagnostic 的构建与格式化逻辑。
package bootstrap

import (
	"errors"
	"testing"

	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Runtime 初始化与 Governance Override 测试
// =============================================================================

func TestNewHTTPServiceRuntimeLeavesGovernanceOverrideEmptyByDefault(t *testing.T) {
	originProviders := buildHTTPProvidersFunc
	origin := registerSelectedMicroserviceProvidersWithMode
	originWithOptions := registerSelectedMicroserviceProvidersWithOptionsFunc
	defer func() {
		buildHTTPProvidersFunc = originProviders
		registerSelectedMicroserviceProvidersWithMode = origin
		registerSelectedMicroserviceProvidersWithOptionsFunc = originWithOptions
	}()

	var gotMode string
	sentinel := errors.New("stop after capture")
	buildHTTPProvidersFunc = func(opts HTTPServiceOptions) []runtimecontract.ServiceProvider {
		return nil
	}
	registerSelectedMicroserviceProvidersWithOptionsFunc = func(c runtimecontract.Container, modeOverride string, disabled []string, enabled []string, providers map[string]string) error {
		gotMode = modeOverride
		return sentinel
	}

	_, err := NewHTTPServiceRuntime(HTTPServiceOptions{})
	require.Error(t, err)
	require.ErrorIs(t, err, sentinel)
	require.Empty(t, gotMode)
}

func TestNewHTTPServiceRuntimeForwardsMicroserviceGovernanceOverride(t *testing.T) {
	originProviders := buildHTTPProvidersFunc
	origin := registerSelectedMicroserviceProvidersWithMode
	originWithOptions := registerSelectedMicroserviceProvidersWithOptionsFunc
	defer func() {
		buildHTTPProvidersFunc = originProviders
		registerSelectedMicroserviceProvidersWithMode = origin
		registerSelectedMicroserviceProvidersWithOptionsFunc = originWithOptions
	}()

	var gotMode string
	sentinel := errors.New("stop after capture")
	buildHTTPProvidersFunc = func(opts HTTPServiceOptions) []runtimecontract.ServiceProvider {
		return nil
	}
	registerSelectedMicroserviceProvidersWithOptionsFunc = func(c runtimecontract.Container, modeOverride string, disabled []string, enabled []string, providers map[string]string) error {
		gotMode = modeOverride
		return sentinel
	}

	_, err := NewHTTPServiceRuntime(HTTPServiceOptions{
		GovernanceMode: string(resiliencecontract.GovernanceModeMicro),
	})
	require.Error(t, err)
	require.ErrorIs(t, err, sentinel)
	require.Equal(t, string(resiliencecontract.GovernanceModeMicro), gotMode)
}

func TestNewHTTPServiceRuntimeForwardsGovernanceDisableAndProviderOverrides(t *testing.T) {
	originProviders := buildHTTPProvidersFunc
	originWithOptions := registerSelectedMicroserviceProvidersWithOptionsFunc
	defer func() {
		buildHTTPProvidersFunc = originProviders
		registerSelectedMicroserviceProvidersWithOptionsFunc = originWithOptions
	}()

	var (
		gotDisabled  []string
		gotProviders map[string]string
	)
	sentinel := errors.New("stop after capture")
	buildHTTPProvidersFunc = func(opts HTTPServiceOptions) []runtimecontract.ServiceProvider { return nil }
	registerSelectedMicroserviceProvidersWithOptionsFunc = func(c runtimecontract.Container, modeOverride string, disabled []string, enabled []string, providers map[string]string) error {
		gotDisabled = append([]string(nil), disabled...)
		gotProviders = providers
		return sentinel
	}

	_, err := NewHTTPServiceRuntime(HTTPServiceOptions{
		GovernanceDisable:   []string{"tracing"},
		GovernanceProviders: map[string]string{"serviceauth": "mtls"},
	})
	require.Error(t, err)
	require.ErrorIs(t, err, sentinel)
	require.Equal(t, []string{"tracing"}, gotDisabled)
	require.Equal(t, "mtls", gotProviders["serviceauth"])
}
