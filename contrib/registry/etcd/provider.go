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
// - 支持服务元数据与负载均衡策略；
// - 当前已从 framework/provider 真实下沉到 contrib 层。
type Provider struct{}

func NewProvider() *Provider { return &Provider{} }
func (p *Provider) Name() string     { return "discovery.etcd" }
func (p *Provider) IsDefer() bool    { return true }
func (p *Provider) Provides() []string { return []string{contract.RPCRegistryKey} }

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
func (p *Provider) Boot(c contract.Container) error { return nil }

type DiscoveryConfig struct {
	EtcdEndpoints []string
	EtcdUsername  string
	EtcdPassword  string
	ServicePath   string
	LeaseTTL      int64
	ServiceName   string
	ServiceAddr   string
	ServicePort   int
	ServiceMeta   map[string]string
	LoadBalance   string
}

func getDiscoveryConfig(c contract.Container) (*DiscoveryConfig, error) {
	cfgAny, err := c.Make(contract.ConfigKey)
	if err != nil {
		return nil, err
	}
	cfg, ok := cfgAny.(contract.Config)
	if !ok {
		return nil, errors.New("discovery: invalid config service")
	}
	discCfg := &DiscoveryConfig{ServicePath: "/services/", LeaseTTL: 10, LoadBalance: "random"}
	endpoints := configprovider.GetStringSliceAny(cfg, "discovery.etcd.endpoints", "discovery.etcd_endpoints")
	if len(endpoints) == 0 { endpoints = []string{"localhost:2379"} }
	discCfg.EtcdEndpoints = endpoints
	if username := configprovider.GetStringAny(cfg, "discovery.etcd.username", "discovery.etcd_username"); username != "" { discCfg.EtcdUsername = username }
	if password := configprovider.GetStringAny(cfg, "discovery.etcd.password", "discovery.etcd_password"); password != "" { discCfg.EtcdPassword = password }
	if servicePath := configprovider.GetStringAny(cfg, "discovery.service.path", "discovery.service_path"); servicePath != "" { discCfg.ServicePath = servicePath }
	if ttl := configprovider.GetIntAny(cfg, "discovery.etcd.lease_ttl", "discovery.lease_ttl"); ttl > 0 { discCfg.LeaseTTL = int64(ttl) }
	if name := configprovider.GetStringAny(cfg, "discovery.service.name", "discovery.service_name"); name != "" { discCfg.ServiceName = name }
	if addr := configprovider.GetStringAny(cfg, "discovery.service.addr", "discovery.service.address", "discovery.service_addr"); addr != "" { discCfg.ServiceAddr = addr }
	if port := configprovider.GetIntAny(cfg, "discovery.service.port", "discovery.service_port"); port > 0 { discCfg.ServicePort = port }
	if lb := configprovider.GetStringAny(cfg, "selector.algorithm", "discovery.load_balance"); lb != "" { discCfg.LoadBalance = lb }
	return discCfg, nil
}

type Registry struct {
	cfg        *DiscoveryConfig
	client     *clientv3.Client
	registered sync.Map
	mu         sync.Mutex
	closed     bool
}

type registeredService struct {
	serviceID string
	leaseID   clientv3.LeaseID
	stopCh    chan struct{}
}

func NewRegistry(cfg *DiscoveryConfig) (*Registry, error) {
	clientCfg := clientv3.Config{Endpoints: cfg.EtcdEndpoints, DialTimeout: 5 * time.Second}
	if cfg.EtcdUsername != "" && cfg.EtcdPassword != "" {
		clientCfg.Username = cfg.EtcdUsername
		clientCfg.Password = cfg.EtcdPassword
	}
	client, err := clientv3.New(clientCfg)
	if err != nil {
		return nil, fmt.Errorf("discovery.etcd: create client failed: %w", err)
	}
	return &Registry{cfg: cfg, client: client}, nil
}

func (r *Registry) Register(ctx context.Context, name, addr string, meta map[string]string) error {
	r.mu.Lock(); defer r.mu.Unlock()
	if r.closed { return errors.New("discovery.etcd: registry closed") }
	host, port := parseAddr(addr)
	if port == 0 { port = r.cfg.ServicePort }
	serviceID := generateServiceID(name, host, port)
	if _, ok := r.registered.Load(serviceID); ok { return nil }
	leaseResp, err := r.client.Lease.Grant(ctx, r.cfg.LeaseTTL)
	if err != nil { return fmt.Errorf("discovery.etcd: create lease failed: %w", err) }
	leaseID := leaseResp.ID
	serviceInfo := map[string]any{"name": name, "address": host, "port": port, "meta": meta, "healthy": true, "registered_at": time.Now().Unix()}
	serviceData, _ := json.Marshal(serviceInfo)
	serviceKey := path.Join(r.cfg.ServicePath, name, serviceID)
	_, err = r.client.KV.Put(ctx, serviceKey, string(serviceData), clientv3.WithLease(leaseID))
	if err != nil { return fmt.Errorf("discovery.etcd: put service failed: %w", err) }
	stopCh := make(chan struct{})
	keepAliveCh, err := r.client.Lease.KeepAlive(ctx, leaseID)
	if err != nil { return fmt.Errorf("discovery.etcd: start keepalive failed: %w", err) }
	go r.keepAliveLoop(keepAliveCh, stopCh)
	r.registered.Store(serviceID, &registeredService{serviceID: serviceID, leaseID: leaseID, stopCh: stopCh})
	return nil
}

func (r *Registry) Deregister(ctx context.Context, name, addr string) error {
	r.mu.Lock(); defer r.mu.Unlock()
	host, port := parseAddr(addr)
	if port == 0 { port = r.cfg.ServicePort }
	serviceID := generateServiceID(name, host, port)
	if cached, ok := r.registered.Load(serviceID); ok {
		reg := cached.(*registeredService)
		close(reg.stopCh)
		_, _ = r.client.Lease.Revoke(ctx, reg.leaseID)
		r.registered.Delete(serviceID)
	}
	return nil
}

func (r *Registry) Discover(ctx context.Context, name string) ([]contract.ServiceInstance, error) {
	servicePrefix := path.Join(r.cfg.ServicePath, name) + "/"
	resp, err := r.client.KV.Get(ctx, servicePrefix, clientv3.WithPrefix())
	if err != nil { return nil, fmt.Errorf("discovery.etcd: get services failed: %w", err) }
	instances := make([]contract.ServiceInstance, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		var info map[string]any
		if err := json.Unmarshal(kv.Value, &info); err != nil { continue }
		serviceName := getString(info, "name")
		address := getString(info, "address")
		port := getInt(info, "port")
		healthy := getBool(info, "healthy")
		meta := getMap(info, "meta")
		fullAddr := fmt.Sprintf("%s:%d", address, port)
		serviceID := strings.TrimPrefix(string(kv.Key), servicePrefix)
		instances = append(instances, contract.ServiceInstance{ID: serviceID, Name: serviceName, Address: fullAddr, Metadata: meta, Healthy: healthy})
	}
	if len(instances) > 1 { instances = r.applyLoadBalance(instances) }
	return instances, nil
}

func (r *Registry) Close() error {
	r.mu.Lock(); defer r.mu.Unlock()
	if r.closed { return nil }
	r.closed = true
	r.registered.Range(func(key, value any) bool {
		reg := value.(*registeredService)
		close(reg.stopCh)
		_, _ = r.client.Lease.Revoke(context.Background(), reg.leaseID)
		return true
	})
	return r.client.Close()
}

func (r *Registry) keepAliveLoop(keepAliveCh <-chan *clientv3.LeaseKeepAliveResponse, stopCh chan struct{}) {
	for {
		select {
		case <-stopCh:
			return
		case _, ok := <-keepAliveCh:
			if !ok { return }
		}
	}
}

func (r *Registry) applyLoadBalance(instances []contract.ServiceInstance) []contract.ServiceInstance {
	switch r.cfg.LoadBalance { case "random": rand.Shuffle(len(instances), func(i, j int) { instances[i], instances[j] = instances[j], instances[i] }) }
	return instances
}
func parseAddr(addr string) (host string, port int) {
	parts := strings.Split(addr, ":")
	if len(parts) == 2 { host = parts[0]; port, _ = strconv.Atoi(parts[1]) } else { host = addr }
	return host, port
}
func generateServiceID(name, host string, port int) string { return fmt.Sprintf("%s-%s-%d", name, host, port) }
func getString(m map[string]any, key string) string { if v, ok := m[key]; ok { return fmt.Sprintf("%v", v) }; return "" }
func getInt(m map[string]any, key string) int { if v, ok := m[key]; ok { switch val := v.(type) { case int: return val; case int64: return int(val); case float64: return int(val) } }; return 0 }
func getBool(m map[string]any, key string) bool { if v, ok := m[key]; ok { if b, ok := v.(bool); ok { return b } }; return true }
func getMap(m map[string]any, key string) map[string]string {
	result := make(map[string]string)
	if v, ok := m[key]; ok {
		if meta, ok := v.(map[string]any); ok {
			for k, val := range meta { result[k] = fmt.Sprintf("%v", val) }
		}
	}
	return result
}
