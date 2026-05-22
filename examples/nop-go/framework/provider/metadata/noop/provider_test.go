// Package noop_test provides unit tests for the metadata noop provider.
//
// 适用场景：
// - 验证 Metadata noop provider 的注册与空操作行为。
package noop

import (
	"testing"

	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	"github.com/stretchr/testify/require"
)

// TestProviderContract verifies metadata noop provider registration and no-op behavior.
//
// TestProviderContract 验证 metadata noop provider 的注册与空操作行为。
func TestProviderContract(t *testing.T) {
	p := NewProvider()
	require.Equal(t, "metadata.noop", p.Name())
	require.True(t, p.IsDefer())
	require.ElementsMatch(t, []string{transportcontract.MetadataKey, transportcontract.MetadataPropagatorKey}, p.Provides())
}
