// Package noop_test provides unit tests for the RPC noop provider.
//
// 适用场景：
// - 验证 RPC noop provider 的注册与空操作行为。
package noop

import (
	"context"
	"testing"

	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	"github.com/stretchr/testify/assert"
)

func TestNoopClient(t *testing.T) {
	client := &noopClient{}

	// 测试 Call 返回错误
	err := client.Call(context.Background(), "user-service", "GetUser", nil, nil)
	assert.ErrorIs(t, err, ErrNoopRPC)

	// 测试 CallRaw 返回错误
	data, err := client.CallRaw(context.Background(), "user-service", "GetUser", nil)
	assert.Nil(t, data)
	assert.ErrorIs(t, err, ErrNoopRPC)

	// 测试 Close 无错误
	err = client.Close()
	assert.NoError(t, err)
}

func TestNoopServer(t *testing.T) {
	server := &noopServer{}

	// 测试 Register 无错误
	err := server.Register("user-service", nil)
	assert.NoError(t, err)

	// 测试 Start 返回错误
	err = server.Start(context.Background())
	assert.ErrorIs(t, err, ErrNoopRPC)

	// 测试 Stop 无错误
	err = server.Stop(context.Background())
	assert.NoError(t, err)

	// 测试 Addr 返回空
	assert.Empty(t, server.Addr())
}

func TestNoopRegistry(t *testing.T) {
	registry := &noopRegistry{}

	// 测试 Register 无错误
	err := registry.Register(context.Background(), "user-service", "localhost:8080", nil)
	assert.NoError(t, err)

	// 测试 Deregister 无错误
	err = registry.Deregister(context.Background(), "user-service", "localhost:8080")
	assert.NoError(t, err)

	// 测试 Discover 返回空列表
	instances, err := registry.Discover(context.Background(), "user-service")
	assert.NoError(t, err)
	assert.Empty(t, instances)

	// 测试 Close 无错误
	err = registry.Close()
	assert.NoError(t, err)
}

// TestProvider verifies that the rpc noop provider provides the correct keys and deferred status.
//
// TestProvider 验证 rpc noop provider 提供正确的键和延迟状态。
func TestProvider(t *testing.T) {
	p := NewProvider()

	assert.Equal(t, "rpc.noop", p.Name())
	assert.True(t, p.IsDefer())
	assert.ElementsMatch(t, []string{
		transportcontract.RPCClientKey,
		transportcontract.RPCServerKey,
		transportcontract.RPCRegistryKey,
	}, p.Provides())
}
