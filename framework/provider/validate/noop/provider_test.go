// Package noop_test provides unit tests for the validate noop provider.
//
// 适用场景：
// - 验证校验 noop provider 的注册与空操作行为。
package noop

import (
	"testing"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	"github.com/stretchr/testify/require"
)

// TestProviderContract verifies the noop provider implements the Provider interface correctly.
//
// TestProviderContract 验证 noop provider 正确实现 Provider 接口。
func TestProviderContract(t *testing.T) {
	p := NewProvider()
	require.Equal(t, "validate.noop", p.Name())
	require.True(t, p.IsDefer())
	require.Equal(t, []string{datacontract.ValidatorKey}, p.Provides())
}
