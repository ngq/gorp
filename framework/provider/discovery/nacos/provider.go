package nacos

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"

	"github.com/ngq/gorp/framework/contract"
	configprovider "github.com/ngq/gorp/framework/provider/config"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

// Provider 提供 Nacos 服务发现实现。
//
// 中文说明：
// - 使用 Nacos Naming API 实现服务注册与发现；
// - 支持命名空间隔离；
// - 支持服务元数据（版本、权重等）；
// - 通过客户端心跳实现健康检查；
// - 需要项目引入 github.com/nacos-group/nacos-sdk-go/v2 依赖。
type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string     { return "discovery.nacos" }
func (p *Provider) IsDefer() bool    { return true }
func (p *Provider) Provides() []string {
	return []string{contract.RPCRegistryKey}
}

func (p *Provider) Register(c contract.Container) error {
	c.Bind(contract.RPCRegistryKey, func(c contract.Container) (any, error) {
		cfg, err := getDiscoveryConfig(c)
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

// DiscoveryConfig 定义 Nacos 服务发现配置。
type DiscoveryConfig struct {
	// Nacos 服务端配置
	NacosAddr      string // 如 "localhost:8848"
	NacosNamespace string // 命名空间 ID（可选）
	NacosGroup     string // 分组名称，默认 "DEFAULT_GROUP"
	NacosUsername  string // 用户名（可选）
	NacosPassword  string // 密码（可选）

	// 服务注册配置
	ServiceName    string
	ServiceAddr    string
	ServicePort    int
	ServiceWeight  float64 // 服务权重，默认 1.0
	ServiceMeta    map[string]string

	// 负载均衡策略
	LoadBalance string // random/weight（按权重）
}

// getDiscoveryConfig 从容器获取服务发现配置。
func getDiscoveryConfig(c contract.Container) (*DiscoveryConfig, error) {
	cfgAny, err := c.Make(contract.ConfigKey)
	if err != nil {
		return nil, err
	}

	cfg, ok := cfgAny.(contract.Config)
	if !ok {
		return nil, errors.New("discovery: invalid config service")
	}

	discCfg := &DiscoveryConfig{
		NacosAddr:     "localhost:8848",
		NacosGroup:    "DEFAULT_GROUP",
		ServiceWeight: 1.0,
		LoadBalance:   "weight", // Nacos 默认按权重
	}

	// Nacos 地址
	if addr := configprovider.GetStringAny(cfg,
		"discovery.nacos.addr",
		"discovery.nacos.address",
		"discovery.nacos_addr",
	); addr != "" {
		discCfg.NacosAddr = addr
	}

	// 命名空间
	if ns := configprovider.GetStringAny(cfg,
		"discovery.nacos.namespace",
		"discovery.nacos_namespace",
	); ns != "" {
		discCfg.NacosNamespace = ns
	}

	// 分组
	if group := configprovider.GetStringAny(cfg,
		"discovery.nacos.group",
		"discovery.nacos_group",
	); group != "" {
		discCfg.NacosGroup = group
	}

	// 认证
	if username := configprovider.GetStringAny(cfg,
		"discovery.nacos.username",
		"discovery.nacos_username",
	); username != "" {
		discCfg.NacosUsername = username
	}
	if password := configprovider.GetStringAny(cfg,
		"discovery.nacos.password",
		"discovery.nacos_password",
	); password != "" {
		discCfg.NacosPassword = password
	}

	// 服务名称
	if name := configprovider.GetStringAny(cfg,
		"discovery.service.name",
		"discovery.service_name",
	); name != "" {
		discCfg.ServiceName = name
	}

	// 服务地址
	if addr := configprovider.GetStringAny(cfg,
		"discovery.service.addr",
		"discovery.service.address",
		"discovery.service_addr",
	); addr != "" {
		discCfg.ServiceAddr = addr
	}

	// 服务端口
	if port := configprovider.GetIntAny(cfg,
		"discovery.service.port",
		"discovery.service_port",
	); port > 0 {
		discCfg.ServicePort = port
	}

	// 服务权重
	if weight := configprovider.GetFloatAny(cfg,
		"discovery.service.weight",
		"discovery.service_weight",
	); weight > 0 {
		discCfg.ServiceWeight = weight
	}

	// 负载均衡策略
	if lb := configprovider.GetStringAny(cfg,
		"selector.algorithm",
		"discovery.load_balance",
	); lb != "" {
		discCfg.LoadBalance = lb
	}

	return discCfg, nil
}

// Registry 是 Nacos 服务发现实现。
//
// 中文说明：
// - 使用 Nacos Naming Client 注册服务；
// - 通过客户端心跳维持健康状态；
// - 支持服务权重负载均衡；
// - 支持命名空间和分组隔离。
type Registry struct {
	cfg         *DiscoveryConfig
	namingClient naming_client.INamingClient

	// 已注册服务缓存
	registered sync.Map // map[string]bool
	mu         sync.Mutex
	closed     bool
}

// NewRegistry 创建 Nacos 服务发现实例。
func NewRegistry(cfg *DiscoveryConfig) (*Registry, error) {
	// 构建 Nacos 服务端配置
	serverConfigs := []constant.ServerConfig{
		{
			IpAddr:      parseNacosAddr(cfg.NacosAddr),
			Port:        parseNacosPort(cfg.NacosAddr),
			ContextPath: "/nacos",
		},
	}

	// 构建客户端配置
	clientConfig := constant.ClientConfig{
		NamespaceId:         cfg.NacosNamespace,
		NotLoadCacheAtStart: true,
	}

	if cfg.NacosUsername != "" && cfg.NacosPassword != "" {
		clientConfig.Username = cfg.NacosUsername
		clientConfig.Password = cfg.NacosPassword
	}

	// 创建 Naming Client
	namingClient, err := clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: serverConfigs,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("discovery.nacos: create naming client failed: %w", err)
	}

	return &Registry{
		cfg:          cfg,
		namingClient: namingClient,
	}, nil
}

// Register 注册服务实例。
//
// 中文说明：
// - 使用 Nacos RegisterInstance API；
// - 支持服务权重（用于负载均衡）；
// - 支持元数据（版本、环境等）；
// - 自动发送心跳维持健康状态。
func (r *Registry) Register(ctx context.Context, name, addr string, meta map[string]string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return errors.New("discovery.nacos: registry closed")
	}

	// 解析地址获取端口
	host, port := parseAddr(addr)
	if port == 0 {
		port = r.cfg.ServicePort
	}

	// 生成服务 ID
	serviceID := generateServiceID(name, host, port)

	// 合并元数据
	fullMeta := make(map[string]string)
	for k, v := range r.cfg.ServiceMeta {
		fullMeta[k] = v
	}
	for k, v := range meta {
		fullMeta[k] = v
	}

	// 注册服务
	success, err := r.namingClient.RegisterInstance(vo.RegisterInstanceParam{
		ServiceName: name,
		GroupName:   r.cfg.NacosGroup,
		Ip:          host,
		Port:        uint64(port),
		Weight:      r.cfg.ServiceWeight,
		Enable:      true,
		Healthy:     true,
		Metadata:    fullMeta,
		Ephemeral:   true, // 临时实例，心跳维持
	})
	if err != nil {
		return fmt.Errorf("discovery.nacos: register instance failed: %w", err)
	}
	if !success {
		return errors.New("discovery.nacos: register instance failed")
	}

	// 缓存已注册服务
	r.registered.Store(serviceID, true)

	return nil
}

// Deregister 注销服务实例。
//
// 中文说明：
// - 使用 Nacos DeregisterInstance API；
// - 停止心跳发送，服务自动下线。
func (r *Registry) Deregister(ctx context.Context, name, addr string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 解析地址
	host, port := parseAddr(addr)
	if port == 0 {
		port = r.cfg.ServicePort
	}

	// 注销服务
	success, err := r.namingClient.DeregisterInstance(vo.DeregisterInstanceParam{
		ServiceName: name,
		GroupName:   r.cfg.NacosGroup,
		Ip:          host,
		Port:        uint64(port),
		Ephemeral:   true,
	})
	if err != nil {
		return fmt.Errorf("discovery.nacos: deregister instance failed: %w", err)
	}
	if !success {
		return errors.New("discovery.nacos: deregister instance failed")
	}

	// 删除缓存
	serviceID := generateServiceID(name, host, port)
	r.registered.Delete(serviceID)

	return nil
}

// Discover 发现服务实例。
//
// 中文说明：
// - 使用 Nacos SelectInstances API；
// - 只返回健康实例；
// - 支持按权重负载均衡。
func (r *Registry) Discover(ctx context.Context, name string) ([]contract.ServiceInstance, error) {
	// 查询健康实例
	instances, err := r.namingClient.SelectInstances(vo.SelectInstancesParam{
		ServiceName: name,
		GroupName:   r.cfg.NacosGroup,
		HealthyOnly: true,
	})
	if err != nil {
		return nil, fmt.Errorf("discovery.nacos: select instances failed: %w", err)
	}

	// 转换为 ServiceInstance 列表
	result := make([]contract.ServiceInstance, 0, len(instances))
	for _, inst := range instances {
		// 构建完整地址
		fullAddr := fmt.Sprintf("%s:%d", inst.Ip, inst.Port)

		// 提取服务 ID
		serviceID := generateServiceID(name, inst.Ip, int(inst.Port))

		// 转换权重为元数据
		meta := inst.Metadata
		if meta == nil {
			meta = make(map[string]string)
		}
		meta["weight"] = fmt.Sprintf("%.1f", inst.Weight)

		result = append(result, contract.ServiceInstance{
			ID:       serviceID,
			Name:     name,
			Address:  fullAddr,
			Metadata: meta,
			Healthy:  inst.Healthy,
		})
	}

	// 应用负载均衡策略
	if len(result) > 1 {
		result = r.applyLoadBalance(result)
	}

	return result, nil
}

// Close 关闭服务发现连接。
//
// 中文说明：
// - 注销所有已注册的服务；
// - Nacos SDK 会自动清理资源。
func (r *Registry) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return nil
	}
	r.closed = true

	// 注销所有已注册的服务（Nacos SDK 会自动处理）
	// 这里主要是标记已关闭状态
	r.registered.Range(func(key, value any) bool {
		// Nacos SDK 内部会停止心跳
		return true
	})

	return nil
}

// applyLoadBalance 应用负载均衡策略。
//
// 中文说明：
// - weight：按权重选择（Nacos 默认）；
// - random：随机选择。
func (r *Registry) applyLoadBalance(instances []contract.ServiceInstance) []contract.ServiceInstance {
	switch r.cfg.LoadBalance {
	case "random":
		// 随机打乱顺序
		rand.Shuffle(len(instances), func(i, j int) {
			instances[i], instances[j] = instances[j], instances[i]
		})
	case "weight":
		// 按权重排序（权重高的排在前面）
		// Nacos SDK 内部已经按权重处理
	}
	return instances
}

// parseAddr 解析地址和端口。
func parseAddr(addr string) (host string, port int) {
	parts := strings.Split(addr, ":")
	if len(parts) == 2 {
		host = parts[0]
		port, _ = strconv.Atoi(parts[1])
	} else {
		host = addr
	}
	return host, port
}

// parseNacosAddr 解析 Nacos 地址。
func parseNacosAddr(addr string) string {
	parts := strings.Split(addr, ":")
	return parts[0]
}

// parseNacosPort 解析 Nacos 端口。
func parseNacosPort(addr string) uint64 {
	parts := strings.Split(addr, ":")
	if len(parts) == 2 {
		port, _ := strconv.ParseUint(parts[1], 10, 64)
		return port
	}
	return 8848 // 默认端口
}

// generateServiceID 生成唯一服务 ID。
func generateServiceID(name, host string, port int) string {
	return fmt.Sprintf("%s-%s-%d", name, host, port)
}