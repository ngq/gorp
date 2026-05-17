// Package websocket provides WebSocket capability provider for gorp framework.
// This file implements Prometheus metrics for WebSocket connections.
//
// WebSocket 包提供 gorp 框架的 WebSocket 能力 provider。
// 本文件实现 WebSocket 连接的 Prometheus 指标。
package websocket

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// wsConnectionsGauge 当前 WebSocket 连接数
	wsConnectionsGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "gorp_websocket_connections",
		Help: "Current number of WebSocket connections.",
	})

	// wsConnectionsTotal WebSocket 连接总数
	wsConnectionsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "gorp_websocket_connections_total",
		Help: "Total number of WebSocket connections established.",
	})

	// wsDisconnectionsTotal WebSocket 断开连接总数
	wsDisconnectionsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "gorp_websocket_disconnections_total",
		Help: "Total number of WebSocket connections closed.",
	})

	// wsMessagesReceivedTotal 接收的消息总数
	wsMessagesReceivedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gorp_websocket_messages_received_total",
		Help: "Total number of WebSocket messages received.",
	}, []string{"type"})

	// wsMessagesSentTotal 发送的消息总数
	wsMessagesSentTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gorp_websocket_messages_sent_total",
		Help: "Total number of WebSocket messages sent.",
	}, []string{"type"})

	// wsMessageLatency 消息处理延迟
	wsMessageLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "gorp_websocket_message_latency_seconds",
		Help:    "WebSocket message processing latency in seconds.",
		Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
	}, []string{"type"})

	// wsErrorsTotal 错误总数
	wsErrorsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gorp_websocket_errors_total",
		Help: "Total number of WebSocket errors.",
	}, []string{"type"})

	// wsBroadcastTotal 广播消息总数
	wsBroadcastTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gorp_websocket_broadcast_total",
		Help: "Total number of WebSocket broadcast messages.",
	}, []string{"type"})
)

// WSMetricsRecorder records WebSocket metrics.
// Use this to record connection and message events.
//
// WSMetricsRecorder 记录 WebSocket 指标。
// 用于记录连接和消息事件。
type WSMetricsRecorder struct{}

// NewWSMetricsRecorder creates a new WebSocket metrics recorder.
//
// NewWSMetricsRecorder 创建新的 WebSocket 指标记录器。
func NewWSMetricsRecorder() *WSMetricsRecorder {
	return &WSMetricsRecorder{}
}

// OnConnect records a new WebSocket connection.
//
// OnConnect 记录新的 WebSocket 连接。
func (r *WSMetricsRecorder) OnConnect() {
	wsConnectionsGauge.Inc()
	wsConnectionsTotal.Inc()
}

// OnDisconnect records a WebSocket disconnection.
//
// OnDisconnect 记录 WebSocket 断开连接。
func (r *WSMetricsRecorder) OnDisconnect() {
	wsConnectionsGauge.Dec()
	wsDisconnectionsTotal.Inc()
}

// OnMessageReceived records a received message.
// messageType: "text" or "binary"
//
// OnMessageReceived 记录接收的消息。
// messageType: "text" 或 "binary"
func (r *WSMetricsRecorder) OnMessageReceived(messageType string, latency time.Duration) {
	wsMessagesReceivedTotal.WithLabelValues(messageType).Inc()
	wsMessageLatency.WithLabelValues(messageType).Observe(latency.Seconds())
}

// OnMessageSent records a sent message.
// messageType: "text" or "binary"
//
// OnMessageSent 记录发送的消息。
// messageType: "text" 或 "binary"
func (r *WSMetricsRecorder) OnMessageSent(messageType string) {
	wsMessagesSentTotal.WithLabelValues(messageType).Inc()
}

// OnError records an error.
// errorType: "read", "write", "upgrade", etc.
//
// OnError 记录错误。
// errorType: "read", "write", "upgrade" 等。
func (r *WSMetricsRecorder) OnError(errorType string) {
	wsErrorsTotal.WithLabelValues(errorType).Inc()
}

// OnBroadcast records a broadcast message.
// messageType: "text" or "binary"
//
// OnBroadcast 记录广播消息。
// messageType: "text" 或 "binary"
func (r *WSMetricsRecorder) OnBroadcast(messageType string) {
	wsBroadcastTotal.WithLabelValues(messageType).Inc()
}

// SetConnections sets the current number of connections.
// Use this for cluster mode where connections are tracked globally.
//
// SetConnections 设置当前连接数。
// 用于集群模式下全局跟踪连接数。
func (r *WSMetricsRecorder) SetConnections(count float64) {
	wsConnectionsGauge.Set(count)
}
