package captcha

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/mojocn/base64Captcha"
	datacontract "github.com/ngq/gorp/framework/contract/data"
)

// stubRedis implements datacontract.Redis in-memory for testing.
type stubRedis struct {
	datacontract.Redis
	mu    sync.RWMutex
	data  map[string]string
	errOn bool
}

func newStubRedis() *stubRedis {
	return &stubRedis{data: make(map[string]string)}
}

func (s *stubRedis) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.errOn {
		return errors.New("redis connection down")
	}
	s.data[key] = value
	return nil
}

func (s *stubRedis) Get(ctx context.Context, key string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.errOn {
		return "", errors.New("redis connection down")
	}
	val, ok := s.data[key]
	if !ok {
		return "", errors.New("key not found")
	}
	return val, nil
}

func (s *stubRedis) Del(ctx context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.errOn {
		return errors.New("redis connection down")
	}
	delete(s.data, key)
	return nil
}

func TestNewRedisCaptchaStore_NilRedisFallback(t *testing.T) {
	memFallback := base64Captcha.NewMemoryStore(10, 5*time.Minute)
	store := NewRedisCaptchaStore(nil, WithFallback(memFallback))

	id := "test-id-1"
	ans := "X9yz"

	if err := store.Set(id, ans); err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}

	// Verify case-insensitive match without clear
	if !store.Verify(id, "x9YZ", false) {
		t.Fatalf("expected verify success")
	}

	// Value should still exist
	val := store.Get(id, true)
	if val != ans {
		t.Fatalf("expected %s, got %s", ans, val)
	}

	// Value should now be cleared
	valAfter := store.Get(id, false)
	if valAfter != "" {
		t.Fatalf("expected empty after clear, got %s", valAfter)
	}
}

func TestRedisCaptchaStore_WithStubRedis(t *testing.T) {
	redis := newStubRedis()
	store := NewRedisCaptchaStore(
		redis,
		WithPrefix("test_cap:"),
		WithExpiration(3*time.Minute),
		WithTimeout(500*time.Millisecond),
	)

	id := "captcha-2"
	ans := "4321"

	if err := store.Set(id, ans); err != nil {
		t.Fatalf("failed to set captcha: %v", err)
	}

	// Check underlying redis key has prefix
	redis.mu.RLock()
	rawVal, ok := redis.data["test_cap:"+id]
	redis.mu.RUnlock()
	if !ok || rawVal != ans {
		t.Fatalf("expected redis to have raw key 'test_cap:%s', got ok=%v, val=%v", id, ok, rawVal)
	}

	// Verify with clear = true
	if !store.Verify(id, ans, true) {
		t.Fatalf("expected verify to pass")
	}

	// Verify it was cleared
	if store.Get(id, false) != "" {
		t.Fatalf("expected captcha to be deleted from redis")
	}
}

func TestRedisCaptchaStore_RedisErrorFallback(t *testing.T) {
	redis := newStubRedis()
	memFallback := base64Captcha.NewMemoryStore(10, 5*time.Minute)
	store := NewRedisCaptchaStore(redis, WithFallback(memFallback))

	// Enable error on redis
	redis.mu.Lock()
	redis.errOn = true
	redis.mu.Unlock()

	id := "fallback-id"
	ans := "8888"

	// Set should fallback to memory store without failing
	if err := store.Set(id, ans); err != nil {
		t.Fatalf("expected fallback set to succeed, got %v", err)
	}

	// Get should fallback to memory store
	if store.Get(id, false) != ans {
		t.Fatalf("expected fallback get to return %s", ans)
	}

	// Verify should succeed via fallback
	if !store.Verify(id, ans, true) {
		t.Fatalf("expected fallback verify to succeed")
	}
}

func TestRedisCaptchaStore_ConcurrentRace(t *testing.T) {
	redis := newStubRedis()
	store := NewRedisCaptchaStore(redis)

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			id := "concurrent-id-" + string(rune('a'+idx))
			_ = store.Set(id, "1234")
			_ = store.Get(id, false)
			_ = store.Verify(id, "1234", true)
		}(i)
	}
	wg.Wait()
}
