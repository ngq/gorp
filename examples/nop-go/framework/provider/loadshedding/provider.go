// Package loadshedding 提供过载保护（LoadShedding）实现。
//
// 提供两种策略：
// - semaphore：固定并发数控制（默认）
// - bbr：基于 CPU + inflight 的自适应过载保护
//
// 适用场景：
// - 微服务模式下默认启用，提供 HTTP/gRPC/RPC 统一的过载保护能力；
// - 当系统过载时立即拒绝请求，防止级联故障。
//
// 配置示例：
//
//	load_shedding:
//	  enabled: true
//	  strategy: bbr  # 或 semaphore
//	  bbr:
//	    cpu_threshold: 0.8
//	    window_size: 10s
//	    bucket_count: 100
package loadshedding

import (
	"context"
	"runtime"
	"sync"
	"time"

	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"

	"github.com/ngq/gorp/framework/provider/loadshedding/bbr"
)

// Provider 是 LoadShedder 能力 provider。
// 将 LoadShedding 契约实现注册到容器中，供 HTTP 中间件和 RPC 客户端使用。
//
// 中文说明：
// - 根据 strategy 配置选择 semaphore 或 bbr 实现；
// - semaphore：固定并发数控制，默认最大并发 = GOMAXPROCS * 100；
// - bbr：自适应过载保护，根据 CPU + inflight 动态调整。
type Provider struct{}

// NewProvider 创建 LoadShedding provider。
func NewProvider() *Provider { return &Provider{} }

// Name 返回 provider 唯一名称。
func (p *Provider) Name() string { return "loadshedding.provider" }

// IsDefer 标记此 provider 延迟装载。
func (p *Provider) IsDefer() bool { return true }

// Provides 返回该 provider 提供的容器 key 列表。
func (p *Provider) Provides() []string {
	return []string{resiliencecontract.LoadShedderKey}
}

// DependsOn returns the keys this provider depends on.
// LoadShedding provider has no dependencies.
//
// DependsOn 返回该 provider 依赖的 key。
// LoadShedding provider 无依赖。
func (p *Provider) DependsOn() []string { return nil }

// Requires 返回该 provider 依赖的容器 key 列表（无外部依赖）。
func (p *Provider) Requires() []string { return nil }

// Register 将 LoadShedder 实例注册到容器。
//
// 根据 strategy 配置选择实现：
// - "semaphore" 或空：固定并发数控制
// - "bbr"：自适应过载保护
func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(resiliencecontract.LoadShedderKey, func(c runtimecontract.Container) (any, error) {
		cfg := loadSheddingConfigFromContainer(c)

		// 根据策略选择实现
		switch cfg.Strategy {
		case "bbr":
			bbrCfg := loadBBRConfigFromContainer(c)
			return bbr.NewLoadShedder(bbrCfg), nil
		default:
			// 默认使用信号量实现
			return newSemaphoreLoadShedder(cfg), nil
		}
	}, true)
	return nil
}

// Boot 启动期初始化（无额外操作）。
func (p *Provider) Boot(c runtimecontract.Container) error { return nil }

// --- 信号量式 LoadShedder 实现 ---

// semaphoreLoadShedder 是 LoadShedder 契约的信号量实现。
// 每个资源（resource）维护独立的加权信号量，并发已满时 TryAcquire 失败并返回过载错误。
type semaphoreLoadShedder struct {
	semaphores    sync.Map // resource string -> *semaphoreEntry
	defaultMaxCon int      // 默认最大并发数
	config        resiliencecontract.LoadSheddingConfig
}

// newSemaphoreLoadShedder 根据配置创建信号量式 LoadShedder。
func newSemaphoreLoadShedder(cfg resiliencecontract.LoadSheddingConfig) *semaphoreLoadShedder {
	defaultMaxCon := cfg.MaxConcurrency
	if defaultMaxCon <= 0 {
		// 默认值：GOMAXPROCS * 100，约 100-800 并发
		defaultMaxCon = runtime.GOMAXPROCS(0) * 100
	}
	return &semaphoreLoadShedder{
		defaultMaxCon: defaultMaxCon,
		config:        cfg,
	}
}

// Allow 尝试获取一个并发槽位。如果当前并发已满，立即返回过载错误。
//
// 中文说明：
// - 根据 resource 查找或创建对应的信号量；
// - 使用 TryAcquire 非阻塞获取，失败则返回 ErrLoadShedded；
// - 支持按资源粒度的独立 MaxConcurrency 配置。
func (s *semaphoreLoadShedder) Allow(ctx context.Context, resource string) error {
	entry := s.getOrCreateEntry(resource)
	if !entry.tryAcquire() {
		return ErrLoadShedded
	}
	return nil
}

// Done 释放一个之前占用的并发槽位。
//
// 中文说明：
// - 根据 resource 查找对应的信号量并释放；
// - 即使 err != nil 也正常释放（过载保护不区分成功/失败）。
func (s *semaphoreLoadShedder) Done(ctx context.Context, resource string, err error) {
	entry := s.getOrCreateEntry(resource)
	entry.release()
}

// getOrCreateEntry 获取或创建资源对应的信号量条目。
// 使用 sync.Map + double-check 保证并发安全且不重复创建。
func (s *semaphoreLoadShedder) getOrCreateEntry(resource string) *semaphoreEntry {
	// 快速路径：已有条目直接返回
	if v, ok := s.semaphores.Load(resource); ok {
		return v.(*semaphoreEntry)
	}

	// 确定该资源的最大并发数
	maxCon := s.defaultMaxCon
	if policy, ok := s.config.ResourcePolicies[resource]; ok && policy.MaxConcurrency > 0 {
		maxCon = policy.MaxConcurrency
	}

	// 慢路径：创建新条目
	entry := newSemaphoreEntry(maxCon)
	actual, _ := s.semaphores.LoadOrStore(resource, entry)
	return actual.(*semaphoreEntry)
}

// --- 信号量条目 ---

// semaphoreEntry 封装一个加权信号量和对应的最大并发数。
type semaphoreEntry struct {
	sem    chan struct{} // 用 buffered channel 模拟信号量，比 semaphore.Weighted 更轻量
	maxCon int
}

// newSemaphoreEntry 创建指定容量的信号量条目。
func newSemaphoreEntry(maxCon int) *semaphoreEntry {
	return &semaphoreEntry{
		sem:    make(chan struct{}, maxCon),
		maxCon: maxCon,
	}
}

// tryAcquire 非阻塞地尝试获取一个并发槽位。
func (e *semaphoreEntry) tryAcquire() bool {
	select {
	case e.sem <- struct{}{}:
		return true
	default:
		return false
	}
}

// release 释放一个并发槽位。
func (e *semaphoreEntry) release() {
	select {
	case <-e.sem:
	default:
		// 防御性编程：避免在无人持有时 panic
	}
}

// --- 错误定义 ---

// ErrLoadShedded 表示请求因过载保护被丢弃。
var ErrLoadShedded = resiliencecontract.ServiceUnavailable("server is busy: load shedding active")

// --- 配置读取辅助 ---

// loadSheddingConfigFromContainer 从容器中读取 load_shedding 配置并构建 LoadSheddingConfig。
func loadSheddingConfigFromContainer(c runtimecontract.Container) resiliencecontract.LoadSheddingConfig {
	cfg := resiliencecontract.LoadSheddingConfig{}

	// 安全地尝试读取配置
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

	// 全局开关
	cfg.Enabled = getter.GetBool("load_shedding.enabled")

	// 策略类型
	if strategy := getter.GetString("load_shedding.strategy"); strategy != "" {
		cfg.Strategy = strategy
	}

	// 全局最大并发数
	cfg.MaxConcurrency = getter.GetInt("load_shedding.max_concurrency")

	// 全局默认策略
	cfg.DefaultPolicy = resiliencecontract.LoadSheddingPolicy{
		Enabled:        cfg.Enabled,
		Strategy:       cfg.Strategy,
		MaxConcurrency: cfg.MaxConcurrency,
	}

	// 按资源粒度的策略配置
	// 配置格式：load_shedding.resource_policies.<resource>.max_concurrency
	cfg.ResourcePolicies = loadResourcePolicies(getter)

	return cfg
}

// loadBBRConfigFromContainer 从容器中读取 BBR 策略配置。
func loadBBRConfigFromContainer(c runtimecontract.Container) *bbr.Config {
	cfg := bbr.DefaultConfig()

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
		GetFloat64(key string) float64
		Get(key string) any
	}

	getter, ok := configAny.(configGetter)
	if !ok {
		return cfg
	}

	// BBR 特定配置
	if v := getter.GetFloat64("load_shedding.bbr.cpu_threshold"); v > 0 {
		cfg.CPUThreshold = v
	}
	if v := getter.GetString("load_shedding.bbr.window_size"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.WindowSize = d
		}
	}
	if v := getter.GetInt("load_shedding.bbr.bucket_count"); v > 0 {
		cfg.BucketCount = v
	}
	if v := getter.GetString("load_shedding.bbr.cool_down"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.CoolDown = d
		}
	}
	if v := getter.GetString("load_shedding.bbr.min_rt_threshold"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.MinRTThreshold = d
		}
	}

	return cfg
}

// loadResourcePolicies 从配置中读取按资源粒度的过载保护策略。
func loadResourcePolicies(getter interface {
	Get(key string) any
	GetString(key string) string
	GetInt(key string) int
	GetBool(key string) bool
}) map[string]resiliencecontract.LoadSheddingPolicy {
	policies := make(map[string]resiliencecontract.LoadSheddingPolicy)

	// 尝试读取 resource_policies 作为 map
	raw := getter.Get("load_shedding.resource_policies")
	if raw == nil {
		return policies
	}

	// 支持 map[string]map[string]any 格式
	if m, ok := raw.(map[string]any); ok {
		for resource, policyRaw := range m {
			if pm, ok := policyRaw.(map[string]any); ok {
				policy := resiliencecontract.LoadSheddingPolicy{}
				if v, ok := pm["enabled"].(bool); ok {
					policy.Enabled = v
				}
				if v, ok := pm["strategy"].(string); ok {
					policy.Strategy = v
				}
				if v, ok := pm["max_concurrency"].(int); ok {
					policy.MaxConcurrency = v
				}
				// float64 是 JSON 反序列化后的常见类型
				if v, ok := pm["max_concurrency"].(float64); ok && v > 0 {
					policy.MaxConcurrency = int(v)
				}
				policies[resource] = policy
			}
		}
	}

	return policies
}
