package local

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ngq/gorp/framework/contract"
	configprovider "github.com/ngq/gorp/framework/provider/config"
	"github.com/spf13/viper"
)

// Provider 提供本地文件配置源实现。
//
// 中文说明：
// - 读取 config/*.yaml 文件；
// - 支持多层覆盖（base + env + env 目录）；
// - 不支持热更新（文件变更需要手动 Reload）。
type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string  { return "configsource.local" }
func (p *Provider) IsDefer() bool { return true }
func (p *Provider) Provides() []string {
	return []string{contract.ConfigSourceKey}
}

func (p *Provider) Register(c contract.Container) error {
	c.Bind(contract.ConfigSourceKey, func(c contract.Container) (any, error) {
		cfg, _ := getConfigSourceConfig(c)
		return NewSource(cfg), nil
	}, true)

	return nil
}

func (p *Provider) Boot(c contract.Container) error {
	return nil
}

// getConfigSourceConfig 从容器获取配置源配置。
func getConfigSourceConfig(c contract.Container) (*contract.ConfigSourceConfig, error) {
	cfgAny, err := c.Make(contract.ConfigKey)
	if err != nil {
		return &contract.ConfigSourceConfig{Type: contract.ConfigSourceLocal}, nil
	}

	cfg, ok := cfgAny.(contract.Config)
	if !ok {
		return &contract.ConfigSourceConfig{Type: contract.ConfigSourceLocal}, nil
	}

	sourceCfg := &contract.ConfigSourceConfig{Type: contract.ConfigSourceLocal}
	if typ := configprovider.GetStringAny(cfg,
		"configsource.type",
		"config_source.type",
	); typ != "" {
		sourceCfg.Type = contract.ConfigSourceType(typ)
	}

	return sourceCfg, nil
}

// Source 是本地文件配置源实现。
//
// 中文说明：
// - 复用 framework/provider/config 的本地配置加载主链；
// - 不支持 Watch 热更新。
type Source struct {
	cfg *contract.ConfigSourceConfig
	env string
}

// NewSource 创建本地配置源。
func NewSource(cfg *contract.ConfigSourceConfig) *Source {
	env := configprovider.NormalizeEnv(os.Getenv("APP_ENV"))
	if env == "" {
		env = configprovider.EnvDev
	}
	return &Source{cfg: cfg, env: env}
}

// Load 从本地文件加载配置。
//
// 中文说明：
// - 统一复用 config.LoadLocalConfigToViper，避免两套本地 YAML 加载逻辑漂移。
func (s *Source) Load(ctx context.Context) (map[string]any, error) {
	root := projectRoot()
	v := viper.New()
	if err := configprovider.LoadLocalConfigToViper(v, s.env, root); err != nil {
		return nil, err
	}
	return v.AllSettings(), nil
}

// Get 获取单个配置项。
//
// 中文说明：
// - 本地文件不支持单独读取，需要全部加载后查询。
func (s *Source) Get(ctx context.Context, key string) (any, error) {
	cfg, err := s.Load(ctx)
	if err != nil {
		return nil, err
	}

	keys := strings.Split(key, ".")
	var current any = cfg
	for _, k := range keys {
		if m, ok := current.(map[string]any); ok {
			if v, exists := m[k]; exists {
				current = v
			} else {
				return nil, fmt.Errorf("config: key %s not found", key)
			}
		} else {
			return nil, fmt.Errorf("config: cannot traverse key %s", key)
		}
	}
	return current, nil
}

// Set 设置单个配置项。
func (s *Source) Set(ctx context.Context, key string, value any) error {
	return errors.New("configsource.local: Set not supported, please modify yaml file directly")
}

// Watch 监听配置变化。
func (s *Source) Watch(ctx context.Context, key string) (contract.ConfigWatcher, error) {
	return nil, errors.New("configsource.local: Watch not supported, use Reload to update config")
}

// Close 关闭配置源。
func (s *Source) Close() error { return nil }

// projectRoot 返回本地配置源的项目根目录。
//
// 中文说明：
// - 与 config provider 共享 APP_BASE_PATH 约定；
// - 保留在 local 包内，避免跨包访问非导出 helper。
func projectRoot() string {
	if base := strings.TrimSpace(os.Getenv("APP_BASE_PATH")); base != "" {
		if filepath.IsAbs(base) {
			return filepath.Clean(base)
		}
		wd, _ := os.Getwd()
		return filepath.Clean(filepath.Join(wd, base))
	}
	wd, _ := os.Getwd()
	return wd
}
