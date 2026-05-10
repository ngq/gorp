// Package nacos provides Nacos configuration center provider for the gorp framework.
// This provider implements ConfigSource contract with Nacos SDK integration.
//
// 本包提供 gorp 框架 Nacos 配置中心 provider。
// 本 provider 实现 ConfigSource 契约，集成 Nacos SDK。
package nacos

import (
	"errors"
	"time"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

// Provider 提供 Nacos 配置中心实现。
//
// 中文说明：
//   - 使用阿里 Nacos 配置中心；
//   - 支持多命名空间；
//   - 支持配置热更新；
//   - 支持分组管理。
//   - 当前状态：部分可用
//   - 说明：已完成 P1 最小闭环，具备 Load / Watch / Set 主流程与 fake client 行为测试；
//     但当前仍是最小配置中心闭环，尚未进入完整产品化治理能力。
type Provider struct{}

// NewProvider creates a new Nacos provider instance.
//
// NewProvider 创建新的 Nacos provider 实例。
func NewProvider() *Provider { return &Provider{} }

// Name returns the provider identifier "configsource.nacos".
//
// Name 返回 provider 标识符 "configsource.nacos"。
func (p *Provider) Name() string  { return "configsource.nacos" }

// IsDefer returns true for lazy initialization.
//
// IsDefer 返回 true，延迟初始化。
func (p *Provider) IsDefer() bool { return true }

// Provides returns the contract keys this provider satisfies.
//
// Provides 返回此 provider 满足的契约键。
func (p *Provider) Provides() []string {
	return []string{datacontract.ConfigSourceKey}
}

// Register binds the Nacos config source to the container.
//
// Register 将 Nacos 配置源绑定到容器。
// 该方法注册一个延迟构造函数，在首次使用时创建 ConfigSource 实例。
func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(datacontract.ConfigSourceKey, func(c runtimecontract.Container) (any, error) {
		cfg, err := getNacosConfig(c)
		if err != nil {
			return nil, err
		}
		return NewConfigSource(cfg)
	}, true)

	return nil
}

// Boot does nothing for lazy providers.
//
// Boot 延迟 provider 不需要 boot 操作。
func (p *Provider) Boot(c runtimecontract.Container) error {
	return nil
}

// NacosConfig 定义 Nacos 配置。
// 该结构体包含连接 Nacos 服务器和访问配置所需的所有参数。
//
// 配置项说明：
//   - ServerAddr: Nacos 服务器地址，支持 http:// 或 https:// 前缀
//   - Port: Nacos 服务器端口，默认 8848
//   - Namespace: 命名空间 ID，用于隔离不同环境的配置
//   - Group: 配置分组，默认 DEFAULT_GROUP
//   - DataID: 配置数据 ID，必需项，标识具体的配置文件
//   - Username: Nacos 认证用户名（可选）
//   - Password: Nacos 认证密码（可选）
//   - PollInterval: 配置轮询间隔，用于 Watch 的定期检查，默认 5s
type NacosConfig struct {
	ServerAddr   string
	Port         int
	Namespace    string
	Group        string
	DataID       string
	Username     string
	Password     string
	PollInterval time.Duration
}

// getNacosConfig extracts Nacos configuration from the container's config binding.
//
// getNacosConfig 从容器的 config binding 中提取 Nacos 配置。
// 该函数从框架配置中读取 Nacos 相关配置项，支持两种配置路径：
//   - configsource.nacos.* （优先）
//   - config.nacos.* （备选）
// 这种双重路径设计兼容历史配置格式，便于平滑迁移。
func getNacosConfig(c runtimecontract.Container) (*NacosConfig, error) {
	cfgAny, err := c.Make(datacontract.ConfigKey)
	if err != nil {
		return nil, err
	}

	cfg, ok := cfgAny.(datacontract.Config)
	if !ok {
		return nil, errors.New("nacos: invalid config service")
	}

	// 初始化默认配置
	nacosCfg := &NacosConfig{
		Port:         8848,
		Group:        "DEFAULT_GROUP",
		PollInterval: defaultNacosPollInterval,
	}

	// 读取服务器地址（优先 configsource.nacos.*，备选 config.nacos.*）
	if v := cfg.Get("configsource.nacos.server_addr"); v != nil {
		nacosCfg.ServerAddr = cfg.GetString("configsource.nacos.server_addr")
	} else if v := cfg.Get("config.nacos.server_addr"); v != nil {
		nacosCfg.ServerAddr = cfg.GetString("config.nacos.server_addr")
	}

	// 读取端口
	if v := cfg.Get("configsource.nacos.port"); v != nil {
		nacosCfg.Port = cfg.GetInt("configsource.nacos.port")
	} else if v := cfg.Get("config.nacos.port"); v != nil {
		nacosCfg.Port = cfg.GetInt("config.nacos.port")
	}

	// 读取命名空间
	if v := cfg.Get("configsource.nacos.namespace"); v != nil {
		nacosCfg.Namespace = cfg.GetString("configsource.nacos.namespace")
	} else if v := cfg.Get("config.nacos.namespace"); v != nil {
		nacosCfg.Namespace = cfg.GetString("config.nacos.namespace")
	}

	// 读取分组
	if v := cfg.Get("configsource.nacos.group"); v != nil {
		nacosCfg.Group = cfg.GetString("configsource.nacos.group")
	} else if v := cfg.Get("config.nacos.group"); v != nil {
		nacosCfg.Group = cfg.GetString("config.nacos.group")
	}

	// 读取数据 ID
	if v := cfg.Get("configsource.nacos.data_id"); v != nil {
		nacosCfg.DataID = cfg.GetString("configsource.nacos.data_id")
	} else if v := cfg.Get("config.nacos.data_id"); v != nil {
		nacosCfg.DataID = cfg.GetString("config.nacos.data_id")
	}

	// 读取认证信息（用户名和密码）
	if v := cfg.Get("configsource.nacos.username"); v != nil {
		nacosCfg.Username = cfg.GetString("configsource.nacos.username")
	} else if v := cfg.Get("config.nacos.username"); v != nil {
		nacosCfg.Username = cfg.GetString("config.nacos.username")
	}
	if v := cfg.Get("configsource.nacos.password"); v != nil {
		nacosCfg.Password = cfg.GetString("configsource.nacos.password")
	} else if v := cfg.Get("config.nacos.password"); v != nil {
		nacosCfg.Password = cfg.GetString("config.nacos.password")
	}

	// 读取轮询间隔（秒）
	if v := cfg.Get("configsource.nacos.poll_interval_seconds"); v != nil {
		if seconds := cfg.GetInt("configsource.nacos.poll_interval_seconds"); seconds > 0 {
			nacosCfg.PollInterval = time.Duration(seconds) * time.Second
		}
	}

	return nacosCfg, nil
}