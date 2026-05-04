package observability

import "time"

type Metrics interface {
	Counter(name string, labels map[string]string, delta float64)
	Gauge(name string, labels map[string]string, value float64)
	Histogram(name string, labels map[string]string, value float64)
	Timing(name string, labels map[string]string, duration time.Duration)
}
