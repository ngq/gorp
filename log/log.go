package log

import (
	"context"

	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	frameworklog "github.com/ngq/gorp/framework/log"
)

// SetDefault sets the process-wide default logger.
// SetDefault 设置进程级默认 logger。
func SetDefault(l observabilitycontract.Logger) {
	frameworklog.SetDefault(l)
}

// Default returns the process-wide default logger.
// Default 返回进程级默认 logger。
func Default() observabilitycontract.Logger {
	return frameworklog.Default()
}

// Ctx returns the logger associated with the context.
// Ctx 返回当前 context 关联的 logger。
func Ctx(ctx context.Context) observabilitycontract.Logger {
	return frameworklog.Ctx(ctx)
}

// WithContext stores a request-scoped logger into the context.
// WithContext 把请求级 logger 写入 context。
func WithContext(ctx context.Context, l observabilitycontract.Logger) context.Context {
	return frameworklog.WithContext(ctx, l)
}

// WithContextFields appends fields to the logger associated with the context.
// WithContextFields 基于 context 关联的 logger 追加字段。
func WithContextFields(ctx context.Context, fields ...observabilitycontract.Field) observabilitycontract.Logger {
	return frameworklog.WithContextFields(ctx, fields...)
}

// String creates a string log field.
// String 构造字符串字段。
func String(key, value string) observabilitycontract.Field {
	return frameworklog.String(key, value)
}

// Int creates an int log field.
// Int 构造 int 字段。
func Int(key string, value int) observabilitycontract.Field {
	return frameworklog.Int(key, value)
}

// Int64 creates an int64 log field.
// Int64 构造 int64 字段。
func Int64(key string, value int64) observabilitycontract.Field {
	return frameworklog.Int64(key, value)
}

// Bool creates a bool log field.
// Bool 构造布尔字段。
func Bool(key string, value bool) observabilitycontract.Field {
	return frameworklog.Bool(key, value)
}

// Any creates a generic log field.
// Any 构造任意类型字段。
func Any(key string, value any) observabilitycontract.Field {
	return frameworklog.Any(key, value)
}

// Err creates an error log field.
// Err 构造错误字段。
func Err(err error) observabilitycontract.Field {
	return frameworklog.Err(err)
}

// Debug writes a debug log with the default logger.
// Debug 使用默认 logger 输出 debug 日志。
func Debug(msg string, fields ...observabilitycontract.Field) {
	frameworklog.Debug(msg, fields...)
}

// Info writes an info log with the default logger.
// Info 使用默认 logger 输出 info 日志。
//
// Example:
//
//	log.Info("user created", log.Int64("user_id", 42), log.String("source", "signup"))
func Info(msg string, fields ...observabilitycontract.Field) {
	frameworklog.Info(msg, fields...)
}

// Warn writes a warn log with the default logger.
// Warn 使用默认 logger 输出 warn 日志。
func Warn(msg string, fields ...observabilitycontract.Field) {
	frameworklog.Warn(msg, fields...)
}

// Error writes an error log with the default logger.
// Error 使用默认 logger 输出 error 日志。
func Error(msg string, fields ...observabilitycontract.Field) {
	frameworklog.Error(msg, fields...)
}

// With returns a logger derived from the default logger with extra fields.
// With 在默认 logger 基础上附加字段。
func With(fields ...observabilitycontract.Field) observabilitycontract.Logger {
	return frameworklog.With(fields...)
}
