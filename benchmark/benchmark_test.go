// Package benchmark 提供了 gorp 框架核心模块的性能基准测试。
//
// 运行方式：
//
//	go test ./benchmark/... -bench=. -benchmem
//
// 中文说明：
// - 集中管理高频调用模块的性能测试；
// - 便于横向对比和性能回归检测；
// - 测试结果用于文档和优化决策。
//
// 覆盖模块（按性能重要性排序）：
// - Selector: 负载均衡选择器（每请求调用，直接影响吞吐量）
// - Metadata: 元数据传递（每请求调用）
// - Errors: 错误处理（业务代码高频使用）
// - Validate: 数据验证（每请求调用）
// - Retry: 重试延迟计算（失败场景调用）
// - Cache: 缓存操作（高频读写场景）
//
// 未包含模块说明：
// - Container: 只在启动时使用，不是性能瓶颈
// - Event: 取决于业务场景，一般不是高频操作
// - CircuitBreaker: 状态切换频率取决于失败率，非高频
package benchmark

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ngq/gorp/framework/contract"
	"github.com/ngq/gorp/framework/provider/selector/p2c"
	"github.com/ngq/gorp/framework/provider/selector/random"
	"github.com/ngq/gorp/framework/provider/selector/wrr"
	"github.com/ngq/gorp/framework/provider/validate"
)

// ============================================================
// Selector 选择器性能测试
// 说明：负载均衡选择器在每请求时调用，性能直接影响吞吐量。
// ============================================================

func BenchmarkRandomSelector_Select(b *testing.B) {
	selector := random.NewRandomSelector()
	scenarios := []int{1, 10, 100, 1000}

	for _, n := range scenarios {
		b.Run(fmt.Sprintf("instances_%d", n), func(b *testing.B) {
			instances := makeInstances(n)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _, _ = selector.Select(context.Background(), instances)
			}
		})
	}
}

func BenchmarkWRRSelector_Select(b *testing.B) {
	selector := wrr.NewWRRSelector()
	scenarios := []int{1, 10, 100, 1000}

	for _, n := range scenarios {
		b.Run(fmt.Sprintf("instances_%d", n), func(b *testing.B) {
			instances := makeInstancesWithWeight(n)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _, _ = selector.Select(context.Background(), instances)
			}
		})
	}
}

func BenchmarkP2CSelector_Select(b *testing.B) {
	selector := p2c.NewP2CSelector()
	scenarios := []int{1, 10, 100, 1000}

	for _, n := range scenarios {
		b.Run(fmt.Sprintf("instances_%d", n), func(b *testing.B) {
			instances := makeInstances(n)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				selected, done, _ := selector.Select(context.Background(), instances)
				done(context.Background(), contract.DoneInfo{Err: nil})
				_ = selected
			}
		})
	}
}

// ============================================================
// Metadata 元数据性能测试
// 说明：元数据在服务间传递时高频调用。
// ============================================================

func BenchmarkMetadata_Get(b *testing.B) {
	md := contract.NewMetadata()
	md.Set("x-request-id", "test-12345")
	md.Set("x-trace-id", "trace-67890")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = md.Get("x-request-id")
	}
}

func BenchmarkMetadata_Set(b *testing.B) {
	md := contract.NewMetadata()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		md.Set("x-key", "value")
	}
}

func BenchmarkMetadata_Clone(b *testing.B) {
	md := contract.NewMetadata()
	for i := 0; i < 10; i++ {
		md.Set(fmt.Sprintf("x-key-%d", i), "value")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = md.Clone()
	}
}

// ============================================================
// Errors 错误处理性能测试
// 说明：错误创建在业务代码中高频使用。
// ============================================================

func BenchmarkNewError(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = contract.NewError(404, contract.ErrorReasonNotFound, "user not found")
	}
}

func BenchmarkError_WithCause(b *testing.B) {
	cause := contract.NewError(500, contract.ErrorReasonInternal, "database error")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = contract.NewError(404, contract.ErrorReasonNotFound, "user not found").WithCause(cause)
	}
}

func BenchmarkFromError(b *testing.B) {
	err := contract.NewError(404, contract.ErrorReasonNotFound, "user not found")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = contract.FromError(err)
	}
}

// ============================================================
// Retry 重试性能测试
// 说明：延迟计算在重试逻辑中调用。
// ============================================================

func BenchmarkCalculateDelay(b *testing.B) {
	policy := contract.RetryPolicy{
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     10 * time.Second,
		Multiplier:   2.0,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = policy.CalculateDelay(i%10, 0.5)
	}
}

// ============================================================
// Validate 验证性能测试
// 说明：数据验证在请求处理时调用。
// ============================================================

type BenchmarkUser struct {
	Username string `json:"username" validate:"required,min=3,max=20"`
	Email    string `json:"email" validate:"required,email"`
	Age      int    `json:"age" validate:"gte=0,lte=150"`
}

func BenchmarkValidator_Validate_Valid(b *testing.B) {
	cfg := &contract.ValidatorConfig{Locale: "zh"}
	svc, _ := validate.NewValidatorService(cfg)

	user := &BenchmarkUser{
		Username: "testuser",
		Email:    "test@example.com",
		Age:      25,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = svc.Validate(context.Background(), user)
	}
}

func BenchmarkValidator_Validate_Invalid(b *testing.B) {
	cfg := &contract.ValidatorConfig{Locale: "zh"}
	svc, _ := validate.NewValidatorService(cfg)

	user := &BenchmarkUser{
		Username: "ab",
		Email:    "invalid",
		Age:      200,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = svc.Validate(context.Background(), user)
	}
}

// ============================================================
// 辅助函数
// ============================================================

func makeInstances(n int) []contract.ServiceInstance {
	instances := make([]contract.ServiceInstance, n)
	for i := 0; i < n; i++ {
		instances[i] = contract.ServiceInstance{
			ID:      fmt.Sprintf("instance-%d", i),
			Name:    "test-service",
			Address: fmt.Sprintf("192.168.1.%d:8080", i%256),
			Healthy: true,
		}
	}
	return instances
}

func makeInstancesWithWeight(n int) []contract.ServiceInstance {
	instances := make([]contract.ServiceInstance, n)
	for i := 0; i < n; i++ {
		instances[i] = contract.ServiceInstance{
			ID:      fmt.Sprintf("instance-%d", i),
			Name:    "test-service",
			Address: fmt.Sprintf("192.168.1.%d:8080", i%256),
			Healthy: true,
			Metadata: map[string]string{
				"weight": fmt.Sprintf("%d", (i%10)+1),
			},
		}
	}
	return instances
}