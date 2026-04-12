package zookeeper

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/ngq/gorp/framework/contract"
)

// Provider 提供 Zookeeper 服务发现实现。
//
// 中文说明：
// - 使用 Zookeeper 实现服务注册与发现；
// - 支持临时节点（Ephemeral ZNode）实现健康检查；
// - 支持服务元数据；
// - 适用于已有 Zookeeper 集群的环境。
type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string     { return "discovery.zookeeper" }
func (p *Provider) IsDefer() bool    { return true }
func (p *Provider) Provides() []string {
	return []string{contract.RPCRegistryKey}
}

func (p *Provider) Register(c contract.Container) error {
	c.Bind(contract.RPCRegistryKey, func(c contract.Container) (any, error) {
		cfg, err := getZookeeperConfig(c)
		if err != nil {
			return nil, err
		}
		return NewRegistry(cfg)
	}, true)

	return nil
}

func (p *Provider) Boot(c contract.Container) error {
	return nil
}

// ZookeeperConfig 定义 Zookeeper 配置。
type ZookeeperConfig struct {
	// Servers Zookeeper 服务器列表
	Servers []string

	// SessionTimeout 会话超时时间
	SessionTimeout time.Duration

	// BasePath 服务注册基础路径（默认 "/services"）
	BasePath string

	// 服务注册配置
	ServiceName string
	ServiceAddr string
	ServicePort int
	ServiceMeta map[string]string
}

// getZookeeperConfig 从容器获取配置。
func getZookeeperConfig(c contract.Container) (*ZookeeperConfig, error) {
	cfgAny, err := c.Make(contract.ConfigKey)
	if err != nil {
		return nil, err
	}

	cfg, ok := cfgAny.(contract.Config)
	if !ok {
		return nil, errors.New("zookeeper: invalid config service")
	}

	zkCfg := &ZookeeperConfig{
		BasePath:       "/services",
		SessionTimeout: 30 * time.Second,
	}

	// 读取服务器列表
	if v := cfg.Get("discovery.zookeeper.servers"); v != nil {
		if servers, ok := v.([]string); ok {
			zkCfg.Servers = servers
		}
	}
	if v := cfg.Get("discovery.zookeeper.base_path"); v != nil {
		zkCfg.BasePath = cfg.GetString("discovery.zookeeper.base_path")
	}
	if v := cfg.Get("discovery.zookeeper.session_timeout"); v != nil {
		// 从配置读取超时时间（秒）
		zkCfg.SessionTimeout = time.Duration(cfg.GetInt("discovery.zookeeper.session_timeout")) * time.Second
	}

	return zkCfg, nil
}

// Registry Zookeeper 服务注册中心实现。
type Registry struct {
	config *ZookeeperConfig

	// mu 保护状态
	mu sync.RWMutex

	// registeredInstances 已注册的实例
	registeredInstances map[string]string // name -> znode path

	// conn Zookeeper 连接（需要引入 go-zookeeper/zk）
	// conn *zk.Conn
}

// NewRegistry 创建 Zookeeper Registry。
func NewRegistry(cfg *ZookeeperConfig) (*Registry, error) {
	if len(cfg.Servers) == 0 {
		return nil, errors.New("zookeeper: servers is required")
	}

	return &Registry{
		config:              cfg,
		registeredInstances: make(map[string]string),
	}, nil
}

// Register 注册服务实例。
//
// 中文说明：
// - 在 BasePath 下创建临时节点；
// - 节点路径：/services/{name}/{address}；
// - 节点数据：服务元数据 JSON；
// - 会话断开时节点自动删除。
func (r *Registry) Register(ctx context.Context, name, addr string, meta map[string]string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// TODO: 实现真实的 Zookeeper 注册
	// 需要引入 github.com/go-zookeeper/zk
	//
	// 示例流程：
	// 1. 确保 /services/{name} 路径存在
	// 2. 创建临时节点 /services/{name}/{addr}
	// 3. 设置节点数据（元数据 JSON）

	// 记录已注册的实例
	instancePath := r.config.BasePath + "/" + name + "/" + addr
	r.registeredInstances[name] = instancePath

	return nil
}

// Deregister 注销服务实例。
//
// 中文说明：
// - 删除临时节点。
func (r *Registry) Deregister(ctx context.Context, name, addr string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// TODO: 实现真实的 Zookeeper 注销
	// 删除临时节点

	delete(r.registeredInstances, name)
	return nil
}

// Discover 发现服务实例。
//
// 中文说明：
// - 获取 /services/{name} 下的所有子节点；
// - 解析节点数据和地址；
// - 返回 ServiceInstance 列表。
func (r *Registry) Discover(ctx context.Context, name string) ([]contract.ServiceInstance, error) {
	// TODO: 实现真实的 Zookeeper 发现
	// 需要引入 github.com/go-zookeeper/zk
	//
	// 示例流程：
	// 1. 获取 /services/{name} 的子节点列表
	// 2. 读取每个节点的数据
	// 3. 解析为 ServiceInstance

	return []contract.ServiceInstance{}, nil
}

// Close 关闭 Zookeeper 连接。
func (r *Registry) Close() error {
	return nil
}

// ErrNoServers 表示未配置服务器。
var ErrNoServers = errors.New("zookeeper: no servers configured")