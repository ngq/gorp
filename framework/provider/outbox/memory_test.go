// Package outbox_test provides unit tests for the outbox memory implementation.
//
// 适用场景：
// - 验证 Outbox 内存实现的发布、查询和清理行为。
package outbox

import (
	"context"
	"errors"
	"testing"

	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
)

type senderStub struct {
	err  error
	sent int
	last *integrationcontract.OutboxMessage
}

func (s *senderStub) Send(_ context.Context, msg *integrationcontract.OutboxMessage) error {
	s.sent++
	s.last = msg
	return s.err
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
	if sender.sent != 1 {
		t.Fatalf("expected sender to be called once, got %d", sender.sent)
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
