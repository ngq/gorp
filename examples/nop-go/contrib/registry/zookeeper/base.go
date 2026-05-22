// Package zookeeper 提供 Zookeeper 服务注册中心的基础能力。
//
// 本文件内联了 baseregistry / native 的公共逻辑，使 zookeeper 包成为完全独立的模块，
// 不再依赖 contrib/internal 下的任何包。
package zookeeper

import (
	"reflect"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	configprovider "github.com/ngq/gorp/framework/provider/config"
)

// As 尝试通过 reflect 将 source 断言/转换为 target 指向的类型。
//
// 支持三种匹配路径：
//   1. source 直接可赋值给 target 类型
//   2. target 是接口且 source 实现了该接口
//   3. source 可转换为目标类型
//
// 返回 true 表示转换成功且已写入 target，false 表示不匹配。
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

// ServiceConfig 保存服务注册的通用字段。
//
// 各注册中心组件的配置结构体可嵌入此结构体，或通过 ReadServiceConfig 读取后映射到自己的字段。
type ServiceConfig struct {
	ServiceName string
	ServiceAddr string
	ServicePort int
	ServiceMeta map[string]string
	LoadBalance string
}

// ReadServiceConfig 从配置中心读取 discovery.service.* 下的通用服务配置。
//
// 支持多种配置键名风格（点分/下划线），适配不同团队的配置习惯。
func ReadServiceConfig(cfg datacontract.Config) ServiceConfig {
	sc := ServiceConfig{}
	sc.ServiceName = configprovider.GetStringAny(cfg, "discovery.service.name", "discovery.service_name")
	sc.ServiceAddr = configprovider.GetStringAny(cfg, "discovery.service.addr", "discovery.service.address", "discovery.service_addr")
	sc.ServicePort = configprovider.GetIntAny(cfg, "discovery.service.port", "discovery.service_port")
	sc.LoadBalance = configprovider.GetStringAny(cfg, "selector.algorithm", "discovery.load_balance")
	return sc
}

// BaseRegistryProvider 消除注册中心 provider 之间的结构性重复。
//
// 各注册中心组件只需提供 NameStr / GetConfig / NewRegistry 三个字段，
// 即可自动获得完整的 ServiceProvider 行为（Boot / Register / Provides / DependsOn）。
//
// 使用方式：
//
//	p := &Provider{}
//	p.NameStr = "registry.xxx"
//	p.GetConfig = func(c Container) (any, error) { return getConfig(c) }
//	p.NewRegistry = func(cfg any) (ServiceRegistry, error) { return NewRegistry(cfg) }
type BaseRegistryProvider struct {
	NameStr     string
	GetConfig   func(c runtimecontract.Container) (any, error)
	NewRegistry func(cfg any) (transportcontract.ServiceRegistry, error)
}

// Name 返回 provider 名称。
func (p *BaseRegistryProvider) Name() string { return p.NameStr }

// IsDefer 返回是否延迟注册，注册中心统一为 true。
func (p *BaseRegistryProvider) IsDefer() bool { return true }

// Provides 返回 provider 提供的能力键，固定为 RPCRegistryKey。
func (p *BaseRegistryProvider) Provides() []string {
	return []string{transportcontract.RPCRegistryKey}
}

// DependsOn 返回 provider 的依赖键，固定依赖 ConfigKey。
func (p *BaseRegistryProvider) DependsOn() []string {
	return []string{datacontract.ConfigKey}
}

// Boot 启动阶段无操作，直接返回 nil。
func (p *BaseRegistryProvider) Boot(runtimecontract.Container) error { return nil }

// Register 执行服务注册，将 ServiceRegistry 绑定到容器。
//
// 流程：GetConfig -> NewRegistry -> Bind 到容器 + 注册 Close 钩子。
func (p *BaseRegistryProvider) Register(c runtimecontract.Container) error {
	c.Bind(transportcontract.RPCRegistryKey, func(c runtimecontract.Container) (any, error) {
		cfg, err := p.GetConfig(c)
		if err != nil {
			return nil, err
		}
		reg, err := p.NewRegistry(cfg)
		if err != nil {
			return nil, err
		}
		c.RegisterCloser(transportcontract.RPCRegistryKey, reg)
		return reg, nil
	}, true)
	return nil
}