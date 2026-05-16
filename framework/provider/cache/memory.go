// Package cache provides in-memory cache implementation.
// The memory cache uses a simple map with TTL support. It is thread-safe via RWMutex.
// Note: Memory cache is NOT suitable for distributed systems or production environments
// requiring data persistence. Use Redis cache driver instead.
//
// 本文件提供内存缓存实现，适用于单机开发测试或无需持久化的场景。
// 内存缓存使用简单的 map 实现，支持 TTL 过期，通过 RWMutex 保证线程安全。
// 注意：内存缓存不适用于分布式系统或需要数据持久化的生产环境，请使用 Redis 驱动。
package cache

import (
	"context"
	"sync"
	"time"

	datacontract "github.com/ngq/gorp/framework/contract/data"
)

// memoryStore implements cacheDriver using an in-memory map with TTL support.
//
// memoryStore 使用内存 map 实现缓存驱动，支持 TTL 过期。
type memoryStore struct {
	mu sync.RWMutex // mu protects concurrent access to the map.
	//
	// mu 保护并发访问。
	m map[string]memItem // m stores cache items.
	//
	// m 存储缓存项。
	stopCh chan struct{} // stopCh signals the cleanup goroutine to stop.
	//
	// stopCh 通知清理 goroutine 停止。
}

// memItem represents a cached item with value and expiration time.
//
// memItem 表示缓存项，包含值和过期时间。
type memItem struct {
	v string // v is the cached value.
	//
	// v 缓存值。
	expired time.Time // expired is the TTL expiration time.
	//
	// expired TTL 过期时间。
}

// newMemoryStore creates a new in-memory cache store.
// Starts a background goroutine to periodically clean up expired keys.
//
// newMemoryStore 创建新的内存缓存存储实例。
// 启动后台 goroutine 定期清理过期键。
func newMemoryStore() *memoryStore {
	s := &memoryStore{
		m:      make(map[string]memItem),
		stopCh: make(chan struct{}),
	}
	go s.cleanup()
	return s
}

// Close stops the cleanup goroutine. Implements io.Closer.
//
// Close 停止清理 goroutine。实现 io.Closer。
func (s *memoryStore) Close() error {
	select {
	case <-s.stopCh:
		// Already closed.
	default:
		close(s.stopCh)
	}
	return nil
}

// cleanup periodically removes expired keys from the store.
//
// cleanup 定期清理存储中的过期键。
func (s *memoryStore) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.mu.Lock()
			now := time.Now()
			for k, item := range s.m {
				if !item.expired.IsZero() && now.After(item.expired) {
					delete(s.m, k)
				}
			}
			s.mu.Unlock()
		}
	}
}

// Get retrieves a value by key. Returns ErrCacheMiss if key not found or expired.
// Core logic: Check existence first with RLock, then check expiration and cleanup if expired.
//
// Get 根据键获取值，未找到或已过期返回 ErrCacheMiss。
// 核心逻辑：先用 RLock 检查存在性，再检查过期时间，过期则删除并返回 ErrCacheMiss。
func (s *memoryStore) Get(ctx context.Context, key string) (string, error) {
	_ = ctx

	s.mu.RLock()
	item, ok := s.m[key]
	s.mu.RUnlock()
	if !ok {
		return "", datacontract.ErrCacheMiss
	}
	if !item.expired.IsZero() && time.Now().After(item.expired) {
		s.mu.Lock()
		// Re-check: another goroutine may have Set a new non-expired value for this key.
		if current, ok := s.m[key]; ok && current.expired.Equal(item.expired) {
			delete(s.m, key)
		}
		s.mu.Unlock()
		return "", datacontract.ErrCacheMiss
	}
	return item.v, nil
}

// Set stores a key-value pair with optional TTL.
// Core logic: Calculate expiration time from TTL, then store with Lock.
//
// Set 存储键值对，支持可选的 TTL 过期时间。
// 核心逻辑：根据 TTL 计算过期时间，使用 Lock 存储数据。
func (s *memoryStore) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	_ = ctx

	var exp time.Time
	if ttl > 0 {
		exp = time.Now().Add(ttl)
	}

	s.mu.Lock()
	s.m[key] = memItem{v: value, expired: exp}
	s.mu.Unlock()
	return nil
}

// Del deletes a key from the cache.
//
// Del 删除指定键。
func (s *memoryStore) Del(ctx context.Context, key string) error {
	_ = ctx
	s.mu.Lock()
	delete(s.m, key)
	s.mu.Unlock()
	return nil
}

// MGet retrieves multiple keys at once, skipping missing keys.
// Core logic: Acquire lock once, iterate keys, return found non-expired values.
//
// MGet 批量获取多个键的值，跳过不存在的键。
// 核心逻辑：获取一次锁，遍历所有键，返回找到的未过期值。
func (s *memoryStore) MGet(ctx context.Context, keys ...string) (map[string]string, error) {
	_ = ctx

	s.mu.RLock()
	out := make(map[string]string, len(keys))
	now := time.Now()
	for _, k := range keys {
		item, ok := s.m[k]
		if !ok {
			continue
		}
		if !item.expired.IsZero() && now.After(item.expired) {
			continue
		}
		out[k] = item.v
	}
	s.mu.RUnlock()

	// Clean up expired keys found during read (lazy eviction).
	// 在读取期间发现过期键时执行惰性删除。
	var expired []string
	s.mu.RLock()
	for _, k := range keys {
		if item, ok := s.m[k]; ok && !item.expired.IsZero() && now.After(item.expired) {
			expired = append(expired, k)
		}
	}
	s.mu.RUnlock()
	if len(expired) > 0 {
		s.mu.Lock()
		for _, k := range expired {
			if current, ok := s.m[k]; ok && !current.expired.IsZero() && now.After(current.expired) {
				delete(s.m, k)
			}
		}
		s.mu.Unlock()
	}

	return out, nil
}
