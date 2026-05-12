// Package noop_test provides unit tests for proto noop provider contract.
//
// 适用场景：
// - 验证 proto noop provider 的 Name、IsDefer、Provides 接口契约。
// - 确保空实现的 provider 符合 ServiceProvider 规范。
package noop

import (
	"testing"

	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
	"github.com/stretchr/testify/require"
)

func TestProviderContract(t *testing.T) {
	p := NewProvider()
	require.Equal(t, "proto.noop", p.Name())
	require.True(t, p.IsDefer())
	require.Equal(t, []string{integrationcontract.ProtoGeneratorKey}, p.Provides())
}
