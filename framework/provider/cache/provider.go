package cache

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

// Provider registers the unified cache contract.
type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string { return "cache" }

func (p *Provider) IsDefer() bool { return false }

func (p *Provider) Provides() []string { return []string{datacontract.CacheKey} }

type config struct {
	Driver string `mapstructure:"driver"`
}

func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(datacontract.CacheKey, func(c runtimecontract.Container) (any, error) {
		cfgAny, err := c.Make(datacontract.ConfigKey)
		if err != nil {
			return nil, err
		}
		cfg := cfgAny.(datacontract.Config)

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
			rAny, err := c.Make(datacontract.RedisKey)
			if err != nil {
				return nil, err
			}
			r := rAny.(datacontract.Redis)
			d = newRedisCache(r)
		default:
			return nil, fmt.Errorf("invalid cache.driver: %s", driver)
		}

		return &service{d: d}, nil
	}, true)
	return nil
}

func (p *Provider) Boot(runtimecontract.Container) error { return nil }

type cacheDriver interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string, ttl time.Duration) error
	Del(ctx context.Context, key string) error
	MGet(ctx context.Context, keys ...string) (map[string]string, error)
}

type service struct {
	d cacheDriver
}

func (s *service) Get(ctx context.Context, key string) (string, error) {
	return s.d.Get(ctx, key)
}

func (s *service) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	return s.d.Set(ctx, key, value, ttl)
}

func (s *service) Del(ctx context.Context, key string) error {
	return s.d.Del(ctx, key)
}

func (s *service) MGet(ctx context.Context, keys ...string) (map[string]string, error) {
	return s.d.MGet(ctx, keys...)
}

func (s *service) Remember(ctx context.Context, key string, ttl time.Duration, fn func(context.Context) (string, error)) (string, error) {
	v, err := s.d.Get(ctx, key)
	if err == nil {
		return v, nil
	}
	if err != datacontract.ErrCacheMiss {
		return "", err
	}

	computed, err := fn(ctx)
	if err != nil {
		return "", err
	}
	_ = s.d.Set(ctx, key, computed, ttl)
	return computed, nil
}
