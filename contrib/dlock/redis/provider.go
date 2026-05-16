// Package redis provides Redis-based distributed lock implementation for gorp.
//
// Redis 分布式锁 Provider，实现 datacontract.DistributedLock 契约。
// 支持基于 Redis 的分布式锁获取、释放、续期。
//
// 使用示例：
//
//  cfg := &LockConfig{
//      RedisAddr: "localhost:6379",
//  }
//  lock, err := NewDistributedLock(cfg)
//  if err != nil {
//      panic(err)
//  }
//
//  acquired, err := lock.Acquire(ctx, "my-resource", 10*time.Second)
//  if acquired {
//      defer lock.Release(ctx, "my-resource")
//      // 执行临界区操作
//  }
//
// 配置路径：dlock.redis.*
package redis

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string  { return "dlock.redis" }
func (p *Provider) IsDefer() bool { return true }
func (p *Provider) Provides() []string {
	return []string{datacontract.DistributedLockKey}
}

// DependsOn returns the keys this provider depends on.
// Redis dlock depends on Config for Redis configuration.
//
// DependsOn 返回该 provider 依赖的 key。
// Redis dlock 依赖 Config 获取 Redis 配置。
func (p *Provider) DependsOn() []string { return []string{datacontract.ConfigKey} }

func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(datacontract.DistributedLockKey, func(c runtimecontract.Container) (any, error) {
		cfg, err := getLockConfig(c)
		if err != nil {
			return nil, err
		}
		lock, err := NewLock(cfg)
		if err != nil {
			return nil, err
		}
		// Register closer to stop watchdog goroutines and close Redis client on container destroy.
		c.RegisterCloser(datacontract.DistributedLockKey, lock)
		return lock, nil
	}, true)
	return nil
}

func (p *Provider) Boot(c runtimecontract.Container) error { return nil }

func getLockConfig(c runtimecontract.Container) (*datacontract.DistributedLockConfig, error) {
	cfgAny, err := c.Make(datacontract.ConfigKey)
	if err != nil {
		return nil, err
	}
	cfg, ok := cfgAny.(datacontract.Config)
	if !ok {
		return nil, errors.New("dlock: invalid config service")
	}
	lockCfg := &datacontract.DistributedLockConfig{
		Type:          "redis",
		DefaultTTL:    30 * time.Second,
		RetryInterval: 100 * time.Millisecond,
		MaxRetry:      50,
		KeyPrefix:     "lock:",
	}
	if addr := cfg.GetString("distributed_lock.redis.addr"); addr != "" {
		lockCfg.RedisAddr = addr
	} else {
		lockCfg.RedisAddr = "localhost:6379"
	}
	if password := cfg.GetString("distributed_lock.redis.password"); password != "" {
		lockCfg.RedisPassword = password
	}
	if db := cfg.GetInt("distributed_lock.redis.db"); db > 0 {
		lockCfg.RedisDB = db
	}
	if prefix := cfg.GetString("distributed_lock.key_prefix"); prefix != "" {
		lockCfg.KeyPrefix = prefix
	} else if prefix := cfg.GetString("dlock.key_prefix"); prefix != "" {
		lockCfg.KeyPrefix = prefix
	}
	return lockCfg, nil
}

type Lock struct {
	cfg       *datacontract.DistributedLockConfig
	client    *redis.Client
	heldLocks sync.Map
	watchdog  sync.Map
	closed    bool
	mu        sync.Mutex
}

type heldLock struct {
	value string
	token string
	held  time.Time
}

func NewLock(cfg *datacontract.DistributedLockConfig) (*Lock, error) {
	client := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr, Password: cfg.RedisPassword, DB: cfg.RedisDB})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("dlock.redis: connect failed: %w", err)
	}
	return &Lock{cfg: cfg, client: client}, nil
}

func (l *Lock) Lock(ctx context.Context, key string, ttl time.Duration) error {
	fullKey := l.cfg.KeyPrefix + key
	token := generateToken()
	for i := 0; i < l.cfg.MaxRetry; i++ {
		ok, err := l.tryLock(ctx, fullKey, token, ttl)
		if err != nil {
			return err
		}
		if ok {
			l.heldLocks.Store(fullKey, &heldLock{value: fullKey, token: token, held: time.Now()})
			l.startWatchdog(fullKey, ttl)
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(l.cfg.RetryInterval):
		}
	}
	return errors.New("dlock: failed to acquire lock after max retries")
}

func (l *Lock) TryLock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	fullKey := l.cfg.KeyPrefix + key
	token := generateToken()
	ok, err := l.tryLock(ctx, fullKey, token, ttl)
	if err != nil {
		return false, err
	}
	if ok {
		l.heldLocks.Store(fullKey, &heldLock{value: fullKey, token: token, held: time.Now()})
		l.startWatchdog(fullKey, ttl)
	}
	return ok, nil
}

func (l *Lock) tryLock(ctx context.Context, key string, token string, ttl time.Duration) (bool, error) {
	ok, err := l.client.SetNX(ctx, key, token, ttl).Result()
	if err != nil {
		return false, fmt.Errorf("dlock: setnx failed: %w", err)
	}
	return ok, nil
}

func (l *Lock) Unlock(ctx context.Context, key string) error {
	fullKey := l.cfg.KeyPrefix + key
	held, ok := l.heldLocks.Load(fullKey)
	if !ok {
		return errors.New("dlock: lock not held")
	}
	// Delete from heldLocks first to prevent watchdog from renewing an unlocked key.
	l.heldLocks.Delete(fullKey)
	l.stopWatchdog(fullKey)
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`
	result, err := l.client.Eval(ctx, script, []string{fullKey}, held.(*heldLock).token).Result()
	if err != nil {
		return fmt.Errorf("dlock: unlock failed: %w", err)
	}
	if n, ok := result.(int64); !ok || n == 0 {
		return errors.New("dlock: lock not owned")
	}
	return nil
}

func (l *Lock) Renew(ctx context.Context, key string, ttl time.Duration) error {
	fullKey := l.cfg.KeyPrefix + key
	held, ok := l.heldLocks.Load(fullKey)
	if !ok {
		return errors.New("dlock: lock not held")
	}
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("expire", KEYS[1], ARGV[2])
		else
			return 0
		end
	`
	result, err := l.client.Eval(ctx, script, []string{fullKey}, held.(*heldLock).token, int(ttl/time.Second)).Result()
	if err != nil {
		return fmt.Errorf("dlock: renew failed: %w", err)
	}
	if n, ok := result.(int64); !ok || n == 0 {
		return errors.New("dlock: lock not owned")
	}
	return nil
}

func (l *Lock) IsLocked(ctx context.Context, key string) (bool, error) {
	if l == nil || l.client == nil {
		return false, errors.New("dlock: client not initialized")
	}
	fullKey := l.cfg.KeyPrefix + key
	result, err := l.client.Exists(ctx, fullKey).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

func (l *Lock) WithLock(ctx context.Context, key string, ttl time.Duration, fn func() error) error {
	if err := l.Lock(ctx, key, ttl); err != nil {
		return err
	}
	defer l.Unlock(ctx, key)
	return fn()
}

func (l *Lock) startWatchdog(fullKey string, ttl time.Duration) {
	interval := ttl / 3
	ctx, cancel := context.WithCancel(context.Background())
	l.watchdog.Store(fullKey, cancel)
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		consecutiveFailures := 0
		const maxConsecutiveFailures = 3
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				renewCtx, done := context.WithTimeout(context.Background(), time.Second)
				lockKey := strings.TrimPrefix(fullKey, l.cfg.KeyPrefix)
				if err := l.Renew(renewCtx, lockKey, ttl); err != nil {
					consecutiveFailures++
					fmt.Printf("[dlock:warn] watchdog renew failed for key %s (attempt %d/%d): %v\n",
						fullKey, consecutiveFailures, maxConsecutiveFailures, err)
					if consecutiveFailures >= maxConsecutiveFailures {
						fmt.Printf("[dlock:error] watchdog giving up on key %s after %d consecutive failures; lock may expire\n",
							fullKey, maxConsecutiveFailures)
						done()
						return
					}
				} else {
					consecutiveFailures = 0
				}
				done()
			}
		}
	}()
}

func (l *Lock) stopWatchdog(key string) {
	if cancel, ok := l.watchdog.Load(key); ok {
		cancel.(context.CancelFunc)()
		l.watchdog.Delete(key)
	}
}

func generateToken() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// Close stops all watchdog goroutines and closes the Redis client.
// Implements io.Closer for container lifecycle management.
//
// Close 停止所有 watchdog goroutine 并关闭 Redis 客户端。
// 实现 io.Closer 供容器生命周期管理。
func (l *Lock) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed {
		return nil
	}
	l.closed = true

	// Stop all watchdog goroutines
	l.watchdog.Range(func(key, value any) bool {
		if cancel, ok := value.(context.CancelFunc); ok {
			cancel()
		}
		l.watchdog.Delete(key)
		return true
	})

	// Close Redis client
	if l.client != nil {
		return l.client.Close()
	}
	return nil
}
