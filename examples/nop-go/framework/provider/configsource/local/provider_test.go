// Package local_test provides unit tests for the local config source provider.
//
// 适用场景：
// - 验证本地配置源 provider 的注册与本地配置读取行为。
package local

import (
	"testing"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	"github.com/stretchr/testify/require"
)

// TestProviderContract verifies that the local config source provider has correct name, defer behavior, and provided keys.
//
// TestProviderContract 验证本地配置源 provider 的名称、延迟加载行为和提供的键。
func TestProviderContract(t *testing.T) {
	p := NewProvider()
	require.Equal(t, "configsource.local", p.Name())
	require.True(t, p.IsDefer())
	require.Equal(t, []string{datacontract.ConfigSourceKey}, p.Provides())
}
