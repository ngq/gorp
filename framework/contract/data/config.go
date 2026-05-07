// Application scenarios:
// - Define the configuration contracts used by config providers and business code.
// - Standardize config loading, watching, reloading, and typed reads.
// - Provide a shared model for external config source selection and connection settings.
//
// 适用场景：
// - 定义配置 provider 和业务代码共同依赖的配置契约。
// - 统一配置加载、监听、重载和类型化读取语义。
// - 为外部配置源选择和连接参数提供共享模型。
package data

import "context"

// ConfigKey is the container key for the config capability.
//
// ConfigKey 是配置能力的容器键。
const ConfigKey = "framework.config"

// ConfigSourceKey is the container key for the config source capability.
//
// ConfigSourceKey 是配置源能力的容器键。
const ConfigSourceKey = "framework.config.source"

// ConfigWatcherKey is the container key for the config watcher capability.
//
// ConfigWatcherKey 是配置监听能力的容器键。
const ConfigWatcherKey = "framework.config.watcher"

// Config defines the runtime configuration view exposed by the framework.
//
// Config 定义框架对外暴露的运行时配置视图。
type Config interface {
	// Env returns the current application environment name.
	//
	// Env 返回当前应用环境名。
	Env() string

	// Get reads a raw config value by key.
	//
	// Get 按 key 读取原始配置值。
	Get(key string) any

	// GetString reads a config value as string.
	//
	// GetString 以字符串形式读取配置值。
	GetString(key string) string

	// GetInt reads a config value as int.
	//
	// GetInt 以整数形式读取配置值。
	GetInt(key string) int

	// GetBool reads a config value as bool.
	//
	// GetBool 以布尔形式读取配置值。
	GetBool(key string) bool

	// GetFloat reads a config value as float64.
	//
	// GetFloat 以 float64 形式读取配置值。
	GetFloat(key string) float64

	// Unmarshal decodes a config subtree into the target object.
	//
	// Unmarshal 将配置子树解码到目标对象。
	Unmarshal(key string, out any) error

	// Watch subscribes to changes of a given config key.
	//
	// Watch 订阅指定配置 key 的变更事件。
	Watch(ctx context.Context, key string) (ConfigWatcher, error)

	// Reload forces a config refresh from the underlying source.
	//
	// Reload 强制从底层配置源刷新配置。
	Reload(ctx context.Context) error
}

// ConfigSource defines the pluggable backing source for configuration.
//
// ConfigSource 定义可插拔的底层配置源契约。
type ConfigSource interface {
	// Load loads the full config snapshot.
	//
	// Load 加载完整配置快照。
	Load(ctx context.Context) (map[string]any, error)

	// Get reads a single config value from the source.
	//
	// Get 从配置源读取单个配置值。
	Get(ctx context.Context, key string) (any, error)

	// Set writes a config value back to the source when supported.
	//
	// Set 在配置源支持时回写配置值。
	Set(ctx context.Context, key string, value any) error

	// Watch subscribes to source-side changes of a config key.
	//
	// Watch 订阅配置源侧的指定 key 变更。
	Watch(ctx context.Context, key string) (ConfigWatcher, error)

	// Close releases resources held by the source.
	//
	// Close 释放配置源持有的资源。
	Close() error
}

// ConfigWatcher defines the change callback contract for watched config keys.
//
// ConfigWatcher 定义配置 key 监听后的变更回调契约。
type ConfigWatcher interface {
	// OnChange registers a callback for key changes.
	//
	// OnChange 为 key 变更注册回调。
	OnChange(key string, callback func(value any))

	// Stop stops the watcher and releases related resources.
	//
	// Stop 停止监听并释放相关资源。
	Stop() error
}

// ConfigSourceType identifies the implementation type of a config source.
//
// ConfigSourceType 标识配置源实现类型。
type ConfigSourceType string

const (
	ConfigSourceLocal  ConfigSourceType = "local"
	ConfigSourceConsul ConfigSourceType = "consul"
	ConfigSourceEtcd   ConfigSourceType = "etcd"
	ConfigSourceNacos  ConfigSourceType = "nacos"
	ConfigSourceApollo ConfigSourceType = "apollo"
	ConfigSourceNoop   ConfigSourceType = "noop"
)

// ConfigSourceConfig describes the external connection settings of a config source.
//
// ConfigSourceConfig 描述配置源的外部连接配置。
type ConfigSourceConfig struct {
	Type ConfigSourceType `mapstructure:"type"`

	ConsulAddr  string `mapstructure:"consul_addr"`
	ConsulPath  string `mapstructure:"consul_path"`
	ConsulToken string `mapstructure:"consul_token"`

	EtcdEndpoints []string `mapstructure:"etcd_endpoints"`
	EtcdPath      string   `mapstructure:"etcd_path"`
	EtcdUsername  string   `mapstructure:"etcd_username"`
	EtcdPassword  string   `mapstructure:"etcd_password"`

	NacosAddr      string `mapstructure:"nacos_addr"`
	NacosNamespace string `mapstructure:"nacos_namespace"`
	NacosGroup     string `mapstructure:"nacos_group"`
	NacosDataID    string `mapstructure:"nacos_data_id"`

	ApolloAddr      string `mapstructure:"apollo_addr"`
	ApolloNamespace string `mapstructure:"apollo_namespace"`
	ApolloAppID     string `mapstructure:"apollo_app_id"`
}
