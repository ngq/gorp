// Application scenarios:
// - Define the outbox contract used for reliable event/message delivery.
// - Standardize pending, sent, failed, and retrying message lifecycle states.
// - Keep outbox storage, sender, and processing abstractions provider-neutral.
//
// 适用场景：
// - 定义可靠消息投递使用的 outbox 契约。
// - 统一 pending、sent、failed 和 retrying 的消息生命周期状态。
// - 保持 outbox 存储、发送器和处理流程与具体 provider 解耦。
package integration

import (
	"context"
	"time"
)

// OutboxKey is the container key for the outbox capability.
//
// OutboxKey 是 outbox 能力的容器键。
const OutboxKey = "framework.outbox"

// OutboxMessage describes one stored outbox message.
//
// OutboxMessage 描述一条已存储的 outbox 消息。
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

// OutboxStatus describes the lifecycle status of one outbox message.
//
// OutboxStatus 描述 outbox 消息的生命周期状态。
type OutboxStatus string

const (
	OutboxStatusPending  OutboxStatus = "pending"
	OutboxStatusSent     OutboxStatus = "sent"
	OutboxStatusFailed   OutboxStatus = "failed"
	OutboxStatusRetrying OutboxStatus = "retrying"
)

// OutboxStore defines the storage contract for outbox messages.
//
// OutboxStore 定义 outbox 消息存储契约。
type OutboxStore interface {
	Save(ctx context.Context, msg *OutboxMessage) error
	GetPending(ctx context.Context, limit int) ([]*OutboxMessage, error)
	MarkSent(ctx context.Context, id string) error
	MarkFailed(ctx context.Context, id string, err error) error
}

// OutboxSender defines the send contract for one outbox message.
//
// OutboxSender 定义单条 outbox 消息的发送契约。
type OutboxSender interface {
	Send(ctx context.Context, msg *OutboxMessage) error
}

// Outbox defines the reliable message emit/process contract.
//
// Outbox 定义可靠消息发射/处理契约。
type Outbox interface {
	Emit(ctx context.Context, topic string, payload interface{}) error
	EmitSync(ctx context.Context, topic string, payload interface{}) error
	Process(ctx context.Context) error
}

// OutboxConfig describes outbox runtime configuration.
//
// OutboxConfig 描述 outbox 运行时配置。
type OutboxConfig struct {
	Enabled    bool
	BatchSize  int
	RetryLimit int
	RetryDelay time.Duration
}
