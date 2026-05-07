// Application scenarios:
// - Define the framework-wide logging contract used across runtime, middleware, and providers.
// - Keep structured log fields portable across different logger implementations.
// - Expose one minimal logger surface for common application and framework logging flows.
//
// 适用场景：
// - 定义 runtime、中间件和 provider 共同使用的框架级日志契约。
// - 让结构化日志字段在不同 logger 实现之间保持可移植。
// - 为常见应用和框架日志流程提供最小 logger 接口面。
package observability

// LogKey is the container key for the logging capability.
//
// LogKey 是日志能力的容器键。
const LogKey = "framework.log"

// Logger defines the framework logging contract.
//
// Logger 定义框架日志契约。
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)

	With(fields ...Field) Logger
}

// Field describes one structured log field.
//
// Field 描述一个结构化日志字段。
type Field struct {
	Key   string
	Value any
}
