package log

import (
	"context"

	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	frameworklog "github.com/ngq/gorp/framework/log"
)

// SetDefault 设置进程级默认 logger。
func SetDefault(l observabilitycontract.Logger) {
	frameworklog.SetDefault(l)
}

// Default 返回进程级默认 logger。
func Default() observabilitycontract.Logger {
	return frameworklog.Default()
}

// Ctx 返回当前 context 关联的 logger。
func Ctx(ctx context.Context) observabilitycontract.Logger {
	return frameworklog.Ctx(ctx)
}

// WithContext 把请求级 logger 写入 context。
func WithContext(ctx context.Context, l observabilitycontract.Logger) context.Context {
	return frameworklog.WithContext(ctx, l)
}

// WithContextFields 基于 context 关联的 logger 追加字段。
func WithContextFields(ctx context.Context, fields ...observabilitycontract.Field) observabilitycontract.Logger {
	return frameworklog.WithContextFields(ctx, fields...)
}

// String 构造字符串字段。
func String(key, value string) observabilitycontract.Field {
	return frameworklog.String(key, value)
}

// Int 构造 int 字段。
func Int(key string, value int) observabilitycontract.Field {
	return frameworklog.Int(key, value)
}

// Int64 构造 int64 字段。
func Int64(key string, value int64) observabilitycontract.Field {
	return frameworklog.Int64(key, value)
}

// Bool 构造布尔字段。
func Bool(key string, value bool) observabilitycontract.Field {
	return frameworklog.Bool(key, value)
}

// Any 构造任意类型字段。
func Any(key string, value any) observabilitycontract.Field {
	return frameworklog.Any(key, value)
}

// Err 构造错误字段。
func Err(err error) observabilitycontract.Field {
	return frameworklog.Err(err)
}

// Debug 使用默认 logger 输出 debug 日志。
func Debug(msg string, fields ...observabilitycontract.Field) {
	frameworklog.Debug(msg, fields...)
}

// Info 使用默认 logger 输出 info 日志。
func Info(msg string, fields ...observabilitycontract.Field) {
	frameworklog.Info(msg, fields...)
}

// Warn 使用默认 logger 输出 warn 日志。
func Warn(msg string, fields ...observabilitycontract.Field) {
	frameworklog.Warn(msg, fields...)
}

// Error 使用默认 logger 输出 error 日志。
func Error(msg string, fields ...observabilitycontract.Field) {
	frameworklog.Error(msg, fields...)
}

// With 在默认 logger 基础上附加字段。
func With(fields ...observabilitycontract.Field) observabilitycontract.Logger {
	return frameworklog.With(fields...)
}
