// Package cron_test provides unit tests for the cron provider.
//
// 适用场景：
// - 验证 Cron provider 的注册与契约实现。
package cron

import (
	"testing"

	"github.com/ngq/gorp/framework/container"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	"github.com/stretchr/testify/require"
)

// TestProviderContract verifies that the cron provider has correct name, defer behavior, and provided keys.
//
// TestProviderContract 验证 cron provider 的名称、延迟加载行为和提供的键。
func TestProviderContract(t *testing.T) {
	p := NewProvider()
	require.Equal(t, "cron", p.Name())
	require.False(t, p.IsDefer())
	require.Equal(t, []string{runtimecontract.CronKey}, p.Provides())
}

// TestCronProvider_RegisterCloserAndDestroy verifies that cron closer is registered and stopped on container destroy.
func TestCronProvider_RegisterCloserAndDestroy(t *testing.T) {
	c := container.New()
	p := NewProvider()
	require.NoError(t, p.Register(c))

	cronAny, err := c.Make(runtimecontract.CronKey)
	require.NoError(t, err)
	cronSvc, ok := cronAny.(runtimecontract.Cron)
	require.True(t, ok)
	cronSvc.Start()

	// Destroying container should stop cron without errors
	require.NoError(t, c.Destroy())
}
