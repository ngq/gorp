package event

import (
	"context"
	"sync"
	"time"

	"github.com/ngq/gorp/framework/contract"
	"github.com/ngq/gorp/framework/goroutine"
)

// BaseEvent 基础事件实现。
//
// 中文说明：
// - 提供事件的默认实现，可直接使用或作为嵌入；
// - 包含事件名称、载荷和时间戳。
type BaseEvent struct {
	name       string
	payload    interface{}
	occurredAt time.Time
}

// NewBaseEvent 创建基础事件。
func NewBaseEvent(name string, payload interface{}) *BaseEvent {
	return &BaseEvent{
		name:       name,
		payload:    payload,
		occurredAt: time.Now(),
	}
}

func (e *BaseEvent) Name() string          { return e.name }
func (e *BaseEvent) Payload() interface{}  { return e.payload }
func (e *BaseEvent) OccurredAt() time.Time { return e.occurredAt }

// LocalEventBus 本地内存事件总线。
//
// 中文说明：
// - 基于 goroutine 的本地事件总线实现；
// - 支持同步和异步发布；
// - 支持一个事件多个处理器（广播模式）；
// - 适合单体应用内部事件通信，后续可演进为 MQ。
type LocalEventBus struct {
	mu          sync.RWMutex
	subscribers map[string][]contract.EventHandler
}

// NewLocalEventBus 创建本地事件总线。
func NewLocalEventBus() *LocalEventBus {
	return &LocalEventBus{
		subscribers: make(map[string][]contract.EventHandler),
	}
}

// Subscribe 订阅事件。
//
// 中文说明：
// - 注册事件处理器；
// - 同一事件可以有多个处理器，按注册顺序执行；
// - 处理器执行失败不会影响其他处理器的执行。
func (b *LocalEventBus) Subscribe(eventName string, handler contract.EventHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.subscribers[eventName] = append(b.subscribers[eventName], handler)
}

// Publish 同步发布事件。
//
// 中文说明：
// - 阻塞直到所有处理器执行完毕；
// - 处理器按注册顺序依次执行；
// - 如果某个处理器失败，记录错误但继续执行其他处理器；
// - 返回第一个遇到的错误。
func (b *LocalEventBus) Publish(ctx context.Context, event contract.Event) error {
	b.mu.RLock()
	handlers := b.subscribers[event.Name()]
	b.mu.RUnlock()

	if len(handlers) == 0 {
		return nil
	}

	var firstErr error
	for _, handler := range handlers {
		if err := handler(ctx, event); err != nil {
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

// PublishAsync 异步发布事件。
//
// 中文说明：
// - 立即返回，事件在后台 goroutine 中处理；
// - 使用 goroutine.SafeGo 包装，确保 panic 不会导致程序崩溃；
// - 适合不需要等待处理完成的场景。
func (b *LocalEventBus) PublishAsync(ctx context.Context, event contract.Event) error {
	b.mu.RLock()
	handlers := b.subscribers[event.Name()]
	b.mu.RUnlock()

	if len(handlers) == 0 {
		return nil
	}

	// 在后台 goroutine 中执行处理器
	goroutine.SafeGo(ctx, nil, func(ctx context.Context) {
		for _, handler := range handlers {
			_ = handler(ctx, event)
		}
	})

	return nil
}

// Unsubscribe 取消订阅事件。
//
// 中文说明：
// - 移除指定事件的所有处理器；
// - 用于清理或重新配置事件处理。
func (b *LocalEventBus) Unsubscribe(eventName string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.subscribers, eventName)
}

// HasSubscribers 检查事件是否有订阅者。
func (b *LocalEventBus) HasSubscribers(eventName string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	handlers, ok := b.subscribers[eventName]
	return ok && len(handlers) > 0
}