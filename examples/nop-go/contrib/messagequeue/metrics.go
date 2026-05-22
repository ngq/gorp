// Package messagequeue provides shared metrics for MQ implementations.
// This file implements Prometheus metrics for message queue operations.
//
// 消息队列包提供 MQ 实现的共享指标。
// 本文件实现消息队列操作的 Prometheus 指标。
package messagequeue

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// MetricsRecorder records Prometheus metrics for message queue operations.
// This provides a reusable metrics layer for all MQ provider implementations.
//
// MetricsRecorder 记录消息队列操作的 Prometheus 指标。
// 这为所有 MQ provider 实现提供可复用的指标层。
type MetricsRecorder struct {
	// Published messages counter, labeled by topic and status (success/error).
	// 已发布消息计数器，按 topic 和 status 标签分类。
	messagesPublished *prometheus.CounterVec

	// Consumed messages counter, labeled by topic and status (success/error).
	// 已消费消息计数器，按 topic 和 status 标签分类。
	messagesConsumed *prometheus.CounterVec

	// Current active subscriptions gauge, labeled by topic.
	// 当前活跃订阅数，按 topic 标签分类。
	activeSubscriptions *prometheus.GaugeVec

	// Publish latency histogram, labeled by topic.
	// 发布延迟直方图，按 topic 标签分类。
	publishLatency *prometheus.HistogramVec

	// Consume latency histogram, labeled by topic.
	// 消费延迟直方图，按 topic 标签分类。
	consumeLatency *prometheus.HistogramVec

	// Message size histogram, labeled by operation (publish/consume).
	// 消息大小直方图，按 operation 标签分类。
	messageSize *prometheus.HistogramVec

	// Retry counter, labeled by topic and operation (publish/consume).
	// 重试计数器，按 topic 和 operation 标签分类。
	retriesTotal *prometheus.CounterVec

	// Dead letter queue counter, labeled by topic.
	// 死信队列计数器，按 topic 标签分类。
	deadLetterTotal *prometheus.CounterVec
}

// NewMetricsRecorder creates a new MQ metrics recorder.
// Metrics are registered with the default Prometheus registry.
//
// NewMetricsRecorder 创建新的 MQ 指标记录器。
// 指标注册到默认 Prometheus 注册表。
func NewMetricsRecorder() *MetricsRecorder {
	return &MetricsRecorder{
		messagesPublished: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "gorp_mq_messages_published_total",
				Help: "Total number of messages published to the message queue.",
			},
			[]string{"topic", "status"},
		),
		messagesConsumed: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "gorp_mq_messages_consumed_total",
				Help: "Total number of messages consumed from the message queue.",
			},
			[]string{"topic", "status"},
		),
		activeSubscriptions: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "gorp_mq_active_subscriptions",
				Help: "Current number of active subscriptions.",
			},
			[]string{"topic"},
		),
		publishLatency: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "gorp_mq_publish_latency_seconds",
				Help:    "Latency of message publishing in seconds.",
				Buckets: prometheus.ExponentialBuckets(0.001, 2, 15), // 1ms to ~16s
			},
			[]string{"topic"},
		),
		consumeLatency: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "gorp_mq_consume_latency_seconds",
				Help:    "Latency of message consumption in seconds.",
				Buckets: prometheus.ExponentialBuckets(0.001, 2, 15), // 1ms to ~16s
			},
			[]string{"topic"},
		),
		messageSize: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "gorp_mq_message_size_bytes",
				Help:    "Size of messages in bytes.",
				Buckets: prometheus.ExponentialBuckets(64, 2, 15), // 64B to ~2MB
			},
			[]string{"operation"},
		),
		retriesTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "gorp_mq_retries_total",
				Help: "Total number of message retries.",
			},
			[]string{"topic", "operation"},
		),
		deadLetterTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "gorp_mq_dead_letter_total",
				Help: "Total number of messages sent to dead letter queue.",
			},
			[]string{"topic"},
		),
	}
}

// OnPublish records a publish operation.
// status should be "success" or "error".
//
// OnPublish 记录发布操作。
// status 应为 "success" 或 "error"。
func (m *MetricsRecorder) OnPublish(topic string, status string, messageSize int, latencySeconds float64) {
	m.messagesPublished.WithLabelValues(topic, status).Inc()
	m.publishLatency.WithLabelValues(topic).Observe(latencySeconds)
	if messageSize > 0 {
		m.messageSize.WithLabelValues("publish").Observe(float64(messageSize))
	}
}

// OnConsume records a consume operation.
// status should be "success" or "error".
//
// OnConsume 记录消费操作。
// status 应为 "success" 或 "error"。
func (m *MetricsRecorder) OnConsume(topic string, status string, messageSize int, latencySeconds float64) {
	m.messagesConsumed.WithLabelValues(topic, status).Inc()
	m.consumeLatency.WithLabelValues(topic).Observe(latencySeconds)
	if messageSize > 0 {
		m.messageSize.WithLabelValues("consume").Observe(float64(messageSize))
	}
}

// OnSubscribe records a new subscription.
//
// OnSubscribe 记录新订阅。
func (m *MetricsRecorder) OnSubscribe(topic string) {
	m.activeSubscriptions.WithLabelValues(topic).Inc()
}

// OnUnsubscribe records an unsubscribe event.
//
// OnUnsubscribe 记录取消订阅事件。
func (m *MetricsRecorder) OnUnsubscribe(topic string) {
	m.activeSubscriptions.WithLabelValues(topic).Dec()
}

// OnRetry records a retry operation.
// operation should be "publish" or "consume".
//
// OnRetry 记录重试操作。
// operation 应为 "publish" 或 "consume"。
func (m *MetricsRecorder) OnRetry(topic string, operation string) {
	m.retriesTotal.WithLabelValues(topic, operation).Inc()
}

// OnDeadLetter records a message sent to dead letter queue.
//
// OnDeadLetter 记录发送到死信队列的消息。
func (m *MetricsRecorder) OnDeadLetter(topic string) {
	m.deadLetterTotal.WithLabelValues(topic).Inc()
}
