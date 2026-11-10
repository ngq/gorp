package gorm_test

import (
	"context"
	"errors"
	"testing"

	gormoutbox "github.com/ngq/gorp/contrib/outbox/gorm"
	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestGormOutboxStore_Lifecycle(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	store := gormoutbox.NewGormOutboxStore(db)
	err = store.AutoMigrate()
	require.NoError(t, err)

	ctx := t.Context()

	// 1. Save Pending Message inside Transaction
	tx := db.Begin()
	txCtx := context.WithValue(ctx, "tx", tx)
	msg := &integrationcontract.OutboxMessage{
		ID:      "msg-1001",
		Topic:   "user.registered",
		Payload: map[string]any{"user_id": "u1"},
		Status:  integrationcontract.OutboxStatusPending,
	}
	err = store.Save(txCtx, msg)
	require.NoError(t, err)
	tx.Commit()

	// 2. Fetch Pending
	pending, err := store.GetPending(ctx, 10)
	require.NoError(t, err)
	require.Len(t, pending, 1)
	require.Equal(t, "msg-1001", pending[0].ID)

	// 3. Mark Sent
	err = store.MarkSent(ctx, "msg-1001")
	require.NoError(t, err)

	// 4. Verify no more pending
	pendingAfter, err := store.GetPending(ctx, 10)
	require.NoError(t, err)
	require.Len(t, pendingAfter, 0)
}

// TestGormOutboxStore_MarkFailedRetriesUntilLimit 锁定 FEAT-02 修复语义：
// 失败先置 retrying（可重新投递），达到重试上限才进入终态 failed——
// 恢复 at-least-once，杜绝"broker 一次抖动即永久判死"。
func TestGormOutboxStore_MarkFailedRetriesUntilLimit(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	require.NoError(t, err)

	store := gormoutbox.NewGormOutboxStoreWithRetryLimit(db, 2)
	err = store.AutoMigrate()
	require.NoError(t, err)

	ctx := context.Background()
	msg := &integrationcontract.OutboxMessage{
		ID:     "msg-retry",
		Topic:  "user.registered",
		Status: integrationcontract.OutboxStatusPending,
	}
	require.NoError(t, store.Save(ctx, msg))

	// 第 1 次失败 → retrying（仍可被 GetPending 捞到）
	require.NoError(t, store.MarkFailed(ctx, "msg-retry", errors.New("transient")))
	pending, err := store.GetPending(ctx, 10)
	require.NoError(t, err)
	require.Len(t, pending, 1)
	require.Equal(t, integrationcontract.OutboxStatusRetrying, pending[0].Status)

	// 第 2 次失败（retry_count=2 >= limit=2）→ 终态 failed
	require.NoError(t, store.MarkFailed(ctx, "msg-retry", errors.New("transient again")))
	pending, err = store.GetPending(ctx, 10)
	require.NoError(t, err)
	require.Empty(t, pending, "failed message must NOT be returned by GetPending")

	// 确认库内终态为 failed 且错误信息已记录
	var row gormoutbox.GormOutboxModel
	require.NoError(t, db.Where("id = ?", "msg-retry").First(&row).Error)
	require.Equal(t, string(integrationcontract.OutboxStatusFailed), row.Status)
	require.Contains(t, row.Error, "transient")
}
