// Application scenarios:
// - Multi-instance shared captcha store backed by Redis.
// - Graceful automatic fallback to in-memory store when Redis is unavailable.
// - High-availability verification with timeout control.
//
// 适用场景：
// - 基于 Redis 的多实例共享验证码存储。
// - Redis 不可用时自动优雅降级为单机内存存储。
// - 带超时控制的高可用验证码校验。
package captcha

import (
	"context"
	"strings"
	"time"

	"github.com/mojocn/base64Captcha"
	datacontract "github.com/ngq/gorp/framework/contract/data"
	securitycontract "github.com/ngq/gorp/framework/contract/security"
)

var (
	_ base64Captcha.Store          = (*RedisCaptchaStore)(nil)
	_ securitycontract.CaptchaStore = (*RedisCaptchaStore)(nil)
)

const (
	defaultPrefix     = "captcha:"
	defaultExpiration = 5 * time.Minute
	defaultTimeout    = 2 * time.Second
)

// Option configures RedisCaptchaStore options.
type Option func(*RedisCaptchaStore)

// WithPrefix configures the Redis key prefix for captcha entries.
func WithPrefix(prefix string) Option {
	return func(s *RedisCaptchaStore) {
		if prefix != "" {
			s.prefix = prefix
		}
	}
}

// WithExpiration configures the captcha TTL.
func WithExpiration(expiration time.Duration) Option {
	return func(s *RedisCaptchaStore) {
		if expiration > 0 {
			s.expiration = expiration
		}
	}
}

// WithTimeout configures the timeout for Redis operations.
func WithTimeout(timeout time.Duration) Option {
	return func(s *RedisCaptchaStore) {
		if timeout > 0 {
			s.timeout = timeout
		}
	}
}

// WithFallback configures the fallback store when Redis is unavailable or errors.
func WithFallback(fallback base64Captcha.Store) Option {
	return func(s *RedisCaptchaStore) {
		if fallback != nil {
			s.fallback = fallback
		}
	}
}

// RedisCaptchaStore implements base64Captcha.Store and securitycontract.CaptchaStore backed by Redis,
// with automatic graceful fallback to an in-memory store.
type RedisCaptchaStore struct {
	redis      datacontract.Redis
	fallback   base64Captcha.Store
	prefix     string
	expiration time.Duration
	timeout    time.Duration
}

// NewRedisCaptchaStore creates a captcha store instance.
// If redis is nil, it automatically delegates to base64Captcha.DefaultMemStore.
func NewRedisCaptchaStore(redis datacontract.Redis, opts ...Option) *RedisCaptchaStore {
	store := &RedisCaptchaStore{
		redis:      redis,
		fallback:   base64Captcha.DefaultMemStore,
		prefix:     defaultPrefix,
		expiration: defaultExpiration,
		timeout:    defaultTimeout,
	}
	for _, opt := range opts {
		opt(store)
	}
	return store
}

// Set stores the captcha value with expiration.
func (s *RedisCaptchaStore) Set(id string, value string) error {
	if s.redis == nil {
		if s.fallback != nil {
			return s.fallback.Set(id, value)
		}
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	err := s.redis.Set(ctx, s.prefix+id, value, s.expiration)
	if err != nil && s.fallback != nil {
		// Redis error fallback
		return s.fallback.Set(id, value)
	}
	return err
}

// Get retrieves the captcha value by id, optionally clearing it.
func (s *RedisCaptchaStore) Get(id string, clear bool) string {
	if s.redis == nil {
		if s.fallback != nil {
			return s.fallback.Get(id, clear)
		}
		return ""
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	val, err := s.redis.Get(ctx, s.prefix+id)
	if err != nil {
		if s.fallback != nil {
			return s.fallback.Get(id, clear)
		}
		return ""
	}

	if clear {
		delCtx, delCancel := context.WithTimeout(context.Background(), s.timeout)
		defer delCancel()
		_ = s.redis.Del(delCtx, s.prefix+id)
	}
	return val
}

// Verify verifies the given captcha answer, optionally clearing the entry.
func (s *RedisCaptchaStore) Verify(id, answer string, clear bool) bool {
	val := s.Get(id, clear)
	if val == "" {
		return false
	}
	return strings.EqualFold(val, answer)
}
