package noop

import (
	"context"
	"testing"

	"github.com/ngq/gorp/framework/contract"
	"github.com/stretchr/testify/assert"
)

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

func TestNoopTCCBuilder(t *testing.T) {
	builder := &noopTCCBuilder{}

	// 测试 Add
	result := builder.Add("/try", "/confirm", "/cancel", nil)
	assert.Equal(t, builder, result)

	// 测试 Submit
	err := builder.Submit(context.Background())
	assert.ErrorIs(t, err, ErrNoopDTM)
}

func TestNoopXABuilder(t *testing.T) {
	builder := &noopXABuilder{}

	// 测试 Add
	result := builder.Add("/url", nil)
	assert.Equal(t, builder, result)

	// 测试 Submit
	err := builder.Submit(context.Background())
	assert.ErrorIs(t, err, ErrNoopDTM)
}

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

func TestProvider(t *testing.T) {
	p := NewProvider()

	assert.Equal(t, "dtm.noop", p.Name())
	assert.True(t, p.IsDefer())
	assert.ElementsMatch(t, []string{contract.DTMKey}, p.Provides())
}