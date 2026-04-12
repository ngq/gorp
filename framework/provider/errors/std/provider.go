package std

import (
	"context"

	"github.com/ngq/gorp/framework/contract"
)

// Provider 提供统一错误处理实现。
//
// 中文说明：
// - 支持错误创建、转换；
// - 支持 HTTP/gRPC 错误码转换；
// - 自己实现，不抄袭 Kratos。
type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string     { return "errors.default" }
func (p *Provider) IsDefer() bool    { return true }
func (p *Provider) Provides() []string {
	return []string{contract.ErrorsKey}
}

func (p *Provider) Register(c contract.Container) error {
	// 注册错误处理器
	c.Bind(contract.ErrorsKey, func(c contract.Container) (any, error) {
		return &errorHandler{}, nil
	}, true)
	return nil
}

func (p *Provider) Boot(c contract.Container) error {
	return nil
}

// ErrorHandler 错误处理器。
type errorHandler struct{}

// HTTPToGRPC 将 HTTP 错误码转换为 gRPC 错误码。
//
// 中文说明：
// - HTTP 400 -> gRPC InvalidArgument；
// - HTTP 401 -> gRPC Unauthenticated；
// - HTTP 403 -> gRPC PermissionDenied；
// - HTTP 404 -> gRPC NotFound；
// - HTTP 409 -> gRPC AlreadyExists；
// - HTTP 429 -> gRPC ResourceExhausted；
// - HTTP 500 -> gRPC Internal；
// - HTTP 503 -> gRPC Unavailable；
// - HTTP 504 -> gRPC DeadlineExceeded。
func (h *errorHandler) HTTPToGRPC(httpCode int) int {
	// gRPC 错误码定义
	const (
		CodeOK                  = 0
		CodeCanceled            = 1
		CodeUnknown             = 2
		CodeInvalidArgument     = 3
		CodeDeadlineExceeded    = 4
		CodeNotFound            = 5
		CodeAlreadyExists       = 6
		CodePermissionDenied    = 7
		CodeResourceExhausted   = 8
		CodeFailedPrecondition  = 9
		CodeAborted             = 10
		CodeOutOfRange          = 11
		CodeUnimplemented       = 12
		CodeInternal            = 13
		CodeUnavailable         = 14
		CodeDataLoss            = 15
		CodeUnauthenticated     = 16
	)

	switch httpCode {
	case 200:
		return CodeOK
	case 400:
		return CodeInvalidArgument
	case 401:
		return CodeUnauthenticated
	case 403:
		return CodePermissionDenied
	case 404:
		return CodeNotFound
	case 409:
		return CodeAlreadyExists
	case 429:
		return CodeResourceExhausted
	case 500:
		return CodeInternal
	case 503:
		return CodeUnavailable
	case 504:
		return CodeDeadlineExceeded
	default:
		return CodeUnknown
	}
}

// GRPCToHTTP 将 gRPC 错误码转换为 HTTP 错误码。
func (h *errorHandler) GRPCToHTTP(grpcCode int) int {
	switch grpcCode {
	case 0: // OK
		return 200
	case 1: // Canceled
		return 499 // Client Closed Request
	case 2: // Unknown
		return 500
	case 3: // InvalidArgument
		return 400
	case 4: // DeadlineExceeded
		return 504
	case 5: // NotFound
		return 404
	case 6: // AlreadyExists
		return 409
	case 7: // PermissionDenied
		return 403
	case 8: // ResourceExhausted
		return 429
	case 9: // FailedPrecondition
		return 400
	case 10: // Aborted
		return 409
	case 11: // OutOfRange
		return 400
	case 12: // Unimplemented
		return 501
	case 13: // Internal
		return 500
	case 14: // Unavailable
		return 503
	case 15: // DataLoss
		return 500
	case 16: // Unauthenticated
		return 401
	default:
		return 500
	}
}

// WrapError 包装错误为 AppError。
func (h *errorHandler) WrapError(ctx context.Context, err error, code int, reason contract.ErrorReason, message string) contract.AppError {
	if err == nil {
		return nil
	}

	// 如果已经是 AppError，添加原因
	if appErr, ok := err.(contract.AppError); ok {
		return appErr.WithCause(err)
	}

	// 创建新的 AppError
	return contract.NewError(code, reason, message).WithCause(err)
}

// IsBusinessError 判断是否为业务错误。
//
// 中文说明：
// - 业务错误通常是 4xx 错误；
// - 系统错误通常是 5xx 错误。
func (h *errorHandler) IsBusinessError(err error) bool {
	code := contract.Code(err)
	return code >= 400 && code < 500
}

// IsSystemError 判断是否为系统错误。
func (h *errorHandler) IsSystemError(err error) bool {
	code := contract.Code(err)
	return code >= 500
}