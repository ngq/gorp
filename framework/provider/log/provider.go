// Package log provides zap-based logging service for gorp framework.
// Supports multiple output drivers: stdout, single file, rotating file.
// Configurable log level, format, and rotation policy.
//
// 日志包提供基于 zap 的日志服务，用于 gorp 框架。
// 支持多种输出驱动：stdout、单文件、滚动文件。
// 可配置日志级别、格式和滚动策略。
package log

import (
	"path/filepath"
	"strings"
	"time"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	configprovider "github.com/ngq/gorp/framework/provider/config"
)

// Provider registers the logging service contract.
// Core logic: Read log config, create zap logger with sink, bind to container.
//
// Provider 注册日志服务契约。
// 核心逻辑：读取日志配置、创建带 sink 的 zap logger、绑定到容器。
type Provider struct{}

// NewProvider creates a new log provider.
//
// NewProvider 创建新的日志 provider。
func NewProvider() *Provider { return &Provider{} }

// Name returns provider name for identification.
//
// Name 返回 provider 名称，用于标识。
func (p *Provider) Name() string { return "log" }

// IsDefer indicates log provider should not defer loading.
// Must be available early for other providers to log.
//
// IsDefer 表示日志 provider 不应延迟加载。
// 必须尽早可用以便其他 provider 记录日志。
func (p *Provider) IsDefer() bool { return false }

// Provides returns the capability keys this provider exposes.
// Exposes LogKey for logging service.
//
// Provides 返回 provider 暴露的能力键。
// 暴露 LogKey 用于日志服务。
func (p *Provider) Provides() []string { return []string{observabilitycontract.LogKey} }

// DependsOn returns the keys this provider depends on.
// Log provider depends on Config for log configuration.
//
// DependsOn 返回该 provider 依赖的 key。
// Log provider 依赖 Config 获取日志配置。
func (p *Provider) DependsOn() []string { return []string{datacontract.ConfigKey} }

// Register binds the log factory to the container.
// Core logic: Read config, configure zap logger with sink, bind singleton.
//
// Register 将日志工厂绑定到容器。
// 核心逻辑：读取配置、配置带 sink 的 zap logger、绑定单例。
func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(observabilitycontract.LogKey, func(c runtimecontract.Container) (any, error) {
		var cfg datacontract.Config
		if c.IsBind(datacontract.ConfigKey) {
			if v, err := c.Make(datacontract.ConfigKey); err == nil {
				cfg, _ = v.(datacontract.Config)
			}
		}

		level := "info"
		format := "console"
		driver := "stdout"
		file := ""
		rotatePattern := ""
		rotateMaxAge := "168h"
		rotateTime := "24h"
		ljMaxSize := 100
		ljMaxBackups := 7
		ljMaxAgeDays := 7
		ljCompress := false
		ljLocalTime := true

		if cfg != nil {
			if s := cfg.GetString("log.level"); s != "" {
				level = s
			}
			if s := cfg.GetString("log.format"); s != "" {
				format = s
			}
			if s := cfg.GetString("log.driver"); s != "" {
				driver = s
			}
			if s := cfg.GetString("log.file"); s != "" {
				file = s
			}
			if s := cfg.GetString("log.rotate_pattern"); s != "" {
				rotatePattern = s
			}
			if s := cfg.GetString("log.rotate_max_age"); s != "" {
				rotateMaxAge = s
			}
			if s := cfg.GetString("log.rotate_time"); s != "" {
				rotateTime = s
			}
			if v := cfg.GetInt("log.max_size_mb"); v > 0 {
				ljMaxSize = v
			}
			if v := cfg.GetInt("log.max_backups"); v > 0 {
				ljMaxBackups = v
			}
			if v := cfg.GetInt("log.max_age_days"); v > 0 {
				ljMaxAgeDays = v
			}
			if v, ok := configprovider.GetBoolAny(cfg, "log.compress"); ok {
				ljCompress = v
			}
			if v, ok := configprovider.GetBoolAny(cfg, "log.local_time"); ok {
				ljLocalTime = v
			}
		}

		level = strings.ToLower(level)
		format = strings.ToLower(format)
		driver = strings.ToLower(driver)

		if driver != "stdout" {
			if file == "" {
				if rootAny, err := c.Make(runtimecontract.RootKey); err == nil {
					if rootSvc, ok := rootAny.(runtimecontract.Root); ok {
						file = filepath.Join(rootSvc.LogPath(), "gorp.log")
					}
				}
			}
			if file == "" {
				file = filepath.Join(".", "gorp.log")
			}
		}
		if rotatePattern == "" && file != "" {
			rotatePattern = file + ".%Y%m%d"
		}

		sink := SinkConfig{
			Driver:        driver,
			Filename:      file,
			RotatePattern: rotatePattern,
			MaxSizeMB:     ljMaxSize,
			MaxBackups:    ljMaxBackups,
			MaxAgeDays:    ljMaxAgeDays,
			Compress:      ljCompress,
			LocalTime:     ljLocalTime,
		}
		if d, err := time.ParseDuration(rotateMaxAge); err == nil {
			sink.RotateMaxAge = d
		}
		if d, err := time.ParseDuration(rotateTime); err == nil {
			sink.RotateTime = d
		}

		return newZapLoggerWithSink(level, format, sink)
	}, true)
	return nil
}

// Boot initializes the log provider.
// No additional startup logic required.
//
// Boot 初始化日志 provider。
// 无需额外启动逻辑。
func (p *Provider) Boot(runtimecontract.Container) error { return nil }
