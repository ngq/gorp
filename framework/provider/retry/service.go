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

	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
)

type RetryService struct {
	cfg *resiliencecontract.RetryConfig
	rng *rand.Rand
}

func NewRetryService(cfg *resiliencecontract.RetryConfig) *RetryService {
	return &RetryService{
		cfg: cfg,
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (r *RetryService) Do(ctx context.Context, fn func() error) error {
	return r.doWithPolicy(ctx, r.cfg.DefaultPolicy, fn)
}

func (r *RetryService) doWithPolicy(ctx context.Context, policy resiliencecontract.RetryPolicy, fn func() error) error {
	var lastErr error
	for attempt := 0; attempt < policy.MaxAttempts; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err
		if !r.IsRetryable(err) {
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

func (r *RetryService) IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	policy := r.cfg.DefaultPolicy

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
	}

	grpcStatus, ok := status.FromError(err)
	if ok {
		for _, code := range policy.RetryableGRPCCodes {
			if grpcStatus.Code().String() == code {
				return true
			}
		}
		switch grpcStatus.Code() {
		case codes.Unavailable, codes.DeadlineExceeded, codes.ResourceExhausted, codes.Aborted:
			return true
		}
		return false
	}

	if isNetworkError(err) {
		return true
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return false
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

func (r *RetryService) GetConfig() *resiliencecontract.RetryConfig {
	return r.cfg
}

func (r *RetryService) SetPolicy(resource string, policy resiliencecontract.RetryPolicy) {
	if r.cfg.ResourcePolicies == nil {
		r.cfg.ResourcePolicies = make(map[string]resiliencecontract.RetryPolicy)
	}
	r.cfg.ResourcePolicies[resource] = policy
}

func (r *RetryService) DoForResource(ctx context.Context, resource string, fn func() error) error {
	return r.doWithPolicy(ctx, r.cfg.GetPolicy(resource), fn)
}
