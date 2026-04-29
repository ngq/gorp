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
// - 当前已从 framework/provider 真实下沉到 contrib 层。
type Provider struct{}

func NewProvider() *Provider { return &Provider{} }
func (p *Provider) Name() string     { return "discovery.nacos" }
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
	NacosAddr      string
	NacosNamespace string
	NacosGroup     string
	NacosUsername  string
	NacosPassword  string
	ServiceName    string
	ServiceAddr    string
	ServicePort    int
	ServiceWeight  float64
	ServiceMeta    map[string]string
	LoadBalance    string
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
	discCfg := &DiscoveryConfig{NacosAddr: "localhost:8848", NacosGroup: "DEFAULT_GROUP", ServiceWeight: 1.0, LoadBalance: "weight"}
	if addr := configprovider.GetStringAny(cfg, "discovery.nacos.addr", "discovery.nacos.address", "discovery.nacos_addr"); addr != "" { discCfg.NacosAddr = addr }
	if ns := configprovider.GetStringAny(cfg, "discovery.nacos.namespace", "discovery.nacos_namespace"); ns != "" { discCfg.NacosNamespace = ns }
	if group := configprovider.GetStringAny(cfg, "discovery.nacos.group", "discovery.nacos_group"); group != "" { discCfg.NacosGroup = group }
	if username := configprovider.GetStringAny(cfg, "discovery.nacos.username", "discovery.nacos_username"); username != "" { discCfg.NacosUsername = username }
	if password := configprovider.GetStringAny(cfg, "discovery.nacos.password", "discovery.nacos_password"); password != "" { discCfg.NacosPassword = password }
	if name := configprovider.GetStringAny(cfg, "discovery.service.name", "discovery.service_name"); name != "" { discCfg.ServiceName = name }
	if addr := configprovider.GetStringAny(cfg, "discovery.service.addr", "discovery.service.address", "discovery.service_addr"); addr != "" { discCfg.ServiceAddr = addr }
	if port := configprovider.GetIntAny(cfg, "discovery.service.port", "discovery.service_port"); port > 0 { discCfg.ServicePort = port }
	if weight := configprovider.GetFloatAny(cfg, "discovery.service.weight", "discovery.service_weight"); weight > 0 { discCfg.ServiceWeight = weight }
	if lb := configprovider.GetStringAny(cfg, "selector.algorithm", "discovery.load_balance"); lb != "" { discCfg.LoadBalance = lb }
	return discCfg, nil
}

type Registry struct {
	cfg          *DiscoveryConfig
	namingClient naming_client.INamingClient
	registered   sync.Map
	mu           sync.Mutex
	closed       bool
}

func NewRegistry(cfg *DiscoveryConfig) (*Registry, error) {
	serverConfigs := []constant.ServerConfig{{IpAddr: parseNacosAddr(cfg.NacosAddr), Port: parseNacosPort(cfg.NacosAddr), ContextPath: "/nacos"}}
	clientConfig := constant.ClientConfig{NamespaceId: cfg.NacosNamespace, NotLoadCacheAtStart: true}
	if cfg.NacosUsername != "" && cfg.NacosPassword != "" {
		clientConfig.Username = cfg.NacosUsername
		clientConfig.Password = cfg.NacosPassword
	}
	namingClient, err := clients.NewNamingClient(vo.NacosClientParam{ClientConfig: &clientConfig, ServerConfigs: serverConfigs})
	if err != nil {
		return nil, fmt.Errorf("discovery.nacos: create naming client failed: %w", err)
	}
	return &Registry{cfg: cfg, namingClient: namingClient}, nil
}

func (r *Registry) Register(ctx context.Context, name, addr string, meta map[string]string) error {
	r.mu.Lock(); defer r.mu.Unlock()
	if r.closed { return errors.New("discovery.nacos: registry closed") }
	host, port := parseAddr(addr)
	if port == 0 { port = r.cfg.ServicePort }
	serviceID := generateServiceID(name, host, port)
	fullMeta := make(map[string]string)
	for k, v := range r.cfg.ServiceMeta { fullMeta[k] = v }
	for k, v := range meta { fullMeta[k] = v }
	success, err := r.namingClient.RegisterInstance(vo.RegisterInstanceParam{ServiceName: name, GroupName: r.cfg.NacosGroup, Ip: host, Port: uint64(port), Weight: r.cfg.ServiceWeight, Enable: true, Healthy: true, Metadata: fullMeta, Ephemeral: true})
	if err != nil { return fmt.Errorf("discovery.nacos: register instance failed: %w", err) }
	if !success { return errors.New("discovery.nacos: register instance failed") }
	r.registered.Store(serviceID, true)
	return nil
}
func (r *Registry) Deregister(ctx context.Context, name, addr string) error {
	r.mu.Lock(); defer r.mu.Unlock()
	host, port := parseAddr(addr)
	if port == 0 { port = r.cfg.ServicePort }
	success, err := r.namingClient.DeregisterInstance(vo.DeregisterInstanceParam{ServiceName: name, GroupName: r.cfg.NacosGroup, Ip: host, Port: uint64(port), Ephemeral: true})
	if err != nil { return fmt.Errorf("discovery.nacos: deregister instance failed: %w", err) }
	if !success { return errors.New("discovery.nacos: deregister instance failed") }
	serviceID := generateServiceID(name, host, port)
	r.registered.Delete(serviceID)
	return nil
}
func (r *Registry) Discover(ctx context.Context, name string) ([]contract.ServiceInstance, error) {
	instances, err := r.namingClient.SelectInstances(vo.SelectInstancesParam{ServiceName: name, GroupName: r.cfg.NacosGroup, HealthyOnly: true})
	if err != nil { return nil, fmt.Errorf("discovery.nacos: select instances failed: %w", err) }
	result := make([]contract.ServiceInstance, 0, len(instances))
	for _, inst := range instances {
		fullAddr := fmt.Sprintf("%s:%d", inst.Ip, inst.Port)
		serviceID := generateServiceID(name, inst.Ip, int(inst.Port))
		meta := inst.Metadata
		if meta == nil { meta = make(map[string]string) }
		meta["weight"] = fmt.Sprintf("%.1f", inst.Weight)
		result = append(result, contract.ServiceInstance{ID: serviceID, Name: name, Address: fullAddr, Metadata: meta, Healthy: inst.Healthy})
	}
	if len(result) > 1 { result = r.applyLoadBalance(result) }
	return result, nil
}
func (r *Registry) Close() error { r.closed = true; return nil }

func (r *Registry) applyLoadBalance(instances []contract.ServiceInstance) []contract.ServiceInstance {
	switch r.cfg.LoadBalance { case "random": rand.Shuffle(len(instances), func(i, j int) { instances[i], instances[j] = instances[j], instances[i] }) }
	return instances
}
func parseAddr(addr string) (host string, port int) {
	parts := strings.Split(addr, ":")
	if len(parts) == 2 { host = parts[0]; port, _ = strconv.Atoi(parts[1]) } else { host = addr }
	return host, port
}
func parseNacosAddr(addr string) string { parts := strings.Split(addr, ":"); return parts[0] }
func parseNacosPort(addr string) uint64 { parts := strings.Split(addr, ":"); if len(parts) == 2 { port, _ := strconv.ParseUint(parts[1], 10, 64); return port }; return 8848 }
func generateServiceID(name, host string, port int) string { return fmt.Sprintf("%s-%s-%d", name, host, port) }
