// Package baseconfigsource provides config reading helpers for ConfigSource providers.
// These utilities standardize config key fallback patterns and eliminate
// the cfg.Get() + cfg.GetString() double-read redundancy found in
// Apollo, Nacos, and Polaris providers.
//
// 本包提供 ConfigSource provider 的配置读取辅助工具。
// 这些工具标准化配置键回退模式，消除 Apollo、Nacos、Polaris 中
// cfg.Get() + cfg.GetString() 的双读冗余。
package baseconfigsource

import (
	"errors"
	"time"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	configprovider "github.com/ngq/gorp/framework/provider/config"
)

// ReadConfig extracts the Config service from the container with unified validation.
//
// ReadConfig 从容器提取 Config 服务，统一验证逻辑。
func ReadConfig(c interface {
	Make(key string) (any, error)
}) (datacontract.Config, error) {
	cfgAny, err := c.Make(datacontract.ConfigKey)
	if err != nil {
		return nil, err
	}
	cfg, ok := cfgAny.(datacontract.Config)
	if !ok {
		return nil, errors.New("invalid config service")
	}
	return cfg, nil
}

// GetStringFallback reads a string config value with standardized key fallback.
// Primary key: configsource.<name>.<field>
// Fallback key: config.<name>.<field>
//
// GetStringFallback 读取字符串配置值，标准化 key 回退。
// 主路径: configsource.<name>.<field>
// 回退路径: config.<name>.<field>
func GetStringFallback(cfg datacontract.Config, name, field string) string {
	return configprovider.GetStringAny(cfg,
		"configsource."+name+"."+field,
		"config."+name+"."+field,
	)
}

// GetIntFallback reads an int config value with standardized key fallback.
//
// GetIntFallback 读取整数配置值，标准化 key 回退。
func GetIntFallback(cfg datacontract.Config, name, field string) int {
	return configprovider.GetIntAny(cfg,
		"configsource."+name+"."+field,
		"config."+name+"."+field,
	)
}

// GetBoolFallback reads a bool config value with standardized key fallback.
//
// GetBoolFallback 读取布尔配置值，标准化 key 回退。
func GetBoolFallback(cfg datacontract.Config, name, field string) (bool, bool) {
	return configprovider.GetBoolAny(cfg,
		"configsource."+name+"."+field,
		"config."+name+"."+field,
	)
}

// GetDurationSecondsFallback reads a seconds config value and converts to time.Duration.
// Returns 0 if the key is not set or the value is non-positive.
//
// GetDurationSecondsFallback 读取秒数配置并转为 time.Duration。
// 如果 key 未设置或值非正，返回 0。
func GetDurationSecondsFallback(cfg datacontract.Config, name, field string) time.Duration {
	if seconds := GetIntFallback(cfg, name, field); seconds > 0 {
		return time.Duration(seconds) * time.Second
	}
	return 0
}

// GetDurationMillisFallback reads a milliseconds config value and converts to time.Duration.
// Returns 0 if the key is not set or the value is non-positive.
//
// GetDurationMillisFallback 读取毫秒数配置并转为 time.Duration。
// 如果 key 未设置或值非正，返回 0。
func GetDurationMillisFallback(cfg datacontract.Config, name, field string) time.Duration {
	if ms := GetIntFallback(cfg, name, field); ms > 0 {
		return time.Duration(ms) * time.Millisecond
	}
	return 0
}
