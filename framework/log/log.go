package log

import observabilitycontract "github.com/ngq/gorp/framework/contract/observability"

func String(key, value string) observabilitycontract.Field {
	return observabilitycontract.Field{Key: key, Value: value}
}

func Int(key string, value int) observabilitycontract.Field {
	return observabilitycontract.Field{Key: key, Value: value}
}

func Int64(key string, value int64) observabilitycontract.Field {
	return observabilitycontract.Field{Key: key, Value: value}
}

func Bool(key string, value bool) observabilitycontract.Field {
	return observabilitycontract.Field{Key: key, Value: value}
}

func Any(key string, value any) observabilitycontract.Field {
	return observabilitycontract.Field{Key: key, Value: value}
}

func Err(err error) observabilitycontract.Field {
	return observabilitycontract.Field{Key: "err", Value: err}
}

func Debug(msg string, fields ...observabilitycontract.Field) {
	Default().Debug(msg, fields...)
}

func Info(msg string, fields ...observabilitycontract.Field) {
	Default().Info(msg, fields...)
}

func Warn(msg string, fields ...observabilitycontract.Field) {
	Default().Warn(msg, fields...)
}

func Error(msg string, fields ...observabilitycontract.Field) {
	Default().Error(msg, fields...)
}

func With(fields ...observabilitycontract.Field) observabilitycontract.Logger {
	return Default().With(fields...)
}
