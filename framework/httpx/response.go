// Package httpx provides HTTP response helpers and error mappings for gorp framework.
package httpx

import (
	"errors"
	"fmt"
	"net/http"
	"sync/atomic"

	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
)

// JSONResponder represents any context capable of rendering a JSON response,
// seamlessly supporting both *gin.Context and transportcontract.Context.
type JSONResponder interface {
	JSON(code int, obj any)
}

// Code represents canonical semantic status codes aligned with gRPC AIP-193.
// It determines the HTTP transmission status, while business domain semantics use Reason.
type Code int

const (
	CodeOK                 Code = 0
	CodeCancelled          Code = 1
	CodeUnknown            Code = 2
	CodeInvalidArgument    Code = 3  // -> 400
	CodeDeadlineExceeded   Code = 4  // -> 504
	CodeNotFound           Code = 5  // -> 404
	CodeAlreadyExists      Code = 6  // -> 409
	CodePermissionDenied   Code = 7  // -> 403
	CodeResourceExhausted  Code = 8  // -> 429
	CodeFailedPrecondition Code = 9  // -> 400
	CodeAborted            Code = 10 // -> 409
	CodeOutOfRange         Code = 11 // -> 400
	CodeUnimplemented      Code = 12 // -> 501
	CodeInternal           Code = 13 // -> 500
	CodeUnavailable        Code = 14 // -> 503
	CodeDataLoss           Code = 15 // -> 500
	CodeUnauthenticated    Code = 16 // -> 401
)

// CodeSuccess is an alias for CodeOK to maintain compatibility where needed.
const CodeSuccess = CodeOK

// codeTable is the single source of truth for Code -> HTTP status code mapping.
var codeTable = map[Code]int{
	CodeOK:                 http.StatusOK,                  // 200
	CodeCancelled:          499,                            // 499
	CodeUnknown:            http.StatusInternalServerError, // 500
	CodeInvalidArgument:    http.StatusBadRequest,          // 400
	CodeDeadlineExceeded:   http.StatusGatewayTimeout,      // 504
	CodeNotFound:           http.StatusNotFound,            // 404
	CodeAlreadyExists:      http.StatusConflict,            // 409
	CodePermissionDenied:   http.StatusForbidden,           // 403
	CodeResourceExhausted:  http.StatusTooManyRequests,     // 429
	CodeFailedPrecondition: http.StatusBadRequest,          // 400
	CodeAborted:            http.StatusConflict,            // 409
	CodeOutOfRange:         http.StatusBadRequest,          // 400
	CodeUnimplemented:      http.StatusNotImplemented,      // 501
	CodeInternal:           http.StatusInternalServerError, // 500
	CodeUnavailable:        http.StatusServiceUnavailable,  // 503
	CodeDataLoss:           http.StatusInternalServerError, // 500
	CodeUnauthenticated:    http.StatusUnauthorized,        // 401
}

// Response is the default HTTP response envelope used by the framework.
type Response struct {
	Code    Code   `json:"code"`
	Reason  string `json:"reason,omitempty"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// Problem represents an RFC 9457 Problem Details payload (application/problem+json).
type Problem struct {
	Type   string `json:"type"`
	Title  string `json:"title"`
	Status int    `json:"status"`
	Detail string `json:"detail,omitempty"`
	Reason string `json:"reason,omitempty"`
}

var useProblemJSON atomic.Bool

// UseProblemJSON enables or disables RFC 9457 Problem Details (application/problem+json) error responses.
func UseProblemJSON(enabled bool) {
	useProblemJSON.Store(enabled)
}

// IsProblemJSONEnabled reports whether RFC 9457 Problem Details is enabled.
func IsProblemJSONEnabled() bool {
	return useProblemJSON.Load()
}

// PaginatedData is the standard pagination payload shape.
type PaginatedData struct {
	Items    any   `json:"items"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}

// BusinessError describes an error carrying canonical Code, Reason, and message semantics.
type BusinessError interface {
	error
	Code() Code
	Reason() string
	Message() string
}

// BizError is the standard business error carrying canonical Code and domain Reason.
type BizError struct {
	code    Code
	reason  string // 业务子码，如 "payment.balance_not_enough"（可空）
	message string
}

func (e *BizError) Error() string {
	if e.message != "" {
		return e.message
	}
	return e.reason
}

func (e *BizError) Code() Code      { return e.code }
func (e *BizError) Reason() string  { return e.reason }
func (e *BizError) Message() string { return e.message }

// NewBizError creates a new BizError with constructor-level fail-fast verification.
func NewBizError(code Code, reason, message string) *BizError {
	if _, ok := codeTable[code]; !ok {
		panic(fmt.Sprintf("httpx: 非法语义码 %d，必须使用 Code 常量", code))
	}
	if code == CodeOK {
		panic("httpx: CodeOK 不能构造错误")
	}
	return &BizError{code: code, reason: reason, message: message}
}

// Convenient constructors for canonical codes (covers 90%+ business scenarios).
func InvalidArgument(reason, msg string) *BizError    { return NewBizError(CodeInvalidArgument, reason, msg) }
func NotFound(reason, msg string) *BizError           { return NewBizError(CodeNotFound, reason, msg) }
func Unauthenticated(reason, msg string) *BizError    { return NewBizError(CodeUnauthenticated, reason, msg) }
func PermissionDenied(reason, msg string) *BizError   { return NewBizError(CodePermissionDenied, reason, msg) }
func ResourceExhausted(reason, msg string) *BizError  { return NewBizError(CodeResourceExhausted, reason, msg) }
func FailedPrecondition(reason, msg string) *BizError { return NewBizError(CodeFailedPrecondition, reason, msg) }
func Internal(msg string) *BizError                   { return NewBizError(CodeInternal, "", msg) }

func Cancelled(reason, msg string) *BizError        { return NewBizError(CodeCancelled, reason, msg) }
func Unknown(reason, msg string) *BizError          { return NewBizError(CodeUnknown, reason, msg) }
func DeadlineExceeded(reason, msg string) *BizError { return NewBizError(CodeDeadlineExceeded, reason, msg) }
func AlreadyExists(reason, msg string) *BizError    { return NewBizError(CodeAlreadyExists, reason, msg) }
func Aborted(reason, msg string) *BizError          { return NewBizError(CodeAborted, reason, msg) }
func OutOfRange(reason, msg string) *BizError       { return NewBizError(CodeOutOfRange, reason, msg) }
func Unimplemented(reason, msg string) *BizError    { return NewBizError(CodeUnimplemented, reason, msg) }
func Unavailable(reason, msg string) *BizError      { return NewBizError(CodeUnavailable, reason, msg) }
func DataLoss(reason, msg string) *BizError         { return NewBizError(CodeDataLoss, reason, msg) }

// Success writes a standard success response.
func Success(c JSONResponder, data any) {
	c.JSON(http.StatusOK, Response{Code: CodeOK, Message: "success", Data: data})
}

// OK is an alias for Success.
func OK(c JSONResponder, data any) {
	Success(c, data)
}

// SuccessWithMessage writes a success response with a custom message.
func SuccessWithMessage(c JSONResponder, message string, data any) {
	c.JSON(http.StatusOK, Response{Code: CodeOK, Message: message, Data: data})
}

// SuccessWithStatus writes a success response with a custom HTTP status.
func SuccessWithStatus(c JSONResponder, status int, data any) {
	c.JSON(status, Response{Code: CodeOK, Message: "success", Data: data})
}

// SuccessPaginated writes a standard paginated success response.
func SuccessPaginated(c JSONResponder, items any, total int64, page, pageSize int) {
	Success(c, PaginatedData{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

// Error writes an error response, automatically resolving AppError and BizError.
func Error(c JSONResponder, err error) {
	bizErr := parseError(err)
	status := codeToHTTPStatus(bizErr.code)
	writeError(c, status, bizErr, nil)
}

// Fail is an alias for Error.
func Fail(c JSONResponder, err error) {
	Error(c, err)
}

// ErrorWithData writes an error response and attaches extra response data.
func ErrorWithData(c JSONResponder, err error, data any) {
	bizErr := parseError(err)
	status := codeToHTTPStatus(bizErr.code)
	writeError(c, status, bizErr, data)
}

// ErrorWithStatus writes an error response using a caller-provided HTTP status.
func ErrorWithStatus(c JSONResponder, status int, err error) {
	bizErr := parseError(err)
	writeError(c, status, bizErr, nil)
}

func writeError(c JSONResponder, status int, bizErr *BizError, data any) {
	if useProblemJSON.Load() {
		if hs, ok := c.(interface{ Header(string, string) }); ok {
			hs.Header("Content-Type", "application/problem+json")
		} else if shs, ok := c.(interface{ SetHeader(string, string) }); ok {
			shs.SetHeader("Content-Type", "application/problem+json")
		}
		problemType := "about:blank"
		if bizErr.reason != "" {
			problemType = bizErr.reason
		}
		prob := Problem{
			Type:   problemType,
			Title:  http.StatusText(status),
			Status: status,
			Detail: bizErr.message,
			Reason: bizErr.reason,
		}
		c.JSON(status, prob)
		return
	}

	c.JSON(status, Response{
		Code:    bizErr.code,
		Reason:  bizErr.reason,
		Message: bizErr.message,
		Data:    data,
	})
}

func parseError(err error) *BizError {
	if err == nil {
		return &BizError{code: CodeOK, message: "success"}
	}
	var bizErr *BizError
	if errors.As(err, &bizErr) {
		return bizErr
	}
	var bErr BusinessError
	if errors.As(err, &bErr) {
		return &BizError{
			code:    bErr.Code(),
			reason:  bErr.Reason(),
			message: bErr.Message(),
		}
	}
	var appErr resiliencecontract.AppError
	if errors.As(err, &appErr) {
		return fromAppError(appErr)
	}
	return &BizError{code: CodeInternal, message: err.Error()}
}

func fromAppError(appErr resiliencecontract.AppError) *BizError {
	st := appErr.GetStatus()
	if st == nil {
		return &BizError{code: CodeInternal, message: appErr.Error()}
	}
	code := mapAppErrorCode(int(st.Code), st.Reason)
	reason := string(st.Reason)
	if st.Reason == resiliencecontract.ErrorReasonUnknown {
		reason = ""
	}
	return &BizError{
		code:    code,
		reason:  reason,
		message: st.Message,
	}
}

func mapAppErrorCode(code int, reason resiliencecontract.ErrorReason) Code {
	switch {
	case code == 400 || reason == resiliencecontract.ErrorReasonBadRequest:
		return CodeInvalidArgument
	case code == 401 || reason == resiliencecontract.ErrorReasonUnauthorized:
		return CodeUnauthenticated
	case code == 403 || reason == resiliencecontract.ErrorReasonForbidden:
		return CodePermissionDenied
	case code == 404 || reason == resiliencecontract.ErrorReasonNotFound:
		return CodeNotFound
	case code == 409 || reason == resiliencecontract.ErrorReasonConflict:
		return CodeAlreadyExists
	case code == 429 || reason == resiliencecontract.ErrorReasonRateLimited:
		return CodeResourceExhausted
	case code == 503 || reason == resiliencecontract.ErrorReasonServiceUnavailable:
		return CodeUnavailable
	case code == 504 || reason == resiliencecontract.ErrorReasonTimeout:
		return CodeDeadlineExceeded
	case code == 500 || reason == resiliencecontract.ErrorReasonInternal:
		return CodeInternal
	default:
		return CodeInternal
	}
}

func codeToHTTPStatus(code Code) int {
	if status, ok := codeTable[code]; ok {
		return status
	}
	return codeTable[CodeUnknown]
}
