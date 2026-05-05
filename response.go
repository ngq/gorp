package gorp

import (
	"net/http"

	"github.com/gin-gonic/gin"
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

// ========== Gin 原生版本辅助函数 ==========

// GinSuccess 返回成功响应（原生 gin.Context 版本）。
func GinSuccess(c *gin.Context, data any) {
	c.JSON(http.StatusOK, ginprovider.Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// GinSuccessWithMessage 返回成功响应带自定义消息（原生 gin.Context 版本）。
func GinSuccessWithMessage(c *gin.Context, message string, data any) {
	c.JSON(http.StatusOK, ginprovider.Response{
		Code:    0,
		Message: message,
		Data:    data,
	})
}

// GinSuccessWithStatus 返回成功响应带自定义状态码（原生 gin.Context 版本）。
func GinSuccessWithStatus(c *gin.Context, status int, data any) {
	c.JSON(status, ginprovider.Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// GinError 返回错误响应（原生 gin.Context 版本）。
func GinError(c *gin.Context, err error) {
	code, message := ginprovider.ParseError(err)
	httpStatus := ginprovider.CodeToHTTPStatus(code)
	c.JSON(httpStatus, ginprovider.Response{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}

// GinBadRequest 返回 400 错误响应（原生 gin.Context 版本）。
func GinBadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, ginprovider.Response{
		Code:    400,
		Message: message,
		Data:    nil,
	})
}

// GinInternalError 返回 500 错误响应（原生 gin.Context 版本）。
func GinInternalError(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, ginprovider.Response{
		Code:    500,
		Message: message,
		Data:    nil,
	})
}
