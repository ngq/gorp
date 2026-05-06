// Application scenarios:
// - Prevent a single panic from crashing the whole HTTP process.
// - Convert panic output into the framework's unified error response.
// - Keep panic records observable in request logs.
//
// 适用场景：
// - 防止单次 panic 直接打垮整个 HTTP 进程。
// - 把 panic 转换为框架统一错误响应。
// - 让 panic 记录仍然能在请求日志中被观测到。
package middleware

import (
	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	frameworkbizlog "github.com/ngq/gorp/framework/log"
)

// RecoveryMiddleware recovers panics and returns a unified internal-error response.
//
// RecoveryMiddleware 捕获 panic，并返回统一的内部错误响应。
func RecoveryMiddleware() transportcontract.HTTPMiddleware {
	return func(next transportcontract.HTTPHandler) transportcontract.HTTPHandler {
		return func(c transportcontract.HTTPContext) {
			defer func() {
				if rec := recover(); rec != nil {
					frameworkbizlog.Ctx(c.Context()).Error("http panic recovered",
						observabilitycontract.Field{Key: "panic", Value: rec},
					)
					responderFor(c).InternalError(c, "internal server error")
				}
			}()
			if next != nil {
				next(c)
			}
		}
	}
}
