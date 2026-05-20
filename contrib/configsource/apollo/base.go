// Package apollo provides Apollo configuration center provider for the gorp framework.
// 本文件内联 baseconfigsource 和 native.As，使 apollo 成为完全独立的模块。
package apollo

import (
	"errors"
	"reflect"
	"time"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	configprovider "github.com/ngq/gorp/framework/provider/config"
)

// As 尝试将 source 通过反射投射到 target。
// 支持直接赋值、接口实现和类型转换三种路径。
// 当 target 为 nil、非指针、不可寻址或投射失败时返回 false。
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

// BaseConfigSourceProvider 消除 ConfigSource provider 之间的结构重复。
// 每个具体 provider 只需设置 NameStr/GetConfig/NewSource 三个字段即可。
type BaseConfigSourceProvider struct {
	// NameStr 是 provider 的名称标识。
	NameStr string
	// GetConfig 从容器中提取配置源所需的配置结构。
	GetConfig func(c runtimecontract.Container) (any, error)
	// NewSource 根据配置创建具体的 ConfigSource 实例。
	NewSource func(cfg any) (datacontract.ConfigSource, error)
}

// Name 返回 provider 名称。
func (p *BaseConfigSourceProvider) Name() string { return p.NameStr }

// IsDefer 标记 ConfigSource 为延迟加载。
func (p *BaseConfigSourceProvider) IsDefer() bool { return true }

// Provides 声明本 provider 提供 ConfigSourceKey 绑定。
func (p *BaseConfigSourceProvider) Provides() []string {
	return []string{datacontract.ConfigSourceKey}
}

// DependsOn 声明本 provider 依赖 ConfigKey 绑定。
func (p *BaseConfigSourceProvider) DependsOn() []string {
	return []string{datacontract.ConfigKey}
}

// Boot 启动阶段无额外操作。
func (p *BaseConfigSourceProvider) Boot(runtimecontract.Container) error { return nil }

// Register 将 ConfigSource 工厂注册到容器，并在关闭时调用 ConfigSource.Close。
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

// ReadConfig 从容器中提取 Config 服务实例。
func ReadConfig(c interface{ Make(key string) (any, error) }) (datacontract.Config, error) {
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

// GetStringFallback 读取字符串配置值，支持标准化 key 回退。
// 优先读取 configsource.<name>.<field>，回退到 config.<name>.<field>。
func GetStringFallback(cfg datacontract.Config, name, field string) string {
	return configprovider.GetStringAny(cfg, "configsource."+name+"."+field, "config."+name+"."+field)
}

// GetIntFallback 读取整数配置值，支持标准化 key 回退。
// 优先读取 configsource.<name>.<field>，回退到 config.<name>.<field>。
func GetIntFallback(cfg datacontract.Config, name, field string) int {
	return configprovider.GetIntAny(cfg, "configsource."+name+"."+field, "config."+name+"."+field)
}

// GetBoolFallback 读取布尔配置值，支持标准化 key 回退。
// 优先读取 configsource.<name>.<field>，回退到 config.<name>.<field>。
// 第二个返回值表示是否成功读取到配置。
func GetBoolFallback(cfg datacontract.Config, name, field string) (bool, bool) {
	return configprovider.GetBoolAny(cfg, "configsource."+name+"."+field, "config."+name+"."+field)
}

// GetDurationSecondsFallback 读取秒数配置值并转换为 time.Duration。
// 优先读取 configsource.<name>.<field>，回退到 config.<name>.<field>。
func GetDurationSecondsFallback(cfg datacontract.Config, name, field string) time.Duration {
	if seconds := GetIntFallback(cfg, name, field); seconds > 0 {
		return time.Duration(seconds) * time.Second
	}
	return 0
}

// GetDurationMillisFallback 读取毫秒配置值并转换为 time.Duration。
// 优先读取 configsource.<name>.<field>，回退到 config.<name>.<field>。
func GetDurationMillisFallback(cfg datacontract.Config, name, field string) time.Duration {
	if ms := GetIntFallback(cfg, name, field); ms > 0 {
		return time.Duration(ms) * time.Millisecond
	}
	return 0
}
