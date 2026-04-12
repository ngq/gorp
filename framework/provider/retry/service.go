package retry

import (
	"context"
	"errors"
	"math/rand"
	"net"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ngq/gorp/framework/contract"
)

// RetryService 实现 Retry 接口。
//
// 中文说明：
// - 支持指数退避重试策略；
// - 支持可重试错误判断；
// - 每次重试前检查 context。
type RetryService struct {
	cfg *contract.RetryConfig
	rng *rand.Rand
}

// NewRetryService 创建 RetryService。
func NewRetryService(cfg *contract.RetryConfig) *RetryService {
	return &RetryService{
		cfg: cfg,
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Do 执行带重试的函数。
//
// 中文说明：
// - 根据 RetryPolicy 进行重试；
// - 使用指数退避延迟；
// - 最多重试 MaxAttempts 次；
// - 每次重试前检查 context 取消。
func (r *RetryService) Do(ctx context.Context, fn func() error) error {
	policy := r.cfg.DefaultPolicy

	var lastErr error

	for attempt := 0; attempt < policy.MaxAttempts; attempt++ {
		// 执行函数
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// 判断是否可重试
		if !r.IsRetryable(err) {
			return err
		}

		// 最后一次尝试不等待
		if attempt == policy.MaxAttempts-1 {
			break
		}

		// 检查 context
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// 计算延迟并等待
		jitter := r.rng.Float64()
		delay := policy.CalculateDelay(attempt, jitter)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}

	return lastErr
}

// DoWithResult 执行带重试和返回值的函数。
//
// 中文说明：
// - 支持返回值的重试执行；
// - 成功时返回结果和 nil；
// - 失败时返回 nil 和最后一个错误。
func (r *RetryService) DoWithResult(ctx context.Context, fn func() (any, error)) (any, error) {
	var result any

	err := r.Do(ctx, func() error {
		res, e := fn()
		if e != nil {
			return e
		}
		result = res
		return nil
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

// IsRetryable 判断错误是否可重试。
//
// 中文说明：
// - 检查 AppError 错误码；
// - 检查 HTTP 状态码；
// - 检查 gRPC 状态码；
// - 检查网络错误。
func (r *RetryService) IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	policy := r.cfg.DefaultPolicy

	// 检查 AppError
	var appErr contract.AppError
	if errors.As(err, &appErr) {
		st := appErr.GetStatus()
		if st == nil {
			return false
		}

		// 检查错误原因
		for _, reason := range policy.RetryableErrors {
			if st.Reason == reason {
				return true
			}
		}

		// 检查 HTTP 状态码
		for _, code := range policy.RetryableCodes {
			if int(st.Code) == code {
				return true
			}
		}
	}

	// 检查 gRPC status
	grpcStatus, ok := status.FromError(err)
	if ok {
		// 检查配置的 gRPC 状态码
		for _, code := range policy.RetryableGRPCCodes {
			if grpcStatus.Code().String() == code {
				return true
			}
		}

		// 默认可重试的 gRPC 状态码
		switch grpcStatus.Code() {
		case codes.Unavailable, codes.DeadlineExceeded, codes.ResourceExhausted, codes.Aborted:
			return true
		}
		return false
	}

	// 检查网络错误
	if isNetworkError(err) {
		return true
	}

	// 检查 context 错误
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return false // context 错误不重试
	}

	return false
}

// isNetworkError 判断是否为网络错误。
//
// 中文说明：
// - 检查常见的网络错误类型；
// - 连接拒绝、超时、重置等。
func isNetworkError(err error) bool {
	if err == nil {
		return false
	}

	// 检查 net.Error
	var netErr net.Error
	if errors.As(err, &netErr) {
		// 超时和临时错误可重试
		if netErr.Timeout() || netErr.Temporary() {
			return true
		}
	}

	// 检查特定错误类型
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		// 连接被拒绝、重置等
		if opErr.Op == "dial" || opErr.Op == "read" || opErr.Op == "write" {
			return true
		}
	}

	// 检查错误消息
	errMsg := err.Error()
	retryableMessages := []string{
		"connection refused",
		"connection reset",
		"broken pipe",
		"timeout",
		"EOF",
		"temporary failure",
	}

	for _, msg := range retryableMessages {
		if strings.Contains(strings.ToLower(errMsg), msg) {
			return true
		}
	}

	return false
}

// GetConfig 获取当前配置。
func (r *RetryService) GetConfig() *contract.RetryConfig {
	return r.cfg
}

// SetPolicy 设置指定资源的重试策略。
//
// 中文说明：
// - 动态添加或更新资源策略；
// - 用于运行时调整重试行为。
func (r *RetryService) SetPolicy(resource string, policy contract.RetryPolicy) {
	if r.cfg.ResourcePolicies == nil {
		r.cfg.ResourcePolicies = make(map[string]contract.RetryPolicy)
	}
	r.cfg.ResourcePolicies[resource] = policy
}

// DoForResource 为指定资源执行重试。
//
// 中文说明：
// - 使用资源特定的策略进行重试；
// - 如果没有特定策略则使用默认策略。
func (r *RetryService) DoForResource(ctx context.Context, resource string, fn func() error) error {
	// 暂存默认策略
	originalPolicy := r.cfg.DefaultPolicy

	// 获取资源策略
	if policy, ok := r.cfg.ResourcePolicies[resource]; ok {
		r.cfg.DefaultPolicy = policy
	}

	// 恢复默认策略
	defer func() {
		r.cfg.DefaultPolicy = originalPolicy
	}()

	return r.Do(ctx, fn)
}