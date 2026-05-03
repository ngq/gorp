package log

import (
	"path/filepath"
	"strings"
	"time"

	"github.com/ngq/gorp/framework/contract"
	configprovider "github.com/ngq/gorp/framework/provider/config"
)

// Provider 把日志服务注册进容器。
//
// 中文说明：
// - 对外统一暴露 contract.LogKey，业务层不需要直接依赖 zap；
// - zap 是 framework 的必需依赖核（P1），直接内化在本包（zap_backend.go），不再经由 contrib/log/zap；
// - 业务若需深度定制 zap（多 sink、动态 level、自定义 encoder），使用 contrib/log/zap 扩展层注册替换；
// - 日志配置项较多，在 Register 中集中做默认值解析与驱动选择；
// - 当前阶段正式冻结的口径是：
//  1. 对外承诺面优先是"统一日志能力"
//  2. 文件路径推导仍属于 runtime convention 语义的一部分
//  3. 当 file/rotate 且未显式提供 `log.file` 时：优先走 `contract.Root` 的 `LogPath()`，否则最小回退到 `./gorp.log`
type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string      { return "log" }
func (p *Provider) IsDefer() bool     { return false }
func (p *Provider) Provides() []string { return []string{contract.LogKey} }

// Register 绑定统一日志服务。
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
				if rootAny, err := c.Make(contract.RootKey); err == nil {
					if rootSvc, ok := rootAny.(contract.Root); ok {
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

// Boot log provider 无额外启动逻辑。
func (p *Provider) Boot(contract.Container) error { return nil }
