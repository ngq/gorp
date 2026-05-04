package nacos

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	configclient "github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	internalnative "github.com/ngq/gorp/contrib/internal/native"
	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	"gopkg.in/yaml.v3"
)

const defaultNacosPollInterval = 5 * time.Second

// Provider 提供 Nacos 配置中心实现。
//
// 中文说明：
//   - 使用 Nacos 配置中心；
//   - 支持多命名空间；
//   - 支持配置热更新；
//   - 支持分组管理。
//   - 当前状态：部分可用
//   - 说明：已完成 P1 最小闭环，具备 Load / Watch / Set 主流程与 fake client 行为测试；
//     但当前仍是最小配置中心闭环，尚未进入完整产品化治理能力。
type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string  { return "configsource.nacos" }
func (p *Provider) IsDefer() bool { return true }
func (p *Provider) Provides() []string {
	return []string{datacontract.ConfigSourceKey}
}

func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(datacontract.ConfigSourceKey, func(c runtimecontract.Container) (any, error) {
		cfg, err := getNacosConfig(c)
		if err != nil {
			return nil, err
		}
		return NewConfigSource(cfg)
	}, true)

	return nil
}

func (p *Provider) Boot(c runtimecontract.Container) error {
	return nil
}

// NacosConfig 定义 Nacos 配置。
type NacosConfig struct {
	ServerAddr   string
	Port         int
	Namespace    string
	Group        string
	DataID       string
	Username     string
	Password     string
	PollInterval time.Duration
}

// getNacosConfig 从容器获取配置。
func getNacosConfig(c runtimecontract.Container) (*NacosConfig, error) {
	cfgAny, err := c.Make(datacontract.ConfigKey)
	if err != nil {
		return nil, err
	}

	cfg, ok := cfgAny.(datacontract.Config)
	if !ok {
		return nil, errors.New("nacos: invalid config service")
	}

	nacosCfg := &NacosConfig{
		Port:         8848,
		Group:        "DEFAULT_GROUP",
		PollInterval: defaultNacosPollInterval,
	}

	if v := cfg.Get("configsource.nacos.server_addr"); v != nil {
		nacosCfg.ServerAddr = cfg.GetString("configsource.nacos.server_addr")
	} else if v := cfg.Get("config.nacos.server_addr"); v != nil {
		nacosCfg.ServerAddr = cfg.GetString("config.nacos.server_addr")
	}
	if v := cfg.Get("configsource.nacos.port"); v != nil {
		nacosCfg.Port = cfg.GetInt("configsource.nacos.port")
	} else if v := cfg.Get("config.nacos.port"); v != nil {
		nacosCfg.Port = cfg.GetInt("config.nacos.port")
	}
	if v := cfg.Get("configsource.nacos.namespace"); v != nil {
		nacosCfg.Namespace = cfg.GetString("configsource.nacos.namespace")
	} else if v := cfg.Get("config.nacos.namespace"); v != nil {
		nacosCfg.Namespace = cfg.GetString("config.nacos.namespace")
	}
	if v := cfg.Get("configsource.nacos.group"); v != nil {
		nacosCfg.Group = cfg.GetString("configsource.nacos.group")
	} else if v := cfg.Get("config.nacos.group"); v != nil {
		nacosCfg.Group = cfg.GetString("config.nacos.group")
	}
	if v := cfg.Get("configsource.nacos.data_id"); v != nil {
		nacosCfg.DataID = cfg.GetString("configsource.nacos.data_id")
	} else if v := cfg.Get("config.nacos.data_id"); v != nil {
		nacosCfg.DataID = cfg.GetString("config.nacos.data_id")
	}
	if v := cfg.Get("configsource.nacos.username"); v != nil {
		nacosCfg.Username = cfg.GetString("configsource.nacos.username")
	} else if v := cfg.Get("config.nacos.username"); v != nil {
		nacosCfg.Username = cfg.GetString("config.nacos.username")
	}
	if v := cfg.Get("configsource.nacos.password"); v != nil {
		nacosCfg.Password = cfg.GetString("configsource.nacos.password")
	} else if v := cfg.Get("config.nacos.password"); v != nil {
		nacosCfg.Password = cfg.GetString("config.nacos.password")
	}
	if v := cfg.Get("configsource.nacos.poll_interval_seconds"); v != nil {
		if seconds := cfg.GetInt("configsource.nacos.poll_interval_seconds"); seconds > 0 {
			nacosCfg.PollInterval = time.Duration(seconds) * time.Second
		}
	}

	return nacosCfg, nil
}

type nacosConfigClient interface {
	GetConfig(ctx context.Context, cfg *NacosConfig) (string, error)
	PublishConfig(ctx context.Context, cfg *NacosConfig, content string) error
	WatchConfig(ctx context.Context, cfg *NacosConfig, onUpdate func(string)) error
}

// ConfigSource Nacos 配置源实现。
type ConfigSource struct {
	config   *NacosConfig
	client   nacosConfigClient
	mu       sync.RWMutex
	cache    map[string]any
	watchers sync.Map
	closed   bool
	closeMu  sync.Mutex
}

// NewConfigSource 创建 Nacos 配置源。
func NewConfigSource(cfg *NacosConfig) (*ConfigSource, error) {
	client, err := newOfficialNacosClient(cfg)
	if err != nil {
		return nil, err
	}
	return NewConfigSourceWithClient(cfg, client)
}

// NewConfigSourceWithClient 创建带自定义 client 的配置源。
func NewConfigSourceWithClient(cfg *NacosConfig, client nacosConfigClient) (*ConfigSource, error) {
	if cfg.ServerAddr == "" {
		return nil, ErrServerAddrRequired
	}
	if cfg.DataID == "" {
		return nil, ErrDataIDRequired
	}
	if client == nil {
		return nil, errors.New("nacos: config client is required")
	}

	return &ConfigSource{
		config: cfg,
		client: client,
		cache:  make(map[string]any),
	}, nil
}

// Load 加载配置。
func (s *ConfigSource) Load(ctx context.Context) (map[string]any, error) {
	content, err := s.client.GetConfig(ctx, s.config)
	if err != nil {
		return nil, err
	}

	loaded, err := decodeContent(content, s.config.DataID)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	s.cache = cloneMap(loaded)
	s.mu.Unlock()

	return cloneMap(loaded), nil
}

// Get 获取配置值。
func (s *ConfigSource) Get(ctx context.Context, key string) (any, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, ok := lookupNestedValue(s.cache, key)
	if !ok {
		return nil, fmt.Errorf("nacos: key %s not found", key)
	}
	return value, nil
}

// Set 发布配置。
func (s *ConfigSource) Set(ctx context.Context, key string, value any) error {
	if key != "" && key != s.config.DataID {
		return fmt.Errorf("nacos: set only supports data_id %s", s.config.DataID)
	}

	content, err := encodeContent(value)
	if err != nil {
		return err
	}
	if err := s.client.PublishConfig(ctx, s.config, content); err != nil {
		return err
	}

	loaded, err := decodeContent(content, s.config.DataID)
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.cache = cloneMap(loaded)
	s.mu.Unlock()

	return nil
}

// Underlying returns the native client currently backing this config source.
func (s *ConfigSource) Underlying() any {
	return unwrapNacosNativeClient(s.client)
}

// As projects the native client into the requested target when possible.
func (s *ConfigSource) As(target any) bool {
	return internalnative.As(unwrapNacosNativeClient(s.client), target)
}

// Watch 监听配置变化。
func (s *ConfigSource) Watch(ctx context.Context, key string) (datacontract.ConfigWatcher, error) {
	s.closeMu.Lock()
	defer s.closeMu.Unlock()

	if s.closed {
		return nil, errors.New("nacos: config source closed")
	}

	watchCtx, cancel := context.WithCancel(ctx)
	watcher := &nacosWatcher{
		ctx:       watchCtx,
		cancel:    cancel,
		source:    s,
		callbacks: &sync.Map{},
	}
	s.watchers.Store(watcher, struct{}{})
	go func() {
		err := s.client.WatchConfig(watchCtx, s.config, func(content string) {
			loaded, decodeErr := decodeContent(content, s.config.DataID)
			if decodeErr != nil {
				return
			}
			s.mu.Lock()
			s.cache = cloneMap(loaded)
			s.mu.Unlock()
			watcher.dispatch()
		})
		if err != nil {
			cancel()
		}
	}()

	return watcher, nil
}

// Close 关闭配置源。
func (s *ConfigSource) Close() error {
	s.closeMu.Lock()
	defer s.closeMu.Unlock()

	if s.closed {
		return nil
	}
	s.closed = true
	s.watchers.Range(func(key, value any) bool {
		if watcher, ok := key.(*nacosWatcher); ok {
			_ = watcher.Stop()
		}
		return true
	})
	if closer, ok := s.client.(interface{ CloseClient() }); ok {
		closer.CloseClient()
	}
	return nil
}

// nacosWatcher 配置监听器。
type nacosWatcher struct {
	ctx       context.Context
	cancel    context.CancelFunc
	source    *ConfigSource
	callbacks *sync.Map
}

// OnChange 配置变更回调。
func (w *nacosWatcher) OnChange(key string, callback func(value any)) {
	w.callbacks.Store(key, callback)
	w.source.mu.RLock()
	current, exists := lookupNestedValue(w.source.cache, key)
	w.source.mu.RUnlock()
	if exists {
		callback(current)
	}
}

// Stop 停止监听。
func (w *nacosWatcher) Stop() error {
	w.cancel()
	w.source.watchers.Delete(w)
	return nil
}

func (w *nacosWatcher) dispatch() {
	w.callbacks.Range(func(key, value any) bool {
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

type sdkNacosClient struct {
	client configclient.IConfigClient
}

func newOfficialNacosClient(cfg *NacosConfig) (nacosConfigClient, error) {
	clientConfig := constant.ClientConfig{
		NamespaceId:         cfg.Namespace,
		NotLoadCacheAtStart: true,
	}
	if cfg.Username != "" && cfg.Password != "" {
		clientConfig.Username = cfg.Username
		clientConfig.Password = cfg.Password
	}

	client, err := clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig: &clientConfig,
			ServerConfigs: []constant.ServerConfig{
				{
					IpAddr:      parseNacosHost(cfg.ServerAddr),
					Port:        uint64(cfg.Port),
					ContextPath: "/nacos",
				},
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("nacos: create config client failed: %w", err)
	}

	return &sdkNacosClient{client: client}, nil
}

func (c *sdkNacosClient) GetConfig(ctx context.Context, cfg *NacosConfig) (string, error) {
	type result struct {
		content string
		err     error
	}
	done := make(chan result, 1)
	go func() {
		content, err := c.client.GetConfig(toNacosConfigParam(cfg))
		done <- result{content: content, err: translateNacosError("load config", err)}
	}()

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case res := <-done:
		return res.content, res.err
	}
}

func (c *sdkNacosClient) PublishConfig(ctx context.Context, cfg *NacosConfig, content string) error {
	type result struct {
		ok  bool
		err error
	}
	done := make(chan result, 1)
	go func() {
		param := toNacosConfigParam(cfg)
		param.Content = content
		ok, err := c.client.PublishConfig(param)
		done <- result{ok: ok, err: err}
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case res := <-done:
		if res.err != nil {
			return translateNacosError("publish config", res.err)
		}
		if !res.ok {
			return errors.New("nacos: publish config failed")
		}
		return nil
	}
}

func (c *sdkNacosClient) WatchConfig(ctx context.Context, cfg *NacosConfig, onUpdate func(string)) error {
	param := toNacosConfigParam(cfg)
	param.OnChange = func(namespace, group, dataID, data string) {
		select {
		case <-ctx.Done():
			return
		default:
			onUpdate(data)
		}
	}

	if err := c.client.ListenConfig(param); err != nil {
		return translateNacosError("listen config", err)
	}

	<-ctx.Done()
	_ = c.client.CancelListenConfig(param)
	return nil
}

func (c *sdkNacosClient) CloseClient() {
	c.client.CloseClient()
}

func (c *sdkNacosClient) Underlying() any {
	return c.client
}

func toNacosConfigParam(cfg *NacosConfig) vo.ConfigParam {
	return vo.ConfigParam{
		DataId: cfg.DataID,
		Group:  cfg.Group,
	}
}

func parseNacosHost(addr string) string {
	trimmed := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(addr, "http://"), "https://"))
	if idx := strings.Index(trimmed, "/"); idx >= 0 {
		trimmed = trimmed[:idx]
	}
	if host, _, found := strings.Cut(trimmed, ":"); found {
		return host
	}
	return trimmed
}

func translateNacosError(action string, err error) error {
	if err == nil {
		return nil
	}
	message := strings.ToLower(err.Error())
	switch {
	case strings.Contains(message, "config data not exist"),
		strings.Contains(message, "data not exist"),
		strings.Contains(message, "404"):
		return ErrConfigNotFound
	default:
		return fmt.Errorf("nacos: %s failed: %w", action, err)
	}
}

func unwrapNacosNativeClient(client nacosConfigClient) any {
	if provider, ok := client.(interface{ Underlying() any }); ok {
		return provider.Underlying()
	}
	return client
}

func decodeContent(content, dataID string) (map[string]any, error) {
	value, err := decodeStructuredContent(content)
	if err != nil {
		return nil, err
	}
	if asMap, ok := value.(map[string]any); ok {
		return asMap, nil
	}
	return map[string]any{dataID: value}, nil
}

func encodeContent(value any) (string, error) {
	switch typed := value.(type) {
	case string:
		return typed, nil
	case []byte:
		return string(typed), nil
	default:
		data, err := yaml.Marshal(value)
		if err != nil {
			return "", fmt.Errorf("nacos: encode config content failed: %w", err)
		}
		return string(data), nil
	}
}

func normalizeYAMLValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		result := make(map[string]any, len(typed))
		for key, nested := range typed {
			result[key] = normalizeYAMLValue(nested)
		}
		return result
	case map[any]any:
		result := make(map[string]any, len(typed))
		for key, nested := range typed {
			result[fmt.Sprint(key)] = normalizeYAMLValue(nested)
		}
		return result
	case []any:
		result := make([]any, 0, len(typed))
		for _, nested := range typed {
			result = append(result, normalizeYAMLValue(nested))
		}
		return result
	default:
		return typed
	}
}

func decodeStructuredContent(content string) (any, error) {
	var value any
	if err := yaml.Unmarshal([]byte(content), &value); err == nil {
		return normalizeYAMLValue(value), nil
	}
	return strings.TrimSpace(content), nil
}

func lookupNestedValue(data map[string]any, path string) (any, bool) {
	if path == "" {
		return cloneMap(data), true
	}

	current := any(data)
	for _, segment := range strings.Split(path, ".") {
		asMap, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		next, exists := asMap[segment]
		if !exists {
			return nil, false
		}
		current = next
	}
	return current, true
}

func cloneMap(source map[string]any) map[string]any {
	result := make(map[string]any, len(source))
	for key, value := range source {
		result[key] = value
	}
	return result
}

var ErrServerAddrRequired = errors.New("nacos: server_addr is required")
var ErrDataIDRequired = errors.New("nacos: data_id is required")
var ErrConfigNotFound = errors.New("nacos: config not found")
