// Package retry provides retry service implementation.
// Implements retry logic with exponential backoff, jitter, and error classification.
//
// 重试包提供重试服务实现。
// 实现带指数退避、抖动和错误分类的重试逻辑。
package retry

import (
	"context"
	"errors"
	"math/rand"
	"net"
	"reflect"
	"strings"
	"sync"
	"time"

	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
)

// RetryService implements retry logic with configurable policy.
// Core logic: Execute function with retry, calculate delay with jitter, classify retryable errors.
//
// RetryService 实现带可配置策略的重试逻辑。
// 核心逻辑：带重试执行函数、带抖动计算延迟、分类可重试错误。
type RetryService struct {
	cfg *resiliencecontract.RetryConfig
	mu  sync.RWMutex
}

// NewRetryService creates a retry service with configuration.
// Core logic: Initialize random source for jitter.
//
// NewRetryService 创建带配置的重试服务。
// 核心逻辑：初始化随机源用于抖动。
func NewRetryService(cfg *resiliencecontract.RetryConfig) *RetryService {
	if cfg == nil {
		cfg = &resiliencecontract.RetryConfig{
			Enabled:       true,
			DefaultPolicy: resiliencecontract.DefaultRetryPolicy(),
		}
	}
	return &RetryService{cfg: cfg}
}

// Do executes function with default retry policy.
// Core logic: Call doWithPolicy with default config.
//
// Do 使用默认重试策略执行函数。
// 核心逻辑：调用 doWithPolicy 并使用默认配置。
func (r *RetryService) Do(ctx context.Context, fn func() error) error {
	r.mu.RLock()
	policy := r.cfg.DefaultPolicy
	r.mu.RUnlock()
	return r.doWithPolicy(ctx, policy, fn)
}

func (r *RetryService) doWithPolicy(ctx context.Context, policy resiliencecontract.RetryPolicy, fn func() error) error {
	if policy.MaxAttempts < 1 {
		// A non-positive attempt count must still execute the operation once
		// instead of silently skipping it and reporting success.
		policy.MaxAttempts = 1
	}
	var lastErr error
	for attempt := 0; attempt < policy.MaxAttempts; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err
		if !isRetryableWithPolicy(err, policy) {
			return err
		}
		if attempt == policy.MaxAttempts-1 {
			break
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		jitter := rand.Float64()
		delay := policy.CalculateDelay(attempt, jitter)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}

	return lastErr
}

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

// IsRetryable checks if error is retryable based on policy.
// Core logic: Check AppError reason/code, gRPC status, network error type.
//
// IsRetryable 根据策略检查错误是否可重试。
// 核心逻辑：检查 AppError reason/code、gRPC status、网络错误类型。
func (r *RetryService) IsRetryable(err error) bool {
	r.mu.RLock()
	policy := r.cfg.DefaultPolicy
	r.mu.RUnlock()
	return isRetryableWithPolicy(err, policy)
}

func isRetryableWithPolicy(err error, policy resiliencecontract.RetryPolicy) bool {
	if err == nil {
		return false
	}

	var appErr resiliencecontract.AppError
	if errors.As(err, &appErr) {
		st := appErr.GetStatus()
		if st == nil {
			return false
		}

		for _, reason := range policy.RetryableErrors {
			if st.Reason == reason {
				return true
			}
		}
		for _, code := range policy.RetryableCodes {
			if int(st.Code) == code {
				return true
			}
		}
		// A classified business error that misses the policy is deterministic:
		// do not fall through to the string-heuristic network checks below.
		return false
	}

	// gRPC 错误分类：通过接口方法提取 code 字符串（如 "Unavailable"），
	// 不引入 grpc 依赖。与 grpc/status.FromError 语义一致——非 gRPC 错误返回 ""。
	if code := grpcErrorCodeString(err); code != "" {
		for _, rc := range policy.RetryableGRPCCodes {
			if code == rc {
				return true
			}
		}
		switch code {
		case "Unavailable", "DeadlineExceeded", "ResourceExhausted", "Aborted":
			return true
		}
		return false
	}

	// Context errors are never retryable, and must be checked before
	// isNetworkError: context.DeadlineExceeded implements net.Error and would
	// otherwise be misclassified as a retryable timeout.
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return false
	}
	if isNetworkError(err) {
		return true
	}
	return false
}

func isNetworkError(err error) bool {
	if err == nil {
		return false
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		if netErr.Timeout() || netErr.Temporary() {
			return true
		}
	}

	var opErr *net.OpError
	if errors.As(err, &opErr) {
		if opErr.Op == "dial" || opErr.Op == "read" || opErr.Op == "write" {
			return true
		}
	}

	errMsg := strings.ToLower(err.Error())
	retryableMessages := []string{
		"connection refused",
		"connection reset",
		"broken pipe",
		"timeout",
		"eof",
		"temporary failure",
	}

	for _, msg := range retryableMessages {
		if strings.Contains(errMsg, msg) {
			return true
		}
	}

	return false
}

func (r *RetryService) GetConfig() *resiliencecontract.RetryConfig {
	r.mu.RLock()
	defer r.mu.RUnlock()
	copyCfg := *r.cfg
	if r.cfg.ResourcePolicies != nil {
		copyCfg.ResourcePolicies = make(map[string]resiliencecontract.RetryPolicy, len(r.cfg.ResourcePolicies))
		for key, policy := range r.cfg.ResourcePolicies {
			copyCfg.ResourcePolicies[key] = policy
		}
	}
	return &copyCfg
}

func (r *RetryService) SetPolicy(resource string, policy resiliencecontract.RetryPolicy) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.cfg.ResourcePolicies == nil {
		r.cfg.ResourcePolicies = make(map[string]resiliencecontract.RetryPolicy)
	}
	r.cfg.ResourcePolicies[resource] = policy
}

func (r *RetryService) DoForResource(ctx context.Context, resource string, fn func() error) error {
	r.mu.RLock()
	policy := r.cfg.GetPolicy(resource)
	r.mu.RUnlock()
	return r.doWithPolicy(ctx, policy, fn)
}

// grpcErrorCodeString 从错误中提取 gRPC 状态码字符串（如 "Unavailable"）。
// 用反射访问 GRPCStatus()/Code()/String()，避免引入 google.golang.org/grpc 依赖；
// 非 gRPC 错误返回 ""。语义对齐 grpc/status.FromError：遍历 unwrap 链。
// 注：不能用接口断言实现——Go 要求接口方法返回类型与实现完全相同（不可协变），
// 而 codes.Code 是具体类型，无法用本地接口收窄。
func grpcErrorCodeString(err error) string {
	for e := err; e != nil; e = errors.Unwrap(e) {
		if code := grpcCodeViaReflect(e); code != "" {
			return code
		}
	}
	return ""
}

// grpcCodeViaReflect 通过反射调用 GRPCStatus().Code().String()。
func grpcCodeViaReflect(err error) string {
	v := reflect.ValueOf(err)
	m := v.MethodByName("GRPCStatus")
	if !m.IsValid() {
		return ""
	}
	out := m.Call(nil)
	if len(out) != 1 || !out[0].IsValid() {
		return ""
	}
	codeM := out[0].MethodByName("Code")
	if !codeM.IsValid() {
		return ""
	}
	codeOut := codeM.Call(nil)
	if len(codeOut) != 1 || !codeOut[0].IsValid() {
		return ""
	}
	strM := codeOut[0].MethodByName("String")
	if !strM.IsValid() {
		return ""
	}
	strOut := strM.Call(nil)
	if len(strOut) != 1 {
		return ""
	}
	return strOut[0].String()
}
