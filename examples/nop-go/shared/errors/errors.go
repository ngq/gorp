// Package errors 鎻愪緵缁熶竴鐨勯敊璇爜瀹氫箟
//
// 涓枃璇存槑锛?
// - 瀹氫箟鎵€鏈夋湇鍔＄殑涓氬姟閿欒鐮侊紱
// - 涓?gorp 妗嗘灦鐨?AppError 瀵归綈锛?
// - 鏀寔鍥介檯鍖栭敊璇秷鎭€?
package errors

import (
	"fmt"

	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
)

// 閿欒鐮佽寖鍥村垎閰嶏細
// - 1000-1999: 閫氱敤閿欒
// - 2000-2999: 瀹㈡埛鏈嶅姟閿欒
// - 3000-3999: 鍟嗗搧鏈嶅姟閿欒
// - 4000-4999: 璁㈠崟鏈嶅姟閿欒
// - 5000-5999: 鏀粯鏈嶅姟閿欒
// - 6000-6999: 搴撳瓨鏈嶅姟閿欒
// - 7000-7999: 鐗╂祦鏈嶅姟閿欒

// 閫氱敤閿欒鐮?
var (
	ErrBadRequest   = NewBizError(1001, "璇锋眰鍙傛暟閿欒")
	ErrUnauthorized = NewBizError(1002, "鏈巿鏉冭闂?)
	ErrForbidden    = NewBizError(1003, "绂佹璁块棶")
	ErrNotFound     = NewBizError(1004, "璧勬簮涓嶅瓨鍦?)
	ErrConflict     = NewBizError(1005, "璧勬簮鍐茬獊")
	ErrInternal     = NewBizError(1006, "鏈嶅姟鍣ㄥ唴閮ㄩ敊璇?)
	ErrServiceUnavailable = NewBizError(1007, "鏈嶅姟鏆傛椂涓嶅彲鐢?)
)

// 瀹㈡埛鏈嶅姟閿欒鐮?
var (
	ErrCustomerNotFound      = NewBizError(2001, "瀹㈡埛涓嶅瓨鍦?)
	ErrCustomerAlreadyExists = NewBizError(2002, "瀹㈡埛宸插瓨鍦?)
	ErrInvalidCredentials    = NewBizError(2003, "鐢ㄦ埛鍚嶆垨瀵嗙爜閿欒")
	ErrEmailNotVerified      = NewBizError(2004, "閭鏈獙璇?)
	ErrPhoneNotVerified      = NewBizError(2005, "鎵嬫満鏈獙璇?)
	ErrCustomerDisabled      = NewBizError(2006, "璐︽埛宸茶绂佺敤")
	ErrInvalidPassword       = NewBizError(2007, "瀵嗙爜鏍煎紡涓嶆纭?)
	ErrPasswordMismatch      = NewBizError(2008, "瀵嗙爜涓嶅尮閰?)
	ErrAddressNotFound       = NewBizError(2009, "鍦板潃涓嶅瓨鍦?)
)

// 鍟嗗搧鏈嶅姟閿欒鐮?
var (
	ErrProductNotFound       = NewBizError(3001, "鍟嗗搧涓嶅瓨鍦?)
	ErrProductNotPublished   = NewBizError(3002, "鍟嗗搧鏈笂鏋?)
	ErrCategoryNotFound      = NewBizError(3003, "鍒嗙被涓嶅瓨鍦?)
	ErrManufacturerNotFound  = NewBizError(3004, "鍝佺墝涓嶅瓨鍦?)
	ErrProductReviewNotFound = NewBizError(3005, "鍟嗗搧璇勮涓嶅瓨鍦?)
	ErrDuplicateSku          = NewBizError(3006, "SKU宸插瓨鍦?)
)

// 璁㈠崟鏈嶅姟閿欒鐮?
var (
	ErrOrderNotFound         = NewBizError(4001, "璁㈠崟涓嶅瓨鍦?)
	ErrOrderAlreadyPaid      = NewBizError(4002, "璁㈠崟宸叉敮浠?)
	ErrOrderAlreadyCancelled = NewBizError(4003, "璁㈠崟宸插彇娑?)
	ErrOrderCannotCancel     = NewBizError(4004, "璁㈠崟鏃犳硶鍙栨秷")
	ErrOrderCannotModify     = NewBizError(4005, "璁㈠崟鏃犳硶淇敼")
	ErrInvalidOrderStatus    = NewBizError(4006, "璁㈠崟鐘舵€佹棤鏁?)
	ErrGiftCardNotFound      = NewBizError(4007, "绀煎搧鍗′笉瀛樺湪")
	ErrGiftCardExpired       = NewBizError(4008, "绀煎搧鍗″凡杩囨湡")
	ErrGiftCardUsed          = NewBizError(4009, "绀煎搧鍗″凡浣跨敤")
	ErrReturnRequestNotFound = NewBizError(4010, "閫€璐ц姹備笉瀛樺湪")
)

// 鏀粯鏈嶅姟閿欒鐮?
var (
	ErrPaymentNotFound       = NewBizError(5001, "鏀粯璁板綍涓嶅瓨鍦?)
	ErrPaymentFailed         = NewBizError(5002, "鏀粯澶辫触")
	ErrPaymentTimeout        = NewBizError(5003, "鏀粯瓒呮椂")
	ErrPaymentAlreadyRefunded = NewBizError(5004, "鏀粯宸查€€娆?)
	ErrRefundFailed          = NewBizError(5005, "閫€娆惧け璐?)
	ErrInvalidPaymentMethod  = NewBizError(5006, "鏃犳晥鐨勬敮浠樻柟寮?)
)

// 搴撳瓨鏈嶅姟閿欒鐮?
var (
	ErrInventoryNotFound     = NewBizError(6001, "搴撳瓨璁板綍涓嶅瓨鍦?)
	ErrInsufficientStock     = NewBizError(6002, "搴撳瓨涓嶈冻")
	ErrStockReservationFailed = NewBizError(6003, "搴撳瓨棰勭暀澶辫触")
	ErrWarehouseNotFound     = NewBizError(6004, "浠撳簱涓嶅瓨鍦?)
	ErrInventoryLocked       = NewBizError(6005, "搴撳瓨琚攣瀹?)
)

// 鐗╂祦鏈嶅姟閿欒鐮?
var (
	ErrShipmentNotFound     = NewBizError(7001, "鍙戣揣鍗曚笉瀛樺湪")
	ErrShippingMethodNotFound = NewBizError(7002, "閰嶉€佹柟寮忎笉瀛樺湪")
	ErrInvalidTrackingNumber = NewBizError(7003, "鏃犳晥鐨勭墿娴佸崟鍙?)
	ErrShipmentAlreadyDelivered = NewBizError(7004, "鍙戣揣鍗曞凡閫佽揪")
)

// BizError 涓氬姟閿欒
type BizError struct {
	Code    int
	Message string
}

// NewBizError 鍒涘缓涓氬姟閿欒
func NewBizError(code int, message string) *BizError {
	return &BizError{Code: code, Message: message}
}

// Error 瀹炵幇 error 鎺ュ彛
func (e *BizError) Error() string {
	return e.Message
}

// ToAppError 杞崲涓?gorp AppError
// 涓枃璇存槑: 灏嗚嚜瀹氫箟 BizError 杞崲涓烘鏋舵爣鍑嗙殑 AppError
func (e *BizError) ToAppError() resiliencecontract.AppError {
	// 鏍规嵁閿欒鐮佹槧灏勫埌瀵瑰簲鐨?HTTP 鐘舵€佺爜鍜?Reason
	var code int
	var reason resiliencecontract.ErrorReason

	// 閿欒鐮佽寖鍥存槧灏?
	// - 1000-1999: 閫氱敤閿欒,榛樿 400
	// - 2000-2999: 瀹㈡埛鏈嶅姟閿欒,榛樿 400
	// - 3000-3999: 鍟嗗搧鏈嶅姟閿欒,榛樿 404
	// - 4000-3999: 璁㈠崟鏈嶅姟閿欒,榛樿 400
	// - 5000-5999: 鏀粯鏈嶅姟閿欒,榛樿 400
	// - 6000-6999: 搴撳瓨鏈嶅姟閿欒,榛樿 400
	// - 7000-7999: 鐗╂祦鏈嶅姟閿欒,榛樿 404
	switch {
	case e.Code >= 1000 && e.Code < 2000:
		code = resiliencecontract.ErrorCodeBadRequest
		reason = resiliencecontract.ErrorReasonBadRequest
	case e.Code >= 2000 && e.Code < 3000:
		// 瀹㈡埛鐩稿叧閿欒,澶氭暟涓?400/401/404
		switch e.Code {
		case 2001, 2009: // Not Found
			code = resiliencecontract.ErrorCodeNotFound
			reason = resiliencecontract.ErrorReasonNotFound
		case 2003, 2006: // Unauthorized/Forbidden
			code = resiliencecontract.ErrorCodeUnauthorized
			reason = resiliencecontract.ErrorReasonUnauthorized
		default:
			code = resiliencecontract.ErrorCodeBadRequest
			reason = resiliencecontract.ErrorReasonBadRequest
		}
	case e.Code >= 3000 && e.Code < 4000:
		code = resiliencecontract.ErrorCodeNotFound
		reason = resiliencecontract.ErrorReasonNotFound
	case e.Code >= 4000 && e.Code < 5000:
		code = resiliencecontract.ErrorCodeBadRequest
		reason = resiliencecontract.ErrorReasonBadRequest
	case e.Code >= 5000 && e.Code < 6000:
		code = resiliencecontract.ErrorCodeBadRequest
		reason = resiliencecontract.ErrorReasonBadRequest
	case e.Code >= 6000 && e.Code < 7000:
		// 搴撳瓨涓嶈冻绛夐敊璇睘浜庝笟鍔″啿绐?
		code = resiliencecontract.ErrorCodeConflict
		reason = resiliencecontract.ErrorReasonConflict
	case e.Code >= 7000 && e.Code < 8000:
		code = resiliencecontract.ErrorCodeNotFound
		reason = resiliencecontract.ErrorReasonNotFound
	default:
		code = resiliencecontract.ErrorCodeInternalServerError
		reason = resiliencecontract.ErrorReasonInternal
	}

	return resiliencecontract.NewError(code, reason, e.Message)
}

// Is 鍒ゆ柇閿欒鏄惁鐩哥瓑
func (e *BizError) Is(target error) bool {
	t, ok := target.(*BizError)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

// WrapError 鍖呰閿欒
func WrapError(err error, code int, message string) *BizError {
	return &BizError{
		Code:    code,
		Message: fmt.Sprintf("%s: %v", message, err),
	}
}