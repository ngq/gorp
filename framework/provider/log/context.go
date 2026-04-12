package log

import (
	"github.com/ngq/gorp/framework/contract"
)

// WithTraceID 返回带 trace_id 字段的 logger。
//
// 中文说明：
// - 用于业务代码中快速获取带 trace id 的 logger；
// - 如果 traceID 为空，则返回原始 logger；
// - 示例用法：
//
//	logger := log.WithTraceID(container, traceID)
//	logger.Info("处理请求", contract.Field{Key: "user_id", Value: 123})
func WithTraceID(container contract.Container, traceID string) contract.Logger {
	logAny, err := container.Make(contract.LogKey)
	if err != nil {
		return nil
	}
	logger := logAny.(contract.Logger)

	if traceID == "" {
		return logger
	}

	return logger.With(contract.Field{Key: "trace_id", Value: traceID})
}

// WithRequestID 返回带 request_id 字段的 logger。
//
// 中文说明：
// - 用于业务代码中快速获取带 request id 的 logger；
// - 如果 requestID 为空，则返回原始 logger。
func WithRequestID(container contract.Container, requestID string) contract.Logger {
	logAny, err := container.Make(contract.LogKey)
	if err != nil {
		return nil
	}
	logger := logAny.(contract.Logger)

	if requestID == "" {
		return logger
	}

	return logger.With(contract.Field{Key: "request_id", Value: requestID})
}

// WithTraceAndRequestID 返回带 trace_id 和 request_id 字段的 logger。
//
// 中文说明：
// - 用于业务代码中快速获取带 trace id 和 request id 的 logger；
// - 这是推荐的用法，可以同时关联单次请求和分布式链路。
func WithTraceAndRequestID(container contract.Container, traceID, requestID string) contract.Logger {
	logAny, err := container.Make(contract.LogKey)
	if err != nil {
		return nil
	}
	logger := logAny.(contract.Logger)

	var fields []contract.Field
	if traceID != "" {
		fields = append(fields, contract.Field{Key: "trace_id", Value: traceID})
	}
	if requestID != "" {
		fields = append(fields, contract.Field{Key: "request_id", Value: requestID})
	}

	if len(fields) == 0 {
		return logger
	}
	return logger.With(fields...)
}