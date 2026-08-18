package retry

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
)

// 本测试锁定 gRPC 错误分类语义：retry 实现不再 import grpc，
// 但必须与 grpc/status.FromError 的行为完全一致（实现已解耦，测试仍可引 grpc 验证）。
func TestGrpcErrorCodeString_ExtractsCode(t *testing.T) {
	// 非 gRPC 错误 → ""
	if got := grpcErrorCodeString(errors.New("plain error")); got != "" {
		t.Fatalf("expected empty for non-grpc error, got %q", got)
	}
	// context 错误 → ""（且不误判为可重试）
	if got := grpcErrorCodeString(context.DeadlineExceeded); got != "" {
		t.Fatalf("expected empty for context error, got %q", got)
	}
	// grpc 错误 → 提取 code 字符串
	for _, c := range []codes.Code{codes.Unavailable, codes.DeadlineExceeded, codes.ResourceExhausted, codes.Aborted, codes.NotFound} {
		if got := grpcErrorCodeString(status.Error(c, "x")); got != c.String() {
			t.Fatalf("code %s: expected %q, got %q", c, c.String(), got)
		}
	}
	// 包装的 grpc 错误（Unwrap 链）同样可提取
	wrapped := fmt.Errorf("outer: %w", status.Error(codes.Unavailable, "down"))
	if got := grpcErrorCodeString(wrapped); got != "Unavailable" {
		t.Fatalf("expected Unavailable from wrapped grpc error, got %q", got)
	}
}

func TestIsRetryable_GrpcCodes(t *testing.T) {
	policy := resiliencecontract.DefaultRetryPolicy()
	policy.RetryableGRPCCodes = []string{"Unavailable"}

	// Unavailable 命中配置
	if !isRetryableWithPolicy(status.Error(codes.Unavailable, "down"), policy) {
		t.Fatal("expected Unavailable to be retryable")
	}
	// Internal 不在配置也不在默认集 → 不可重试
	if isRetryableWithPolicy(status.Error(codes.Internal, "bug"), policy) {
		t.Fatal("expected Internal NOT to be retryable")
	}
	// 默认重试码（policy 默认含 Unavailable/DeadlineExceeded/ResourceExhausted/Aborted）
	pol2 := resiliencecontract.DefaultRetryPolicy()
	for _, code := range []codes.Code{codes.Unavailable, codes.DeadlineExceeded, codes.ResourceExhausted, codes.Aborted} {
		if !isRetryableWithPolicy(status.Error(code, "x"), pol2) {
			t.Fatalf("expected default retryable grpc code %s", code)
		}
	}
	if isRetryableWithPolicy(status.Error(codes.NotFound, "x"), pol2) {
		t.Fatal("expected NotFound NOT retryable by default")
	}
}
