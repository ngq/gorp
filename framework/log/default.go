// Application scenarios:
// - Manage the process-wide default logger used across the framework.
// - Keep default logger reads and writes concurrency-safe.
// - Provide a noop fallback so logging calls remain safe before explicit initialization.
//
// 适用场景：
// - 管理框架全局使用的进程级默认 logger。
// - 保证默认 logger 的读写具备并发安全。
// - 提供 noop 回退，确保在显式初始化前日志调用也始终安全。
package log

import (
	"sync"

	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
)

var (
	defaultLogger   observabilitycontract.Logger = noopLogger{}
	defaultLoggerMu sync.RWMutex
)

// SetDefault replaces the process-wide default logger.
//
// SetDefault 替换进程级默认 logger。
func SetDefault(l observabilitycontract.Logger) {
	defaultLoggerMu.Lock()
	defer defaultLoggerMu.Unlock()
	if l == nil {
		defaultLogger = noopLogger{}
		return
	}
	defaultLogger = l
}

// Default returns the process-wide default logger.
//
// Default 返回进程级默认 logger。
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
