package apollo

import (
	"context"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/apolloconfig/agollo/v4"
	apolloagcache "github.com/apolloconfig/agollo/v4/agcache"
	apolloconfig "github.com/apolloconfig/agollo/v4/env/config"
	apollostorage "github.com/apolloconfig/agollo/v4/storage"
	internalnative "github.com/ngq/gorp/contrib/internal/native"
	"github.com/ngq/gorp/framework/contract"
	"gopkg.in/yaml.v3"
)

const defaultApolloPollInterval = 5 * time.Second

var (
	ErrAppIDRequired      = errors.New("apollo: app_id is required")
	ErrMetaRequired       = errors.New("apollo: meta_server is required")
	ErrConfigNotFound     = errors.New("apollo: config not found")
	ErrAuthFailed         = errors.New("apollo: auth failed")
	ErrSourceUnavailable  = errors.New("apollo: source unavailable")
	ErrSetNotSupported    = errors.New("apollo: set is not supported")
	ErrConfigSourceClosed = errors.New("apollo: config source closed")
)

// Provider 提供 Apollo 配置中心实现。
//
// 中文说明：
//   - 使用携程 Apollo 配置中心；
//   - 支持多命名空间；
//   - 支持配置热更新；
//   - 支持灰度发布。
//   - 当前状态：部分可用
//   - 说明：已完成 P2 第一版最小 HTTP 配置闭环，具备 Load / Watch 与 fake client 行为测试；
//     但当前仍是轮询桥接态，尚未进入完整 Apollo SDK 产品化能力。
type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string  { return "configsource.apollo" }
func (p *Provider) IsDefer() bool { return true }
func (p *Provider) Provides() []string {
	return []string{contract.ConfigSourceKey}
}

func (p *Provider) Register(c contract.Container) error {
	c.Bind(contract.ConfigSourceKey, func(c contract.Container) (any, error) {
		cfg, err := getApolloConfig(c)
		if err != nil {
			return nil, err
		}
		return NewConfigSource(cfg)
	}, true)

	return nil
}

func (p *Provider) Boot(c contract.Container) error { return nil }

// ApolloConfig 定义 Apollo 配置。
type ApolloConfig struct {
	AppID              string
	Cluster            string
	Namespace          string
	MetaServer         string
	AccessKey          string
	PollInterval       time.Duration
	WatchRetryInterval time.Duration
}

func getApolloConfig(c contract.Container) (*ApolloConfig, error) {
	cfgAny, err := c.Make(contract.ConfigKey)
	if err != nil {
		return nil, err
	}

	cfg, ok := cfgAny.(contract.Config)
	if !ok {
		return nil, errors.New("apollo: invalid config service")
	}

	apolloCfg := &ApolloConfig{
		Cluster:            "default",
		Namespace:          "application",
		PollInterval:       defaultApolloPollInterval,
		WatchRetryInterval: time.Second,
	}

	if v := cfg.Get("configsource.apollo.app_id"); v != nil {
		apolloCfg.AppID = cfg.GetString("configsource.apollo.app_id")
	} else if v := cfg.Get("config.apollo.app_id"); v != nil {
		apolloCfg.AppID = cfg.GetString("config.apollo.app_id")
	}
	if v := cfg.Get("configsource.apollo.cluster"); v != nil {
		apolloCfg.Cluster = cfg.GetString("configsource.apollo.cluster")
	} else if v := cfg.Get("config.apollo.cluster"); v != nil {
		apolloCfg.Cluster = cfg.GetString("config.apollo.cluster")
	}
	if v := cfg.Get("configsource.apollo.namespace"); v != nil {
		apolloCfg.Namespace = cfg.GetString("configsource.apollo.namespace")
	} else if v := cfg.Get("config.apollo.namespace"); v != nil {
		apolloCfg.Namespace = cfg.GetString("config.apollo.namespace")
	}
	if v := cfg.Get("configsource.apollo.meta_server"); v != nil {
		apolloCfg.MetaServer = cfg.GetString("configsource.apollo.meta_server")
	} else if v := cfg.Get("config.apollo.meta_server"); v != nil {
		apolloCfg.MetaServer = cfg.GetString("config.apollo.meta_server")
	}
	if v := cfg.Get("configsource.apollo.access_key"); v != nil {
		apolloCfg.AccessKey = cfg.GetString("configsource.apollo.access_key")
	} else if v := cfg.Get("config.apollo.access_key"); v != nil {
		apolloCfg.AccessKey = cfg.GetString("config.apollo.access_key")
	}
	if v := cfg.Get("configsource.apollo.poll_interval_seconds"); v != nil {
		if seconds := cfg.GetInt("configsource.apollo.poll_interval_seconds"); seconds > 0 {
			apolloCfg.PollInterval = time.Duration(seconds) * time.Second
		}
	}
	if v := cfg.Get("configsource.apollo.watch_retry_interval_ms"); v != nil {
		if ms := cfg.GetInt("configsource.apollo.watch_retry_interval_ms"); ms > 0 {
			apolloCfg.WatchRetryInterval = time.Duration(ms) * time.Millisecond
		}
	}

	return apolloCfg, nil
}

type apolloConfigClient interface {
	GetConfig(ctx context.Context, cfg *ApolloConfig) (apolloConfigSnapshot, error)
	WatchConfig(ctx context.Context, cfg *ApolloConfig, lastRevision string, onUpdate func(snapshot apolloConfigSnapshot)) error
}

type apolloNativeClient interface {
	Underlying() any
}

type apolloConfigSnapshot struct {
	Content  string
	Revision string
}

// ConfigSource Apollo 配置源实现。
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

func NewConfigSource(cfg *ApolloConfig) (*ConfigSource, error) {
	return NewConfigSourceWithClient(cfg, newOfficialApolloClient())
}

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

func (s *ConfigSource) Get(ctx context.Context, key string) (any, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, ok := lookupNestedValue(s.cache, key)
	if !ok {
		return nil, fmt.Errorf("apollo: key %s not found", key)
	}
	return value, nil
}

func (s *ConfigSource) Set(ctx context.Context, key string, value any) error {
	return ErrSetNotSupported
}

// Underlying returns the current native client object used by this config source.
func (s *ConfigSource) Underlying() any {
	if provider, ok := s.client.(apolloNativeClient); ok {
		if native := provider.Underlying(); native != nil {
			return native
		}
	}
	return s.client
}

// As projects the current native client into the requested target when possible.
func (s *ConfigSource) As(target any) bool {
	return internalnative.As(s.Underlying(), target)
}

func (s *ConfigSource) Watch(ctx context.Context, key string) (contract.ConfigWatcher, error) {
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

func (s *ConfigSource) currentRevision() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.revision
}

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

type apolloWatcher struct {
	cancel    context.CancelFunc
	source    *ConfigSource
	callbacks *sync.Map
	stopped   bool
	stopMu    sync.RWMutex
}

func (w *apolloWatcher) OnChange(key string, callback func(value any)) {
	w.stopMu.RLock()
	stopped := w.stopped
	w.stopMu.RUnlock()
	if stopped {
		return
	}
	w.callbacks.Store(key, callback)
	w.source.mu.RLock()
	current, exists := lookupNestedValue(w.source.cache, key)
	w.source.mu.RUnlock()
	if exists {
		callback(current)
	}
}

func (w *apolloWatcher) Stop() error {
	w.stopMu.Lock()
	if w.stopped {
		w.stopMu.Unlock()
		return nil
	}
	w.stopped = true
	w.stopMu.Unlock()
	w.cancel()
	w.source.watchers.Delete(w)
	return nil
}

func (w *apolloWatcher) dispatch() {
	w.stopMu.RLock()
	stopped := w.stopped
	w.stopMu.RUnlock()
	if stopped {
		return
	}
	w.callbacks.Range(func(key, value any) bool {
		w.stopMu.RLock()
		stopped := w.stopped
		w.stopMu.RUnlock()
		if stopped {
			return false
		}
		path, ok := key.(string)
		if !ok {
			return true
		}
		callback, ok := value.(func(value any))
		if !ok {
			return true
		}

		w.source.mu.RLock()
		current, exists := lookupNestedValue(w.source.cache, path)
		w.source.mu.RUnlock()
		if exists {
			callback(current)
		}
		return true
	})
}

type officialApolloClient struct {
	mu     sync.Mutex
	client agollo.Client
}

func newOfficialApolloClient() apolloConfigClient {
	return &officialApolloClient{}
}

func (c *officialApolloClient) Underlying() any {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.client
}

func (c *officialApolloClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.client != nil {
		c.client.Close()
		c.client = nil
	}
	return nil
}

func (c *officialApolloClient) GetConfig(ctx context.Context, cfg *ApolloConfig) (apolloConfigSnapshot, error) {
	client, err := c.ensureClient(cfg)
	if err != nil {
		return apolloConfigSnapshot{}, err
	}
	cache := client.GetConfigCache(cfg.Namespace)
	if cache == nil {
		return apolloConfigSnapshot{}, ErrConfigNotFound
	}
	content, err := buildApolloSnapshotContent(cache)
	if err != nil {
		return apolloConfigSnapshot{}, err
	}
	return apolloConfigSnapshot{
		Content:  content,
		Revision: normalizeApolloRevision(apolloConfigSnapshot{Content: content}),
	}, nil
}

func (c *officialApolloClient) WatchConfig(ctx context.Context, cfg *ApolloConfig, lastRevision string, onUpdate func(snapshot apolloConfigSnapshot)) error {
	client, err := c.ensureClient(cfg)
	if err != nil {
		return err
	}

	listener := &apolloChangeListener{
		namespace: cfg.Namespace,
		onEvent: func() {
			snapshot, getErr := c.GetConfig(ctx, cfg)
			if getErr != nil {
				return
			}
			revision := normalizeApolloRevision(snapshot)
			if revision == "" || revision == lastRevision {
				return
			}
			lastRevision = revision
			onUpdate(snapshot)
		},
	}
	client.AddChangeListener(listener)
	defer client.RemoveChangeListener(listener)

	<-ctx.Done()
	return nil
}

func (c *officialApolloClient) ensureClient(cfg *ApolloConfig) (agollo.Client, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.client != nil {
		return c.client, nil
	}

	client, err := agollo.StartWithConfig(func() (*apolloconfig.AppConfig, error) {
		return &apolloconfig.AppConfig{
			AppID:             cfg.AppID,
			Cluster:           cfg.Cluster,
			NamespaceName:     cfg.Namespace,
			IP:                cfg.MetaServer,
			Secret:            cfg.AccessKey,
			IsBackupConfig:    true,
			MustStart:         false,
			SyncServerTimeout: int((10 * time.Second) / time.Millisecond),
		}, nil
	})
	if err != nil {
		return nil, translateApolloSDKError(err)
	}
	c.client = client
	return c.client, nil
}

type apolloChangeListener struct {
	namespace string
	onEvent   func()
}

func (l *apolloChangeListener) OnChange(event *apollostorage.ChangeEvent) {
	if event == nil || event.Namespace != l.namespace {
		return
	}
	l.onEvent()
}

func (l *apolloChangeListener) OnNewestChange(event *apollostorage.FullChangeEvent) {
	if event == nil || event.Namespace != l.namespace {
		return
	}
	l.onEvent()
}

func buildApolloConfigURL(cfg *ApolloConfig) string {
	base := strings.TrimRight(cfg.MetaServer, "/")
	return fmt.Sprintf("%s/configs/%s/%s/%s", base, cfg.AppID, cfg.Cluster, cfg.Namespace)
}

func buildApolloSnapshotContent(cache apolloagcache.CacheInterface) (string, error) {
	if cache == nil || cache.EntryCount() == 0 {
		return "", ErrConfigNotFound
	}

	loaded := make(map[string]any)
	cache.Range(func(key, value interface{}) bool {
		keyString, ok := key.(string)
		if !ok || keyString == "" {
			return true
		}
		assignNestedValue(loaded, keyString, value)
		return true
	})
	if len(loaded) == 0 {
		return "", ErrConfigNotFound
	}

	content, err := yaml.Marshal(loaded)
	if err != nil {
		return "", fmt.Errorf("apollo: marshal config failed: %w", err)
	}
	return string(content), nil
}

func assignNestedValue(target map[string]any, path string, value any) {
	parts := strings.Split(path, ".")
	current := target
	for i, part := range parts {
		if i == len(parts)-1 {
			current[part] = value
			return
		}
		next, ok := current[part].(map[string]any)
		if !ok {
			next = make(map[string]any)
			current[part] = next
		}
		current = next
	}
}

func translateApolloSDKError(err error) error {
	if err == nil {
		return nil
	}
	message := strings.ToLower(err.Error())
	switch {
	case strings.Contains(message, "401"), strings.Contains(message, "403"),
		strings.Contains(message, "forbidden"), strings.Contains(message, "unauthorized"):
		return ErrAuthFailed
	case strings.Contains(message, "404"), strings.Contains(message, "not found"):
		return ErrConfigNotFound
	case strings.Contains(message, "connection refused"),
		strings.Contains(message, "no such host"),
		strings.Contains(message, "timeout"),
		strings.Contains(message, "dial tcp"),
		strings.Contains(message, "server unavailable"):
		return ErrSourceUnavailable
	default:
		return err
	}
}

func decodeContent(content string, namespace string) (map[string]any, error) {
	var decoded map[string]any
	if err := yaml.Unmarshal([]byte(content), &decoded); err == nil && len(decoded) > 0 {
		return normalizeMap(decoded), nil
	}

	var object map[string]any
	if err := json.Unmarshal([]byte(content), &object); err == nil && len(object) > 0 {
		return normalizeMap(object), nil
	}

	return map[string]any{
		namespace: content,
	}, nil
}

func cloneMap(input map[string]any) map[string]any {
	if input == nil {
		return nil
	}
	cloned := make(map[string]any, len(input))
	for k, v := range input {
		if nested, ok := v.(map[string]any); ok {
			cloned[k] = cloneMap(nested)
			continue
		}
		cloned[k] = v
	}
	return cloned
}

func lookupNestedValue(data map[string]any, path string) (any, bool) {
	if path == "" {
		return data, true
	}

	current := any(data)
	for _, part := range strings.Split(path, ".") {
		nested, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		current, ok = nested[part]
		if !ok {
			return nil, false
		}
	}
	return current, true
}

func normalizeMap(input map[string]any) map[string]any {
	normalized := make(map[string]any, len(input))
	for k, v := range input {
		switch typed := v.(type) {
		case map[string]any:
			normalized[k] = normalizeMap(typed)
		case map[any]any:
			converted := make(map[string]any, len(typed))
			for nestedKey, nestedValue := range typed {
				keyString, ok := nestedKey.(string)
				if !ok {
					continue
				}
				converted[keyString] = nestedValue
			}
			normalized[k] = normalizeMap(converted)
		default:
			normalized[k] = v
		}
	}
	return normalized
}

func isRetryableApolloWatchError(err error) bool {
	return errors.Is(err, ErrSourceUnavailable)
}

func normalizeApolloRevision(snapshot apolloConfigSnapshot) string {
	revision := strings.TrimSpace(snapshot.Revision)
	if revision != "" {
		return revision
	}
	sum := sha1.Sum([]byte(snapshot.Content))
	return fmt.Sprintf("%x", sum[:])
}
