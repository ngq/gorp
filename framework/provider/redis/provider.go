// Package redis provides Redis service for gorp framework.
// Supports connection configuration, basic operations (Get/Set/Del/MGet).
// Includes Prometheus metrics hook for monitoring.
//
// Redis 包提供 Redis 服务，用于 gorp 框架。
// 支持连接配置、基本操作（Get/Set/Del/MGet）。
// 包含 Prometheus 指标 hook 用于监控。
package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

// Provider registers Redis service.
// Core logic: Read redis config, create client with metrics hook, bind to container.
//
// Provider 注册 Redis 服务。
// 核心逻辑：读取 redis 配置、创建带指标 hook 的客户端、绑定到容器。
type Provider struct{}

// NewProvider creates a new Redis provider.
//
// NewProvider 创建新的 Redis provider。
func NewProvider() *Provider { return &Provider{} }

// Name returns provider name for identification.
//
// Name 返回 provider 名称，用于标识。
func (p *Provider) Name() string { return "redis" }

// IsDefer indicates Redis provider should not defer loading.
// Redis may be needed early for caching/session.
//
// IsDefer 表示 Redis provider 不应延迟加载。
// Redis 可能早期就被用于缓存/session。
func (p *Provider) IsDefer() bool { return false }

// Provides returns the capability keys this provider exposes.
// Exposes RedisKey for Redis service.
//
// Provides 返回 provider 暴露的能力键。
// 暴露 RedisKey 用于 Redis 服务。
func (p *Provider) Provides() []string { return []string{datacontract.RedisKey} }

type config struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// Register binds the Redis factory to the container.
// Core logic: Read config, create Redis client with metrics hook, bind service.
//
// Register 将 Redis 工厂绑定到容器。
// 核心逻辑：读取配置、创建带指标 hook 的 Redis client、绑定服务。
func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(datacontract.RedisKey, func(c runtimecontract.Container) (any, error) {
		cfgAny, err := c.Make(datacontract.ConfigKey)
		if err != nil {
			return nil, err
		}
		cfg := cfgAny.(datacontract.Config)

		rc := config{
			Addr:     cfg.GetString("redis.addr"),
			Password: cfg.GetString("redis.password"),
			DB:       cfg.GetInt("redis.db"),
		}
		if rc.Addr == "" {
			rc.Addr = "127.0.0.1:6379"
		}

		client := redis.NewClient(&redis.Options{
			Addr:     rc.Addr,
			Password: rc.Password,
			DB:       rc.DB,
		})
		client.AddHook(NewRedisMetricsHook())

		return &service{c: client}, nil
	}, true)
	return nil
}

// Boot initializes the Redis provider.
// No additional startup logic required.
//
// Boot 初始化 Redis provider。
// 无需额外启动逻辑。
func (p *Provider) Boot(runtimecontract.Container) error { return nil }

type service struct {
	c *redis.Client
}

func (s *service) Ping(ctx context.Context) error {
	return s.c.Ping(ctx).Err()
}

func (s *service) Get(ctx context.Context, key string) (string, error) {
	return s.c.Get(ctx, key).Result()
}

func (s *service) Set(ctx context.Context, key, value string, ttlSeconds int) error {
	var ttl time.Duration
	if ttlSeconds > 0 {
		ttl = time.Duration(ttlSeconds) * time.Second
	}
	return s.c.Set(ctx, key, value, ttl).Err()
}

func (s *service) Del(ctx context.Context, key string) error {
	return s.c.Del(ctx, key).Err()
}

func (s *service) MGet(ctx context.Context, keys ...string) (map[string]string, error) {
	if len(keys) == 0 {
		return map[string]string{}, nil
	}
	vals, err := s.c.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}
	out := make(map[string]string, len(keys))
	for i, k := range keys {
		if i >= len(vals) {
			break
		}
		if vals[i] == nil {
			continue
		}
		out[k] = fmt.Sprint(vals[i])
	}
	return out, nil
}
