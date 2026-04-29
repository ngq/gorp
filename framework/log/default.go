package log

import (
	"sync"

	"github.com/ngq/gorp/framework/contract"
)

var (
	defaultLogger   contract.Logger = noopLogger{}
	defaultLoggerMu sync.RWMutex
)

// SetDefault 设置进程级默认 logger。
//
// 中文说明：
// - 只应由 bootstrap / runtime 初始化阶段调用；
// - 传入 nil 时会回退到安全的 noop logger，避免 Default() 返回 nil。
func SetDefault(l contract.Logger) {
	defaultLoggerMu.Lock()
	defer defaultLoggerMu.Unlock()
	if l == nil {
		defaultLogger = noopLogger{}
		return
	}
	defaultLogger = l
}

// Default 返回进程级默认 logger。
//
// 中文说明：
// - 业务层无上下文时走这条主线；
// - 未显式设置时自动回退到 noop logger。
func Default() contract.Logger {
	defaultLoggerMu.RLock()
	defer defaultLoggerMu.RUnlock()
	if defaultLogger != nil {
		return defaultLogger
	}
	return noopLogger{}
}

type noopLogger struct{}

func (noopLogger) Debug(string, ...contract.Field) {}
func (noopLogger) Info(string, ...contract.Field)  {}
func (noopLogger) Warn(string, ...contract.Field)  {}
func (noopLogger) Error(string, ...contract.Field) {}
func (noopLogger) With(...contract.Field) contract.Logger {
	return noopLogger{}
}
