// Package metadata_test provides unit tests for the metadata provider.
//
// 适用场景：
// - 验证 Metadata provider 的注册与契约实现。
package metadata

import (
	"testing"

	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	"github.com/stretchr/testify/require"
)

// TestProviderContract verifies metadata default provider registration and contract compliance.
//
// TestProviderContract 验证 metadata 默认 provider 的注册与契约实现。
func TestProviderContract(t *testing.T) {
	p := NewProvider()
	require.Equal(t, "metadata.default", p.Name())
	require.True(t, p.IsDefer())
	require.ElementsMatch(t, []string{transportcontract.MetadataKey, transportcontract.MetadataPropagatorKey}, p.Provides())
}
