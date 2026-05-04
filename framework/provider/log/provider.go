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

type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string       { return "log" }
func (p *Provider) IsDefer() bool      { return false }
func (p *Provider) Provides() []string { return []string{observabilitycontract.LogKey} }

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

func (p *Provider) Boot(runtimecontract.Container) error { return nil }
