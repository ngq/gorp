package gin

import (
	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	frameworkbizlog "github.com/ngq/gorp/framework/log"
)

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
