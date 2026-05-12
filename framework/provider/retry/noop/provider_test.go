// Package noop_test provides unit tests for the retry noop provider.
//
// 适用场景：
// - 验证重试 noop provider 的注册与空操作行为。
package noop

import (
	"testing"

	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	"github.com/stretchr/testify/require"
)

func TestProviderContract(t *testing.T) {
	p := NewProvider()
	require.Equal(t, "retry.noop", p.Name())
	require.True(t, p.IsDefer())
	require.Equal(t, []string{resiliencecontract.RetryKey}, p.Provides())
}
