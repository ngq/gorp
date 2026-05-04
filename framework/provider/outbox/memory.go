package outbox

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
)

type MemoryOutbox struct {
	mu       sync.RWMutex
	messages map[string]*integrationcontract.OutboxMessage
	sender   integrationcontract.OutboxSender
	config   integrationcontract.OutboxConfig
}

func NewMemoryOutbox(sender integrationcontract.OutboxSender, config integrationcontract.OutboxConfig) *MemoryOutbox {
	return &MemoryOutbox{
		messages: make(map[string]*integrationcontract.OutboxMessage),
		sender:   sender,
		config:   config,
	}
}

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

	go o.Process(context.Background())

	return nil
}

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

func (o *MemoryOutbox) Process(ctx context.Context) error {
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
		return fmt.Errorf("outbox sender not configured")
	}

	for _, msg := range pending {
		if msg.RetryCount >= o.config.RetryLimit {
			o.mu.Lock()
			if m, ok := o.messages[msg.ID]; ok {
				m.Status = integrationcontract.OutboxStatusFailed
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
		}
		o.mu.Unlock()
	}

	return nil
}

type MemoryOutboxStore struct {
	mu       sync.RWMutex
	messages map[string]*integrationcontract.OutboxMessage
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
	}
	return nil
}

func (s *MemoryOutboxStore) MarkFailed(ctx context.Context, id string, err error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if msg, ok := s.messages[id]; ok {
		msg.Status = integrationcontract.OutboxStatusFailed
		msg.Error = err.Error()
	}
	return nil
}
