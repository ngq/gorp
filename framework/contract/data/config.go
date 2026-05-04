package data

import "context"

const ConfigKey = "framework.config"

const ConfigSourceKey = "framework.config.source"

const ConfigWatcherKey = "framework.config.watcher"

type Config interface {
	Env() string

	Get(key string) any
	GetString(key string) string
	GetInt(key string) int
	GetBool(key string) bool
	GetFloat(key string) float64

	Unmarshal(key string, out any) error
	Watch(ctx context.Context, key string) (ConfigWatcher, error)
	Reload(ctx context.Context) error
}

type ConfigSource interface {
	Load(ctx context.Context) (map[string]any, error)
	Get(ctx context.Context, key string) (any, error)
	Set(ctx context.Context, key string, value any) error
	Watch(ctx context.Context, key string) (ConfigWatcher, error)
	Close() error
}

type ConfigWatcher interface {
	OnChange(key string, callback func(value any))
	Stop() error
}

type ConfigSourceType string

const (
	ConfigSourceLocal  ConfigSourceType = "local"
	ConfigSourceConsul ConfigSourceType = "consul"
	ConfigSourceEtcd   ConfigSourceType = "etcd"
	ConfigSourceNacos  ConfigSourceType = "nacos"
	ConfigSourceApollo ConfigSourceType = "apollo"
	ConfigSourceNoop   ConfigSourceType = "noop"
)

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
