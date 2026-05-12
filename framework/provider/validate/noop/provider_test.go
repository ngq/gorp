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

func TestProviderContract(t *testing.T) {
	p := NewProvider()
	require.Equal(t, "validate.noop", p.Name())
	require.True(t, p.IsDefer())
	require.Equal(t, []string{datacontract.ValidatorKey}, p.Provides())
}
