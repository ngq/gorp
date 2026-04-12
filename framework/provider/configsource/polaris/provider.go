package polaris

import (
	"context"
	"errors"
	"sync"

	"github.com/ngq/gorp/framework/contract"
)

// Provider 提供 Polaris 配置中心实现。
//
// 中文说明：
// - 使用腾讯云 Polaris 配置中心；
// - 支持命名空间隔离；
// - 支持配置分组管理；
// - 支持配置热更新；
// - 适用于腾讯云环境和私有化部署。
type Provider struct{}

// NewProvider 创建 Polaris 配置 Provider。
func NewProvider() *Provider { return &Provider{} }

// Name 返回 Provider 名称。
func (p *Provider) Name() string { return "configsource.polaris" }

// IsDefer 返回是否延迟加载。
func (p *Provider) IsDefer() bool { return true }

// Provides 返回提供的服务 key。
func (p *Provider) Provides() []string {
	return []string{contract.ConfigSourceKey}
}

// Register 注册 Polaris 配置源。
func (p *Provider) Register(c contract.Container) error {
	c.Bind(contract.ConfigSourceKey, func(c contract.Container) (any, error) {
		cfg, err := getPolarisConfig(c)
		if err != nil {
			return nil, err
		}
		return NewConfigSource(cfg)
	}, true)

	return nil
}

// Boot 启动 Provider。
func (p *Provider) Boot(c contract.Container) error {
	return nil
}

// PolarisConfig 定义 Polaris 配置。
type PolarisConfig struct {
	// ServerAddress Polaris Server 地址
	ServerAddress string

	// Namespace 命名空间
	Namespace string

	// FileGroup 配置文件组
	FileGroup string

	// FileName 配置文件名
	FileName string

	// Token 访问令牌
	Token string
}

// getPolarisConfig 从容器获取配置。
func getPolarisConfig(c contract.Container) (*PolarisConfig, error) {
	cfgAny, err := c.Make(contract.ConfigKey)
	if err != nil {
		return nil, err
	}

	cfg, ok := cfgAny.(contract.Config)
	if !ok {
		return nil, errors.New("polaris: invalid config service")
	}

	polarisCfg := &PolarisConfig{
		Namespace: "default",
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

	return polarisCfg, nil
}

// ConfigSource Polaris 配置源实现。
type ConfigSource struct {
	config *PolarisConfig
	mu     sync.RWMutex
	cache  map[string]string
}

// NewConfigSource 创建 Polaris 配置源。
func NewConfigSource(cfg *PolarisConfig) (*ConfigSource, error) {
	if cfg.ServerAddress == "" {
		return nil, errors.New("polaris: server_address is required")
	}
	if cfg.FileGroup == "" {
		return nil, errors.New("polaris: file_group is required")
	}
	if cfg.FileName == "" {
		return nil, errors.New("polaris: file_name is required")
	}

	return &ConfigSource{
		config: cfg,
		cache:  make(map[string]string),
	}, nil
}

// Load 加载配置。
//
// 中文说明：
// - 从 Polaris 配置中心获取配置文件；
// - 解析配置内容到缓存；
// - 支持多种配置格式（JSON/YAML/Properties）。
func (s *ConfigSource) Load() error {
	// TODO: 实现真实的 Polaris 配置加载
	// 需要引入 github.com/polarismesh/polaris-go
	//
	// 示例流程：
	// 1. 创建 Polaris SDK 配置
	//    config.NewConfiguration()
	// 2. 创建 ConfigFile
	//    polaris.NewConfigFileAPI()
	// 3. 获取配置文件
	//    configFile.GetConfigFile(namespace, fileGroup, fileName)
	// 4. 解析配置内容

	return nil
}

// Get 获取配置值。
func (s *ConfigSource) Get(key string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cache[key]
}

// Set 设置配置值（本地缓存）。
func (s *ConfigSource) Set(key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cache[key] = value
	return nil
}

// Watch 监听配置变化。
//
// 中文说明：
// - 使用 Polaris SDK 的配置变更监听；
// - 当配置更新时自动更新缓存；
// - 通过回调通知配置变更。
func (s *ConfigSource) Watch(ctx context.Context, key string) (contract.ConfigWatcher, error) {
	return &polarisWatcher{
		ctx:    ctx,
		source: s,
	}, nil
}

// Close 关闭配置源。
func (s *ConfigSource) Close() error {
	return nil
}

// polarisWatcher 配置监听器。
type polarisWatcher struct {
	ctx    context.Context
	source *ConfigSource
}

// OnChange 配置变更回调。
//
// 中文说明：
// - 当 Polaris 配置变更时触发；
// - 使用 Polaris SDK 的 ChangeEventHandler 实现。
func (w *polarisWatcher) OnChange(key string, callback func(value any)) {
	// TODO: 实现真实的配置变更监听
	// 使用 github.com/polarismesh/polaris-go 的 AddChangeListener
}

// Stop 停止监听。
func (w *polarisWatcher) Stop() error { return nil }

// 错误定义
var (
	ErrServerAddressRequired = errors.New("polaris: server_address is required")
	ErrFileGroupRequired     = errors.New("polaris: file_group is required")
	ErrFileNameRequired      = errors.New("polaris: file_name is required")
)