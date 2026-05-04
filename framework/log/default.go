package log

import (
	"sync"

	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
)

var (
	defaultLogger   observabilitycontract.Logger = noopLogger{}
	defaultLoggerMu sync.RWMutex
)

func SetDefault(l observabilitycontract.Logger) {
	defaultLoggerMu.Lock()
	defer defaultLoggerMu.Unlock()
	if l == nil {
		defaultLogger = noopLogger{}
		return
	}
	defaultLogger = l
}

func Default() observabilitycontract.Logger {
	defaultLoggerMu.RLock()
	defer defaultLoggerMu.RUnlock()
	if defaultLogger != nil {
		return defaultLogger
	}
	return noopLogger{}
}

type noopLogger struct{}

func (noopLogger) Debug(string, ...observabilitycontract.Field) {}
func (noopLogger) Info(string, ...observabilitycontract.Field)  {}
func (noopLogger) Warn(string, ...observabilitycontract.Field)  {}
func (noopLogger) Error(string, ...observabilitycontract.Field) {}
func (noopLogger) With(...observabilitycontract.Field) observabilitycontract.Logger {
	return noopLogger{}
}
