// Package outbox provides in-memory outbox pattern implementation for gorp framework.
// Implements transactional outbox pattern for reliable message delivery.
// Suitable for testing and single-instance deployments.
//
// Outbox 包提供内存 Outbox 模式实现，用于 gorp 框架。
// 实现事务 Outbox 模式，确保可靠的消息投递。
// 适用于测试和单实例部署场景。
package outbox

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
)

// maxRetainedTerminalMessages 是终态（sent/failed）消息的保留上限。
// 超过后按进入终态的顺序逐出最旧的消息，防止内存实现在长期运行中
// 无界增长。需要完整历史时请使用持久化存储实现。
const maxRetainedTerminalMessages = 1000

// MemoryOutbox is the in-memory outbox pattern implementation.
// Core logic: Store messages, process with retry, track status.
//
// MemoryOutbox 是内存 Outbox 模式实现。
// 核心逻辑：存储消息、带重试处理、跟踪状态。
type MemoryOutbox struct {
	mu             sync.RWMutex
	messages       map[string]*integrationcontract.OutboxMessage
	terminalOrder  []string // 进入终态的消息 ID，按顺序，用于限量逐出
	sender         integrationcontract.OutboxSender
	config         integrationcontract.OutboxConfig
	processMu      sync.Mutex
	workerRunning  atomic.Bool
	pendingVersion atomic.Uint64
}

// NewMemoryOutbox creates a new in-memory outbox.
// Core logic: Initialize message map and store sender/config.
//
// NewMemoryOutbox 创建新的内存 Outbox。
// 核心逻辑：初始化消息 map 并存储 sender/config。
func NewMemoryOutbox(sender integrationcontract.OutboxSender, config integrationcontract.OutboxConfig) *MemoryOutbox {
	return &MemoryOutbox{
		messages: make(map[string]*integrationcontract.OutboxMessage),
		sender:   sender,
		config:   config,
	}
}

// Emit emits a message asynchronously to the outbox.
// Core logic: Create message with UUID, store, spawn background processor.
//
// Emit 异步将消息发送到 Outbox。
// 核心逻辑：创建带 UUID 的消息、存储、启动后台处理器。
func (o *MemoryOutbox) Emit(ctx context.Context, topic string, payload interface{}) error {
	msg := &integrationcontract.OutboxMessage{
		ID:        uuid.New().String(),
		Topic:     topic,
		Payload:   payload,
		Status:    integrationcontract.OutboxStatusPending,
		CreatedAt: time.Now(),
	}

	o.mu.Lock()
	o.messages[msg.ID] = msg
	o.mu.Unlock()
	o.pendingVersion.Add(1)
	o.scheduleProcess()

	return nil
}

// EmitSync emits a message synchronously and processes immediately.
// Core logic: Create message, store, call Process synchronously.
//
// EmitSync 同步发送消息并立即处理。
// 核心逻辑：创建消息、存储、同步调用 Process。
func (o *MemoryOutbox) EmitSync(ctx context.Context, topic string, payload interface{}) error {
	msg := &integrationcontract.OutboxMessage{
		ID:        uuid.New().String(),
		Topic:     topic,
		Payload:   payload,
		Status:    integrationcontract.OutboxStatusPending,
		CreatedAt: time.Now(),
	}

	o.mu.Lock()
	o.messages[msg.ID] = msg
	o.mu.Unlock()

	return o.Process(ctx)
}

// Process processes pending messages with retry logic.
// Core logic: Find pending/retrying messages, send with retry limit, update status.
//
// Process 处理待处理消息，带重试逻辑。
// 核心逻辑：找到待处理/重试中的消息、带重试限制发送、更新状态。
func (o *MemoryOutbox) Process(ctx context.Context) error {
	o.processMu.Lock()
	defer o.processMu.Unlock()

	o.mu.RLock()
	var pending []*integrationcontract.OutboxMessage
	for _, msg := range o.messages {
		if msg.Status == integrationcontract.OutboxStatusPending || msg.Status == integrationcontract.OutboxStatusRetrying {
			pending = append(pending, msg)
		}
	}
	o.mu.RUnlock()

	if len(pending) == 0 {
		return nil
	}

	if o.sender == nil {
		return errors.New("outbox sender not configured")
	}

	for _, msg := range pending {
		if msg.RetryCount >= o.config.RetryLimit {
			o.mu.Lock()
			if m, ok := o.messages[msg.ID]; ok {
				m.Status = integrationcontract.OutboxStatusFailed
				o.retainTerminalLocked(msg.ID)
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
					m.Status = integrationcontract.OutboxStatusFailed
					o.retainTerminalLocked(msg.ID)
				} else {
					m.Status = integrationcontract.OutboxStatusRetrying
				}
			}
			o.mu.Unlock()
			continue
		}

		now := time.Now()
		o.mu.Lock()
		if m, ok := o.messages[msg.ID]; ok {
			m.Status = integrationcontract.OutboxStatusSent
			m.SentAt = &now
			o.retainTerminalLocked(msg.ID)
		}
		o.mu.Unlock()
	}

	return nil
}

// retainTerminalLocked 记录进入终态的消息并逐出超限的最旧终态消息。
// 调用方必须持有 o.mu。
func (o *MemoryOutbox) retainTerminalLocked(id string) {
	o.terminalOrder = append(o.terminalOrder, id)
	if len(o.terminalOrder) <= maxRetainedTerminalMessages {
		return
	}
	evictCount := len(o.terminalOrder) - maxRetainedTerminalMessages
	for _, oldID := range o.terminalOrder[:evictCount] {
		delete(o.messages, oldID)
	}
	o.terminalOrder = append(o.terminalOrder[:0], o.terminalOrder[evictCount:]...)
}

func (o *MemoryOutbox) scheduleProcess() {
	if !o.workerRunning.CompareAndSwap(false, true) {
		return
	}
	go func() {
		for {
			version := o.pendingVersion.Load()
			_ = o.Process(context.Background())
			o.workerRunning.Store(false)
			if o.pendingVersion.Load() == version || !o.workerRunning.CompareAndSwap(false, true) {
				return
			}
		}
	}()
}

// MemoryOutboxStore is the in-memory outbox storage implementation.
// Core logic: Store messages, provide pending query, mark status.
//
// MemoryOutboxStore 是内存 Outbox 存储实现。
// 核心逻辑：存储消息、提供待处理查询、标记状态。
type MemoryOutboxStore struct {
	mu            sync.RWMutex
	messages      map[string]*integrationcontract.OutboxMessage
	terminalOrder []string // 终态消息 ID，按顺序，用于限量逐出
}

func NewMemoryOutboxStore() *MemoryOutboxStore {
	return &MemoryOutboxStore{
		messages: make(map[string]*integrationcontract.OutboxMessage),
	}
}

func (s *MemoryOutboxStore) Save(ctx context.Context, msg *integrationcontract.OutboxMessage) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messages[msg.ID] = msg
	return nil
}

func (s *MemoryOutboxStore) GetPending(ctx context.Context, limit int) ([]*integrationcontract.OutboxMessage, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*integrationcontract.OutboxMessage
	for _, msg := range s.messages {
		if msg.Status == integrationcontract.OutboxStatusPending || msg.Status == integrationcontract.OutboxStatusRetrying {
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
		msg.Status = integrationcontract.OutboxStatusSent
		msg.SentAt = &now
		s.retainTerminalLocked(id)
	}
	return nil
}

func (s *MemoryOutboxStore) MarkFailed(ctx context.Context, id string, err error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if msg, ok := s.messages[id]; ok {
		msg.Status = integrationcontract.OutboxStatusFailed
		if err != nil {
			msg.Error = err.Error()
		}
		s.retainTerminalLocked(id)
	}
	return nil
}

// retainTerminalLocked 记录进入终态的消息并逐出超限的最旧终态消息。
// 调用方必须持有 s.mu。
func (s *MemoryOutboxStore) retainTerminalLocked(id string) {
	s.terminalOrder = append(s.terminalOrder, id)
	if len(s.terminalOrder) <= maxRetainedTerminalMessages {
		return
	}
	evictCount := len(s.terminalOrder) - maxRetainedTerminalMessages
	for _, oldID := range s.terminalOrder[:evictCount] {
		delete(s.messages, oldID)
	}
	s.terminalOrder = append(s.terminalOrder[:0], s.terminalOrder[evictCount:]...)
}
