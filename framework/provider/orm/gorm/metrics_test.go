package gorm

import "testing"

func TestDBMetricsCollectorStartStopIsIdempotent(t *testing.T) {
	collector := NewDBMetricsCollector(nil, "test")
	stopA := collector.StartCollection()
	stopB := collector.StartCollection()
	stopA()
	stopB()
	collector.Stop()
}
