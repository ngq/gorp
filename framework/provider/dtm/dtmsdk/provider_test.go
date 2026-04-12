package dtmsdk

import (
	"context"
	"testing"

	"github.com/ngq/gorp/framework/contract"
	"github.com/stretchr/testify/assert"
)

func TestDTMClient_SAGA(t *testing.T) {
	cfg := &contract.DTMConfig{
		Enabled:  true,
		Endpoint: "http://localhost:36789",
	}
	client, err := NewDTMClient(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, client)

	// 测试 SAGA 构建
	saga := client.SAGA("test-saga")
	assert.NotNil(t, saga)

	// 添加步骤
	saga.Add("/api/action", "/api/compensate", map[string]string{"key": "value"})
	saga.Add("/api/action2", "/api/compensate2", nil)

	// Submit 应返回 SDK 未引入错误（骨架实现）
	err = saga.Submit(context.Background())
	assert.ErrorIs(t, err, ErrDTMSDKNotImported)

	// Build 应返回事务对象
	tx, err := saga.Build()
	assert.NoError(t, err)
	assert.NotNil(t, tx)
	assert.Len(t, tx.Steps, 2)
}

func TestDTMClient_TCC(t *testing.T) {
	cfg := &contract.DTMConfig{
		Enabled:  true,
		Endpoint: "http://localhost:36789",
	}
	client, err := NewDTMClient(cfg)
	assert.NoError(t, err)

	// 测试 TCC 构建
	tcc := client.TCC("test-tcc")
	assert.NotNil(t, tcc)

	// 添加步骤
	tcc.Add("/api/try", "/api/confirm", "/api/cancel", nil)

	// Submit 应返回 SDK 未引入错误
	err = tcc.Submit(context.Background())
	assert.ErrorIs(t, err, ErrDTMSDKNotImported)
}

func TestDTMClient_XA(t *testing.T) {
	cfg := &contract.DTMConfig{
		Enabled:  true,
		Endpoint: "http://localhost:36789",
	}
	client, err := NewDTMClient(cfg)
	assert.NoError(t, err)

	// 测试 XA 构建
	xa := client.XA("test-xa")
	assert.NotNil(t, xa)

	// 添加步骤
	xa.Add("/api/xa-action", nil)

	// Submit 应返回 SDK 未引入错误
	err = xa.Submit(context.Background())
	assert.ErrorIs(t, err, ErrDTMSDKNotImported)
}

func TestDTMClient_Barrier(t *testing.T) {
	cfg := &contract.DTMConfig{
		Enabled:  true,
		Endpoint: "http://localhost:36789",
	}
	client, err := NewDTMClient(cfg)
	assert.NoError(t, err)

	// 测试 Barrier
	barrier := client.Barrier("saga", "test-gid")
	assert.NotNil(t, barrier)

	// Call 应执行函数（骨架实现）
	executed := false
	err = barrier.Call(context.Background(), func(db any) error {
		executed = true
		return nil
	})
	assert.NoError(t, err)
	assert.True(t, executed)
}

func TestDTMClient_Query(t *testing.T) {
	cfg := &contract.DTMConfig{
		Enabled:  true,
		Endpoint: "http://localhost:36789",
	}
	client, err := NewDTMClient(cfg)
	assert.NoError(t, err)

	// Query 应返回基本信息（骨架实现）
	info, err := client.Query(context.Background(), "test-gid")
	assert.NoError(t, err)
	assert.NotNil(t, info)
	assert.Equal(t, "test-gid", info.GID)
	assert.Equal(t, "unknown", info.Status)
}

func TestSAGABuilder_AddBranch(t *testing.T) {
	cfg := &contract.DTMConfig{Enabled: true}
	client, _ := NewDTMClient(cfg)

	saga := client.SAGA("test")
	saga.AddBranch("/action", "/compensate", nil, contract.BranchOptions{
		RetryCount:    3,
		RetryInterval: 5,
	})

	tx, err := saga.Build()
	assert.NoError(t, err)
	assert.Len(t, tx.Steps, 1)
}

func TestProvider_Name(t *testing.T) {
	p := NewProvider()
	assert.Equal(t, "dtm.sdk", p.Name())
	assert.True(t, p.IsDefer())
	assert.ElementsMatch(t, []string{contract.DTMKey}, p.Provides())
}

func TestErrDTMSDKNotImported(t *testing.T) {
	// 验证错误消息
	assert.Contains(t, ErrDTMSDKNotImported.Error(), "dtm-labs/client")
	assert.Contains(t, ErrDTMSDKNotImported.Error(), "dtm.pub")
}