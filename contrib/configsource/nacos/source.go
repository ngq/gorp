// Package nacos provides Nacos configuration source implementation.
// This file implements the ConfigSource contract with cache and watcher support.
//
// 本包提供 Nacos 配置源实现。
// 本文件实现 ConfigSource 契约，包含缓存和监听支持。
package nacos

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	datacontract "github.com/ngq/gorp/framework/contract/data"
)

// ErrServerAddrRequired indicates Nacos server_addr is required.
//
// ErrServerAddrRequired 表示 Nacos server_addr 必需。
var ErrServerAddrRequired = errors.New("configsource.nacos: server_addr is required")

// ErrDataIDRequired indicates Nacos data_id is required.
//
// ErrDataIDRequired 表示 Nacos data_id 必需。
var ErrDataIDRequired = errors.New("configsource.nacos: data_id is required")

// ErrConfigNotFound indicates Nacos config not found.
//
// ErrConfigNotFound 表示 Nacos 配置未找到。
var ErrConfigNotFound = errors.New("configsource.nacos: config not found")

// defaultNacosPollInterval is the default polling interval for config updates.
//
// defaultNacosPollInterval 是配置更新的默认轮询间隔。
const defaultNacosPollInterval = 5 * time.Second

// nacosConfigClient defines the internal client interface for Nacos operations.
//
// nacosConfigClient 定义 Nacos 操作的内部客户端接口。
// 该接口将 SDK 调用与配置源逻辑解耦，便于测试和替换实现。
type nacosConfigClient interface {
	// GetConfig 从 Nacos 获取配置内容。
	GetConfig(ctx context.Context, cfg *NacosConfig) (string, error)
	// PublishConfig 发布配置内容到 Nacos。
	PublishConfig(ctx context.Context, cfg *NacosConfig, content string) error
	// WatchConfig 监听配置变更，onUpdate 在变更时被调用。
	WatchConfig(ctx context.Context, cfg *NacosConfig, onUpdate func(string)) error
}

// ConfigSource Nacos 配置源实现。
// Implements datacontract.ConfigSource with cache, set support and watch support.
//
// ConfigSource 实现 datacontract.ConfigSource。
// 包含缓存、set 支持和 watch 支持。
// 与 Apollo 不同，Nacos 支持 Set 操作，可以回写配置。
type ConfigSource struct {
	// config 保存 Nacos 连接和命名空间配置。
	config *NacosConfig
	// client 是底层 Nacos SDK 客户端包装。
	client nacosConfigClient
	// mu 保护 cache 的读写访问。
	mu sync.RWMutex
	// cache 缓存从 Nacos 加载的配置数据。
	cache map[string]any
	// watchers 记录所有活跃的配置监听器。
	watchers sync.Map
	// closed 标记配置源是否已关闭。
	closed bool
	// closeMu 保护 closed 状态变更。
	closeMu sync.Mutex
}

// NewConfigSource creates a new Nacos config source with default official client.
//
// NewConfigSource 使用默认官方客户端创建新的 Nacos 配置源。
// 该函数从配置创建 SDK 客户端，并初始化配置源实例。
func NewConfigSource(cfg *NacosConfig) (*ConfigSource, error) {
	client, err := newOfficialNacosClient(cfg)
	if err != nil {
		return nil, err
	}
	return NewConfigSourceWithClient(cfg, client)
}

// NewConfigSourceWithClient creates a new Nacos config source with custom client.
//
// NewConfigSourceWithClient 使用自定义客户端创建新的 Nacos 配置源。
// 该函数主要用于测试场景，允许注入 mock 或 fake 客户端。
// 参数：
//   - cfg: Nacos 配置，必须包含 ServerAddr 和 DataID
//   - client: Nacos 客户端实现，不能为 nil
func NewConfigSourceWithClient(cfg *NacosConfig, client nacosConfigClient) (*ConfigSource, error) {
	// 验证必需配置项
	if cfg.ServerAddr == "" {
		return nil, ErrServerAddrRequired
	}
	if cfg.DataID == "" {
		return nil, ErrDataIDRequired
	}
	if client == nil {
		return nil, errors.New("configsource.nacos: config client is required")
	}

	return &ConfigSource{
		config: cfg,
		client: client,
		cache:  make(map[string]any),
	}, nil
}

// Load loads the full config snapshot from Nacos.
// Implements datacontract.ConfigSource.Load.
//
// Load 从 Nacos 加载完整配置快照。
// 实现 datacontract.ConfigSource.Load。
// 该方法从 Nacos 获取配置内容，解析为 map，并更新内部缓存。
func (s *ConfigSource) Load(ctx context.Context) (map[string]any, error) {
	// 从 Nacos SDK 获取配置内容
	content, err := s.client.GetConfig(ctx, s.config)
	if err != nil {
		return nil, err
	}

	// 解析配置内容为结构化数据
	loaded, err := decodeContent(content, s.config.DataID)
	if err != nil {
		return nil, err
	}

	// 更新缓存并返回拷贝（防止外部修改影响缓存）
	s.mu.Lock()
	s.cache = cloneMap(loaded)
	s.mu.Unlock()

	return cloneMap(loaded), nil
}

// Get reads a single config value from the source.
// Implements datacontract.ConfigSource.Get.
//
// Get 从配置源读取单个配置值。
// 实现 datacontract.ConfigSource.Get。
// 支持点分隔的嵌套路径查找，如 "app.name" 查找 app.name 字段。
func (s *ConfigSource) Get(ctx context.Context, key string) (any, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 查找嵌套值
	value, ok := lookupNestedValue(s.cache, key)
	if !ok {
		return nil, fmt.Errorf("configsource.nacos: key %s not found", key)
	}
	return value, nil
}

// Set writes a config value back to Nacos server.
// Implements datacontract.ConfigSource.Set.
//
// Set 回写配置值到 Nacos 服务器。
// 实现 datacontract.ConfigSource.Set。
// 注意：Nacos 支持 Set 操作，但只能对当前 DataID 进行整体更新，
// 不支持单 key 粒度的更新。传入空 key 或等于 DataID 的 key 表示整体更新。
func (s *ConfigSource) Set(ctx context.Context, key string, value any) error {
	// 验证 key 必须为空或等于当前 DataID
	// Nacos SDK 只支持整体配置更新，不支持单 key 更新
	if key != "" && key != s.config.DataID {
		return fmt.Errorf("configsource.nacos: set only supports data_id %s", s.config.DataID)
	}

	// 将值编码为配置内容（YAML 或纯文本）
	content, err := encodeContent(value)
	if err != nil {
		return err
	}

	// 发布配置到 Nacos
	if err := s.client.PublishConfig(ctx, s.config, content); err != nil {
		return err
	}

	// 重新解析并更新缓存
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
//
// Underlying 返回此配置源使用的当前原生客户端对象。
// 该方法用于暴露底层 Nacos SDK 客户端，允许高级用户直接访问 SDK 功能。
// 返回的是 configclient.IConfigClient 实例。
func (s *ConfigSource) Underlying() any {
	return unwrapNacosNativeClient(s.client)
}

// As projects the native client into the requested target when possible.
//
// As 在可能时将当前原生客户端投射到请求的目标。
// 该方法实现了下探机制，允许用户将原生客户端转换为特定类型。
// 例如：var client configclient.IConfigClient; source.As(&client)
func (s *ConfigSource) As(target any) bool {
	return As(unwrapNacosNativeClient(s.client), target)
}

// Watch subscribes to source-side changes of a config key.
// Implements datacontract.ConfigSource.Watch.
//
// Watch 订阅配置源侧的指定 key 变更。
// 实现 datacontract.ConfigSource.Watch。
// Nacos 通过 SDK 的 ListenConfig 实现 watch，配置变更时回调 onUpdate。
// 返回的 ConfigWatcher 可用于注册多个 key 的变更回调。
func (s *ConfigSource) Watch(ctx context.Context, key string) (datacontract.ConfigWatcher, error) {
	s.closeMu.Lock()
	defer s.closeMu.Unlock()

	// 检查配置源是否已关闭
	if s.closed {
		return nil, errors.New("configsource.nacos: config source closed")
	}

	// 创建可取消的监听上下文
	watchCtx, cancel := context.WithCancel(ctx)
	watcher := &nacosWatcher{
		ctx:       watchCtx,
		cancel:    cancel,
		source:    s,
		callbacks: &sync.Map{},
	}
	s.watchers.Store(watcher, struct{}{})

	// 启动后台监听 goroutine
	go func() {
		err := s.client.WatchConfig(watchCtx, s.config, func(content string) {
			// 解析新配置内容
			loaded, decodeErr := decodeContent(content, s.config.DataID)
			if decodeErr != nil {
				return
			}
			// 更新缓存
			s.mu.Lock()
			s.cache = cloneMap(loaded)
			s.mu.Unlock()
			// 通知所有注册的回调
			watcher.dispatch()
		})
		if err != nil {
			cancel()
		}
	}()

	return watcher, nil
}

// Close releases resources held by the source.
// Implements datacontract.ConfigSource.Close.
//
// Close 释放配置源持有的资源。
// 实现 datacontract.ConfigSource.Close。
// 该方法会停止所有活跃的监听器，并关闭 Nacos SDK 客户端连接。
func (s *ConfigSource) Close() error {
	s.closeMu.Lock()
	defer s.closeMu.Unlock()

	// 避免重复关闭
	if s.closed {
		return nil
	}
	s.closed = true

	// 停止所有监听器
	s.watchers.Range(func(key, value any) bool {
		if watcher, ok := key.(*nacosWatcher); ok {
			_ = watcher.Stop()
		}
		return true
	})

	// 关闭底层客户端（如果支持）
	if closer, ok := s.client.(interface{ CloseClient() }); ok {
		closer.CloseClient()
	}
	return nil
}
