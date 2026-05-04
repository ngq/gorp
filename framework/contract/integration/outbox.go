package integration

import (
	"context"
	"time"
)

const OutboxKey = "framework.outbox"

type OutboxMessage struct {
	ID         string
	Topic      string
	Payload    interface{}
	Status     OutboxStatus
	RetryCount int
	CreatedAt  time.Time
	SentAt     *time.Time
	Error      string
}

type OutboxStatus string

const (
	OutboxStatusPending  OutboxStatus = "pending"
	OutboxStatusSent     OutboxStatus = "sent"
	OutboxStatusFailed   OutboxStatus = "failed"
	OutboxStatusRetrying OutboxStatus = "retrying"
)

type OutboxStore interface {
	Save(ctx context.Context, msg *OutboxMessage) error
	GetPending(ctx context.Context, limit int) ([]*OutboxMessage, error)
	MarkSent(ctx context.Context, id string) error
	MarkFailed(ctx context.Context, id string, err error) error
}

type OutboxSender interface {
	Send(ctx context.Context, msg *OutboxMessage) error
}

type Outbox interface {
	Emit(ctx context.Context, topic string, payload interface{}) error
	EmitSync(ctx context.Context, topic string, payload interface{}) error
	Process(ctx context.Context) error
}

type OutboxConfig struct {
	Enabled    bool
	BatchSize  int
	RetryLimit int
	RetryDelay time.Duration
}
