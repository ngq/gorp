// Application scenarios:
// - Define the in-process event contract shared by publishers, subscribers, and event buses.
// - Standardize event payload, occurrence time, and handler signatures.
// - Keep event-driven integration points framework-neutral.
//
// 适用场景：
// - 定义发布者、订阅者和事件总线共享的进程内事件契约。
// - 统一事件载荷、发生时间和处理器签名。
// - 保持事件驱动集成点与具体实现解耦。
package integration

import (
	"context"
	"time"
)

// EventKey is the container key for the event bus capability.
//
// EventKey 是事件总线能力的容器键。
const EventKey = "framework.event"

// Event describes one business or integration event.
//
// Event 描述一个业务或集成事件。
type Event interface {
	Name() string
	Payload() any
	OccurredAt() time.Time
}

// EventHandler handles one event.
//
// EventHandler 定义单个事件处理器。
type EventHandler func(ctx context.Context, event Event) error

// EventSubscriber defines the event subscription contract.
//
// EventSubscriber 定义事件订阅契约。
type EventSubscriber interface {
	Subscribe(eventName string, handler EventHandler)
}

// EventPublisher defines the event publishing contract.
//
// EventPublisher 定义事件发布契约。
type EventPublisher interface {
	Publish(ctx context.Context, event Event) error
	PublishAsync(ctx context.Context, event Event) error
}

// EventBus combines publishing and subscription capabilities.
//
// EventBus 组合发布和订阅能力。
type EventBus interface {
	EventSubscriber
	EventPublisher
}
