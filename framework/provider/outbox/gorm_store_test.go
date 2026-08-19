package outbox_test

import (
	"context"
	"testing"

	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
	outboxprovider "github.com/ngq/gorp/framework/provider/outbox"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestGormOutboxStore_Lifecycle(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	store := outboxprovider.NewGormOutboxStore(db)
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
