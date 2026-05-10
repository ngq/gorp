// Package local provides local file config source implementation for gorp framework.
// Reads config/*.yaml files with layered overlay support (base + env + env directory).
// Does not support hot update (file changes require manual Reload).
//
// 本地文件配置源包，提供 gorp 框架的本地配置源实现。
// 读取 config/*.yaml 文件，支持分层覆盖（基础 + 环境 + 环境目录）。
// 不支持热更新（文件变更需要手动 Reload）。
package local

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	configprovider "github.com/ngq/gorp/framework/provider/config"
	"github.com/spf13/viper"
)

// Provider provides local file config source implementation.
// Core logic: Bind Source factory, delegate loading to Source.Load.
//
// Provider 提供本地文件配置源实现。
// 核心逻辑：绑定 Source 工厂、委托加载给 Source.Load。
type Provider struct{}

// NewProvider creates a new local config source provider.
//
// NewProvider 创建新的本地配置源 provider。
func NewProvider() *Provider { return &Provider{} }

// Name returns provider name for identification.
//
// Name 返回 provider 名称，用于标识。
func (p *Provider) Name() string  { return "configsource.local" }

// IsDefer indicates local config source should defer loading.
// Allows other providers to register before config source initialization.
//
// IsDefer 表示本地配置源应延迟加载。
// 允许其他 provider 在配置源初始化之前注册。
func (p *Provider) IsDefer() bool { return true }

// Provides returns the capability keys this provider exposes.
// Exposes ConfigSourceKey for config source abstraction.
//
// Provides 返回 provider 暴露的能力键。
// 暴露 ConfigSourceKey 用于配置源抽象。
func (p *Provider) Provides() []string {
	return []string{datacontract.ConfigSourceKey}
}

// Register binds the local config source factory to the container.
// Core logic: Create Source instance with config, bind to container.
//
// Register 将本地配置源工厂绑定到容器。
// 核心逻辑：创建 Source 实例并绑定到容器。
func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(datacontract.ConfigSourceKey, func(c runtimecontract.Container) (any, error) {
		cfg, _ := getConfigSourceConfig(c)
		return NewSource(cfg), nil
	}, true)

	return nil
}

// Boot initializes the local config source provider.
// No additional startup logic required.
//
// Boot 初始化本地配置源 provider。
// 无需额外启动逻辑。
func (p *Provider) Boot(c runtimecontract.Container) error {
	return nil
}

// getConfigSourceConfig 从容器获取配置源配置。
func getConfigSourceConfig(c runtimecontract.Container) (*datacontract.ConfigSourceConfig, error) {
	cfgAny, err := c.Make(datacontract.ConfigKey)
	if err != nil {
		return &datacontract.ConfigSourceConfig{Type: datacontract.ConfigSourceLocal}, nil
	}

	cfg, ok := cfgAny.(datacontract.Config)
	if !ok {
		return &datacontract.ConfigSourceConfig{Type: datacontract.ConfigSourceLocal}, nil
	}

	sourceCfg := &datacontract.ConfigSourceConfig{Type: datacontract.ConfigSourceLocal}
	if typ := configprovider.GetStringAny(cfg,
		"configsource.type",
		"config_source.type",
	); typ != "" {
		sourceCfg.Type = datacontract.ConfigSourceType(typ)
	}

	return sourceCfg, nil
}

// Source is the local file config source implementation.
// Reuses framework/provider/config local config loading chain.
// Core logic: Delegate to LoadLocalConfigToViper for consistent loading.
//
// Source 是本地文件配置源实现。
// 复用 framework/provider/config 的本地配置加载主链。
// 核心逻辑：委托给 LoadLocalConfigToViper 实现一致加载。
type Source struct {
	cfg *datacontract.ConfigSourceConfig
	env string
}

// NewSource 创建本地配置源。
func NewSource(cfg *datacontract.ConfigSourceConfig) *Source {
	env := configprovider.NormalizeEnv(os.Getenv("APP_ENV"))
	if env == "" {
		env = configprovider.EnvDev
	}
	return &Source{cfg: cfg, env: env}
}

// Load loads configuration from local files.
// Core logic: Reuse LoadLocalConfigToViper, return all settings.
//
// Load 从本地文件加载配置。
// 核心逻辑：复用 LoadLocalConfigToViper，返回所有配置项。
func (s *Source) Load(ctx context.Context) (map[string]any, error) {
	root := projectRoot()
	v := viper.New()
	if err := configprovider.LoadLocalConfigToViper(v, s.env, root); err != nil {
		return nil, err
	}
	return v.AllSettings(), nil
}

// Get retrieves a single configuration item.
// Local files require full loading before key lookup.
// Core logic: Load all config, traverse nested keys.
//
// Get 获取单个配置项。
// 本地文件需要全部加载后才能查询键。
// 核心逻辑：加载全部配置、遍历嵌套键。
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

// Set sets a single configuration item (not supported for local files).
// Direct YAML file modification required instead.
//
// Set 设置单个配置项（本地文件不支持）。
// 需直接修改 YAML 文件。
func (s *Source) Set(ctx context.Context, key string, value any) error {
	return errors.New("configsource.local: Set not supported, please modify yaml file directly")
}

// Watch watches configuration changes (not supported for local files).
// Use Reload to update configuration manually.
//
// Watch 监听配置变化（本地文件不支持）。
// 使用 Reload 手动更新配置。
func (s *Source) Watch(ctx context.Context, key string) (datacontract.ConfigWatcher, error) {
	return nil, errors.New("configsource.local: Watch not supported, use Reload to update config")
}

// Close closes the config source (no cleanup required for local files).
//
// Close 关闭配置源（本地文件无需清理）。
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
