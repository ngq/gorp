// Application scenarios:
// - Expose the root-package HTTP response helpers built on the framework responder chain.
// - Give business handlers a short path for success and error response output.
// - Preserve unified responder semantics while keeping the fallback behavior easy to consume.
//
// 适用场景：
// - 暴露基于框架 responder 链的根包级 HTTP 响应 helper。
// - 为业务 handler 提供简短的成功/错误响应输出入口。
// - 在保持统一 responder 语义的同时，让默认回退行为更易使用。
package gorp

import (
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	httpmiddleware "github.com/ngq/gorp/framework/http/middleware"
)

// Success writes a successful HTTP response through the current responder chain.
//
// Success 通过当前 responder 链输出成功 HTTP 响应。
//
// Example:
//
//	func Ping(c gorp.HTTPContext) {
//	    gorp.Success(c, map[string]any{"pong": true})
//	}
func Success(c HTTPContext, data any) {
	responderFor(c).Success(c, data)
}

// SuccessWithMessage writes a successful HTTP response with a custom message.
//
// SuccessWithMessage 使用自定义 message 输出成功 HTTP 响应。
func SuccessWithMessage(c HTTPContext, message string, data any) {
	responderFor(c).SuccessWithMessage(c, message, data)
}

// SuccessWithStatus writes a successful HTTP response with a custom HTTP status.
//
// SuccessWithStatus 使用自定义 HTTP status 输出成功响应。
func SuccessWithStatus(c HTTPContext, status int, data any) {
	responderFor(c).SuccessWithStatus(c, status, data)
}

// Error writes an error HTTP response through the current responder chain.
//
// Error 通过当前 responder 链输出错误 HTTP 响应。
func Error(c HTTPContext, err error) {
	responderFor(c).Error(c, err)
}

// BadRequest writes a bad-request HTTP response through the current responder chain.
//
// BadRequest 通过当前 responder 链输出 bad request 响应。
func BadRequest(c HTTPContext, message string) {
	responderFor(c).BadRequest(c, message)
}

// InternalError writes an internal-error HTTP response through the current responder chain.
//
// InternalError 通过当前 responder 链输出内部错误响应。
func InternalError(c HTTPContext, message string) {
	responderFor(c).InternalError(c, message)
}

// responderFor resolves the responder from context and falls back to the default responder.
//
// responderFor 从上下文中解析 responder，缺省时回退到默认 responder。
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
	return httpmiddleware.NewDefaultResponder()
}
