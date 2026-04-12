package event

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ngq/gorp/framework/contract"
	"github.com/stretchr/testify/assert"
)

func TestBaseEvent(t *testing.T) {
	// 测试创建事件
	payload := map[string]string{"key": "value"}
	event := NewBaseEvent("test.event", payload)

	assert.Equal(t, "test.event", event.Name())
	assert.Equal(t, payload, event.Payload())
	assert.NotZero(t, event.OccurredAt())
	assert.WithinDuration(t, time.Now(), event.OccurredAt(), time.Second)
}

func TestLocalEventBus_SubscribePublish(t *testing.T) {
	bus := NewLocalEventBus()

	// 订阅事件
	var count int32
	bus.Subscribe("user.created", func(ctx context.Context, event contract.Event) error {
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

func TestLocalEventBus_MultipleHandlers(t *testing.T) {
	bus := NewLocalEventBus()

	// 订阅多个处理器
	var results []string
	bus.Subscribe("test.event", func(ctx context.Context, event contract.Event) error {
		results = append(results, "handler1")
		return nil
	})
	bus.Subscribe("test.event", func(ctx context.Context, event contract.Event) error {
		results = append(results, "handler2")
		return nil
	})
	bus.Subscribe("test.event", func(ctx context.Context, event contract.Event) error {
		results = append(results, "handler3")
		return nil
	})

	// 发布事件
	event := NewBaseEvent("test.event", nil)
	err := bus.Publish(context.Background(), event)
	assert.NoError(t, err)
	assert.Equal(t, []string{"handler1", "handler2", "handler3"}, results)
}

func TestLocalEventBus_HandlerError(t *testing.T) {
	bus := NewLocalEventBus()

	// 订阅处理器（一个失败，一个成功）
	var successCalled bool
	bus.Subscribe("test.event", func(ctx context.Context, event contract.Event) error {
		return errors.New("handler error")
	})
	bus.Subscribe("test.event", func(ctx context.Context, event contract.Event) error {
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

func TestLocalEventBus_NoSubscribers(t *testing.T) {
	bus := NewLocalEventBus()

	// 发布没有订阅者的事件
	event := NewBaseEvent("no.subscribers", nil)
	err := bus.Publish(context.Background(), event)
	assert.NoError(t, err) // 无订阅者时返回 nil
}

func TestLocalEventBus_PublishAsync(t *testing.T) {
	bus := NewLocalEventBus()

	// 订阅事件
	var count int32
	bus.Subscribe("async.event", func(ctx context.Context, event contract.Event) error {
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

func TestLocalEventBus_Unsubscribe(t *testing.T) {
	bus := NewLocalEventBus()

	// 订阅事件
	var count int32
	bus.Subscribe("test.event", func(ctx context.Context, event contract.Event) error {
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

func TestLocalEventBus_HasSubscribers(t *testing.T) {
	bus := NewLocalEventBus()

	assert.False(t, bus.HasSubscribers("test.event"))

	bus.Subscribe("test.event", func(ctx context.Context, event contract.Event) error {
		return nil
	})

	assert.True(t, bus.HasSubscribers("test.event"))

	bus.Unsubscribe("test.event")

	assert.False(t, bus.HasSubscribers("test.event"))
}