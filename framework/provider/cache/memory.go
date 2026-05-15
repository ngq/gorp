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
	"errors"
	"sync"
	"time"

	datacontract "github.com/ngq/gorp/framework/contract/data"
)

// memoryStore implements cacheDriver using an in-memory map with TTL support.
//
// memoryStore 使用内存 map 实现缓存驱动，支持 TTL 过期。
type memoryStore struct {
	mu sync.RWMutex       // mu protects concurrent access to the map.
	                       //
	                        // mu 保护并发访问。
	m  map[string]memItem // m stores cache items.
	                       //
	                        // m 存储缓存项。
}

// memItem represents a cached item with value and expiration time.
//
// memItem 表示缓存项，包含值和过期时间。
type memItem struct {
	v       string    // v is the cached value.
	                   //
	                    // v 缓存值。
	expired time.Time // expired is the TTL expiration time.
	                   //
	                    // expired TTL 过期时间。
}

// newMemoryStore creates a new in-memory cache store.
//
// newMemoryStore 创建新的内存缓存存储实例。
func newMemoryStore() *memoryStore {
	return &memoryStore{m: make(map[string]memItem)}
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
		delete(s.m, key)
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
// Core logic: Iterate over keys, call Get for each, skip ErrCacheMiss.
//
// MGet 批量获取多个键的值，跳过不存在的键。
// 核心逻辑：遍历所有键，逐个调用 Get，跳过 ErrCacheMiss 错误。
func (s *memoryStore) MGet(ctx context.Context, keys ...string) (map[string]string, error) {
	_ = ctx

	out := make(map[string]string, len(keys))
	for _, k := range keys {
		v, err := s.Get(ctx, k)
		if err != nil {
		if errors.Is(err, datacontract.ErrCacheMiss) {
				continue
			}
			return nil, err
		}
		out[k] = v
	}
	return out, nil
}