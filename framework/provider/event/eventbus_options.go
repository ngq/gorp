// Package event provides EventBus options for retry and DLQ configurations.
package event

import (
	"context"
	"time"

	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
)

// DLQHandler is the callback function invoked when an event exhausts all retry attempts.
type DLQHandler func(ctx context.Context, event integrationcontract.Event, err error)

// EventBusOptions configures the EventBus behavior including exponential retries and DLQ.
type EventBusOptions struct {
	MaxRetries       int           // Maximum retry attempts (default: 3)
	InitialBackoff   time.Duration // Initial retry delay (default: 50ms)
	MaxBackoff       time.Duration // Maximum retry delay (default: 1s)
	EnableDLQ        bool          // Enable Dead-Letter Queue routing (default: true)
	DLQHandler       DLQHandler    // Custom DLQ handler
	AsyncWorkerCount int           // Number of background workers for PublishAsync (default: 8)
	AsyncQueueSize   int           // Buffer size for async event queue (default: 1024)
}

// DefaultEventBusOptions returns sensible defaults for production.
func DefaultEventBusOptions() EventBusOptions {
	return EventBusOptions{
		MaxRetries:       3,
		InitialBackoff:   50 * time.Millisecond,
		MaxBackoff:       1 * time.Second,
		EnableDLQ:        true,
		AsyncWorkerCount: 8,
		AsyncQueueSize:   1024,
	}
}

// Option configures EventBusOptions.
type Option func(*EventBusOptions)

// WithMaxRetries configures max retry attempts.
func WithMaxRetries(maxRetries int) Option {
	return func(o *EventBusOptions) {
		o.MaxRetries = maxRetries
	}
}

// WithBackoff configures exponential backoff delays.
func WithBackoff(initial, max time.Duration) Option {
	return func(o *EventBusOptions) {
		o.InitialBackoff = initial
		o.MaxBackoff = max
	}
}

// WithDLQHandler configures the Dead Letter Queue handler callback.
func WithDLQHandler(handler DLQHandler) Option {
	return func(o *EventBusOptions) {
		o.EnableDLQ = true
		o.DLQHandler = handler
	}
}

// WithAsyncWorkers configures the worker pool and queue size for asynchronous event publishing.
func WithAsyncWorkers(workerCount, queueSize int) Option {
	return func(o *EventBusOptions) {
		if workerCount > 0 {
			o.AsyncWorkerCount = workerCount
		}
		if queueSize > 0 {
			o.AsyncQueueSize = queueSize
		}
	}
}

