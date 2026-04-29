package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/ngq/gorp/framework/contract"

	"github.com/redis/go-redis/v9"
)

// Provider 把 Redis 客户端服务注册进容器。
//
// 中文说明：
// - Redis 常被 cache、session、限流等基础设施复用，因此直接作为独立 provider 暴露。
// - 这里使用单实例 redis.Client，由容器按单例方式缓存，避免重复创建连接池。
// - 已集成 Prometheus 指标收集，通过 /metrics 端点暴露 Redis 命令执行统计。
type Provider struct{}

// NewProvider 创建 redis provider。
func NewProvider() *Provider { return &Provider{} }

// Name 返回 provider 名称。
func (p *Provider) Name() string { return "redis" }

// IsDefer 表示 redis provider 不走延迟加载。
func (p *Provider) IsDefer() bool {
	return false
}

// Provides 返回 redis provider 暴露的能力 key。
func (p *Provider) Provides() []string { return []string{contract.RedisKey} }

// config 对应 redis 配置节点。
//
// 中文说明：
// - Addr 为 host:port。
// - Password 可为空，适用于本地无密码环境。
// - DB 为逻辑库编号，默认由配置中心决定。
type config struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// Register 绑定统一 Redis 服务。
func (p *Provider) Register(c contract.Container) error {
	c.Bind(contract.RedisKey, func(c contract.Container) (any, error) {
		cfgAny, err := c.Make(contract.ConfigKey)
		if err != nil {
			return nil, err
		}
		cfg := cfgAny.(contract.Config)

		// 这里不能直接使用 Unmarshal("redis", &rc)。
		//
		// 中文说明：
		// - viper 的 AutomaticEnv 对嵌套结构反序列化支持有限。
		// - 如果直接整体 Unmarshal，像 REDIS_ADDR 这类环境变量覆盖可能不会生效。
		// - 因此这里逐个读取 redis.addr / redis.password / redis.db，确保配置文件与环境变量都能正确合并。
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

		// 添加 Prometheus 指标收集 hook
		// 中文说明：
		// - 拦截每次 Redis 命令执行，记录命令次数和耗时；
		// - 通过 /metrics 端点暴露给 Prometheus。
		client.AddHook(NewRedisMetricsHook())

		return &service{c: client}, nil
	}, true)
	return nil
}

// Boot redis provider 无额外启动逻辑。
func (p *Provider) Boot(contract.Container) error { return nil }

type service struct {
	c *redis.Client
}

// Ping 用于做 Redis 连通性探测。
func (s *service) Ping(ctx context.Context) error {
	return s.c.Ping(ctx).Err()
}

// Get 读取单个 key。
func (s *service) Get(ctx context.Context, key string) (string, error) {
	return s.c.Get(ctx, key).Result()
}

// Set 写入 key，并把秒级 TTL 转成 go-redis 需要的 time.Duration。
//
// 中文说明：
// - ttlSeconds<=0 时表示不过期。
// - 这里统一在 provider 层做转换，上层 contract.Redis 不必直接暴露 time.Duration。
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

// MGet 批量读取多个 key，并把返回值整理成 key->value map。
//
// 中文说明：
// - go-redis 的 MGet 返回 []interface{}，需要调用方自己按下标解析。
// - 这里转换成 map，减少业务层处理细节。
// - nil 值代表该 key 不存在，会被直接跳过，不会写入结果 map。
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