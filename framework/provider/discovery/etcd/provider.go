package etcd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ngq/gorp/framework/contract"
	configprovider "github.com/ngq/gorp/framework/provider/config"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// Provider 提供 etcd 服务发现实现。
//
// 中文说明：
// - 使用 etcd KV + Lease API 实现服务注册与发现；
// - 通过租约（TTL）实现健康检查，服务下线自动注销；
// - 支持服务元数据（版本、权重等）；
// - 与 configsource.etcd 共用 etcd client，减少连接开销。
type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string     { return "discovery.etcd" }
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

// DiscoveryConfig 定义 etcd 服务发现配置。
type DiscoveryConfig struct {
	// etcd 配置
	EtcdEndpoints []string
	EtcdUsername  string
	EtcdPassword  string

	// 服务注册路径前缀
	ServicePath string // 如 "/services/"

	// 租约配置（健康检查）
	LeaseTTL int64 // 租约 TTL（秒），默认 10

	// 服务注册配置
	ServiceName string
	ServiceAddr string
	ServicePort int
	ServiceMeta map[string]string

	// 负载均衡策略
	LoadBalance string // random/round_robin
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
		ServicePath: "/services/",
		LeaseTTL:    10, // 默认 10 秒
		LoadBalance: "random",
	}

	// etcd endpoints
	endpoints := configprovider.GetStringSliceAny(cfg,
		"discovery.etcd.endpoints",
		"discovery.etcd_endpoints",
	)
	if len(endpoints) == 0 {
		endpoints = []string{"localhost:2379"}
	}
	discCfg.EtcdEndpoints = endpoints

	// etcd 认证
	if username := configprovider.GetStringAny(cfg,
		"discovery.etcd.username",
		"discovery.etcd_username",
	); username != "" {
		discCfg.EtcdUsername = username
	}
	if password := configprovider.GetStringAny(cfg,
		"discovery.etcd.password",
		"discovery.etcd_password",
	); password != "" {
		discCfg.EtcdPassword = password
	}

	// 服务路径
	if servicePath := configprovider.GetStringAny(cfg,
		"discovery.service.path",
		"discovery.service_path",
	); servicePath != "" {
		discCfg.ServicePath = servicePath
	}

	// 租约 TTL
	if ttl := configprovider.GetIntAny(cfg,
		"discovery.etcd.lease_ttl",
		"discovery.lease_ttl",
	); ttl > 0 {
		discCfg.LeaseTTL = int64(ttl)
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

	// 负载均衡策略
	if lb := configprovider.GetStringAny(cfg,
		"selector.algorithm",
		"discovery.load_balance",
	); lb != "" {
		discCfg.LoadBalance = lb
	}

	return discCfg, nil
}

// Registry 是 etcd 服务发现实现。
//
// 中文说明：
// - 使用 etcd KV 存储服务信息；
// - 使用 Lease（租约）实现健康检查；
// - 服务定期续约保持存活状态；
// - 租约过期后服务自动注销。
type Registry struct {
	cfg    *DiscoveryConfig
	client *clientv3.Client

	// 已注册服务缓存（用于续约）
	registered sync.Map // map[string]*registeredService
	mu         sync.Mutex
	closed     bool
}

// registeredService 已注册的服务信息。
type registeredService struct {
	serviceID string
	leaseID   clientv3.LeaseID
	keepAlive clientv3.LeaseKeepAliveResponse
	stopCh    chan struct{}
}

// NewRegistry 创建 etcd 服务发现实例。
func NewRegistry(cfg *DiscoveryConfig) (*Registry, error) {
	clientCfg := clientv3.Config{
		Endpoints:   cfg.EtcdEndpoints,
		DialTimeout: 5 * time.Second,
	}

	if cfg.EtcdUsername != "" && cfg.EtcdPassword != "" {
		clientCfg.Username = cfg.EtcdUsername
		clientCfg.Password = cfg.EtcdPassword
	}

	client, err := clientv3.New(clientCfg)
	if err != nil {
		return nil, fmt.Errorf("discovery.etcd: create client failed: %w", err)
	}

	return &Registry{
		cfg:    cfg,
		client: client,
	}, nil
}

// Register 注册服务实例。
//
// 中文说明：
// - 创建租约（Lease），设置 TTL；
// - 将服务信息写入 etcd KV，绑定租约；
// - 启动后台 goroutine 定期续约（KeepAlive）；
// - 租约过期后服务自动注销。
func (r *Registry) Register(ctx context.Context, name, addr string, meta map[string]string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return errors.New("discovery.etcd: registry closed")
	}

	// 解析地址获取端口
	host, port := parseAddr(addr)
	if port == 0 {
		port = r.cfg.ServicePort
	}

	// 生成唯一服务 ID
	serviceID := generateServiceID(name, host, port)

	// 检查是否已注册
	if _, ok := r.registered.Load(serviceID); ok {
		// 已注册，无需重复注册
		return nil
	}

	// 创建租约
	leaseResp, err := r.client.Lease.Grant(ctx, r.cfg.LeaseTTL)
	if err != nil {
		return fmt.Errorf("discovery.etcd: create lease failed: %w", err)
	}
	leaseID := leaseResp.ID

	// 构建服务信息
	serviceInfo := map[string]any{
		"name":     name,
		"address":  host,
		"port":     port,
		"meta":     meta,
		"healthy":  true,
		"registered_at": time.Now().Unix(),
	}
	serviceData, _ := json.Marshal(serviceInfo)

	// 服务 KV 路径
	serviceKey := path.Join(r.cfg.ServicePath, name, serviceID)

	// 写入 KV（绑定租约）
	_, err = r.client.KV.Put(ctx, serviceKey, string(serviceData), clientv3.WithLease(leaseID))
	if err != nil {
		return fmt.Errorf("discovery.etcd: put service failed: %w", err)
	}

	// 启动续约 goroutine
	stopCh := make(chan struct{})
	keepAliveCh, err := r.client.Lease.KeepAlive(ctx, leaseID)
	if err != nil {
		return fmt.Errorf("discovery.etcd: start keepalive failed: %w", err)
	}

	go r.keepAliveLoop(serviceID, keepAliveCh, stopCh)

	// 缓存已注册服务
	r.registered.Store(serviceID, &registeredService{
		serviceID: serviceID,
		leaseID:   leaseID,
		stopCh:    stopCh,
	})

	return nil
}

// Deregister 注销服务实例。
//
// 中文说明：
// - 撤销租约，服务 KV 自动删除；
// - 停止续约 goroutine。
func (r *Registry) Deregister(ctx context.Context, name, addr string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 查找已注册的服务
	host, port := parseAddr(addr)
	if port == 0 {
		port = r.cfg.ServicePort
	}
	serviceID := generateServiceID(name, host, port)

	// 获取已注册服务信息
	if cached, ok := r.registered.Load(serviceID); ok {
		reg := cached.(*registeredService)

		// 停止续约
		close(reg.stopCh)

		// 撤销租约
		_, _ = r.client.Lease.Revoke(ctx, reg.leaseID)

		// 删除缓存
		r.registered.Delete(serviceID)
	}

	return nil
}

// Discover 发现服务实例。
//
// 中文说明：
// - 从 etcd KV 读取服务列表；
// - 只返回未过期（租约有效）的服务；
// - 支持负载均衡策略选择实例。
func (r *Registry) Discover(ctx context.Context, name string) ([]contract.ServiceInstance, error) {
	// 服务路径前缀
	servicePrefix := path.Join(r.cfg.ServicePath, name) + "/"

	// 查询 KV
	resp, err := r.client.KV.Get(ctx, servicePrefix, clientv3.WithPrefix())
	if err != nil {
		return nil, fmt.Errorf("discovery.etcd: get services failed: %w", err)
	}

	// 转换为 ServiceInstance 列表
	instances := make([]contract.ServiceInstance, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		// 解析服务信息
		var info map[string]any
		if err := json.Unmarshal(kv.Value, &info); err != nil {
			continue
		}

		// 提取字段
		serviceName := getString(info, "name")
		address := getString(info, "address")
		port := getInt(info, "port")
		healthy := getBool(info, "healthy")
		meta := getMap(info, "meta")

		// 构建完整地址
		fullAddr := fmt.Sprintf("%s:%d", address, port)

		// 提取服务 ID（从 Key 路径）
		serviceID := strings.TrimPrefix(string(kv.Key), servicePrefix)

		instances = append(instances, contract.ServiceInstance{
			ID:       serviceID,
			Name:     serviceName,
			Address:  fullAddr,
			Metadata: meta,
			Healthy:  healthy,
		})
	}

	// 应用负载均衡策略
	if len(instances) > 1 {
		instances = r.applyLoadBalance(instances)
	}

	return instances, nil
}

// Close 关闭服务发现连接。
//
// 中文说明：
// - 注销所有已注册的服务（撤销租约）；
// - 关闭 etcd client。
func (r *Registry) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return nil
	}
	r.closed = true

	// 注销所有已注册的服务
	r.registered.Range(func(key, value any) bool {
		reg := value.(*registeredService)
		close(reg.stopCh)
		_, _ = r.client.Lease.Revoke(context.Background(), reg.leaseID)
		return true
	})

	return r.client.Close()
}

// keepAliveLoop 续约循环。
//
// 中文说明：
// - 定期接收 KeepAlive 响应，维持租约有效；
// - 如果 KeepAlive 失败，尝试重新注册；
// - 收到 stopCh 信号后停止。
func (r *Registry) keepAliveLoop(serviceID string, keepAliveCh <-chan *clientv3.LeaseKeepAliveResponse, stopCh chan struct{}) {
	for {
		select {
		case <-stopCh:
			return
		case resp, ok := <-keepAliveCh:
			if !ok {
				// KeepAlive 通道关闭，租约失效
				// TODO: 触发重新注册逻辑
				return
			}
			if resp != nil {
				// 续约成功，租约 TTL 重置
				// 可用于监控租约状态
			}
		}
	}
}

// applyLoadBalance 应用负载均衡策略。
func (r *Registry) applyLoadBalance(instances []contract.ServiceInstance) []contract.ServiceInstance {
	switch r.cfg.LoadBalance {
	case "random":
		rand.Shuffle(len(instances), func(i, j int) {
			instances[i], instances[j] = instances[j], instances[i]
		})
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

// generateServiceID 生成唯一服务 ID。
func generateServiceID(name, host string, port int) string {
	return fmt.Sprintf("%s-%s-%d", name, host, port)
}

// 辅助函数：从 map 提取字段
func getString(m map[string]any, key string) string {
	if v, ok := m[key]; ok {
		return fmt.Sprintf("%v", v)
	}
	return ""
}

func getInt(m map[string]any, key string) int {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case int:
			return val
		case int64:
			return int(val)
		case float64:
			return int(val)
		}
	}
	return 0
}

func getBool(m map[string]any, key string) bool {
	if v, ok := m[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return true // 默认健康
}

func getMap(m map[string]any, key string) map[string]string {
	result := make(map[string]string)
	if v, ok := m[key]; ok {
		if meta, ok := v.(map[string]any); ok {
			for k, val := range meta {
				result[k] = fmt.Sprintf("%v", val)
			}
		}
	}
	return result
}