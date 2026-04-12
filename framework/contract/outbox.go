package contract

import (
	"context"
	"time"
)

const OutboxKey = "framework.outbox"

// OutboxMessage Outbox 消息。
//
// 中文说明：
// - 表示一个待发送的消息；
// - 包含消息 ID、主题、载荷、状态等信息。
type OutboxMessage struct {
	ID          string      // 消息唯一 ID
	Topic       string      // 消息主题/路由键
	Payload     interface{} // 消息载荷
	Status      OutboxStatus // 消息状态
	RetryCount  int         // 重试次数
	CreatedAt   time.Time   // 创建时间
	SentAt      *time.Time  // 发送时间（成功后设置）
	Error       string      // 错误信息（失败时记录）
}

// OutboxStatus Outbox 消息状态。
type OutboxStatus string

const (
	OutboxStatusPending   OutboxStatus = "pending"   // 待发送
	OutboxStatusSent      OutboxStatus = "sent"      // 已发送
	OutboxStatusFailed    OutboxStatus = "failed"    // 发送失败
	OutboxStatusRetrying  OutboxStatus = "retrying"  // 重试中
)

// OutboxStore Outbox 存储接口。
//
// 中文说明：
// - 定义 Outbox 消息的存储接口；
// - Save 保存消息（通常在业务事务中调用）；
// - GetPending 获取待发送的消息；
// - MarkSent 标记消息为已发送；
// - MarkFailed 标记消息为发送失败。
type OutboxStore interface {
	// Save 保存消息到 outbox
	Save(ctx context.Context, msg *OutboxMessage) error
	// GetPending 获取待发送的消息列表
	GetPending(ctx context.Context, limit int) ([]*OutboxMessage, error)
	// MarkSent 标记消息为已发送
	MarkSent(ctx context.Context, id string) error
	// MarkFailed 标记消息为发送失败
	MarkFailed(ctx context.Context, id string, err error) error
}

// OutboxSender Outbox 发送器接口。
//
// 中文说明：
// - 定义消息发送接口；
// - 具体实现可以是 Kafka/RabbitMQ/Redis Stream 等。
type OutboxSender interface {
	// Send 发送消息到目标系统
	Send(ctx context.Context, msg *OutboxMessage) error
}

// Outbox Outbox 接入位接口。
//
// 中文说明：
// - 整合存储和发送能力；
// - 提供统一的消息投递入口；
// - 当前先定义抽象，后续根据需要实现。
type Outbox interface {
	// Emit 发送消息（存入 outbox 后异步发送）
	Emit(ctx context.Context, topic string, payload interface{}) error
	// EmitSync 同步发送消息（存入 outbox 并立即尝试发送）
	EmitSync(ctx context.Context, topic string, payload interface{}) error
	// Process 处理待发送的消息（由后台任务调用）
	Process(ctx context.Context) error
}

// OutboxConfig Outbox 配置。
type OutboxConfig struct {
	// Enabled 是否启用 Outbox
	Enabled bool
	// BatchSize 每次处理的消息数量
	BatchSize int
	// RetryLimit 最大重试次数
	RetryLimit int
	// RetryDelay 重试间隔
	RetryDelay time.Duration
}