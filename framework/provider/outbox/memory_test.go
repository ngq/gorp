// Package outbox_test provides unit tests for the outbox memory implementation.
//
// 适用场景：
// - 验证 Outbox 内存实现的发布、查询和清理行为。
// - 验证持久化主路径、失败重试边界、幂等性边界和恢复路径。
package outbox

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
)

// senderStub 是用于测试的发送器存根，支持控制发送成功/失败。
type senderStub struct {
	err     error               // 发送时返回的错误
	sent    int                 // 成功发送次数
	last    *integrationcontract.OutboxMessage // 最后发送的消息
	mu      sync.Mutex          // 保护 sent 计数
	callIDs []string            // 所有调用过的消息 ID（用于验证幂等性）
}

// Send 实现 OutboxSender 接口。
func (s *senderStub) Send(_ context.Context, msg *integrationcontract.OutboxMessage) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.callIDs = append(s.callIDs, msg.ID)
	s.last = msg
	if s.err != nil {
		return s.err
	}
	s.sent++
	return nil
}

// getSent 返回成功发送次数（线程安全）。
func (s *senderStub) getSent() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.sent
}

// TestMemoryOutboxEmitSyncMarksSent verifies that EmitSync sends message synchronously
// and marks it as sent with proper status and timestamp.
//
// TestMemoryOutboxEmitSyncMarksSent 验证 EmitSync 同步发送消息并正确标记为已发送状态和时间戳。
func TestMemoryOutboxEmitSyncMarksSent(t *testing.T) {
	sender := &senderStub{}
	outbox := NewMemoryOutbox(sender, integrationcontract.OutboxConfig{RetryLimit: 3})

	if err := outbox.EmitSync(context.Background(), "order.created", map[string]any{"id": 1}); err != nil {
		t.Fatalf("EmitSync failed: %v", err)
	}
	if sender.getSent() != 1 {
		t.Fatalf("expected sender to be called once, got %d", sender.getSent())
	}
	if sender.last == nil {
		t.Fatalf("expected last message to be captured")
	}
	msg := outbox.messages[sender.last.ID]
	if msg.Status != integrationcontract.OutboxStatusSent {
		t.Fatalf("expected sent status, got %s", msg.Status)
	}
	if msg.SentAt == nil {
		t.Fatalf("expected SentAt to be set")
	}
}

// TestMemoryOutboxProcessMarksRetryingThenFailed verifies that when sender fails,
// the message status is marked as failed with retry count and error stored.
//
// TestMemoryOutboxProcessMarksRetryingThenFailed 验证当发送器失败时，消息状态标记为失败，并存储重试次数和错误。
func TestMemoryOutboxProcessMarksRetryingThenFailed(t *testing.T) {
	sender := &senderStub{err: errors.New("boom")}
	outbox := NewMemoryOutbox(sender, integrationcontract.OutboxConfig{RetryLimit: 1})

	if err := outbox.EmitSync(context.Background(), "order.created", map[string]any{"id": 2}); err != nil {
		t.Fatalf("EmitSync failed: %v", err)
	}
	if sender.last == nil {
		t.Fatalf("expected last message to be captured")
	}
	msg := outbox.messages[sender.last.ID]
	if msg.Status != integrationcontract.OutboxStatusFailed {
		t.Fatalf("expected failed status, got %s", msg.Status)
	}
	if msg.RetryCount != 1 {
		t.Fatalf("expected retry count 1, got %d", msg.RetryCount)
	}
	if msg.Error != "boom" {
		t.Fatalf("expected stored error boom, got %q", msg.Error)
	}
}

// TestMemoryOutboxProcessWithoutSenderFails verifies that Process returns an error
// when no sender is configured.
//
// TestMemoryOutboxProcessWithoutSenderFails 验证 Process 在未配置发送器时返回错误。
func TestMemoryOutboxProcessWithoutSenderFails(t *testing.T) {
	outbox := NewMemoryOutbox(nil, integrationcontract.OutboxConfig{RetryLimit: 3})
	outbox.messages["msg-1"] = &integrationcontract.OutboxMessage{
		ID:     "msg-1",
		Topic:  "demo",
		Status: integrationcontract.OutboxStatusPending,
	}

	err := outbox.Process(context.Background())
	if err == nil {
		t.Fatalf("expected process to fail without sender")
	}
}

// === 持久化主路径测试 ===

// TestMemoryOutboxStoreSaveAndGetPending verifies that MemoryOutboxStore correctly
// saves messages and retrieves pending ones.
//
// TestMemoryOutboxStoreSaveAndGetPending 验证 MemoryOutboxStore 正确保存消息并检索待处理消息。
func TestMemoryOutboxStoreSaveAndGetPending(t *testing.T) {
	store := NewMemoryOutboxStore()
	ctx := context.Background()

	// 保存待处理消息
	msg1 := &integrationcontract.OutboxMessage{
		ID:        "msg-1",
		Topic:     "order.created",
		Payload:   map[string]any{"id": 1},
		Status:    integrationcontract.OutboxStatusPending,
		CreatedAt: time.Now(),
	}
	if err := store.Save(ctx, msg1); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// 保存已发送消息（不应被 GetPending 返回）
	msg2 := &integrationcontract.OutboxMessage{
		ID:        "msg-2",
		Topic:     "order.updated",
		Payload:   map[string]any{"id": 2},
		Status:    integrationcontract.OutboxStatusSent,
		CreatedAt: time.Now(),
	}
	if err := store.Save(ctx, msg2); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// 获取待处理消息
	pending, err := store.GetPending(ctx, 10)
	if err != nil {
		t.Fatalf("GetPending failed: %v", err)
	}
	if len(pending) != 1 {
		t.Fatalf("expected 1 pending message, got %d", len(pending))
	}
	if pending[0].ID != "msg-1" {
		t.Fatalf("expected msg-1, got %s", pending[0].ID)
	}
}

// TestMemoryOutboxStoreMarkSent verifies that MarkSent correctly updates message status.
//
// TestMemoryOutboxStoreMarkSent 验证 MarkSent 正确更新消息状态。
func TestMemoryOutboxStoreMarkSent(t *testing.T) {
	store := NewMemoryOutboxStore()
	ctx := context.Background()

	msg := &integrationcontract.OutboxMessage{
		ID:        "msg-1",
		Topic:     "order.created",
		Status:    integrationcontract.OutboxStatusPending,
		CreatedAt: time.Now(),
	}
	_ = store.Save(ctx, msg)

	if err := store.MarkSent(ctx, "msg-1"); err != nil {
		t.Fatalf("MarkSent failed: %v", err)
	}

	// 验证状态已更新
	pending, _ := store.GetPending(ctx, 10)
	if len(pending) != 0 {
		t.Fatalf("expected 0 pending messages after MarkSent, got %d", len(pending))
	}

	// 验证消息仍存在但状态为 sent
	if store.messages["msg-1"].Status != integrationcontract.OutboxStatusSent {
		t.Fatalf("expected sent status, got %s", store.messages["msg-1"].Status)
	}
	if store.messages["msg-1"].SentAt == nil {
		t.Fatalf("expected SentAt to be set")
	}
}

// TestMemoryOutboxStoreMarkFailed verifies that MarkFailed correctly updates message status and error.
//
// TestMemoryOutboxStoreMarkFailed 验证 MarkFailed 正确更新消息状态和错误信息。
func TestMemoryOutboxStoreMarkFailed(t *testing.T) {
	store := NewMemoryOutboxStore()
	ctx := context.Background()

	msg := &integrationcontract.OutboxMessage{
		ID:        "msg-1",
		Topic:     "order.created",
		Status:    integrationcontract.OutboxStatusPending,
		CreatedAt: time.Now(),
	}
	_ = store.Save(ctx, msg)

	testErr := errors.New("connection refused")
	if err := store.MarkFailed(ctx, "msg-1", testErr); err != nil {
		t.Fatalf("MarkFailed failed: %v", err)
	}

	// 验证状态已更新
	if store.messages["msg-1"].Status != integrationcontract.OutboxStatusFailed {
		t.Fatalf("expected failed status, got %s", store.messages["msg-1"].Status)
	}
	if store.messages["msg-1"].Error != "connection refused" {
		t.Fatalf("expected error message 'connection refused', got %q", store.messages["msg-1"].Error)
	}
}

// === 重试次数边界测试 ===

// TestMemoryOutboxRetryLimitBoundary verifies that messages are marked as failed
// exactly when retry count reaches the limit through multiple Process calls.
//
// TestMemoryOutboxRetryLimitBoundary 验证消息在重试次数达到限制时被标记为失败。
// 注意：每次 Process 调用只会尝试发送一次，需要多次调用才能达到重试限制。
func TestMemoryOutboxRetryLimitBoundary(t *testing.T) {
	sender := &senderStub{err: errors.New("persistently failing")}
	// 设置重试限制为 3
	outbox := NewMemoryOutbox(sender, integrationcontract.OutboxConfig{RetryLimit: 3})

	// 发送消息
	_ = outbox.EmitSync(context.Background(), "order.created", map[string]any{"id": 1})
	msgID := sender.last.ID

	// EmitSync 只调用一次 Process，所以第一次失败后 RetryCount=1，状态=retrying
	msg := outbox.messages[msgID]
	if msg.Status != integrationcontract.OutboxStatusRetrying {
		t.Fatalf("expected retrying status after first failure, got %s", msg.Status)
	}
	if msg.RetryCount != 1 {
		t.Fatalf("expected retry count 1 after first Process, got %d", msg.RetryCount)
	}

	// 继续调用 Process 直到达到重试限制
	_ = outbox.Process(context.Background()) // 第 2 次尝试
	if msg.RetryCount != 2 {
		t.Fatalf("expected retry count 2, got %d", msg.RetryCount)
	}

	_ = outbox.Process(context.Background()) // 第 3 次尝试
	if msg.Status != integrationcontract.OutboxStatusFailed {
		t.Fatalf("expected failed status after retry limit, got %s", msg.Status)
	}
	if msg.RetryCount != 3 {
		t.Fatalf("expected retry count 3, got %d", msg.RetryCount)
	}

	// sender 应该被调用了 3 次
	if len(sender.callIDs) != 3 {
		t.Fatalf("expected 3 send attempts, got %d", len(sender.callIDs))
	}
}

// TestMemoryOutboxRetryLimitZero verifies that with RetryLimit=0, message is immediately failed
// without any send attempts.
//
// TestMemoryOutboxRetryLimitZero 验证 RetryLimit=0 时消息立即失败，不尝试发送。
func TestMemoryOutboxRetryLimitZero(t *testing.T) {
	sender := &senderStub{err: errors.New("boom")}
	outbox := NewMemoryOutbox(sender, integrationcontract.OutboxConfig{RetryLimit: 0})

	// 手动创建消息并放入 outbox
	msg := &integrationcontract.OutboxMessage{
		ID:        "msg-1",
		Topic:     "order.created",
		Payload:   map[string]any{"id": 1},
		Status:    integrationcontract.OutboxStatusPending,
		CreatedAt: time.Now(),
	}
	outbox.messages["msg-1"] = msg

	// 调用 Process
	_ = outbox.Process(context.Background())

	// RetryLimit=0 时，消息应立即标记为失败，不尝试发送
	if msg.Status != integrationcontract.OutboxStatusFailed {
		t.Fatalf("expected failed status with RetryLimit=0, got %s", msg.Status)
	}
	if msg.RetryCount != 0 {
		t.Fatalf("expected retry count 0 with RetryLimit=0, got %d", msg.RetryCount)
	}
	// sender 不应该被调用
	if len(sender.callIDs) != 0 {
		t.Fatalf("expected 0 send attempts with RetryLimit=0, got %d", len(sender.callIDs))
	}
}

// TestMemoryOutboxRetryingStatusBeforeFinalFailure verifies that message has Retrying status
// before final failure when retry limit is not yet reached.
//
// TestMemoryOutboxRetryingStatusBeforeFinalFailure 验证消息在达到最终失败前处于 Retrying 状态。
func TestMemoryOutboxRetryingStatusBeforeFinalFailure(t *testing.T) {
	sender := &senderStub{err: errors.New("temporary failure")}
	// 设置重试限制为 5，观察中间状态
	outbox := NewMemoryOutbox(sender, integrationcontract.OutboxConfig{RetryLimit: 5})

	msg := &integrationcontract.OutboxMessage{
		ID:        "msg-1",
		Topic:     "order.created",
		Payload:   map[string]any{"id": 1},
		Status:    integrationcontract.OutboxStatusPending,
		CreatedAt: time.Now(),
	}
	outbox.messages["msg-1"] = msg

	// 手动调用一次 Process，模拟中间状态
	_ = outbox.Process(context.Background())

	// 第一次处理后，应该是 Retrying 状态
	if msg.Status != integrationcontract.OutboxStatusRetrying {
		t.Fatalf("expected retrying status after first failure, got %s", msg.Status)
	}
	if msg.RetryCount != 1 {
		t.Fatalf("expected retry count 1, got %d", msg.RetryCount)
	}
}

// === 恢复路径测试 ===

// TestMemoryOutboxProcessRecoversPendingMessages verifies that Process recovers
// pending messages from previous runs.
//
// TestMemoryOutboxProcessRecoversPendingMessages 验证 Process 能恢复之前遗留的待处理消息。
func TestMemoryOutboxProcessRecoversPendingMessages(t *testing.T) {
	sender := &senderStub{}
	outbox := NewMemoryOutbox(sender, integrationcontract.OutboxConfig{RetryLimit: 3})

	// 模拟之前遗留的待处理消息（如应用重启后）
	msg1 := &integrationcontract.OutboxMessage{
		ID:        "msg-1",
		Topic:     "order.created",
		Payload:   map[string]any{"id": 1},
		Status:    integrationcontract.OutboxStatusPending,
		CreatedAt: time.Now().Add(-time.Hour), // 1小时前创建
	}
	msg2 := &integrationcontract.OutboxMessage{
		ID:        "msg-2",
		Topic:     "order.updated",
		Payload:   map[string]any{"id": 2},
		Status:    integrationcontract.OutboxStatusPending,
		CreatedAt: time.Now().Add(-30 * time.Minute),
	}
	outbox.messages["msg-1"] = msg1
	outbox.messages["msg-2"] = msg2

	// 调用 Process 恢复处理
	if err := outbox.Process(context.Background()); err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// 验证两条消息都被发送
	if sender.getSent() != 2 {
		t.Fatalf("expected 2 messages sent, got %d", sender.getSent())
	}

	// 验证消息状态
	if msg1.Status != integrationcontract.OutboxStatusSent {
		t.Fatalf("expected msg-1 sent, got %s", msg1.Status)
	}
	if msg2.Status != integrationcontract.OutboxStatusSent {
		t.Fatalf("expected msg-2 sent, got %s", msg2.Status)
	}
}

// TestMemoryOutboxProcessRecoversRetryingMessages verifies that Process recovers
// messages that were in retrying state from previous runs.
//
// TestMemoryOutboxProcessRecoversRetryingMessages 验证 Process 能恢复之前处于重试中的消息。
func TestMemoryOutboxProcessRecoversRetryingMessages(t *testing.T) {
	sender := &senderStub{}
	outbox := NewMemoryOutbox(sender, integrationcontract.OutboxConfig{RetryLimit: 3})

	// 模拟之前失败过一次的消息
	msg := &integrationcontract.OutboxMessage{
		ID:         "msg-1",
		Topic:      "order.created",
		Payload:    map[string]any{"id": 1},
		Status:     integrationcontract.OutboxStatusRetrying,
		RetryCount: 1,
		CreatedAt:  time.Now().Add(-time.Hour),
	}
	outbox.messages["msg-1"] = msg

	// 调用 Process 恢复处理
	_ = outbox.Process(context.Background())

	// 验证消息被发送（重试成功）
	if sender.getSent() != 1 {
		t.Fatalf("expected 1 message sent, got %d", sender.getSent())
	}
	if msg.Status != integrationcontract.OutboxStatusSent {
		t.Fatalf("expected sent status, got %s", msg.Status)
	}
}

// TestMemoryOutboxProcessSkipsAlreadySent verifies that Process skips messages
// that are already sent.
//
// TestMemoryOutboxProcessSkipsAlreadySent 验证 Process 跳过已发送的消息。
func TestMemoryOutboxProcessSkipsAlreadySent(t *testing.T) {
	sender := &senderStub{}
	outbox := NewMemoryOutbox(sender, integrationcontract.OutboxConfig{RetryLimit: 3})

	now := time.Now()
	// 已发送的消息
	msg := &integrationcontract.OutboxMessage{
		ID:        "msg-1",
		Topic:     "order.created",
		Payload:   map[string]any{"id": 1},
		Status:    integrationcontract.OutboxStatusSent,
		SentAt:    &now,
		CreatedAt: time.Now().Add(-time.Hour),
	}
	outbox.messages["msg-1"] = msg

	// 调用 Process
	_ = outbox.Process(context.Background())

	// 验证不会重复发送
	if sender.getSent() != 0 {
		t.Fatalf("expected 0 messages sent (already sent), got %d", sender.getSent())
	}
}

// TestMemoryOutboxProcessSkipsFailedMessages verifies that Process skips messages
// that have already failed (not retrying).
//
// TestMemoryOutboxProcessSkipsFailedMessages 验证 Process 跳过已失败的消息。
func TestMemoryOutboxProcessSkipsFailedMessages(t *testing.T) {
	sender := &senderStub{}
	outbox := NewMemoryOutbox(sender, integrationcontract.OutboxConfig{RetryLimit: 3})

	// 已失败的消息
	msg := &integrationcontract.OutboxMessage{
		ID:         "msg-1",
		Topic:      "order.created",
		Payload:    map[string]any{"id": 1},
		Status:     integrationcontract.OutboxStatusFailed,
		RetryCount: 3,
		Error:      "max retries exceeded",
		CreatedAt:  time.Now().Add(-time.Hour),
	}
	outbox.messages["msg-1"] = msg

	// 调用 Process
	_ = outbox.Process(context.Background())

	// 验证不会重试已失败的消息
	if sender.getSent() != 0 {
		t.Fatalf("expected 0 messages sent (already failed), got %d", sender.getSent())
	}
}

// === Emit 异步测试 ===

// TestMemoryOutboxEmitAsync verifies that Emit creates a message and spawns background processing.
//
// TestMemoryOutboxEmitAsync 验证 Emit 创建消息并启动后台处理。
func TestMemoryOutboxEmitAsync(t *testing.T) {
	sender := &senderStub{}
	outbox := NewMemoryOutbox(sender, integrationcontract.OutboxConfig{RetryLimit: 3})

	// Emit 是异步的，立即返回
	if err := outbox.Emit(context.Background(), "order.created", map[string]any{"id": 1}); err != nil {
		t.Fatalf("Emit failed: %v", err)
	}

	// 等待后台处理完成
	time.Sleep(100 * time.Millisecond)

	// 验证消息被发送
	if sender.getSent() != 1 {
		t.Fatalf("expected 1 message sent after async processing, got %d", sender.getSent())
	}
}

// === GetPending limit 测试 ===

// TestMemoryOutboxStoreGetPendingLimit verifies that GetPending respects the limit parameter.
//
// TestMemoryOutboxStoreGetPendingLimit 验证 GetPending 遵守 limit 参数。
func TestMemoryOutboxStoreGetPendingLimit(t *testing.T) {
	store := NewMemoryOutboxStore()
	ctx := context.Background()

	// 保存 5 条待处理消息
	for i := 0; i < 5; i++ {
		msg := &integrationcontract.OutboxMessage{
			ID:        string(rune('a' + i)),
			Topic:     "order.created",
			Status:    integrationcontract.OutboxStatusPending,
			CreatedAt: time.Now(),
		}
		_ = store.Save(ctx, msg)
	}

	// 获取最多 3 条
	pending, err := store.GetPending(ctx, 3)
	if err != nil {
		t.Fatalf("GetPending failed: %v", err)
	}
	if len(pending) != 3 {
		t.Fatalf("expected 3 pending messages (limited), got %d", len(pending))
	}
}

// TestMemoryOutboxStoreGetPendingNoPending verifies that GetPending returns empty slice
// when there are no pending messages.
//
// TestMemoryOutboxStoreGetPendingNoPending 验证无待处理消息时 GetPending 返回空切片。
func TestMemoryOutboxStoreGetPendingNoPending(t *testing.T) {
	store := NewMemoryOutboxStore()
	ctx := context.Background()

	// 不保存任何消息
	pending, err := store.GetPending(ctx, 10)
	if err != nil {
		t.Fatalf("GetPending failed: %v", err)
	}
	if len(pending) != 0 {
		t.Fatalf("expected 0 pending messages, got %d", len(pending))
	}
}

// === 幂等性边界说明测试 ===

// TestMemoryOutboxNoIdempotencyGuarantee verifies that the memory outbox does NOT
// provide idempotency guarantees - sender may be called multiple times for the same message
// during retries. This is by design: idempotency is the responsibility of the business layer.
//
// TestMemoryOutboxNoIdempotencyGuarantee 验证内存 outbox 不提供幂等保证 -
// 在重试过程中 sender 可能对同一条消息被调用多次。这是设计决定：幂等是业务层的责任。
func TestMemoryOutboxNoIdempotencyGuarantee(t *testing.T) {
	// 注意：这个测试说明设计原则：
	// 1. Outbox 不保证幂等，sender 可能被多次调用
	// 2. 业务层需要处理重复投递（如消息去重表、唯一键约束等）
	// 3. 这是 outbox 模式的常见设计选择
	// 详见 TestMemoryOutboxSenderCalledMultipleTimesOnRetry
	_ = t // 避免空函数警告
}

// TestMemoryOutboxSenderCalledMultipleTimesOnRetry verifies that sender.Send is called
// multiple times when Process is called repeatedly for retrying messages.
//
// TestMemoryOutboxSenderCalledMultipleTimesOnRetry 验证重复调用 Process 时 sender.Send 被多次调用。
// 注意：每次 Process 调用只尝试发送一次，重试需要多次调用 Process。
func TestMemoryOutboxSenderCalledMultipleTimesOnRetry(t *testing.T) {
	sender := &senderStub{err: errors.New("persistent failure")}
	outbox := NewMemoryOutbox(sender, integrationcontract.OutboxConfig{RetryLimit: 3})

	// 手动创建消息
	msg := &integrationcontract.OutboxMessage{
		ID:        "msg-1",
		Topic:     "order.created",
		Payload:   map[string]any{"id": 1},
		Status:    integrationcontract.OutboxStatusPending,
		CreatedAt: time.Now(),
	}
	outbox.messages["msg-1"] = msg

	// 调用 Process 3 次达到重试限制
	_ = outbox.Process(context.Background())
	_ = outbox.Process(context.Background())
	_ = outbox.Process(context.Background())

	// sender 应该被调用了 3 次（等于 RetryLimit）
	if len(sender.callIDs) != 3 {
		t.Fatalf("expected sender to be called 3 times during retry, got %d", len(sender.callIDs))
	}

	// 所有调用都是同一条消息
	for i, id := range sender.callIDs {
		if id != "msg-1" {
			t.Fatalf("call %d: expected msg-1, got %s", i, id)
		}
	}
}

// TestMemoryOutboxSenderEventuallySucceeds verifies that message is sent when sender
// eventually succeeds after initial failures (through multiple Process calls).
//
// TestMemoryOutboxSenderEventuallySucceeds 验证 sender 在多次失败后最终成功时消息被发送。
// 注意：每次 Process 调用只尝试发送一次，需要多次调用 Process 来实现重试。
func TestMemoryOutboxSenderEventuallySucceeds(t *testing.T) {
	var callCount int
	var mu sync.Mutex
	var lastMsg *integrationcontract.OutboxMessage
	var sentCount int

	// 创建一个第三次调用时成功的 sender（使用 senderStub 模式）
	eventualSender := &struct {
		integrationcontract.OutboxSender // 嵌入接口
	}{
		OutboxSender: senderStubFunc(func(ctx context.Context, msg *integrationcontract.OutboxMessage) error {
			mu.Lock()
			callCount++
			currentCall := callCount
			mu.Unlock()

			if currentCall < 3 {
				return errors.New("temporary failure")
			}
			lastMsg = msg
			sentCount++
			return nil
		}),
	}

	outbox := NewMemoryOutbox(eventualSender, integrationcontract.OutboxConfig{RetryLimit: 5})

	// 手动创建消息
	msg := &integrationcontract.OutboxMessage{
		ID:        "msg-1",
		Topic:     "order.created",
		Payload:   map[string]any{"id": 1},
		Status:    integrationcontract.OutboxStatusPending,
		CreatedAt: time.Now(),
	}
	outbox.messages["msg-1"] = msg

	// 第一次 Process：失败
	_ = outbox.Process(context.Background())
	if msg.Status != integrationcontract.OutboxStatusRetrying {
		t.Fatalf("expected retrying after first failure, got %s", msg.Status)
	}

	// 第二次 Process：失败
	_ = outbox.Process(context.Background())
	if msg.Status != integrationcontract.OutboxStatusRetrying {
		t.Fatalf("expected retrying after second failure, got %s", msg.Status)
	}

	// 第三次 Process：成功
	_ = outbox.Process(context.Background())
	if msg.Status != integrationcontract.OutboxStatusSent {
		t.Fatalf("expected sent status after eventual success, got %s", msg.Status)
	}

	// 验证调用次数
	if callCount != 3 {
		t.Fatalf("expected 3 calls, got %d", callCount)
	}
	if sentCount != 1 {
		t.Fatalf("expected 1 successful send, got %d", sentCount)
	}
	if lastMsg == nil {
		t.Fatalf("expected message to be sent")
	}
}

// senderStubFunc 是一个函数类型，实现 OutboxSender 接口。
type senderStubFunc func(context.Context, *integrationcontract.OutboxMessage) error

func (f senderStubFunc) Send(ctx context.Context, msg *integrationcontract.OutboxMessage) error {
	return f(ctx, msg)
}
