package gin

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 是统一的 API 响应结构。
//
// 中文说明：
// - 所有 HTTP API 响应都应使用此结构；
// - code=0 表示成功，非 0 表示业务错误；
// - message 是人类可读的描述；
// - data 是实际返回数据（成功时）或错误详情（失败时）；
// - request_id 和 trace_id 通过响应头返回（X-Request-Id, X-Trace-Id）。
type Response struct {
	// Code 业务状态码，0=成功，非 0=业务错误
	Code int `json:"code"`

	// Message 人类可读的消息描述
	Message string `json:"message"`

	// Data 实际数据或错误详情
	Data any `json:"data,omitempty"`
}

// writeResponseHeaders 将链路追踪 ID 写入响应头。
func writeResponseHeaders(c *gin.Context) {
	requestID := getRequestID(c)
	traceID := getTraceID(c)

	if requestID != "" {
		c.Header("X-Request-Id", requestID)
	}
	if traceID != "" {
		c.Header("X-Trace-Id", traceID)
	}
}

// Success 返回成功响应。
//
// 中文说明：
// - 用于 Handler 返回成功数据；
// - request_id 和 trace_id 通过响应头返回。
//
// 使用示例：
//
//	func (h *Handler) GetUser(c *gin.Context) {
//	    user, err := h.service.GetUser(c, id)
//	    if err != nil {
//	        Error(c, err)
//	        return
//	    }
//	    Success(c, user)
//	}
func Success(c *gin.Context, data any) {
	writeResponseHeaders(c)
	resp := Response{
		Code:    0,
		Message: "success",
		Data:    data,
	}
	c.JSON(http.StatusOK, resp)
}

// SuccessWithMessage 返回成功响应并自定义消息。
func SuccessWithMessage(c *gin.Context, message string, data any) {
	writeResponseHeaders(c)
	resp := Response{
		Code:    0,
		Message: message,
		Data:    data,
	}
	c.JSON(http.StatusOK, resp)
}

// SuccessWithStatus 返回成功响应并自定义 HTTP 状态码。
func SuccessWithStatus(c *gin.Context, status int, data any) {
	writeResponseHeaders(c)
	resp := Response{
		Code:    0,
		Message: "success",
		Data:    data,
	}
	c.JSON(status, resp)
}

// Error 返回业务错误响应。
//
// 中文说明：
// - 用于 Handler 返回业务错误；
// - 根据 Error 的 Code 自动选择 HTTP 状态码；
// - request_id 和 trace_id 通过响应头返回。
//
// 使用示例：
//
//	if user == nil {
//	    Error(c, errors.NotFound("用户不存在"))
//	    return
//	}
func Error(c *gin.Context, err error) {
	writeResponseHeaders(c)
	code, message := parseError(err)
	httpStatus := codeToHTTPStatus(code)

	resp := Response{
		Code:    code,
		Message: message,
		Data:    nil,
	}
	c.JSON(httpStatus, resp)
}

// ErrorWithData 返回业务错误响应并附带数据。
func ErrorWithData(c *gin.Context, err error, data any) {
	writeResponseHeaders(c)
	code, message := parseError(err)
	httpStatus := codeToHTTPStatus(code)

	resp := Response{
		Code:    code,
		Message: message,
		Data:    data,
	}
	c.JSON(httpStatus, resp)
}

// ErrorWithStatus 返回业务错误响应并自定义 HTTP 状态码。
func ErrorWithStatus(c *gin.Context, status int, err error) {
	writeResponseHeaders(c)
	code, message := parseError(err)

	resp := Response{
		Code:    code,
		Message: message,
		Data:    nil,
	}
	c.JSON(status, resp)
}

// BadRequest 返回 400 参数错误响应。
func BadRequest(c *gin.Context, message string) {
	writeResponseHeaders(c)
	resp := Response{
		Code:    CodeBadRequest,
		Message: message,
		Data:    nil,
	}
	c.JSON(http.StatusBadRequest, resp)
}

// Unauthorized 返回 401 未认证响应。
func Unauthorized(c *gin.Context, message string) {
	writeResponseHeaders(c)
	resp := Response{
		Code:    CodeUnauthorized,
		Message: message,
		Data:    nil,
	}
	c.JSON(http.StatusUnauthorized, resp)
}

// Forbidden 返回 403 禁止访问响应。
func Forbidden(c *gin.Context, message string) {
	writeResponseHeaders(c)
	resp := Response{
		Code:    CodeForbidden,
		Message: message,
		Data:    nil,
	}
	c.JSON(http.StatusForbidden, resp)
}

// NotFound 返回 404 未找到响应。
func NotFound(c *gin.Context, message string) {
	writeResponseHeaders(c)
	resp := Response{
		Code:    CodeNotFound,
		Message: message,
		Data:    nil,
	}
	c.JSON(http.StatusNotFound, resp)
}

// InternalError 返回 500 内部错误响应。
func InternalError(c *gin.Context, message string) {
	writeResponseHeaders(c)
	resp := Response{
		Code:    CodeInternalError,
		Message: message,
		Data:    nil,
	}
	c.JSON(http.StatusInternalServerError, resp)
}

// PaginatedData 是分页数据结构。
type PaginatedData struct {
	// Items 数据列表
	Items any `json:"items"`

	// Total 总数量
	Total int64 `json:"total"`

	// Page 当前页码
	Page int `json:"page"`

	// PageSize 每页数量
	PageSize int `json:"page_size"`
}

// SuccessPaginated 返回分页成功响应。
func SuccessPaginated(c *gin.Context, items any, total int64, page, pageSize int) {
	data := PaginatedData{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}
	Success(c, data)
}

// --- 业务错误码定义 ---
//
// 中文说明：
// - 错误码范围划分：
//   - 0: 成功
//   - 1-9999: 通用错误
//   - 10000-19999: 用户模块错误
//   - 20000-29999: 订单模块错误
//   - 30000-39999: 商品模块错误
//   - 40000-49999: 支付模块错误
//   - 其他模块按需扩展

const (
	// 成功
	CodeSuccess = 0

	// 通用错误 (1-9999)
	CodeBadRequest         = 1001 // 参数错误
	CodeUnauthorized       = 1002 // 未认证
	CodeForbidden          = 1003 // 禁止访问
	CodeNotFound           = 1004 // 未找到
	CodeInternalError      = 1005 // 内部错误
	CodeServiceUnavailable = 1006 // 服务不可用
	CodeTooManyRequests    = 1007 // 请求过多
	CodeConflict           = 1008 // 资源冲突
	CodeValidationFailed   = 1009 // 验证失败

	// 用户模块错误 (10000-19999)
	CodeUserNotFound      = 10001 // 用户不存在
	CodeUserAlreadyExists = 10002 // 用户已存在
	CodeUserPasswordWrong = 10003 // 密码错误
	CodeUserDisabled      = 10004 // 用户已禁用
	CodeUserTokenExpired  = 10005 // Token 已过期
	CodeUserTokenInvalid  = 10006 // Token 无效

	// 订单模块错误 (20000-29999)
	CodeOrderNotFound      = 20001 // 订单不存在
	CodeOrderStatusInvalid = 20002 // 订单状态无效
	CodeOrderAlreadyPaid   = 20003 // 订单已支付
	CodeOrderCanceled      = 20004 // 订单已取消

	// 商品模块错误 (30000-39999)
	CodeProductNotFound   = 30001 // 商品不存在
	CodeProductOutOfStock = 30002 // 商品库存不足
	CodeProductDisabled   = 30003 // 商品已下架

	// 支付模块错误 (40000-49999)
	CodePaymentFailed   = 40001 // 支付失败
	CodePaymentTimeout  = 40002 // 支付超时
	CodePaymentCanceled = 40003 // 支付已取消
)

// --- 内部辅助函数 ---

func getRequestID(c *gin.Context) string {
	return c.GetString("request_id")
}

func getTraceID(c *gin.Context) string {
	return c.GetString("trace_id")
}

// parseError 解析错误获取错误码和消息
func parseError(err error) (int, string) {
	if err == nil {
		return CodeSuccess, "success"
	}

	// 尝试解析自定义错误类型
	if bizErr, ok := err.(BusinessError); ok {
		return bizErr.Code(), bizErr.Message()
	}

	// 默认返回内部错误
	return CodeInternalError, err.Error()
}

// codeToHTTPStatus 将业务错误码映射到 HTTP 状态码
func codeToHTTPStatus(code int) int {
	switch code {
	case CodeSuccess:
		return http.StatusOK
	case CodeBadRequest, CodeValidationFailed:
		return http.StatusBadRequest
	case CodeUnauthorized, CodeUserTokenExpired, CodeUserTokenInvalid:
		return http.StatusUnauthorized
	case CodeForbidden, CodeUserDisabled:
		return http.StatusForbidden
	case CodeNotFound, CodeUserNotFound, CodeOrderNotFound, CodeProductNotFound:
		return http.StatusNotFound
	case CodeConflict, CodeUserAlreadyExists, CodeOrderAlreadyPaid:
		return http.StatusConflict
	case CodeTooManyRequests:
		return http.StatusTooManyRequests
	case CodeServiceUnavailable:
		return http.StatusServiceUnavailable
	default:
		// 根据错误码范围判断
		if code >= 1001 && code <= 9999 {
			return http.StatusBadRequest
		}
		return http.StatusInternalServerError
	}
}

// BusinessError 是业务错误接口。
//
// 中文说明：
// - 业务层应实现此接口返回自定义错误码和消息；
// - 响应中间件会自动解析并转换为统一格式。
type BusinessError interface {
	error
	Code() int
	Message() string
}

// BizError 是简单的业务错误实现。
type BizError struct {
	code    int
	message string
}

// NewBizError 创建业务错误。
func NewBizError(code int, message string) *BizError {
	return &BizError{code: code, message: message}
}

func (e *BizError) Error() string   { return e.message }
func (e *BizError) Code() int       { return e.code }
func (e *BizError) Message() string { return e.message }

// --- 常用错误创建函数 ---
//
// 中文说明：
// - 提供快捷创建常见错误的方法；
// - Handler 中可直接使用这些函数。

// ErrBadRequest 创建参数错误
func ErrBadRequest(message string) *BizError {
	return NewBizError(CodeBadRequest, message)
}

// ErrUnauthorized 创建未认证错误
func ErrUnauthorized(message string) *BizError {
	return NewBizError(CodeUnauthorized, message)
}

// ErrForbidden 创建禁止访问错误
func ErrForbidden(message string) *BizError {
	return NewBizError(CodeForbidden, message)
}

// ErrNotFound 创建未找到错误
func ErrNotFound(message string) *BizError {
	return NewBizError(CodeNotFound, message)
}

// ErrInternal 创建内部错误
func ErrInternal(message string) *BizError {
	return NewBizError(CodeInternalError, message)
}

// ErrUserNotFound 创建用户不存在错误
func ErrUserNotFound() *BizError {
	return NewBizError(CodeUserNotFound, "用户不存在")
}

// ErrOrderNotFound 创建订单不存在错误
func ErrOrderNotFound() *BizError {
	return NewBizError(CodeOrderNotFound, "订单不存在")
}

// ErrProductNotFound 创建商品不存在错误
func ErrProductNotFound() *BizError {
	return NewBizError(CodeProductNotFound, "商品不存在")
}

// ErrOutOfStock 创建库存不足错误
func ErrOutOfStock() *BizError {
	return NewBizError(CodeProductOutOfStock, "商品库存不足")
}
