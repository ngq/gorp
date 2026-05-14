// Package polaris provides Polaris configuration center provider for the gorp framework.
// This provider implements ConfigSource contract with Polaris SDK integration.
//
// 本包提供 gorp 框架 Polaris 配置中心 provider。
// 本 provider 实现 ConfigSource 契约，集成 Polaris SDK。
package polaris

import (
	"time"

	"github.com/ngq/gorp/contrib/internal/baseconfigsource"
	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

// Provider 提供 Polaris 配置中心实现。
type Provider struct {
	baseconfigsource.BaseConfigSourceProvider
}

// NewProvider creates a new Polaris provider instance.
func NewProvider() *Provider {
	p := &Provider{}
	p.NameStr = "configsource.polaris"
	p.GetConfig = func(c runtimecontract.Container) (any, error) {
		return getPolarisConfig(c)
	}
	p.NewSource = func(cfg any) (datacontract.ConfigSource, error) {
		return NewConfigSource(cfg.(*PolarisConfig))
	}
	return p
}

// defaultPolarisPollInterval is the default polling interval for config updates.
const defaultPolarisPollInterval = 5 * time.Second

// PolarisConfig 定义 Polaris 配置。
type PolarisConfig struct {
	ServerAddress      string
	Namespace          string
	FileGroup          string
	FileName           string
	Token              string
	PollInterval       time.Duration
	WatchRetryInterval time.Duration
}

// getPolarisConfig extracts Polaris configuration from the container's config binding.
// Uses GetStringFallback with configsource.polaris.* as primary path and
// config.polaris.* as fallback, unifying with other ConfigSource providers.
//
// getPolarisConfig 从容器的 config binding 中提取 Polaris 配置。
// 使用 GetStringFallback，以 configsource.polaris.* 为主路径，
// config.polaris.* 为回退路径，与其他 ConfigSource provider 统一。
func getPolarisConfig(c runtimecontract.Container) (*PolarisConfig, error) {
	cfg, err := baseconfigsource.ReadConfig(c)
	if err != nil {
		return nil, err
	}

	polarisCfg := &PolarisConfig{
		Namespace:          "default",
		PollInterval:       defaultPolarisPollInterval,
		WatchRetryInterval: time.Second,
	}

	polarisCfg.ServerAddress = baseconfigsource.GetStringFallback(cfg, "polaris", "server_address")
	polarisCfg.Namespace = baseconfigsource.GetStringFallback(cfg, "polaris", "namespace")
	if polarisCfg.Namespace == "" {
		polarisCfg.Namespace = "default"
	}
	polarisCfg.FileGroup = baseconfigsource.GetStringFallback(cfg, "polaris", "file_group")
	polarisCfg.FileName = baseconfigsource.GetStringFallback(cfg, "polaris", "file_name")
	polarisCfg.Token = baseconfigsource.GetStringFallback(cfg, "polaris", "token")

	if d := baseconfigsource.GetDurationSecondsFallback(cfg, "polaris", "poll_interval_seconds"); d > 0 {
		polarisCfg.PollInterval = d
	}
	if d := baseconfigsource.GetDurationMillisFallback(cfg, "polaris", "watch_retry_interval_ms"); d > 0 {
		polarisCfg.WatchRetryInterval = d
	}

	return polarisCfg, nil
}
