// Package event_test provides unit tests for local event bus publish-subscribe behavior.
//
// 适用场景：
// - 验证 BaseEvent 的创建、发布和订阅处理逻辑。
// - 确保事件总线的并发安全和错误处理行为正确。
package event

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
	"github.com/stretchr/testify/assert"
)

// TestBaseEvent 验证 BaseEvent 的创建和字段赋值。
//
// 中文说明：
// - Name、Payload、OccurredAt 字段正确赋值。
// - OccurredAt 时间戳在合理范围内（当前时间 ±1 秒）。
func TestBaseEvent(t *testing.T) {
	payload := map[string]string{"key": "value"}
	event := NewBaseEvent("test.event", payload)

	assert.Equal(t, "test.event", event.Name())
	assert.Equal(t, payload, event.Payload())
	assert.NotZero(t, event.OccurredAt())
	assert.WithinDuration(t, time.Now(), event.OccurredAt(), time.Second)
}

// TestLocalEventBus_SubscribePublish 验证事件订阅与发布的基本流程。
//
// 中文说明：
// - 订阅后发布事件，处理器被正确调用。
// - 同一事件可重复发布，处理器每次都被触发。
func TestLocalEventBus_SubscribePublish(t *testing.T) {
	bus := NewLocalEventBus()

	// 订阅事件
	var count int32
	bus.Subscribe("user.created", func(ctx context.Context, event integrationcontract.Event) error {
		atomic.AddInt32(&count, 1)
		return nil
	})

	// 发布事件
	event := NewBaseEvent("user.created", map[string]string{"id": "123"})
	err := bus.Publish(context.Background(), event)
	assert.NoError(t, err)
	assert.Equal(t, int32(1), atomic.LoadInt32(&count))

	// 再次发布
	err = bus.Publish(context.Background(), event)
	assert.NoError(t, err)
	assert.Equal(t, int32(2), atomic.LoadInt32(&count))
}

// TestLocalEventBus_MultipleHandlers 验证同一事件可注册多个处理器。
//
// 中文说明：
// - 同一事件名可挂载多个 handler。
// - 发布事件时所有处理器按注册顺序依次执行。
func TestLocalEventBus_MultipleHandlers(t *testing.T) {
	bus := NewLocalEventBus()

	// 订阅多个处理器
	var results []string
	bus.Subscribe("test.event", func(ctx context.Context, event integrationcontract.Event) error {
		results = append(results, "handler1")
		return nil
	})
	bus.Subscribe("test.event", func(ctx context.Context, event integrationcontract.Event) error {
		results = append(results, "handler2")
		return nil
	})
	bus.Subscribe("test.event", func(ctx context.Context, event integrationcontract.Event) error {
		results = append(results, "handler3")
		return nil
	})

	// 发布事件
	event := NewBaseEvent("test.event", nil)
	err := bus.Publish(context.Background(), event)
	assert.NoError(t, err)
	assert.Equal(t, []string{"handler1", "handler2", "handler3"}, results)
}

// TestLocalEventBus_HandlerError 验证处理器返回错误时的行为。
//
// 中文说明：
// - Publish 返回第一个错误。
// - 即使某个处理器出错，其他处理器仍继续执行。
func TestLocalEventBus_HandlerError(t *testing.T) {
	bus := NewLocalEventBus()

	// 订阅处理器（一个失败，一个成功）
	var successCalled bool
	bus.Subscribe("test.event", func(ctx context.Context, event integrationcontract.Event) error {
		return errors.New("handler error")
	})
	bus.Subscribe("test.event", func(ctx context.Context, event integrationcontract.Event) error {
		successCalled = true
		return nil
	})

	// 发布事件（返回第一个错误，但其他处理器继续执行）
	event := NewBaseEvent("test.event", nil)
	err := bus.Publish(context.Background(), event)
	assert.Error(t, err)
	assert.Equal(t, "handler error", err.Error())
	assert.True(t, successCalled) // 其他处理器仍然执行
}

// TestLocalEventBus_NoSubscribers 验证发布无订阅者事件时的行为。
//
// 中文说明：
// - 发布到无订阅者的事件名时，Publish 返回 nil，不报错。
func TestLocalEventBus_NoSubscribers(t *testing.T) {
	bus := NewLocalEventBus()

	// 发布没有订阅者的事件
	event := NewBaseEvent("no.subscribers", nil)
	err := bus.Publish(context.Background(), event)
	assert.NoError(t, err) // 无订阅者时返回 nil
}

// TestLocalEventBus_PublishAsync 验证异步发布事件的行为。
//
// 中文说明：
// - PublishAsync 立即返回，处理器在后台异步执行。
// - 等待后处理器确实被触发。
func TestLocalEventBus_PublishAsync(t *testing.T) {
	bus := NewLocalEventBus()

	// 订阅事件
	var count int32
	bus.Subscribe("async.event", func(ctx context.Context, event integrationcontract.Event) error {
		atomic.AddInt32(&count, 1)
		return nil
	})

	// 异步发布
	event := NewBaseEvent("async.event", nil)
	err := bus.PublishAsync(context.Background(), event)
	assert.NoError(t, err)

	// 等待处理器执行
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, int32(1), atomic.LoadInt32(&count))
}

// TestLocalEventBus_Unsubscribe 验证取消订阅后不再接收事件。
//
// 中文说明：
// - 取消订阅后，同名事件发布不再触发处理器。
func TestLocalEventBus_Unsubscribe(t *testing.T) {
	bus := NewLocalEventBus()

	// 订阅事件
	var count int32
	bus.Subscribe("test.event", func(ctx context.Context, event integrationcontract.Event) error {
		atomic.AddInt32(&count, 1)
		return nil
	})

	// 发布（有订阅者）
	event := NewBaseEvent("test.event", nil)
	bus.Publish(context.Background(), event)
	assert.Equal(t, int32(1), atomic.LoadInt32(&count))

	// 取消订阅
	bus.Unsubscribe("test.event")

	// 发布（无订阅者）
	bus.Publish(context.Background(), event)
	assert.Equal(t, int32(1), atomic.LoadInt32(&count)) // 计数不变
}

// TestLocalEventBus_HasSubscribers 验证 HasSubscribers 的准确性。
//
// 中文说明：
// - 有订阅者时返回 true；取消订阅后返回 false。
func TestLocalEventBus_HasSubscribers(t *testing.T) {
	bus := NewLocalEventBus()

	assert.False(t, bus.HasSubscribers("test.event"))

	bus.Subscribe("test.event", func(ctx context.Context, event integrationcontract.Event) error {
		return nil
	})

	assert.True(t, bus.HasSubscribers("test.event"))

	bus.Unsubscribe("test.event")

	assert.False(t, bus.HasSubscribers("test.event"))
}
