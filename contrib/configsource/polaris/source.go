// Package polaris provides Polaris configuration source implementation.
// This file implements the ConfigSource contract with cache and revision tracking.
//
// 本包提供 Polaris 配置源实现。
// 本文件实现 ConfigSource 契约，包含缓存和版本追踪。
package polaris

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	internalnative "github.com/ngq/gorp/contrib/internal/native"
	datacontract "github.com/ngq/gorp/framework/contract/data"
)

// ErrServerAddressRequired indicates Polaris server_address is required.
//
// ErrServerAddressRequired 表示 Polaris server_address 必需。
var ErrServerAddressRequired = errors.New("configsource.polaris: server_address is required")

// ErrFileGroupRequired indicates Polaris file_group is required.
//
// ErrFileGroupRequired 表示 Polaris file_group 必需。
var ErrFileGroupRequired = errors.New("configsource.polaris: file_group is required")

// ErrFileNameRequired indicates Polaris file_name is required.
//
// ErrFileNameRequired 表示 Polaris file_name 必需。
var ErrFileNameRequired = errors.New("configsource.polaris: file_name is required")

// ErrConfigNotFound indicates Polaris config not found.
//
// ErrConfigNotFound 表示 Polaris 配置未找到。
var ErrConfigNotFound = errors.New("configsource.polaris: config not found")

// ErrAuthFailed indicates Polaris authentication failed.
//
// ErrAuthFailed 表示 Polaris 认证失败。
var ErrAuthFailed = errors.New("configsource.polaris: auth failed")

// ErrSourceUnavailable indicates Polaris source unavailable.
//
// ErrSourceUnavailable 表示 Polaris 配置源不可用。
var ErrSourceUnavailable = errors.New("configsource.polaris: source unavailable")

// ErrSetNotSupported indicates Polaris set is not supported.
//
// ErrSetNotSupported 表示 Polaris 不支持 set 操作。
var ErrSetNotSupported = errors.New("configsource.polaris: set is not supported")

// ErrConfigSourceClosed indicates Polaris config source closed.
//
// ErrConfigSourceClosed 表示 Polaris 配置源已关闭。
var ErrConfigSourceClosed = errors.New("configsource.polaris: config source closed")

// polarisConfigClient defines the internal client interface for Polaris operations.
//
// polarisConfigClient 定义 Polaris 操作的内部客户端接口。
type polarisConfigClient interface {
	GetConfig(ctx context.Context, cfg *PolarisConfig) (polarisConfigSnapshot, error)
	WatchConfig(ctx context.Context, cfg *PolarisConfig, lastRevision string, onUpdate func(snapshot polarisConfigSnapshot)) error
}

// polarisNativeClient defines the interface for accessing underlying Polaris client.
//
// polarisNativeClient 定义访问底层 Polaris 客户端的接口。
type polarisNativeClient interface {
	Underlying() any
}

// polarisConfigSnapshot represents a configuration snapshot from Polaris.
//
// polarisConfigSnapshot 表示 Polaris 配置快照。
type polarisConfigSnapshot struct {
	Content  string
	Revision string
}

// ConfigSource Polaris 配置源实现。
// Implements datacontract.ConfigSource with cache, revision tracking and watch support.
//
// ConfigSource 实现 datacontract.ConfigSource。
// 包含缓存、版本追踪和 watch 支持。
type ConfigSource struct {
	config   *PolarisConfig
	client   polarisConfigClient
	mu       sync.RWMutex
	cache    map[string]any
	revision string
	watchers sync.Map
	closeMu  sync.Mutex
	closed   bool
}

// NewConfigSource creates a new Polaris config source with default official client.
//
// NewConfigSource 使用默认官方客户端创建新的 Polaris 配置源。
func NewConfigSource(cfg *PolarisConfig) (*ConfigSource, error) {
	return NewConfigSourceWithClient(cfg, newOfficialPolarisClient())
}

// NewConfigSourceWithClient creates a new Polaris config source with custom client.
//
// NewConfigSourceWithClient 使用自定义客户端创建新的 Polaris 配置源。
func NewConfigSourceWithClient(cfg *PolarisConfig, client polarisConfigClient) (*ConfigSource, error) {
	if cfg.ServerAddress == "" {
		return nil, ErrServerAddressRequired
	}
	if cfg.FileGroup == "" {
		return nil, ErrFileGroupRequired
	}
	if cfg.FileName == "" {
		return nil, ErrFileNameRequired
	}
	if client == nil {
		return nil, errors.New("configsource.polaris: config client is required")
	}

	return &ConfigSource{
		config: cfg,
		client: client,
		cache:  make(map[string]any),
	}, nil
}

// Load loads the full config snapshot from Polaris.
// Implements datacontract.ConfigSource.Load.
//
// Load 从 Polaris 加载完整配置快照。
// 实现 datacontract.ConfigSource.Load。
func (s *ConfigSource) Load(ctx context.Context) (map[string]any, error) {
	content, err := s.client.GetConfig(ctx, s.config)
	if err != nil {
		return nil, err
	}

	loaded, err := decodeContent(content.Content, s.config.FileName)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	s.cache = cloneMap(loaded)
	s.revision = normalizePolarisRevision(content)
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
		return nil, fmt.Errorf("configsource.polaris: key %s not found", key)
	}
	return value, nil
}

// Set is not supported for Polaris config source.
// Implements datacontract.ConfigSource.Set.
//
// Set Polaris 配置源不支持 set 操作。
// 实现 datacontract.ConfigSource.Set。
func (s *ConfigSource) Set(ctx context.Context, key string, value any) error {
	return ErrSetNotSupported
}

// Underlying returns the current native client object used by this config source.
//
// Underlying 返回此配置源使用的当前原生客户端对象。
func (s *ConfigSource) Underlying() any {
	if provider, ok := s.client.(polarisNativeClient); ok {
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
	watcher := &polarisWatcher{
		cancel:    cancel,
		source:    s,
		callbacks: &sync.Map{},
	}
	s.watchers.Store(watcher, struct{}{})

	go func() {
		for {
			lastRevision := s.currentRevision()
			err := s.client.WatchConfig(watchCtx, s.config, lastRevision, func(snapshot polarisConfigSnapshot) {
				if !s.applySnapshot(snapshot) {
					return
				}
				watcher.dispatch()
			})
			if err == nil || watchCtx.Err() != nil {
				return
			}
			if !isRetryablePolarisWatchError(err) {
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
		if watcher, ok := key.(*polarisWatcher); ok {
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
func (s *ConfigSource) pollLoop(ctx context.Context, watcher *polarisWatcher) {
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
func (s *ConfigSource) applySnapshot(snapshot polarisConfigSnapshot) bool {
	loaded, decodeErr := decodeContent(snapshot.Content, s.config.FileName)
	if decodeErr != nil {
		return false
	}

	revision := normalizePolarisRevision(snapshot)
	s.mu.Lock()
	defer s.mu.Unlock()
	if revision != "" && revision == s.revision {
		return false
	}
	s.cache = cloneMap(loaded)
	s.revision = revision
	return true
}
