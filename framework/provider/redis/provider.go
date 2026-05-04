package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string { return "redis" }

func (p *Provider) IsDefer() bool { return false }

func (p *Provider) Provides() []string { return []string{datacontract.RedisKey} }

type config struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

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
