package log

import (
	"path/filepath"
	"strings"
	"time"

	"github.com/ngq/gorp/framework/contract"
	"github.com/ngq/gorp/framework/provider/app"
)

// Provider 把日志服务注册进容器。
//
// 中文说明：
// - 对外统一暴露 contract.LogKey，业务层不需要直接依赖 zap。
// - 日志配置项较多，因此在 Register 中集中做默认值解析与驱动选择。
// - provider 本身不延迟加载，保证框架启动后日志能力立即可用。
type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string { return "log" }
func (p *Provider) IsDefer() bool { return false }
func (p *Provider) Provides() []string { return []string{contract.LogKey} }

func (p *Provider) Register(c contract.Container) error {
	c.Bind(contract.LogKey, func(c contract.Container) (any, error) {
		var cfg contract.Config
		if c.IsBind(contract.ConfigKey) {
			if v, err := c.Make(contract.ConfigKey); err == nil {
				cfg, _ = v.(contract.Config)
			}
		}

		level := "info"
		format := "console"
		driver := "stdout" // stdout|single|rotate
		file := ""
		rotatePattern := ""
		rotateMaxAge := "168h" // 7d
		rotateTime := "24h"
		ljMaxSize := 100
		ljMaxBackups := 7
		ljMaxAgeDays := 7
		ljCompress := false
		ljLocalTime := true

		// 中文说明：
		// - 先定义一套框架默认值，确保即使没有 log 配置也能正常输出。
		// - 后面如果配置中心存在对应键，再逐项覆盖这些默认值。
		// - 这种写法比直接依赖配置文件完整性更稳健。
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
			if v := cfg.GetBool("log.compress"); v {
				ljCompress = v
			}
			if v := cfg.GetBool("log.local_time"); v {
				ljLocalTime = v
			}
		}

		level = strings.ToLower(level)
		format = strings.ToLower(format)
		driver = strings.ToLower(driver)

		// resolve log file path default via app service (app log path)
		//
		// 中文说明：
		// - driver != stdout 时，日志需要落盘；默认路径优先从 app 服务提供的 `LogPath()` 推导。
		// - 这样 framework 的日志默认位置由宿主路径服务决定，而不是由 log provider 自己硬编码项目目录结构。
		if driver != "stdout" {
			if file == "" {
				if appAny, err := c.Make(app.AppKey); err == nil {
					if appSvc, ok := appAny.(app.App); ok {
						file = filepath.Join(appSvc.LogPath(), "gorp.log")
					}
				}
			}
			if file == "" {
				// 中文说明：
				// - 正常情况下，bootstrap 会始终注册 app provider，因此这里大多数时候都会命中 appSvc.LogPath()。
				// - 只有在极少数未注册 app provider 的宿主里，才退化到一个最小、与项目结构无关的相对路径默认值。
				file = filepath.Join(".", "gorp.log")
			}
		}
		if rotatePattern == "" && file != "" {
			rotatePattern = file + ".%Y%m%d"
		}

		cfgSink := sinkConfig{Driver: driver, Filename: file, RotatePattern: rotatePattern, MaxSizeMB: ljMaxSize, MaxBackups: ljMaxBackups, MaxAgeDays: ljMaxAgeDays, Compress: ljCompress, LocalTime: ljLocalTime}
		if d, err := time.ParseDuration(rotateMaxAge); err == nil {
			cfgSink.RotateMaxAge = d
		}
		if d, err := time.ParseDuration(rotateTime); err == nil {
			cfgSink.RotateTime = d
		}

		return NewZapLoggerWithSink(level, format, cfgSink)
	}, true)
	return nil
}

func (p *Provider) Boot(contract.Container) error { return nil }
