// Package event provides local in-memory event bus implementation for gorp framework.
// Suitable for single-application internal event communication.
// Supports synchronous and asynchronous publishing, multiple handlers per event.
//
// 事件包提供本地内存事件总线实现，用于 gorp 框架。
// 适用于单体应用内部事件通信。
// 支持同步和异步发布、单事件多处理器。
package event

import (
	"context"
	"fmt"
	"sync"
	"time"

	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
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
	subscribers map[string][]integrationcontract.EventHandler
	options     EventBusOptions
}

// NewLocalEventBus 创建具备重试与 DLQ 能力的本地事件总线。
func NewLocalEventBus(opts ...Option) *LocalEventBus {
	o := DefaultEventBusOptions()
	for _, opt := range opts {
		opt(&o)
	}
	return &LocalEventBus{
		subscribers: make(map[string][]integrationcontract.EventHandler),
		options:     o,
	}
}

// Subscribe 订阅事件。
func (b *LocalEventBus) Subscribe(eventName string, handler integrationcontract.EventHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.subscribers[eventName] = append(b.subscribers[eventName], handler)
}

// Publish 同步发布事件（带有指数退避重试与死信队列 DLQ 路由）。
func (b *LocalEventBus) Publish(ctx context.Context, event integrationcontract.Event) error {
	b.mu.RLock()
	handlers := b.subscribers[event.Name()]
	b.mu.RUnlock()

	if len(handlers) == 0 {
		return nil
	}

	var firstErr error
	for _, handler := range handlers {
		if err := b.invokeWithRetry(ctx, event.Name(), handler, event); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// invokeWithRetry 包含了 CloudEvent Context 恢复、指数退避重试与死信队列路由。
func (b *LocalEventBus) invokeWithRetry(ctx context.Context, eventName string, handler integrationcontract.EventHandler, event integrationcontract.Event) (err error) {
	execCtx := restoreContextFromEvent(ctx, event)

	backoff := b.options.InitialBackoff
	maxRetries := b.options.MaxRetries
	if maxRetries < 0 {
		maxRetries = 0
	}

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(backoff)
			backoff *= 2
			if backoff > b.options.MaxBackoff && b.options.MaxBackoff > 0 {
				backoff = b.options.MaxBackoff
			}
		}

		err = func() (handlerErr error) {
			defer func() {
				if r := recover(); r != nil {
					handlerErr = fmt.Errorf("event %s: handler panicked: %v", eventName, r)
				}
			}()
			return handler(execCtx, event)
		}()

		if err == nil {
			return nil // 处理成功，退出重试
		}
	}

	// 达到最大重试次数仍失败，触发死信队列 (DLQ) 处理
	if b.options.EnableDLQ && b.options.DLQHandler != nil {
		b.options.DLQHandler(execCtx, event, err)
	}

	return err
}

func restoreContextFromEvent(ctx context.Context, event integrationcontract.Event) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	// 如果是 CloudEvent 类型，自动提取 Extensions 中的 Trace 信息并注入 Context
	if ce, ok := event.(interface{ ExtensionsMap() map[string]any }); ok {
		ext := ce.ExtensionsMap()
		if traceID, ok := ext["trace_id"].(string); ok && traceID != "" {
			ctx = context.WithValue(ctx, "trace_id", traceID)
		}
		if reqID, ok := ext["request_id"].(string); ok && reqID != "" {
			ctx = context.WithValue(ctx, "request_id", reqID)
		}
	}
	return ctx
}

// PublishAsync 异步发布事件。
//
// 中文说明：
// - 立即返回，事件在后台 goroutine 中处理；
// - 使用 goroutine.SafeGo 包装，确保 panic 不会导致程序崩溃；
// - 适合不需要等待处理完成的场景。
func (b *LocalEventBus) PublishAsync(ctx context.Context, event integrationcontract.Event) error {
	b.mu.RLock()
	handlers := b.subscribers[event.Name()]
	b.mu.RUnlock()

	if len(handlers) == 0 {
		return nil
	}

	// 在后台 goroutine 中执行处理器（带有重试与 DLQ 保护）
	goroutine.SafeGo(ctx, nil, func(ctx context.Context) {
		for _, handler := range handlers {
			_ = b.invokeWithRetry(ctx, event.Name(), handler, event)
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
