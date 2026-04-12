package outbox

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/ngq/gorp/framework/contract"
)

// MemoryOutbox 内存 Outbox 实现。
//
// 中文说明：
// - 基于 sync.Map 的内存实现，适合开发测试。
// - 演进方向：生产环境应替换为数据库持久化版本，确保消息不丢失。
// - 本实现演示 Outbox 模式的基本流程。
type MemoryOutbox struct {
	mu      sync.RWMutex
	messages map[string]*contract.OutboxMessage
	sender   contract.OutboxSender
	config   contract.OutboxConfig
}

// NewMemoryOutbox 创建内存 Outbox。
func NewMemoryOutbox(sender contract.OutboxSender, config contract.OutboxConfig) *MemoryOutbox {
	return &MemoryOutbox{
		messages: make(map[string]*contract.OutboxMessage),
		sender:   sender,
		config:   config,
	}
}

// Emit 发送消息（存入 outbox 后异步发送）。
func (o *MemoryOutbox) Emit(ctx context.Context, topic string, payload interface{}) error {
	msg := &contract.OutboxMessage{
		ID:        uuid.New().String(),
		Topic:     topic,
		Payload:   payload,
		Status:    contract.OutboxStatusPending,
		CreatedAt: time.Now(),
	}

	o.mu.Lock()
	o.messages[msg.ID] = msg
	o.mu.Unlock()

	// 异步处理
	go o.Process(context.Background())

	return nil
}

// EmitSync 同步发送消息。
func (o *MemoryOutbox) EmitSync(ctx context.Context, topic string, payload interface{}) error {
	msg := &contract.OutboxMessage{
		ID:        uuid.New().String(),
		Topic:     topic,
		Payload:   payload,
		Status:    contract.OutboxStatusPending,
		CreatedAt: time.Now(),
	}

	o.mu.Lock()
	o.messages[msg.ID] = msg
	o.mu.Unlock()

	return o.Process(ctx)
}

// Process 处理待发送的消息。
func (o *MemoryOutbox) Process(ctx context.Context) error {
	o.mu.RLock()
	var pending []*contract.OutboxMessage
	for _, msg := range o.messages {
		if msg.Status == contract.OutboxStatusPending || msg.Status == contract.OutboxStatusRetrying {
			pending = append(pending, msg)
		}
	}
	o.mu.RUnlock()

	if len(pending) == 0 {
		return nil
	}

	if o.sender == nil {
		return fmt.Errorf("outbox sender not configured")
	}

	for _, msg := range pending {
		if msg.RetryCount >= o.config.RetryLimit {
			o.mu.Lock()
			if m, ok := o.messages[msg.ID]; ok {
				m.Status = contract.OutboxStatusFailed
			}
			o.mu.Unlock()
			continue
		}

		if err := o.sender.Send(ctx, msg); err != nil {
			o.mu.Lock()
			if m, ok := o.messages[msg.ID]; ok {
				m.RetryCount++
				m.Error = err.Error()
				if m.RetryCount >= o.config.RetryLimit {
					m.Status = contract.OutboxStatusFailed
				} else {
					m.Status = contract.OutboxStatusRetrying
				}
			}
			o.mu.Unlock()
			continue
		}

		now := time.Now()
		o.mu.Lock()
		if m, ok := o.messages[msg.ID]; ok {
			m.Status = contract.OutboxStatusSent
			m.SentAt = &now
		}
		o.mu.Unlock()
	}

	return nil
}

// MemoryOutboxStore 内存 Outbox 存储实现。
type MemoryOutboxStore struct {
	mu       sync.RWMutex
	messages map[string]*contract.OutboxMessage
}

// NewMemoryOutboxStore 创建内存存储。
func NewMemoryOutboxStore() *MemoryOutboxStore {
	return &MemoryOutboxStore{
		messages: make(map[string]*contract.OutboxMessage),
	}
}

func (s *MemoryOutboxStore) Save(ctx context.Context, msg *contract.OutboxMessage) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messages[msg.ID] = msg
	return nil
}

func (s *MemoryOutboxStore) GetPending(ctx context.Context, limit int) ([]*contract.OutboxMessage, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*contract.OutboxMessage
	for _, msg := range s.messages {
		if msg.Status == contract.OutboxStatusPending || msg.Status == contract.OutboxStatusRetrying {
			result = append(result, msg)
			if len(result) >= limit {
				break
			}
		}
	}
	return result, nil
}

func (s *MemoryOutboxStore) MarkSent(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if msg, ok := s.messages[id]; ok {
		now := time.Now()
		msg.Status = contract.OutboxStatusSent
		msg.SentAt = &now
	}
	return nil
}

func (s *MemoryOutboxStore) MarkFailed(ctx context.Context, id string, err error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if msg, ok := s.messages[id]; ok {
		msg.Status = contract.OutboxStatusFailed
		msg.Error = err.Error()
	}
	return nil
}