// Application scenarios:
// - Provide lightweight field builders and default-logger helpers for framework-wide logging.
// - Offer one centralized utility layer above the observability logger contract.
// - Keep common logging operations concise for framework and business code.
//
// 适用场景：
// - 为框架级日志提供轻量字段构造器和默认 logger helper。
// - 在 observability logger 契约之上提供统一工具层。
// - 让框架代码和业务代码都能用更简洁的方式进行常见日志操作。
package log

import observabilitycontract "github.com/ngq/gorp/framework/contract/observability"

// String creates a string log field.
//
// String 构造一个字符串日志字段。
func String(key, value string) observabilitycontract.Field {
	return observabilitycontract.Field{Key: key, Value: value}
}

// Int creates an int log field.
//
// Int 构造一个 int 日志字段。
func Int(key string, value int) observabilitycontract.Field {
	return observabilitycontract.Field{Key: key, Value: value}
}

// Int64 creates an int64 log field.
//
// Int64 构造一个 int64 日志字段。
func Int64(key string, value int64) observabilitycontract.Field {
	return observabilitycontract.Field{Key: key, Value: value}
}

// Bool creates a bool log field.
//
// Bool 构造一个 bool 日志字段。
func Bool(key string, value bool) observabilitycontract.Field {
	return observabilitycontract.Field{Key: key, Value: value}
}

// Any creates a generic log field.
//
// Any 构造一个任意类型日志字段。
func Any(key string, value any) observabilitycontract.Field {
	return observabilitycontract.Field{Key: key, Value: value}
}

// Err creates the standard error field.
//
// Err 构造标准错误字段。
func Err(err error) observabilitycontract.Field {
	return observabilitycontract.Field{Key: "err", Value: err}
}

// Debug writes a debug log through the process-wide default logger.
//
// Debug 通过进程级默认 logger 输出 debug 日志。
func Debug(msg string, fields ...observabilitycontract.Field) {
	Default().Debug(msg, fields...)
}

// Info writes an info log through the process-wide default logger.
//
// Info 通过进程级默认 logger 输出 info 日志。
func Info(msg string, fields ...observabilitycontract.Field) {
	Default().Info(msg, fields...)
}

// Warn writes a warn log through the process-wide default logger.
//
// Warn 通过进程级默认 logger 输出 warn 日志。
func Warn(msg string, fields ...observabilitycontract.Field) {
	Default().Warn(msg, fields...)
}

// Error writes an error log through the process-wide default logger.
//
// Error 通过进程级默认 logger 输出 error 日志。
func Error(msg string, fields ...observabilitycontract.Field) {
	Default().Error(msg, fields...)
}

// With derives a logger from the default logger and appends extra fields.
//
// With 基于默认 logger 派生一个追加了字段的新 logger。
func With(fields ...observabilitycontract.Field) observabilitycontract.Logger {
	return Default().With(fields...)
}
