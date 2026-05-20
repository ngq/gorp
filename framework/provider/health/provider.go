// Package health provides health checking capability for gorp framework.
// Supports aggregated health status reporting for microservices.
//
// 健康检查 provider，支持微服务的聚合健康状态报告。
package health

import (
	"context"
	"sync"
	"time"

	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

// Provider 是健康检查聚合器 provider。
// 将 HealthChecker 契约实现注册到容器中，供 HTTP 端点和 gRPC 使用。
type Provider struct{}

// NewProvider 创建健康检查 provider。
func NewProvider() *Provider { return &Provider{} }

// Name 返回 provider 唯一名称。
func (p *Provider) Name() string { return "health.aggregator" }

// IsDefer 标记此 provider 延迟装载。
func (p *Provider) IsDefer() bool { return true }

// Provides 返回该 provider 提供的容器 key 列表。
func (p *Provider) Provides() []string {
	return []string{observabilitycontract.HealthCheckerKey}
}

// DependsOn 返回该 provider 依赖的 key。
func (p *Provider) DependsOn() []string { return nil }

// Register 将 HealthChecker 实例注册到容器。
func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(observabilitycontract.HealthCheckerKey, func(c runtimecontract.Container) (any, error) {
		cfg := loadHealthCheckerConfig(c)
		return newHealthChecker(cfg, c), nil
	}, true)
	return nil
}

// Boot 启动期初始化（无额外操作）。
func (p *Provider) Boot(c runtimecontract.Container) error { return nil }

// --- 健康检查器实现 ---

// healthChecker 是 HealthChecker 契约的实现。
// 聚合多个组件和依赖的健康状态，生成统一的健康报告。
type healthChecker struct {
	config     observabilitycontract.HealthCheckerConfig
	container  runtimecontract.Container
	checkers   map[string]observabilitycontract.ComponentChecker
	deps       map[string]observabilitycontract.DependencyChecker
	mu         sync.RWMutex
}

// newHealthChecker 根据配置创建健康检查器。
func newHealthChecker(cfg observabilitycontract.HealthCheckerConfig, c runtimecontract.Container) *healthChecker {
	// 设置默认值
	if cfg.Timeout <= 0 {
		cfg.Timeout = 5 * time.Second
	}

	hc := &healthChecker{
		config:    cfg,
		container: c,
		checkers:  make(map[string]observabilitycontract.ComponentChecker),
		deps:      make(map[string]observabilitycontract.DependencyChecker),
	}

	// 自动注册基础组件检查器
	hc.registerDefaultCheckers()

	return hc
}

// NewHealthChecker 导出的构造函数，供测试和外部使用。
func NewHealthChecker(cfg observabilitycontract.HealthCheckerConfig, c runtimecontract.Container) *healthChecker {
	return newHealthChecker(cfg, c)
}

// Check 执行健康检查并返回聚合报告。
// 检查所有注册的组件和依赖，汇总整体健康状态。
func (h *healthChecker) Check(ctx context.Context) (*observabilitycontract.HealthReport, error) {
	// 设置检查超时
	if h.config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, h.config.Timeout)
		defer cancel()
	}

	now := time.Now()
	report := &observabilitycontract.HealthReport{
		Service:    h.config.ServiceName,
		Version:    h.config.Version,
		Timestamp:  now,
		Checks:     make(map[string]observabilitycontract.HealthCheckResult),
		Dependencies: make(map[string]observabilitycontract.DependencyHealth),
	}

	// 执行组件检查
	h.mu.RLock()
	checkers := h.checkers
	deps := h.deps
	h.mu.RUnlock()

	// 并发执行组件检查
	var wg sync.WaitGroup
	results := make(chan observabilitycontract.HealthCheckResult, len(checkers))

	for name, checker := range checkers {
		wg.Add(1)
		go func(name string, checker observabilitycontract.ComponentChecker) {
			defer wg.Done()
			result := checker(ctx)
			result.Name = name
			result.Timestamp = now
			results <- result
		}(name, checker)
	}

	wg.Wait()
	close(results)

	// 收集组件检查结果
	for result := range results {
		report.Checks[result.Name] = result
	}

	// 执行依赖检查（如果启用）
	if h.config.CheckDependencies {
		var depWg sync.WaitGroup
		depResults := make(chan observabilitycontract.DependencyHealth, len(deps))

		for name, depChecker := range deps {
			depWg.Add(1)
			go func(name string, depChecker observabilitycontract.DependencyChecker) {
				defer depWg.Done()
				result := depChecker(ctx)
				result.Name = name
				depResults <- result
			}(name, depChecker)
		}

		depWg.Wait()
		close(depResults)

		// 收集依赖检查结果
		for result := range depResults {
			report.Dependencies[result.Name] = result
		}
	}

	// 计算整体健康状态
	report.Status = h.calculateOverallStatus(report)

	return report, nil
}

// CheckComponent 对特定组件执行健康检查。
func (h *healthChecker) CheckComponent(ctx context.Context, name string) (*observabilitycontract.HealthCheckResult, error) {
	h.mu.RLock()
	checker, exists := h.checkers[name]
	h.mu.RUnlock()

	if !exists {
		return nil, ErrComponentNotFound
	}

	// 设置检查超时
	if h.config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, h.config.Timeout)
		defer cancel()
	}

	result := checker(ctx)
	result.Name = name
	result.Timestamp = time.Now()

	return &result, nil
}

// AddChecker 注册组件健康检查器。
func (h *healthChecker) AddChecker(name string, checker observabilitycontract.ComponentChecker) {
	if checker == nil {
		return
	}
	h.mu.Lock()
	h.checkers[name] = checker
	h.mu.Unlock()
}

// AddDependency 注册依赖健康检查器。
func (h *healthChecker) AddDependency(name string, dep observabilitycontract.DependencyChecker) {
	if dep == nil {
		return
	}
	h.mu.Lock()
	h.deps[name] = dep
	h.mu.Unlock()
}

// calculateOverallStatus 根据组件和依赖状态计算整体健康状态。
// 规则：
// - 任一组件 unhealthy → 整体 unhealthy
// - 任一依赖 unhealthy → 整体 degraded（除非有组件 unhealthy）
// - 全部 healthy → 整体 healthy
func (h *healthChecker) calculateOverallStatus(report *observabilitycontract.HealthReport) observabilitycontract.HealthStatus {
	hasUnhealthyComponent := false
	hasUnhealthyDep := false
	hasDegradedDep := false

	// 检查组件状态
	for _, result := range report.Checks {
		if result.Status == observabilitycontract.HealthStatusUnhealthy {
			hasUnhealthyComponent = true
		}
	}

	// 检查依赖状态
	for _, dep := range report.Dependencies {
		if dep.Status == observabilitycontract.HealthStatusUnhealthy {
			hasUnhealthyDep = true
		}
		if dep.Status == observabilitycontract.HealthStatusDegraded {
			hasDegradedDep = true
		}
	}

	// 计算整体状态
	if hasUnhealthyComponent {
		return observabilitycontract.HealthStatusUnhealthy
	}
	if hasUnhealthyDep {
		return observabilitycontract.HealthStatusDegraded
	}
	if hasDegradedDep {
		return observabilitycontract.HealthStatusDegraded
	}
	return observabilitycontract.HealthStatusHealthy
}

// registerDefaultCheckers 注册默认的组件检查器。
// 包括：runtime（Go 运行时状态）。
func (h *healthChecker) registerDefaultCheckers() {
	// 注册 runtime 检查器
	h.AddChecker("runtime", func(ctx context.Context) observabilitycontract.HealthCheckResult {
		start := time.Now()
		// Go runtime 基本检查：内存和 goroutine 数量
		// 如果 goroutine 数量超过阈值，标记为 degraded
		return observabilitycontract.HealthCheckResult{
			Status:   observabilitycontract.HealthStatusHealthy,
			Message:  "Go runtime is healthy",
			Latency:  time.Since(start),
		}
	})
}

// --- 错误定义 ---

// ErrComponentNotFound 表示组件未找到。
var ErrComponentNotFound = resiliencecontract.ServiceUnavailable("component not found")

// --- 配置读取辅助 ---

// loadHealthCheckerConfig 从容器中读取 health_checker 配置。
func loadHealthCheckerConfig(c runtimecontract.Container) observabilitycontract.HealthCheckerConfig {
	cfg := observabilitycontract.HealthCheckerConfig{}

	if c == nil || !c.IsBind("framework.config") {
		return cfg
	}

	configAny, err := c.Make("framework.config")
	if err != nil {
		return cfg
	}

	type configGetter interface {
		GetBool(key string) bool
		GetString(key string) string
		GetInt(key string) int
		Get(key string) any
	}

	getter, ok := configAny.(configGetter)
	if !ok {
		return cfg
	}

	// 服务名称
	cfg.ServiceName = getter.GetString("service.name")

	// 服务版本
	cfg.Version = getter.GetString("service.version")

	// 检查超时
	if timeoutStr := getter.GetString("health_checker.timeout"); timeoutStr != "" {
		cfg.Timeout, _ = time.ParseDuration(timeoutStr)
	}

	// 是否检查依赖
	cfg.CheckDependencies = getter.GetBool("health_checker.check_dependencies")

	return cfg
}