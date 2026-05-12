// Package cron_test provides unit tests for the cron provider.
//
// 适用场景：
// - 验证 Cron provider 的注册与契约实现。
package cron

import (
	"testing"

	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	"github.com/stretchr/testify/require"
)

func TestProviderContract(t *testing.T) {
	p := NewProvider()
	require.Equal(t, "cron", p.Name())
	require.False(t, p.IsDefer())
	require.Equal(t, []string{runtimecontract.CronKey}, p.Provides())
}
