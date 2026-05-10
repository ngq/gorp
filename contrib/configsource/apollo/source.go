// Package apollo provides Apollo configuration source implementation.
// This file implements the ConfigSource contract with cache and revision tracking.
//
// 本包提供 Apollo 配置源实现。
// 本文件实现 ConfigSource 契约，包含缓存和版本追踪。
package apollo

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	internalnative "github.com/ngq/gorp/contrib/internal/native"
	datacontract "github.com/ngq/gorp/framework/contract/data"
)

// ErrAppIDRequired indicates Apollo app_id is required.
//
// ErrAppIDRequired 表示 Apollo app_id 必需。
var ErrAppIDRequired = errors.New("apollo: app_id is required")

// ErrMetaRequired indicates Apollo meta_server is required.
//
// ErrMetaRequired 表示 Apollo meta_server 必需。
var ErrMetaRequired = errors.New("apollo: meta_server is required")

// ErrConfigNotFound indicates Apollo config not found.
//
// ErrConfigNotFound 表示 Apollo 配置未找到。
var ErrConfigNotFound = errors.New("apollo: config not found")

// ErrAuthFailed indicates Apollo authentication failed.
//
// ErrAuthFailed 表示 Apollo 认证失败。
var ErrAuthFailed = errors.New("apollo: auth failed")

// ErrSourceUnavailable indicates Apollo source unavailable.
//
// ErrSourceUnavailable 表示 Apollo 配置源不可用。
var ErrSourceUnavailable = errors.New("apollo: source unavailable")

// ErrSetNotSupported indicates Apollo set is not supported.
//
// ErrSetNotSupported 表示 Apollo 不支持 set 操作。
var ErrSetNotSupported = errors.New("apollo: set is not supported")

// ErrConfigSourceClosed indicates Apollo config source closed.
//
// ErrConfigSourceClosed 表示 Apollo 配置源已关闭。
var ErrConfigSourceClosed = errors.New("apollo: config source closed")

// defaultApolloPollInterval is the default polling interval for config updates.
//
// defaultApolloPollInterval 是配置更新的默认轮询间隔。
const defaultApolloPollInterval = 5 * time.Second

// apolloConfigClient defines the internal client interface for Apollo operations.
//
// apolloConfigClient 定义 Apollo 操作的内部客户端接口。
type apolloConfigClient interface {
	GetConfig(ctx context.Context, cfg *ApolloConfig) (apolloConfigSnapshot, error)
	WatchConfig(ctx context.Context, cfg *ApolloConfig, lastRevision string, onUpdate func(snapshot apolloConfigSnapshot)) error
}

// apolloNativeClient defines the interface for accessing underlying Apollo client.
//
// apolloNativeClient 定义访问底层 Apollo 客户端的接口。
type apolloNativeClient interface {
	Underlying() any
}

// apolloConfigSnapshot represents a configuration snapshot from Apollo.
//
// apolloConfigSnapshot 表示 Apollo 配置快照。
type apolloConfigSnapshot struct {
	Content  string
	Revision string
}

// ConfigSource Apollo 配置源实现。
// Implements datacontract.ConfigSource with cache, revision tracking and watch support.
//
// ConfigSource 实现 datacontract.ConfigSource。
// 包含缓存、版本追踪和 watch 支持。
type ConfigSource struct {
	config   *ApolloConfig
	client   apolloConfigClient
	mu       sync.RWMutex
	cache    map[string]any
	revision string
	watchers sync.Map
	closeMu  sync.Mutex
	closed   bool
}

// NewConfigSource creates a new Apollo config source with default official client.
//
// NewConfigSource 使用默认官方客户端创建新的 Apollo 配置源。
func NewConfigSource(cfg *ApolloConfig) (*ConfigSource, error) {
	return NewConfigSourceWithClient(cfg, newOfficialApolloClient())
}

// NewConfigSourceWithClient creates a new Apollo config source with custom client.
//
// NewConfigSourceWithClient 使用自定义客户端创建新的 Apollo 配置源。
func NewConfigSourceWithClient(cfg *ApolloConfig, client apolloConfigClient) (*ConfigSource, error) {
	if cfg.AppID == "" {
		return nil, ErrAppIDRequired
	}
	if cfg.MetaServer == "" {
		return nil, ErrMetaRequired
	}
	if client == nil {
		return nil, errors.New("apollo: config client is required")
	}

	return &ConfigSource{
		config: cfg,
		client: client,
		cache:  make(map[string]any),
	}, nil
}

// Load loads the full config snapshot from Apollo.
// Implements datacontract.ConfigSource.Load.
//
// Load 从 Apollo 加载完整配置快照。
// 实现 datacontract.ConfigSource.Load。
func (s *ConfigSource) Load(ctx context.Context) (map[string]any, error) {
	content, err := s.client.GetConfig(ctx, s.config)
	if err != nil {
		return nil, err
	}

	loaded, err := decodeContent(content.Content, s.config.Namespace)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	s.cache = cloneMap(loaded)
	s.revision = normalizeApolloRevision(content)
	s.mu.Unlock()

	return cloneMap(loaded), nil
}

// Get reads a single config value from the source.
// Implements datacontract.ConfigSource.Get.
//
// Get 从配置源读取单个配置值。
// 实现 datacontract.ConfigSource.Get。
func (s *ConfigSource) Get(ctx context.Context, key string) (any, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, ok := lookupNestedValue(s.cache, key)
	if !ok {
		return nil, fmt.Errorf("apollo: key %s not found", key)
	}
	return value, nil
}

// Set is not supported for Apollo config source.
// Implements datacontract.ConfigSource.Set.
//
// Set Apollo 配置源不支持 set 操作。
// 实现 datacontract.ConfigSource.Set。
func (s *ConfigSource) Set(ctx context.Context, key string, value any) error {
	return ErrSetNotSupported
}

// Underlying returns the current native client object used by this config source.
//
// Underlying 返回此配置源使用的当前原生客户端对象。
func (s *ConfigSource) Underlying() any {
	if provider, ok := s.client.(apolloNativeClient); ok {
		if native := provider.Underlying(); native != nil {
			return native
		}
	}
	return s.client
}

// As projects the current native client into the requested target when possible.
//
// As 在可能时将当前原生客户端投射到请求的目标。
func (s *ConfigSource) As(target any) bool {
	return internalnative.As(s.Underlying(), target)
}

// Watch subscribes to source-side changes of a config key.
// Implements datacontract.ConfigSource.Watch.
//
// Watch 订阅配置源侧的指定 key 变更。
// 实现 datacontract.ConfigSource.Watch。
func (s *ConfigSource) Watch(ctx context.Context, key string) (datacontract.ConfigWatcher, error) {
	s.closeMu.Lock()
	defer s.closeMu.Unlock()

	if s.closed {
		return nil, ErrConfigSourceClosed
	}
	if err := s.ensureLoaded(ctx); err != nil {
		return nil, err
	}

	watchCtx, cancel := context.WithCancel(ctx)
	watcher := &apolloWatcher{
		cancel:    cancel,
		source:    s,
		callbacks: &sync.Map{},
	}
	s.watchers.Store(watcher, struct{}{})

	go func() {
		for {
			lastRevision := s.currentRevision()
			err := s.client.WatchConfig(watchCtx, s.config, lastRevision, func(snapshot apolloConfigSnapshot) {
				if !s.applySnapshot(snapshot) {
					return
				}
				watcher.dispatch()
			})
			if err == nil || watchCtx.Err() != nil {
				return
			}
			if !isRetryableApolloWatchError(err) {
				cancel()
				return
			}
			select {
			case <-watchCtx.Done():
				return
			case <-time.After(s.config.WatchRetryInterval):
			}
		}
	}()
	go s.pollLoop(watchCtx, watcher)

	return watcher, nil
}

// Close releases resources held by the source.
// Implements datacontract.ConfigSource.Close.
//
// Close 释放配置源持有的资源。
// 实现 datacontract.ConfigSource.Close。
func (s *ConfigSource) Close() error {
	s.closeMu.Lock()
	defer s.closeMu.Unlock()

	if s.closed {
		return nil
	}
	s.closed = true
	s.watchers.Range(func(key, value any) bool {
		if watcher, ok := key.(*apolloWatcher); ok {
			_ = watcher.Stop()
		}
		return true
	})
	if closer, ok := s.client.(interface{ Close() error }); ok {
		return closer.Close()
	}
	return nil
}

// ensureLoaded ensures config is loaded before operations.
//
// ensureLoaded 确保配置在操作前已加载。
func (s *ConfigSource) ensureLoaded(ctx context.Context) error {
	s.mu.RLock()
	loaded := len(s.cache) > 0
	s.mu.RUnlock()
	if loaded {
		return nil
	}

	_, err := s.Load(ctx)
	return err
}

// currentRevision returns the current config revision.
//
// currentRevision 返回当前配置版本。
func (s *ConfigSource) currentRevision() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.revision
}

// pollLoop runs periodic polling for config updates.
//
// pollLoop 运行周期性轮询以更新配置。
func (s *ConfigSource) pollLoop(ctx context.Context, watcher *apolloWatcher) {
	if s.config.PollInterval <= 0 {
		return
	}

	ticker := time.NewTicker(s.config.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			snapshot, err := s.client.GetConfig(ctx, s.config)
			if err != nil {
				continue
			}
			if !s.applySnapshot(snapshot) {
				continue
			}
			watcher.dispatch()
		}
	}
}

// applySnapshot applies a new config snapshot if revision changed.
//
// applySnapshot 如果版本变化则应用新配置快照。
func (s *ConfigSource) applySnapshot(snapshot apolloConfigSnapshot) bool {
	loaded, decodeErr := decodeContent(snapshot.Content, s.config.Namespace)
	if decodeErr != nil {
		return false
	}

	revision := normalizeApolloRevision(snapshot)
	s.mu.Lock()
	defer s.mu.Unlock()
	if revision != "" && revision == s.revision {
		return false
	}
	s.cache = cloneMap(loaded)
	s.revision = revision
	return true
}