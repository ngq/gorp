package polaris

import (
	"context"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	internalnative "github.com/ngq/gorp/contrib/internal/native"
	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	polarissdk "github.com/polarismesh/polaris-go"
	polarismodel "github.com/polarismesh/polaris-go/pkg/model"
	"gopkg.in/yaml.v3"
)

const defaultPolarisPollInterval = 5 * time.Second

var (
	ErrServerAddressRequired = errors.New("polaris: server_address is required")
	ErrFileGroupRequired     = errors.New("polaris: file_group is required")
	ErrFileNameRequired      = errors.New("polaris: file_name is required")
	ErrConfigNotFound        = errors.New("polaris: config not found")
	ErrAuthFailed            = errors.New("polaris: auth failed")
	ErrSourceUnavailable     = errors.New("polaris: source unavailable")
	ErrSetNotSupported       = errors.New("polaris: set is not supported")
	ErrConfigSourceClosed    = errors.New("polaris: config source closed")
)

// Provider 提供 Polaris 配置中心实现。
//
// 中文说明：
//   - 使用腾讯云 Polaris 配置中心；
//   - 支持命名空间隔离；
//   - 支持配置分组管理；
//   - 支持配置热更新；
//   - 适用于腾讯云环境和私有化部署。
//   - 当前状态：部分可用
//   - 说明：已完成 P2 第一版最小 HTTP 配置闭环，具备 Load / Watch 与 fake client 行为测试；
//     但当前仍是轮询桥接态，尚未进入完整 Polaris SDK 产品化能力。
type Provider struct{}

func NewProvider() *Provider           { return &Provider{} }
func (p *Provider) Name() string       { return "configsource.polaris" }
func (p *Provider) IsDefer() bool      { return true }
func (p *Provider) Provides() []string { return []string{datacontract.ConfigSourceKey} }

func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(datacontract.ConfigSourceKey, func(c runtimecontract.Container) (any, error) {
		cfg, err := getPolarisConfig(c)
		if err != nil {
			return nil, err
		}
		return NewConfigSource(cfg)
	}, true)
	return nil
}

func (p *Provider) Boot(c runtimecontract.Container) error { return nil }

type PolarisConfig struct {
	ServerAddress      string
	Namespace          string
	FileGroup          string
	FileName           string
	Token              string
	PollInterval       time.Duration
	WatchRetryInterval time.Duration
}

func getPolarisConfig(c runtimecontract.Container) (*PolarisConfig, error) {
	cfgAny, err := c.Make(datacontract.ConfigKey)
	if err != nil {
		return nil, err
	}
	cfg, ok := cfgAny.(datacontract.Config)
	if !ok {
		return nil, errors.New("polaris: invalid config service")
	}

	polarisCfg := &PolarisConfig{
		Namespace:          "default",
		PollInterval:       defaultPolarisPollInterval,
		WatchRetryInterval: time.Second,
	}
	if v := cfg.Get("config.polaris.server_address"); v != nil {
		polarisCfg.ServerAddress = cfg.GetString("config.polaris.server_address")
	}
	if v := cfg.Get("config.polaris.namespace"); v != nil {
		polarisCfg.Namespace = cfg.GetString("config.polaris.namespace")
	}
	if v := cfg.Get("config.polaris.file_group"); v != nil {
		polarisCfg.FileGroup = cfg.GetString("config.polaris.file_group")
	}
	if v := cfg.Get("config.polaris.file_name"); v != nil {
		polarisCfg.FileName = cfg.GetString("config.polaris.file_name")
	}
	if v := cfg.Get("config.polaris.token"); v != nil {
		polarisCfg.Token = cfg.GetString("config.polaris.token")
	}
	if v := cfg.Get("config.polaris.poll_interval_seconds"); v != nil {
		if seconds := cfg.GetInt("config.polaris.poll_interval_seconds"); seconds > 0 {
			polarisCfg.PollInterval = time.Duration(seconds) * time.Second
		}
	}
	if v := cfg.Get("config.polaris.watch_retry_interval_ms"); v != nil {
		if ms := cfg.GetInt("config.polaris.watch_retry_interval_ms"); ms > 0 {
			polarisCfg.WatchRetryInterval = time.Duration(ms) * time.Millisecond
		}
	}
	return polarisCfg, nil
}

type polarisConfigClient interface {
	GetConfig(ctx context.Context, cfg *PolarisConfig) (polarisConfigSnapshot, error)
	WatchConfig(ctx context.Context, cfg *PolarisConfig, lastRevision string, onUpdate func(snapshot polarisConfigSnapshot)) error
}

type polarisNativeClient interface {
	Underlying() any
}

type polarisConfigSnapshot struct {
	Content  string
	Revision string
}

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

func NewConfigSource(cfg *PolarisConfig) (*ConfigSource, error) {
	return NewConfigSourceWithClient(cfg, newOfficialPolarisClient())
}

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
		return nil, errors.New("polaris: config client is required")
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

func (s *ConfigSource) Get(ctx context.Context, key string) (any, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, ok := lookupNestedValue(s.cache, key)
	if !ok {
		return nil, fmt.Errorf("polaris: key %s not found", key)
	}
	return value, nil
}

func (s *ConfigSource) Set(ctx context.Context, key string, value any) error {
	return ErrSetNotSupported
}

func (s *ConfigSource) Underlying() any {
	if provider, ok := s.client.(polarisNativeClient); ok {
		if native := provider.Underlying(); native != nil {
			return native
		}
	}
	return s.client
}

func (s *ConfigSource) As(target any) bool {
	return internalnative.As(s.Underlying(), target)
}

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

type polarisWatcher struct {
	cancel    context.CancelFunc
	source    *ConfigSource
	callbacks *sync.Map
	stopped   bool
	stopMu    sync.RWMutex
}

func (w *polarisWatcher) OnChange(key string, callback func(value any)) {
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

func (w *polarisWatcher) Stop() error {
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

func (w *polarisWatcher) dispatch() {
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

type officialPolarisClient struct {
	mu      sync.Mutex
	context any
	api     polarissdk.ConfigAPI
}

func newOfficialPolarisClient() polarisConfigClient {
	return &officialPolarisClient{}
}

func (c *officialPolarisClient) Underlying() any {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.api != nil {
		return c.api
	}
	return nil
}

func (c *officialPolarisClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if destroyer, ok := c.context.(interface{ Destroy() }); ok {
		destroyer.Destroy()
	}
	c.context = nil
	c.api = nil
	return nil
}

func (c *officialPolarisClient) GetConfig(ctx context.Context, cfg *PolarisConfig) (polarisConfigSnapshot, error) {
	configFile, err := c.getConfigFile(cfg)
	if err != nil {
		return polarisConfigSnapshot{}, err
	}
	if configFile == nil || !configFile.HasContent() {
		return polarisConfigSnapshot{}, ErrConfigNotFound
	}
	content := configFile.GetContent()
	return polarisConfigSnapshot{
		Content:  content,
		Revision: normalizePolarisRevision(polarisConfigSnapshot{Content: content}),
	}, nil
}

func (c *officialPolarisClient) WatchConfig(ctx context.Context, cfg *PolarisConfig, lastRevision string, onUpdate func(snapshot polarisConfigSnapshot)) error {
	configFile, err := c.getConfigFile(cfg)
	if err != nil {
		return err
	}

	configFile.AddChangeListener(func(event polarismodel.ConfigFileChangeEvent) {
		select {
		case <-ctx.Done():
			return
		default:
		}

		content := event.NewValue
		if strings.TrimSpace(content) == "" {
			snapshot, getErr := c.GetConfig(ctx, cfg)
			if getErr != nil {
				return
			}
			content = snapshot.Content
		}

		snapshot := polarisConfigSnapshot{
			Content:  content,
			Revision: normalizePolarisRevision(polarisConfigSnapshot{Content: content}),
		}
		if snapshot.Revision == "" || snapshot.Revision == lastRevision {
			return
		}
		lastRevision = snapshot.Revision
		onUpdate(snapshot)
	})

	<-ctx.Done()
	return nil
}

func (c *officialPolarisClient) getConfigFile(cfg *PolarisConfig) (polarissdk.ConfigFile, error) {
	api, err := c.ensureAPI(cfg)
	if err != nil {
		return nil, err
	}
	configFile, err := api.GetConfigFile(cfg.Namespace, cfg.FileGroup, cfg.FileName)
	if err != nil {
		return nil, translatePolarisSDKError(err)
	}
	return configFile, nil
}

func (c *officialPolarisClient) ensureAPI(cfg *PolarisConfig) (polarissdk.ConfigAPI, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.api != nil {
		return c.api, nil
	}

	addresses, err := normalizePolarisAddresses(cfg.ServerAddress)
	if err != nil {
		return nil, err
	}
	context, err := polarissdk.NewSDKContextByAddress(addresses...)
	if err != nil {
		return nil, translatePolarisSDKError(err)
	}
	api := polarissdk.NewConfigAPIByContext(context)
	c.context = context
	c.api = api
	return c.api, nil
}

func normalizePolarisAddresses(raw string) ([]string, error) {
	parts := strings.Split(raw, ",")
	addresses := make([]string, 0, len(parts))
	for _, part := range parts {
		candidate := strings.TrimSpace(part)
		if candidate == "" {
			continue
		}
		if strings.Contains(candidate, "://") {
			parsed, err := url.Parse(candidate)
			if err != nil {
				return nil, fmt.Errorf("polaris: invalid server address: %w", err)
			}
			if parsed.Host != "" {
				candidate = parsed.Host
			}
		}
		addresses = append(addresses, candidate)
	}
	if len(addresses) == 0 {
		return nil, ErrServerAddressRequired
	}
	return addresses, nil
}

func translatePolarisSDKError(err error) error {
	if err != nil {
		message := strings.ToLower(err.Error())
		switch {
		case strings.Contains(message, "401"), strings.Contains(message, "403"),
			strings.Contains(message, "forbidden"), strings.Contains(message, "unauthorized"):
			return ErrAuthFailed
		case strings.Contains(message, "not found"), strings.Contains(message, "404"):
			return ErrConfigNotFound
		case strings.Contains(message, "connection refused"),
			strings.Contains(message, "dial tcp"),
			strings.Contains(message, "timeout"),
			strings.Contains(message, "no such host"),
			strings.Contains(message, "unavailable"):
			return ErrSourceUnavailable
		}
	}
	return err
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

func buildPolarisConfigURL(cfg *PolarisConfig) string {
	base := strings.TrimRight(cfg.ServerAddress, "/")
	values := url.Values{}
	values.Set("namespace", cfg.Namespace)
	values.Set("group", cfg.FileGroup)
	values.Set("file", cfg.FileName)
	return base + "/config/v1/files?" + values.Encode()
}

func decodeContent(content string, fallbackKey string) (map[string]any, error) {
	var decoded map[string]any
	if err := yaml.Unmarshal([]byte(content), &decoded); err == nil && len(decoded) > 0 {
		return normalizeMap(decoded), nil
	}
	var object map[string]any
	if err := json.Unmarshal([]byte(content), &object); err == nil && len(object) > 0 {
		return normalizeMap(object), nil
	}
	return map[string]any{fallbackKey: content}, nil
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

func isRetryablePolarisWatchError(err error) bool {
	return errors.Is(err, ErrSourceUnavailable)
}

func normalizePolarisRevision(snapshot polarisConfigSnapshot) string {
	revision := strings.TrimSpace(snapshot.Revision)
	if revision != "" {
		return revision
	}
	sum := sha1.Sum([]byte(snapshot.Content))
	return fmt.Sprintf("%x", sum[:])
}
