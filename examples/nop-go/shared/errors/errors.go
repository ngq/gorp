// Package errors 提供统一的错误码定义
//
// 中文说明：
// - 定义所有服务的业务错误码；
// - 与 gorp 框架的 AppError 对齐；
// - 支持国际化错误消息。
package errors

import (
	"fmt"

	"github.com/ngq/gorp/framework/contract"
)

// 错误码范围分配：
// - 1000-1999: 通用错误
// - 2000-2999: 客户服务错误
// - 3000-3999: 商品服务错误
// - 4000-4999: 订单服务错误
// - 5000-5999: 支付服务错误
// - 6000-6999: 库存服务错误
// - 7000-7999: 物流服务错误

// 通用错误码
var (
	ErrBadRequest   = NewBizError(1001, "请求参数错误")
	ErrUnauthorized = NewBizError(1002, "未授权访问")
	ErrForbidden    = NewBizError(1003, "禁止访问")
	ErrNotFound     = NewBizError(1004, "资源不存在")
	ErrConflict     = NewBizError(1005, "资源冲突")
	ErrInternal     = NewBizError(1006, "服务器内部错误")
	ErrServiceUnavailable = NewBizError(1007, "服务暂时不可用")
)

// 客户服务错误码
var (
	ErrCustomerNotFound      = NewBizError(2001, "客户不存在")
	ErrCustomerAlreadyExists = NewBizError(2002, "客户已存在")
	ErrInvalidCredentials    = NewBizError(2003, "用户名或密码错误")
	ErrEmailNotVerified      = NewBizError(2004, "邮箱未验证")
	ErrPhoneNotVerified      = NewBizError(2005, "手机未验证")
	ErrCustomerDisabled      = NewBizError(2006, "账户已被禁用")
	ErrInvalidPassword       = NewBizError(2007, "密码格式不正确")
	ErrPasswordMismatch      = NewBizError(2008, "密码不匹配")
	ErrAddressNotFound       = NewBizError(2009, "地址不存在")
)

// 商品服务错误码
var (
	ErrProductNotFound       = NewBizError(3001, "商品不存在")
	ErrProductNotPublished   = NewBizError(3002, "商品未上架")
	ErrCategoryNotFound      = NewBizError(3003, "分类不存在")
	ErrManufacturerNotFound  = NewBizError(3004, "品牌不存在")
	ErrProductReviewNotFound = NewBizError(3005, "商品评论不存在")
	ErrDuplicateSku          = NewBizError(3006, "SKU已存在")
)

// 订单服务错误码
var (
	ErrOrderNotFound         = NewBizError(4001, "订单不存在")
	ErrOrderAlreadyPaid      = NewBizError(4002, "订单已支付")
	ErrOrderAlreadyCancelled = NewBizError(4003, "订单已取消")
	ErrOrderCannotCancel     = NewBizError(4004, "订单无法取消")
	ErrOrderCannotModify     = NewBizError(4005, "订单无法修改")
	ErrInvalidOrderStatus    = NewBizError(4006, "订单状态无效")
	ErrGiftCardNotFound      = NewBizError(4007, "礼品卡不存在")
	ErrGiftCardExpired       = NewBizError(4008, "礼品卡已过期")
	ErrGiftCardUsed          = NewBizError(4009, "礼品卡已使用")
	ErrReturnRequestNotFound = NewBizError(4010, "退货请求不存在")
)

// 支付服务错误码
var (
	ErrPaymentNotFound       = NewBizError(5001, "支付记录不存在")
	ErrPaymentFailed         = NewBizError(5002, "支付失败")
	ErrPaymentTimeout        = NewBizError(5003, "支付超时")
	ErrPaymentAlreadyRefunded = NewBizError(5004, "支付已退款")
	ErrRefundFailed          = NewBizError(5005, "退款失败")
	ErrInvalidPaymentMethod  = NewBizError(5006, "无效的支付方式")
)

// 库存服务错误码
var (
	ErrInventoryNotFound     = NewBizError(6001, "库存记录不存在")
	ErrInsufficientStock     = NewBizError(6002, "库存不足")
	ErrStockReservationFailed = NewBizError(6003, "库存预留失败")
	ErrWarehouseNotFound     = NewBizError(6004, "仓库不存在")
	ErrInventoryLocked       = NewBizError(6005, "库存被锁定")
)

// 物流服务错误码
var (
	ErrShipmentNotFound     = NewBizError(7001, "发货单不存在")
	ErrShippingMethodNotFound = NewBizError(7002, "配送方式不存在")
	ErrInvalidTrackingNumber = NewBizError(7003, "无效的物流单号")
	ErrShipmentAlreadyDelivered = NewBizError(7004, "发货单已送达")
)

// BizError 业务错误
type BizError struct {
	Code    int
	Message string
}

// NewBizError 创建业务错误
func NewBizError(code int, message string) *BizError {
	return &BizError{Code: code, Message: message}
}

// Error 实现 error 接口
func (e *BizError) Error() string {
	return e.Message
}

// ToAppError 转换为 gorp AppError
// 中文说明: 将自定义 BizError 转换为框架标准的 AppError
func (e *BizError) ToAppError() contract.AppError {
	// 根据错误码映射到对应的 HTTP 状态码和 Reason
	var code int
	var reason contract.ErrorReason

	// 错误码范围映射:
	// - 1000-1999: 通用错误,默认 400
	// - 2000-2999: 客户服务错误,默认 400
	// - 3000-3999: 商品服务错误,默认 404
	// - 4000-3999: 订单服务错误,默认 400
	// - 5000-5999: 支付服务错误,默认 400
	// - 6000-6999: 库存服务错误,默认 400
	// - 7000-7999: 物流服务错误,默认 404
	switch {
	case e.Code >= 1000 && e.Code < 2000:
		code = contract.ErrorCodeBadRequest
		reason = contract.ErrorReasonBadRequest
	case e.Code >= 2000 && e.Code < 3000:
		// 客户相关错误,多数为 400/401/404
		switch e.Code {
		case 2001, 2009: // Not Found
			code = contract.ErrorCodeNotFound
			reason = contract.ErrorReasonNotFound
		case 2003, 2006: // Unauthorized/Forbidden
			code = contract.ErrorCodeUnauthorized
			reason = contract.ErrorReasonUnauthorized
		default:
			code = contract.ErrorCodeBadRequest
			reason = contract.ErrorReasonBadRequest
		}
	case e.Code >= 3000 && e.Code < 4000:
		code = contract.ErrorCodeNotFound
		reason = contract.ErrorReasonNotFound
	case e.Code >= 4000 && e.Code < 5000:
		code = contract.ErrorCodeBadRequest
		reason = contract.ErrorReasonBadRequest
	case e.Code >= 5000 && e.Code < 6000:
		code = contract.ErrorCodeBadRequest
		reason = contract.ErrorReasonBadRequest
	case e.Code >= 6000 && e.Code < 7000:
		// 库存不足等错误属于业务冲突
		code = contract.ErrorCodeConflict
		reason = contract.ErrorReasonConflict
	case e.Code >= 7000 && e.Code < 8000:
		code = contract.ErrorCodeNotFound
		reason = contract.ErrorReasonNotFound
	default:
		code = contract.ErrorCodeInternalServerError
		reason = contract.ErrorReasonInternal
	}

	return contract.NewError(code, reason, e.Message)
}

// Is 判断错误是否相等
func (e *BizError) Is(target error) bool {
	t, ok := target.(*BizError)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

// WrapError 包装错误
func WrapError(err error, code int, message string) *BizError {
	return &BizError{
		Code:    code,
		Message: fmt.Sprintf("%s: %v", message, err),
	}
}