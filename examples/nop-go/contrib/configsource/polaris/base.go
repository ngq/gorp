// Package polaris 提供 Polaris 配置中心的基础能力。
//
// 本文件内联了 baseconfigsource 的公共逻辑，使 polaris 包成为完全独立的模块，
// 不再依赖 contrib/internal 下的任何包。
package polaris

import (
	"errors"
	"reflect"
	"time"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	configprovider "github.com/ngq/gorp/framework/provider/config"
)

// ReadConfig 从容器中获取 Config 服务。
//
// 统一封装 c.Make(ConfigKey) 调用，简化各 ConfigSource provider 的配置读取代码。
func ReadConfig(c runtimecontract.Container) (datacontract.Config, error) {
	cfgAny, err := c.Make(datacontract.ConfigKey)
	if err != nil {
		return nil, err
	}
	cfg, ok := cfgAny.(datacontract.Config)
	if !ok {
		return nil, ErrInvalidConfigService
	}
	return cfg, nil
}

// ErrInvalidConfigService 表示 Config 服务类型无效。
var ErrInvalidConfigService = errors.New("invalid config service")

// GetStringFallback 从配置中读取字符串，支持主路径和回退路径。
//
// 主路径格式：configsource.{provider}.{key}
// 回退路径格式：config.{provider}.{key}
//
// 这种双路径设计让 ConfigSource provider 可以统一读取自己的配置，
// 同时保持与旧配置路径的兼容性。
func GetStringFallback(cfg datacontract.Config, provider, key string) string {
	primary := "configsource." + provider + "." + key
	fallback := "config." + provider + "." + key
	return configprovider.GetStringAny(cfg, primary, fallback)
}

// GetIntFallback 从配置中读取整数，支持主路径和回退路径。
func GetIntFallback(cfg datacontract.Config, provider, key string) int {
	primary := "configsource." + provider + "." + key
	fallback := "config." + provider + "." + key
	return configprovider.GetIntAny(cfg, primary, fallback)
}

// GetDurationSecondsFallback 从配置中读取秒数并转换为 time.Duration。
func GetDurationSecondsFallback(cfg datacontract.Config, provider, key string) time.Duration {
	seconds := GetIntFallback(cfg, provider, key)
	if seconds <= 0 {
		return 0
	}
	return time.Duration(seconds) * time.Second
}

// GetDurationMillisFallback 从配置中读取毫秒数并转换为 time.Duration。
func GetDurationMillisFallback(cfg datacontract.Config, provider, key string) time.Duration {
	ms := GetIntFallback(cfg, provider, key)
	if ms <= 0 {
		return 0
	}
	return time.Duration(ms) * time.Millisecond
}

// BaseConfigSourceProvider 消除配置中心 provider 之间的结构性重复。
//
// 各配置中心组件只需提供 NameStr / GetConfig / NewSource 三个字段，
// 即可自动获得完整的 ServiceProvider 行为（Boot / Register / Provides / DependsOn）。
//
// 使用方式：
//
//	p := &Provider{}
//	p.NameStr = "configsource.xxx"
//	p.GetConfig = func(c Container) (any, error) { return getConfig(c) }
//	p.NewSource = func(cfg any) (ConfigSource, error) { return NewConfigSource(cfg) }
type BaseConfigSourceProvider struct {
	NameStr   string
	GetConfig func(c runtimecontract.Container) (any, error)
	NewSource func(cfg any) (datacontract.ConfigSource, error)
}

// Name 返回 provider 名称。
func (p *BaseConfigSourceProvider) Name() string { return p.NameStr }

// IsDefer 返回是否延迟注册，配置中心统一为 true。
func (p *BaseConfigSourceProvider) IsDefer() bool { return true }

// Provides 返回 provider 提供的能力键，固定为 ConfigSourceKey。
func (p *BaseConfigSourceProvider) Provides() []string {
	return []string{datacontract.ConfigSourceKey}
}

// DependsOn 返回 provider 的依赖键，配置中心无依赖。
func (p *BaseConfigSourceProvider) DependsOn() []string { return nil }

// Boot 启动阶段无操作，直接返回 nil。
func (p *BaseConfigSourceProvider) Boot(runtimecontract.Container) error { return nil }

// Register 执行配置源注册，将 ConfigSource 绑定到容器。
//
// 流程：GetConfig -> NewSource -> Bind 到容器 + 注册 Close 钩子。
func (p *BaseConfigSourceProvider) Register(c runtimecontract.Container) error {
	c.Bind(datacontract.ConfigSourceKey, func(c runtimecontract.Container) (any, error) {
		cfg, err := p.GetConfig(c)
		if err != nil {
			return nil, err
		}
		src, err := p.NewSource(cfg)
		if err != nil {
			return nil, err
		}
		c.RegisterCloser(datacontract.ConfigSourceKey, src)
		return src, nil
	}, true)
	return nil
}

// As 将 source 投射到目标 target，支持类型断言和接口实现检查。
func As(source any, target any) bool {
	if target == nil {
		return false
	}
	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr || targetValue.IsNil() {
		return false
	}
	sourceValue := reflect.ValueOf(source)
	if !sourceValue.IsValid() {
		return false
	}
	destination := targetValue.Elem()
	if !destination.CanSet() {
		return false
	}
	sourceType := sourceValue.Type()
	destinationType := destination.Type()
	if sourceType.AssignableTo(destinationType) {
		destination.Set(sourceValue)
		return true
	}
	if destinationType.Kind() == reflect.Interface && sourceType.Implements(destinationType) {
		destination.Set(sourceValue)
		return true
	}
	if sourceType.ConvertibleTo(destinationType) {
		destination.Set(sourceValue.Convert(destinationType))
		return true
	}
	return false
}