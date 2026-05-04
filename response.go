package gorp

import (
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	ginprovider "github.com/ngq/gorp/framework/provider/gin"
)

func Success(c HTTPContext, data any) {
	responderFor(c).Success(c, data)
}

func SuccessWithMessage(c HTTPContext, message string, data any) {
	responderFor(c).SuccessWithMessage(c, message, data)
}

func SuccessWithStatus(c HTTPContext, status int, data any) {
	responderFor(c).SuccessWithStatus(c, status, data)
}

func Error(c HTTPContext, err error) {
	responderFor(c).Error(c, err)
}

func BadRequest(c HTTPContext, message string) {
	responderFor(c).BadRequest(c, message)
}

func InternalError(c HTTPContext, message string) {
	responderFor(c).InternalError(c, message)
}

func responderFor(c HTTPContext) transportcontract.HTTPResponder {
	if c != nil && c.Context() != nil {
		if container, ok := FromContainerContext(c.Context()); ok && container != nil && container.IsBind(transportcontract.HTTPResponderKey) {
			if responderAny, err := container.Make(transportcontract.HTTPResponderKey); err == nil {
				if responder, ok := responderAny.(transportcontract.HTTPResponder); ok && responder != nil {
					return responder
				}
			}
		}
	}
	return ginprovider.NewDefaultResponder()
}
