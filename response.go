// Package gorp provides the root-package application startup surface for gorp framework.
// This file exposes HTTP response helpers built on framework responder chain.
// Gives business handlers short path for success and error response output.
//
// Gorp 包提供 gorp 框架的根包层应用启动入口。
// 本文件暴露基于框架 responder 链的 HTTP 响应 helper。
// 为业务 handler 提供简短的成功/错误响应输出入口。
//
// Eg:
//
//	func Ping(c gorp.Context) {
//	    gorp.Success(c, map[string]any{"pong": true})
//	}
package gorp

import (
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	httpmiddleware "github.com/ngq/gorp/framework/http/middleware"
)

// Success writes a successful HTTP response through the current responder chain.
//
// Success 通过当前 responder 链输出成功 HTTP 响应。
func Success(c Context, data any) {
	responderFor(c).Success(c, data)
}

// SuccessWithMessage writes a successful HTTP response with a custom message.
//
// SuccessWithMessage 使用自定义 message 输出成功 HTTP 响应。
func SuccessWithMessage(c Context, message string, data any) {
	responderFor(c).SuccessWithMessage(c, message, data)
}

// SuccessWithStatus writes a successful HTTP response with a custom HTTP status.
//
// SuccessWithStatus 使用自定义 HTTP status 输出成功响应。
func SuccessWithStatus(c Context, status int, data any) {
	responderFor(c).SuccessWithStatus(c, status, data)
}

// Error writes an error HTTP response through the current responder chain.
//
// Error 通过当前 responder 链输出错误 HTTP 响应。
func Error(c Context, err error) {
	responderFor(c).Error(c, err)
}

// BadRequest writes a bad-request HTTP response through the current responder chain.
//
// BadRequest 通过当前 responder 链输出 bad request 响应。
func BadRequest(c Context, message string) {
	responderFor(c).BadRequest(c, message)
}

// InternalError writes an internal-error HTTP response through the current responder chain.
//
// InternalError 通过当前 responder 链输出内部错误响应。
func InternalError(c Context, message string) {
	responderFor(c).InternalError(c, message)
}

// responderFor resolves the responder from context and falls back to the default responder.
//
// responderFor 从上下文中解析 responder，缺省时回退到默认 responder。
func responderFor(c Context) transportcontract.HTTPResponder {
	if c != nil {
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