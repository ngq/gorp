package contract

import "context"

const ConfigKey = "framework.config"

// ConfigSourceKey 是配置源在容器中的绑定 key。
//
// 中文说明：
// - 用于支持远程配置源（Consul KV / etcd / Nacos）；
// - noop 实现使用本地文件，单体项目零依赖。
const ConfigSourceKey = "framework.config.source"

// ConfigWatcherKey 是配置监听器在容器中的绑定 key。
//
// 中文说明：
// - 用于支持配置热更新；
// - noop 实现不支持热更新，单体项目无需监听。
const ConfigWatcherKey = "framework.config.watcher"

type Config interface {
	Env() string

	Get(key string) any
	GetString(key string) string
	GetInt(key string) int
	GetBool(key string) bool
	GetFloat(key string) float64

	// Unmarshal decodes config at key into the provided struct pointer.
	Unmarshal(key string, out any) error

	// Watch 监听配置变化（可选实现）。
	//
	// 中文说明：
	// - noop 实现返回 nil，不支持热更新；
	// - 远程配置源实现返回 ConfigWatcher 实例。
	Watch(ctx context.Context, key string) (ConfigWatcher, error)

	// Reload 强制重新加载配置。
	//
	// 中文说明：
	// - 本地文件实现重新读取 config/*.yaml；
	// - 远程配置源实现重新拉取远程配置。
	Reload(ctx context.Context) error
}

// ConfigSource 定义配置源抽象。
//
// 中文说明：
// - 本地文件实现：读取 config/*.yaml；
// - 远程配置实现：从 Consul KV / etcd / Nacos 拉取；
// - noop 实现：空操作，单体项目使用默认本地文件。
type ConfigSource interface {
	// Load 加载配置。
	//
	// 中文说明：
	// - 从配置源读取配置内容；
	// - 返回的 map 会合并到当前配置实例。
	Load(ctx context.Context) (map[string]any, error)

	// Get 获取单个配置项。
	Get(ctx context.Context, key string) (any, error)

	// Set 设置单个配置项（远程配置源支持）。
	//
	// 中文说明：
	// - noop/本地文件实现返回错误；
	// - 远程配置源实现写入远程存储。
	Set(ctx context.Context, key string, value any) error

	// Watch 监听配置变化。
	//
	// 中文说明：
	// - 返回 ConfigWatcher 实例；
	// - noop 实现返回不支持错误。
	Watch(ctx context.Context, key string) (ConfigWatcher, error)

	// Close 关闭配置源连接。
	Close() error
}

// ConfigWatcher 定义配置监听抽象。
//
// 中文说明：
// - 用于配置热更新；
// - 当配置发生变化时触发回调；
// - noop 实现不支持，远程配置源支持。
type ConfigWatcher interface {
	// OnChange 注册配置变化回调。
	//
	// 中文说明：
	// - key: 配置项路径；
	// - callback: 变化回调函数，接收新值。
	OnChange(key string, callback func(value any))

	// Stop 停止监听。
	Stop() error
}

// ConfigSourceType 定义配置源类型。
type ConfigSourceType string

const (
	// ConfigSourceLocal 本地文件配置源
	ConfigSourceLocal ConfigSourceType = "local"

	// ConfigSourceConsul Consul KV 配置源
	ConfigSourceConsul ConfigSourceType = "consul"

	// ConfigSourceEtcd etcd 配置源
	ConfigSourceEtcd ConfigSourceType = "etcd"

	// ConfigSourceNacos Nacos 配置源
	ConfigSourceNacos ConfigSourceType = "nacos"

	// ConfigSourceApollo Apollo 配置源
	ConfigSourceApollo ConfigSourceType = "apollo"

	// ConfigSourceNoop noop 配置源（单体零依赖）
	ConfigSourceNoop ConfigSourceType = "noop"
)

// ConfigSourceConfig 定义配置源配置。
type ConfigSourceConfig struct {
	// Type 配置源类型：local/consul/etcd/nacos/apollo/noop
	Type ConfigSourceType `mapstructure:"type"`

	// Consul 配置
	ConsulAddr    string `mapstructure:"consul_addr"`
	ConsulPath    string `mapstructure:"consul_path"`
	ConsulToken   string `mapstructure:"consul_token"`

	// etcd 配置
	EtcdEndpoints []string `mapstructure:"etcd_endpoints"`
	EtcdPath      string   `mapstructure:"etcd_path"`
	EtcdUsername  string   `mapstructure:"etcd_username"`
	EtcdPassword  string   `mapstructure:"etcd_password"`

	// Nacos 配置
	NacosAddr      string `mapstructure:"nacos_addr"`
	NacosNamespace string `mapstructure:"nacos_namespace"`
	NacosGroup     string `mapstructure:"nacos_group"`
	NacosDataID    string `mapstructure:"nacos_data_id"`

	// Apollo 配置
	ApolloAddr      string `mapstructure:"apollo_addr"`
	ApolloNamespace string `mapstructure:"apollo_namespace"`
	ApolloAppID     string `mapstructure:"apollo_app_id"`
}
