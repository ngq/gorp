package observability

import "time"

// MetricsKey is the container key for the metrics capability.
//
// MetricsKey 是 metrics 能力的容器键。
const MetricsKey = "framework.metrics"

type Metrics interface {
	Counter(name string, labels map[string]string, delta float64)
	Gauge(name string, labels map[string]string, value float64)
	Histogram(name string, labels map[string]string, value float64)
	Timing(name string, labels map[string]string, duration time.Duration)
}
