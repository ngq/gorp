// Package noop_test provides unit tests for the distributed transaction noop provider.
//
// 适用场景：
// - 验证分布式事务 noop provider 的注册与空操作行为。
package noop

import (
	"context"
	"testing"

	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
	"github.com/stretchr/testify/assert"
)

// TestNoopDTMClient verifies the noop DTM client implementation.
//
// TestNoopDTMClient 验证分布式事务管理客户端的空操作实现。
func TestNoopDTMClient(t *testing.T) {
	client := &noopDTMClient{}

	// 测试 SAGA
	saga := client.SAGA("test-saga")
	assert.NotNil(t, saga)

	// 测试 TCC
	tcc := client.TCC("test-tcc")
	assert.NotNil(t, tcc)

	// 测试 XA
	xa := client.XA("test-xa")
	assert.NotNil(t, xa)

	// 测试 Barrier
	barrier := client.Barrier("saga", "test-gid")
	assert.NotNil(t, barrier)

	// 测试 Query
	info, err := client.Query(context.Background(), "test-gid")
	assert.Nil(t, info)
	assert.ErrorIs(t, err, ErrNoopDTM)
}

// TestNoopSAGABuilder verifies the noop SAGA transaction builder.
//
// TestNoopSAGABuilder 验证 SAGA 事务构建器的空操作实现。
func TestNoopSAGABuilder(t *testing.T) {
	builder := &noopSAGABuilder{}

	// 测试 Add
	result := builder.Add("/action", "/compensate", nil)
	assert.Equal(t, builder, result)

	// 测试 Submit
	err := builder.Submit(context.Background())
	assert.ErrorIs(t, err, ErrNoopDTM)

	// 测试 Build
	tx, err := builder.Build()
	assert.Nil(t, tx)
	assert.ErrorIs(t, err, ErrNoopDTM)
}

// TestNoopTCCBuilder verifies the noop TCC transaction builder.
//
// TestNoopTCCBuilder 验证 TCC 事务构建器的空操作实现。
func TestNoopTCCBuilder(t *testing.T) {
	builder := &noopTCCBuilder{}

	// 测试 Add
	result := builder.Add("/try", "/confirm", "/cancel", nil)
	assert.Equal(t, builder, result)

	// 测试 Submit
	err := builder.Submit(context.Background())
	assert.ErrorIs(t, err, ErrNoopDTM)
}

// TestNoopXABuilder verifies the noop XA transaction builder.
//
// TestNoopXABuilder 验证 XA 事务构建器的空操作实现。
func TestNoopXABuilder(t *testing.T) {
	builder := &noopXABuilder{}

	// 测试 Add
	result := builder.Add("/url", nil)
	assert.Equal(t, builder, result)

	// 测试 Submit
	err := builder.Submit(context.Background())
	assert.ErrorIs(t, err, ErrNoopDTM)
}

// TestNoopBarrierHandler verifies the noop barrier handler.
//
// TestNoopBarrierHandler 验证屏障处理器的空操作实现。
func TestNoopBarrierHandler(t *testing.T) {
	handler := &noopBarrierHandler{}

	// 测试 Call
	executed := false
	err := handler.Call(context.Background(), func(db any) error {
		executed = true
		return nil
	})
	assert.NoError(t, err)
	assert.True(t, executed)
}

// TestDTMProvider verifies the DTM provider registration.
//
// TestDTMProvider 验证分布式事务管理服务提供者的注册。
func TestProvider(t *testing.T) {
	p := NewProvider()

	assert.Equal(t, "dtm.noop", p.Name())
	assert.True(t, p.IsDefer())
	assert.ElementsMatch(t, []string{integrationcontract.DTMKey}, p.Provides())
}
