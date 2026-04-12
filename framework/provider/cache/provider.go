package cache

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ngq/gorp/framework/contract"
)

// Provider 把统一缓存服务注册进容器。
//
// 中文说明：
// - cache 是上层抽象，底层可以切到 memory 或 redis。
// - 业务代码只依赖 contract.Cache，避免直接感知底层实现。
// - 这样后续做本地开发、单测替换、线上切换驱动都更方便。
type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string       { return "cache" }
func (p *Provider) IsDefer() bool      { return false }
func (p *Provider) Provides() []string { return []string{contract.CacheKey} }

// config 描述 cache 顶层配置。
//
// 中文说明：
// - 当前只关心 driver。
// - 如果配置文件没写，会继续回退到环境变量 CACHE_DRIVER。
// - 再没有时默认走 redis。
type config struct {
	Driver string `mapstructure:"driver"`
}

func (p *Provider) Register(c contract.Container) error {
	c.Bind(contract.CacheKey, func(c contract.Container) (any, error) {
		cfgAny, err := c.Make(contract.ConfigKey)
		if err != nil {
			return nil, err
		}
		cfg := cfgAny.(contract.Config)

		var cc config
		_ = cfg.Unmarshal("cache", &cc)
		driver := strings.TrimSpace(cc.Driver)
		if driver == "" {
			driver = strings.TrimSpace(os.Getenv("CACHE_DRIVER"))
		}
		if driver == "" {
			driver = "redis"
		}

		var d cacheDriver
		switch strings.ToLower(driver) {
		case "memory", "mem", "inmemory":
			d = newMemoryStore()
		case "redis":
			// redis-backed driver
			rAny, err := c.Make(contract.RedisKey)
			if err != nil {
				return nil, err
			}
			r := rAny.(contract.Redis)
			d = newRedisCache(r)
		default:
			return nil, fmt.Errorf("invalid cache.driver: %s", driver)
		}

		return &service{d: d}, nil
	}, true)
	return nil
}

func (p *Provider) Boot(contract.Container) error { return nil }

// ---- service impl ----

// cacheDriver 定义底层驱动需要满足的最小能力集。
//
// 中文说明：
// - service 只做统一入口与少量组合逻辑，不关心具体数据存在哪里。
// - memoryStore 与 redisCache 都实现这一组方法，因此可以在运行时自由切换。
// - MGet 返回 map 而不是数组，是为了让上层按 key 直接取值，不必自己维护位置对应关系。
type cacheDriver interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string, ttl time.Duration) error
	Del(ctx context.Context, key string) error
	MGet(ctx context.Context, keys ...string) (map[string]string, error)
}

type service struct {
	d cacheDriver
}

// Get 直接透传到底层驱动。
func (s *service) Get(ctx context.Context, key string) (string, error) {
	return s.d.Get(ctx, key)
}

// Set 直接透传到底层驱动。
func (s *service) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	return s.d.Set(ctx, key, value, ttl)
}

// Del 删除指定 key。
func (s *service) Del(ctx context.Context, key string) error {
	return s.d.Del(ctx, key)
}

// MGet 批量读取多个 key。
func (s *service) MGet(ctx context.Context, keys ...string) (map[string]string, error) {
	return s.d.MGet(ctx, keys...)
}

// Remember 是缓存场景中最常用的“先读缓存，未命中再回源”的组合操作。
//
// 核心流程：
// 1. 先尝试读取缓存。
// 2. 命中则直接返回。
// 3. 如果是明确的缓存未命中，则执行 fn 计算真实值。
// 4. 计算完成后 best-effort 回填缓存，再把结果返回给调用方。
//
// 注意：
// - 只有 ErrCacheMiss 才会触发回源；其他缓存错误会直接返回，避免把真实故障误判成未命中。
// - Set 失败不会中断主流程，因为缓存本质上只是性能优化层。
func (s *service) Remember(ctx context.Context, key string, ttl time.Duration, fn func(context.Context) (string, error)) (string, error) {
	v, err := s.d.Get(ctx, key)
	if err == nil {
		return v, nil
	}
	if err != contract.ErrCacheMiss {
		return "", err
	}

	computed, err := fn(ctx)
	if err != nil {
		return "", err
	}
	// best-effort cache set; cache is an optimization
	_ = s.d.Set(ctx, key, computed, ttl)
	return computed, nil
}
