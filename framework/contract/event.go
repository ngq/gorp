package contract

import (
	"context"
	"time"
)

const EventKey = "framework.event"

// Event 定义事件接口。
//
// 中文说明：
// - 所有事件都需要实现此接口；
// - Name 返回事件名称，用于路由到对应的处理器；
// - Payload 返回事件载荷；
// - OccurredAt 返回事件发生时间。
type Event interface {
	Name() string
	Payload() interface{}
	OccurredAt() time.Time
}

// EventHandler 事件处理函数。
//
// 中文说明：
// - ctx 为请求上下文，可用于获取 trace id 等；
// - event 为事件对象；
// - 返回 error 表示处理失败，事件总线可根据策略决定是否重试。
type EventHandler func(ctx context.Context, event Event) error

// EventSubscriber 事件订阅者接口。
//
// 中文说明：
// - Subscribe 注册事件处理器；
// - 支持一个事件多个处理器（广播模式）。
type EventSubscriber interface {
	Subscribe(eventName string, handler EventHandler)
}

// EventPublisher 事件发布者接口。
//
// 中文说明：
// - Publish 同步发布事件，阻塞直到所有处理器执行完毕；
// - PublishAsync 异步发布事件，立即返回不等待处理器执行。
type EventPublisher interface {
	Publish(ctx context.Context, event Event) error
	PublishAsync(ctx context.Context, event Event) error
}

// EventBus 事件总线接口。
//
// 中文说明：
// - 整合订阅和发布能力；
// - 当前为本地内存实现，后续可演进为 MQ/Outbox。
type EventBus interface {
	EventSubscriber
	EventPublisher
}