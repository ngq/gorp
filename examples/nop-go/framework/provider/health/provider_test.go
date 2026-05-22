// Package health_test provides tests for the health checker provider.
package health_test

import (
	"context"
	"errors"
	"testing"
	"time"

	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	"github.com/ngq/gorp/framework/provider/health"
)

func TestHealthChecker_Check(t *testing.T) {
	cfg := observabilitycontract.HealthCheckerConfig{
		ServiceName:        "test-service",
		Version:            "1.0.0",
		Timeout:            5 * time.Second,
		CheckDependencies:  true,
	}

	checker := health.NewHealthChecker(cfg, nil)

	// 添加一个健康的组件检查器
	checker.AddChecker("database", func(ctx context.Context) observabilitycontract.HealthCheckResult {
		return observabilitycontract.HealthCheckResult{
			Status:  observabilitycontract.HealthStatusHealthy,
			Message: "database connection is healthy",
		}
	})

	// 添加一个健康的依赖检查器
	checker.AddDependency("redis", func(ctx context.Context) observabilitycontract.DependencyHealth {
		return observabilitycontract.DependencyHealth{
			Type:    "redis",
			Status:  observabilitycontract.HealthStatusHealthy,
			Message: "redis connection is healthy",
		}
	})

	// 执行健康检查
	report, err := checker.Check(context.Background())
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}

	// 验证报告
	if report.Service != "test-service" {
		t.Errorf("Service = %v, want test-service", report.Service)
	}
	if report.Version != "1.0.0" {
		t.Errorf("Version = %v, want 1.0.0", report.Version)
	}
	if report.Status != observabilitycontract.HealthStatusHealthy {
		t.Errorf("Status = %v, want healthy", report.Status)
	}
	if len(report.Checks) != 2 { // runtime + database
		t.Errorf("Checks count = %v, want 2", len(report.Checks))
	}
	if len(report.Dependencies) != 1 {
		t.Errorf("Dependencies count = %v, want 1", len(report.Dependencies))
	}
}

func TestHealthChecker_CheckWithUnhealthyComponent(t *testing.T) {
	cfg := observabilitycontract.HealthCheckerConfig{
		ServiceName:        "test-service",
		Version:            "1.0.0",
		Timeout:            5 * time.Second,
		CheckDependencies:  false,
	}

	checker := health.NewHealthChecker(cfg, nil)

	// 添加一个不健康的组件检查器
	checker.AddChecker("database", func(ctx context.Context) observabilitycontract.HealthCheckResult {
		return observabilitycontract.HealthCheckResult{
			Status:  observabilitycontract.HealthStatusUnhealthy,
			Message: "database connection failed",
			Error:   "connection refused",
		}
	})

	// 执行健康检查
	report, err := checker.Check(context.Background())
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}

	// 验证整体状态为 unhealthy
	if report.Status != observabilitycontract.HealthStatusUnhealthy {
		t.Errorf("Status = %v, want unhealthy", report.Status)
	}
}

func TestHealthChecker_CheckWithUnhealthyDependency(t *testing.T) {
	cfg := observabilitycontract.HealthCheckerConfig{
		ServiceName:        "test-service",
		Version:            "1.0.0",
		Timeout:            5 * time.Second,
		CheckDependencies:  true,
	}

	checker := health.NewHealthChecker(cfg, nil)

	// 添加一个不健康的依赖检查器
	checker.AddDependency("redis", func(ctx context.Context) observabilitycontract.DependencyHealth {
		return observabilitycontract.DependencyHealth{
			Type:    "redis",
			Status:  observabilitycontract.HealthStatusUnhealthy,
			Message: "redis connection failed",
		}
	})

	// 执行健康检查
	report, err := checker.Check(context.Background())
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}

	// 验证整体状态为 degraded（依赖不健康）
	if report.Status != observabilitycontract.HealthStatusDegraded {
		t.Errorf("Status = %v, want degraded", report.Status)
	}
}

func TestHealthChecker_CheckComponent(t *testing.T) {
	cfg := observabilitycontract.HealthCheckerConfig{
		ServiceName: "test-service",
		Version:     "1.0.0",
		Timeout:     5 * time.Second,
	}

	checker := health.NewHealthChecker(cfg, nil)

	// 添加组件检查器
	checker.AddChecker("custom", func(ctx context.Context) observabilitycontract.HealthCheckResult {
		return observabilitycontract.HealthCheckResult{
			Status:  observabilitycontract.HealthStatusHealthy,
			Message: "custom component is healthy",
		}
	})

	// 检查特定组件
	result, err := checker.CheckComponent(context.Background(), "custom")
	if err != nil {
		t.Fatalf("CheckComponent() error = %v", err)
	}

	if result.Status != observabilitycontract.HealthStatusHealthy {
		t.Errorf("Status = %v, want healthy", result.Status)
	}

	// 检查不存在的组件
	_, err = checker.CheckComponent(context.Background(), "nonexistent")
	if err == nil {
		t.Error("CheckComponent() should return error for nonexistent component")
	}
}

func TestHealthChecker_Timeout(t *testing.T) {
	cfg := observabilitycontract.HealthCheckerConfig{
		ServiceName: "test-service",
		Version:     "1.0.0",
		Timeout:     100 * time.Millisecond,
	}

	checker := health.NewHealthChecker(cfg, nil)

	// 添加一个会超时的检查器
	checker.AddChecker("slow", func(ctx context.Context) observabilitycontract.HealthCheckResult {
		select {
		case <-time.After(200 * time.Millisecond):
			return observabilitycontract.HealthCheckResult{
				Status:  observabilitycontract.HealthStatusHealthy,
				Message: "slow check completed",
			}
		case <-ctx.Done():
			return observabilitycontract.HealthCheckResult{
				Status:  observabilitycontract.HealthStatusUnhealthy,
				Message: "check timed out",
			}
		}
	})

	// 执行健康检查（应该超时）
	report, err := checker.Check(context.Background())
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}

	// 验证检查结果
	if report.Checks["slow"].Status != observabilitycontract.HealthStatusUnhealthy {
		t.Errorf("slow check status = %v, want unhealthy", report.Checks["slow"].Status)
	}
}

func TestDatabaseHealthChecker(t *testing.T) {
	// 测试数据库健康检查器（无容器）
	checker := health.DatabaseHealthChecker(nil, "mysql")

	result := checker(context.Background())

	if result.Name != "mysql" {
		t.Errorf("Name = %v, want mysql", result.Name)
	}
	if result.Type != "database" {
		t.Errorf("Type = %v, want database", result.Type)
	}
	if result.Status != observabilitycontract.HealthStatusDegraded {
		t.Errorf("Status = %v, want degraded", result.Status)
	}
}

func TestRedisHealthChecker(t *testing.T) {
	// 测试 Redis 健康检查器（无容器）
	checker := health.RedisHealthChecker(nil, "redis")

	result := checker(context.Background())

	if result.Name != "redis" {
		t.Errorf("Name = %v, want redis", result.Name)
	}
	if result.Type != "redis" {
		t.Errorf("Type = %v, want redis", result.Type)
	}
	if result.Status != observabilitycontract.HealthStatusDegraded {
		t.Errorf("Status = %v, want degraded", result.Status)
	}
}

func TestCustomDependencyChecker(t *testing.T) {
	// 测试自定义依赖检查器
	checker := health.CustomDependencyChecker("custom-service", "http", func(ctx context.Context) (bool, string) {
		return true, "custom service is healthy"
	})

	result := checker(context.Background())

	if result.Name != "custom-service" {
		t.Errorf("Name = %v, want custom-service", result.Name)
	}
	if result.Type != "http" {
		t.Errorf("Type = %v, want http", result.Type)
	}
	if result.Status != observabilitycontract.HealthStatusHealthy {
		t.Errorf("Status = %v, want healthy", result.Status)
	}

	// 测试不健康的自定义检查器
	unhealthyChecker := health.CustomDependencyChecker("unhealthy-service", "http", func(ctx context.Context) (bool, string) {
		return false, "service is down"
	})

	unhealthyResult := unhealthyChecker(context.Background())
	if unhealthyResult.Status != observabilitycontract.HealthStatusUnhealthy {
		t.Errorf("Status = %v, want unhealthy", unhealthyResult.Status)
	}

	// 测试 nil 检查函数
	nilChecker := health.CustomDependencyChecker("nil-service", "http", nil)
	nilResult := nilChecker(context.Background())
	if nilResult.Status != observabilitycontract.HealthStatusDegraded {
		t.Errorf("Status = %v, want degraded", nilResult.Status)
	}
}

func TestHealthChecker_ConcurrentChecks(t *testing.T) {
	cfg := observabilitycontract.HealthCheckerConfig{
		ServiceName:       "test-service",
		Version:           "1.0.0",
		Timeout:           5 * time.Second,
		CheckDependencies: true,
	}

	checker := health.NewHealthChecker(cfg, nil)

	// 添加多个组件检查器
	for i := 0; i < 10; i++ {
		idx := i
		checker.AddChecker(string(rune('a'+idx)), func(ctx context.Context) observabilitycontract.HealthCheckResult {
			time.Sleep(10 * time.Millisecond) // 模拟检查延迟
			return observabilitycontract.HealthCheckResult{
				Status:  observabilitycontract.HealthStatusHealthy,
				Message: "component is healthy",
			}
		})
	}

	// 添加多个依赖检查器
	for i := 0; i < 5; i++ {
		idx := i
		checker.AddDependency(string(rune('A'+idx)), func(ctx context.Context) observabilitycontract.DependencyHealth {
			time.Sleep(10 * time.Millisecond) // 模拟检查延迟
			return observabilitycontract.DependencyHealth{
				Type:    "service",
				Status:  observabilitycontract.HealthStatusHealthy,
				Message: "dependency is healthy",
			}
		})
	}

	// 执行健康检查
	start := time.Now()
	report, err := checker.Check(context.Background())
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}

	// 验证并发执行（总时间应该远小于串行执行时间）
	// 串行执行需要 15 * 10ms = 150ms，并发执行应该 < 100ms
	if elapsed > 100*time.Millisecond {
		t.Errorf("Check took %v, should be < 100ms (concurrent execution)", elapsed)
	}

	// 验证所有检查都完成
	if len(report.Checks) != 11 { // 10 + runtime
		t.Errorf("Checks count = %v, want 11", len(report.Checks))
	}
	if len(report.Dependencies) != 5 {
		t.Errorf("Dependencies count = %v, want 5", len(report.Dependencies))
	}
}

func TestHealthChecker_StatusCalculation(t *testing.T) {
	tests := []struct {
		name           string
		componentStatus observabilitycontract.HealthStatus
		depStatus      observabilitycontract.HealthStatus
		wantStatus     observabilitycontract.HealthStatus
	}{
		{
			name:           "all healthy",
			componentStatus: observabilitycontract.HealthStatusHealthy,
			depStatus:      observabilitycontract.HealthStatusHealthy,
			wantStatus:     observabilitycontract.HealthStatusHealthy,
		},
		{
			name:           "component unhealthy",
			componentStatus: observabilitycontract.HealthStatusUnhealthy,
			depStatus:      observabilitycontract.HealthStatusHealthy,
			wantStatus:     observabilitycontract.HealthStatusUnhealthy,
		},
		{
			name:           "dependency unhealthy",
			componentStatus: observabilitycontract.HealthStatusHealthy,
			depStatus:      observabilitycontract.HealthStatusUnhealthy,
			wantStatus:     observabilitycontract.HealthStatusDegraded,
		},
		{
			name:           "dependency degraded",
			componentStatus: observabilitycontract.HealthStatusHealthy,
			depStatus:      observabilitycontract.HealthStatusDegraded,
			wantStatus:     observabilitycontract.HealthStatusDegraded,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := observabilitycontract.HealthCheckerConfig{
				ServiceName:       "test-service",
				Version:           "1.0.0",
				CheckDependencies: true,
			}

			checker := health.NewHealthChecker(cfg, nil)

			checker.AddChecker("test-component", func(ctx context.Context) observabilitycontract.HealthCheckResult {
				return observabilitycontract.HealthCheckResult{
					Status: tt.componentStatus,
				}
			})

			checker.AddDependency("test-dep", func(ctx context.Context) observabilitycontract.DependencyHealth {
				return observabilitycontract.DependencyHealth{
					Status: tt.depStatus,
				}
			})

			report, err := checker.Check(context.Background())
			if err != nil {
				t.Fatalf("Check() error = %v", err)
			}

			if report.Status != tt.wantStatus {
				t.Errorf("Status = %v, want %v", report.Status, tt.wantStatus)
			}
		})
	}
}

// MockContainer 用于测试
type mockContainer struct {
	bindings map[string]any
}

func newMockContainer() *mockContainer {
	return &mockContainer{
		bindings: make(map[string]any),
	}
}

func (m *mockContainer) Bind(key string, factory func(runtimecontract.Container) (any, error), singleton bool) {
	// 简化实现
}

func (m *mockContainer) Make(key string) (any, error) {
	if v, ok := m.bindings[key]; ok {
		return v, nil
	}
	return nil, errors.New("not found")
}

func (m *mockContainer) IsBind(key string) bool {
	_, ok := m.bindings[key]
	return ok
}

func (m *mockContainer) Destroy() {}
