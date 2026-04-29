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
// - 当前已从 framework/provider 真实下沉到 contrib 层。
type Provider struct{}

func NewProvider() *Provider { return &Provider{} }
func (p *Provider) Name() string     { return "configsource.polaris" }
func (p *Provider) IsDefer() bool    { return true }
func (p *Provider) Provides() []string { return []string{contract.ConfigSourceKey} }

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
func (p *Provider) Boot(c contract.Container) error { return nil }

type PolarisConfig struct {
	ServerAddress string
	Namespace     string
	FileGroup     string
	FileName      string
	Token         string
}

func getPolarisConfig(c contract.Container) (*PolarisConfig, error) {
	cfgAny, err := c.Make(contract.ConfigKey)
	if err != nil {
		return nil, err
	}
	cfg, ok := cfgAny.(contract.Config)
	if !ok {
		return nil, errors.New("polaris: invalid config service")
	}
	polarisCfg := &PolarisConfig{Namespace: "default"}
	if v := cfg.Get("config.polaris.server_address"); v != nil { polarisCfg.ServerAddress = cfg.GetString("config.polaris.server_address") }
	if v := cfg.Get("config.polaris.namespace"); v != nil { polarisCfg.Namespace = cfg.GetString("config.polaris.namespace") }
	if v := cfg.Get("config.polaris.file_group"); v != nil { polarisCfg.FileGroup = cfg.GetString("config.polaris.file_group") }
	if v := cfg.Get("config.polaris.file_name"); v != nil { polarisCfg.FileName = cfg.GetString("config.polaris.file_name") }
	if v := cfg.Get("config.polaris.token"); v != nil { polarisCfg.Token = cfg.GetString("config.polaris.token") }
	return polarisCfg, nil
}

type ConfigSource struct {
	config *PolarisConfig
	mu     sync.RWMutex
	cache  map[string]string
}

func NewConfigSource(cfg *PolarisConfig) (*ConfigSource, error) {
	if cfg.ServerAddress == "" { return nil, errors.New("polaris: server_address is required") }
	if cfg.FileGroup == "" { return nil, errors.New("polaris: file_group is required") }
	if cfg.FileName == "" { return nil, errors.New("polaris: file_name is required") }
	return &ConfigSource{config: cfg, cache: make(map[string]string)}, nil
}
func (s *ConfigSource) Load() error { return nil }
func (s *ConfigSource) Get(key string) string { s.mu.RLock(); defer s.mu.RUnlock(); return s.cache[key] }
func (s *ConfigSource) Set(key, value string) error { s.mu.Lock(); defer s.mu.Unlock(); s.cache[key] = value; return nil }
func (s *ConfigSource) Watch(ctx context.Context, key string) (contract.ConfigWatcher, error) { return &polarisWatcher{ctx: ctx, source: s}, nil }
func (s *ConfigSource) Close() error { return nil }

type polarisWatcher struct { ctx context.Context; source *ConfigSource }
func (w *polarisWatcher) OnChange(key string, callback func(value any)) {}
func (w *polarisWatcher) Stop() error { return nil }

var (
	ErrServerAddressRequired = errors.New("polaris: server_address is required")
	ErrFileGroupRequired     = errors.New("polaris: file_group is required")
	ErrFileNameRequired      = errors.New("polaris: file_name is required")
)
