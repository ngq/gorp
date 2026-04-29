package log

import "github.com/ngq/gorp/framework/contract"

// String 构造字符串字段。
func String(key, value string) contract.Field {
	return contract.Field{Key: key, Value: value}
}

// Int 构造 int 字段。
func Int(key string, value int) contract.Field {
	return contract.Field{Key: key, Value: value}
}

// Int64 构造 int64 字段。
func Int64(key string, value int64) contract.Field {
	return contract.Field{Key: key, Value: value}
}

// Bool 构造布尔字段。
func Bool(key string, value bool) contract.Field {
	return contract.Field{Key: key, Value: value}
}

// Any 构造任意类型字段。
func Any(key string, value any) contract.Field {
	return contract.Field{Key: key, Value: value}
}

// Err 构造错误字段。
func Err(err error) contract.Field {
	return contract.Field{Key: "err", Value: err}
}

// Debug 使用默认 logger 输出 debug 日志。
func Debug(msg string, fields ...contract.Field) {
	Default().Debug(msg, fields...)
}

// Info 使用默认 logger 输出 info 日志。
func Info(msg string, fields ...contract.Field) {
	Default().Info(msg, fields...)
}

// Warn 使用默认 logger 输出 warn 日志。
func Warn(msg string, fields ...contract.Field) {
	Default().Warn(msg, fields...)
}

// Error 使用默认 logger 输出 error 日志。
func Error(msg string, fields ...contract.Field) {
	Default().Error(msg, fields...)
}

// With 在默认 logger 基础上附加字段。
func With(fields ...contract.Field) contract.Logger {
	return Default().With(fields...)
}
