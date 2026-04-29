package outbox

import (
	"context"
	"errors"
	"testing"

	"github.com/ngq/gorp/framework/contract"
)

type senderStub struct {
	err   error
	sent  int
	last  *contract.OutboxMessage
}

func (s *senderStub) Send(_ context.Context, msg *contract.OutboxMessage) error {
	s.sent++
	s.last = msg
	return s.err
}

func TestMemoryOutboxEmitSyncMarksSent(t *testing.T) {
	sender := &senderStub{}
	outbox := NewMemoryOutbox(sender, contract.OutboxConfig{RetryLimit: 3})

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
	if msg.Status != contract.OutboxStatusSent {
		t.Fatalf("expected sent status, got %s", msg.Status)
	}
	if msg.SentAt == nil {
		t.Fatalf("expected SentAt to be set")
	}
}

func TestMemoryOutboxProcessMarksRetryingThenFailed(t *testing.T) {
	sender := &senderStub{err: errors.New("boom")}
	outbox := NewMemoryOutbox(sender, contract.OutboxConfig{RetryLimit: 1})

	if err := outbox.EmitSync(context.Background(), "order.created", map[string]any{"id": 2}); err != nil {
		t.Fatalf("EmitSync failed: %v", err)
	}
	if sender.last == nil {
		t.Fatalf("expected last message to be captured")
	}
	msg := outbox.messages[sender.last.ID]
	if msg.Status != contract.OutboxStatusFailed {
		t.Fatalf("expected failed status, got %s", msg.Status)
	}
	if msg.RetryCount != 1 {
		t.Fatalf("expected retry count 1, got %d", msg.RetryCount)
	}
	if msg.Error != "boom" {
		t.Fatalf("expected stored error boom, got %q", msg.Error)
	}
}

func TestMemoryOutboxProcessWithoutSenderFails(t *testing.T) {
	outbox := NewMemoryOutbox(nil, contract.OutboxConfig{RetryLimit: 3})
	outbox.messages["msg-1"] = &contract.OutboxMessage{
		ID:     "msg-1",
		Topic:  "demo",
		Status: contract.OutboxStatusPending,
	}

	err := outbox.Process(context.Background())
	if err == nil {
		t.Fatalf("expected process to fail without sender")
	}
}
