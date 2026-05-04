package integration

import (
	"context"
	"time"
)

const EventKey = "framework.event"

type Event interface {
	Name() string
	Payload() interface{}
	OccurredAt() time.Time
}

type EventHandler func(ctx context.Context, event Event) error

type EventSubscriber interface {
	Subscribe(eventName string, handler EventHandler)
}

type EventPublisher interface {
	Publish(ctx context.Context, event Event) error
	PublishAsync(ctx context.Context, event Event) error
}

type EventBus interface {
	EventSubscriber
	EventPublisher
}
