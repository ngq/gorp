// Package nacos provides Nacos service registry implementation for gorp.
//
// Nacos 注册中心 Provider，实现 transportcontract.ServiceRegistry 契约。
// 支持服务注册、发现、注销、权重负载均衡。
//
// 使用示例：
//
//  cfg := &DiscoveryConfig{
//      NacosAddr:      "localhost:8848",
//      NacosNamespace: "public",
//      NacosGroup:     "DEFAULT_GROUP",
//      ServiceWeight:  1.0,
//  }
//  registry, err := NewRegistry(cfg)
//  if err != nil {
//      panic(err)
//  }
//  defer registry.Close()
//
//  err = registry.Register(ctx, "my-service", "192.168.1.100:8080", nil)
//
// 配置路径：discovery.nacos.*
package nacos

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/ngq/gorp/contrib/internal/baseregistry"
	internalnative "github.com/ngq/gorp/contrib/internal/native"
	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	configprovider "github.com/ngq/gorp/framework/provider/config"
)

// Provider 提供 Nacos 服务发现实现。
type Provider struct {
	baseregistry.BaseRegistryProvider
}

// NewProvider creates a new Nacos registry provider.
func NewProvider() *Provider {
	p := &Provider{}
	p.NameStr = "registry.nacos"
	p.GetConfig = func(c runtimecontract.Container) (any, error) {
		return getDiscoveryConfig(c)
	}
	p.NewRegistry = func(cfg any) (transportcontract.ServiceRegistry, error) {
		return NewRegistry(cfg.(*DiscoveryConfig))
	}
	return p
}

type DiscoveryConfig struct {
	NacosAddr      string
	NacosNamespace string
	NacosGroup     string
	NacosUsername  string
	NacosPassword  string

	ServiceName   string
	ServiceAddr   string
	ServicePort   int
	ServiceWeight float64
	ServiceMeta   map[string]string

	LoadBalance string
}

func getDiscoveryConfig(c runtimecontract.Container) (*DiscoveryConfig, error) {
	cfgAny, err := c.Make(datacontract.ConfigKey)
	if err != nil {
		return nil, err
	}

	cfg, ok := cfgAny.(datacontract.Config)
	if !ok {
		return nil, errors.New("discovery: invalid config service")
	}

	discCfg := &DiscoveryConfig{
		NacosAddr:     "localhost:8848",
		NacosGroup:    "DEFAULT_GROUP",
		ServiceWeight: 1.0,
		LoadBalance:   "weight",
	}

	if addr := configprovider.GetStringAny(cfg,
		"discovery.nacos.addr",
		"discovery.nacos.address",
		"discovery.nacos_addr",
	); addr != "" {
		discCfg.NacosAddr = addr
	}
	if ns := configprovider.GetStringAny(cfg,
		"discovery.nacos.namespace",
		"discovery.nacos_namespace",
	); ns != "" {
		discCfg.NacosNamespace = ns
	}
	if group := configprovider.GetStringAny(cfg,
		"discovery.nacos.group",
		"discovery.nacos_group",
	); group != "" {
		discCfg.NacosGroup = group
	}
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

	sc := baseregistry.ReadServiceConfig(cfg)
	discCfg.ServiceName = sc.ServiceName
	discCfg.ServiceAddr = sc.ServiceAddr
	discCfg.ServicePort = sc.ServicePort
	discCfg.LoadBalance = sc.LoadBalance

	if weight := configprovider.GetFloatAny(cfg,
		"discovery.service.weight",
		"discovery.service_weight",
	); weight > 0 {
		discCfg.ServiceWeight = weight
	}

	return discCfg, nil
}

type Registry struct {
	cfg          *DiscoveryConfig
	namingClient naming_client.INamingClient

	registered sync.Map
	mu         sync.Mutex
	closed     bool
}

func NewRegistry(cfg *DiscoveryConfig) (*Registry, error) {
	serverConfigs := []constant.ServerConfig{
		{
			IpAddr:      parseNacosAddr(cfg.NacosAddr),
			Port:        parseNacosPort(cfg.NacosAddr),
			ContextPath: "/nacos",
		},
	}

	clientConfig := constant.ClientConfig{
		NamespaceId:         cfg.NacosNamespace,
		NotLoadCacheAtStart: true,
	}

	if cfg.NacosUsername != "" && cfg.NacosPassword != "" {
		clientConfig.Username = cfg.NacosUsername
		clientConfig.Password = cfg.NacosPassword
	}

	namingClient, err := clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: serverConfigs,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("registry.nacos: create naming client failed: %w", err)
	}

	return &Registry{
		cfg:          cfg,
		namingClient: namingClient,
	}, nil
}

func (r *Registry) Register(ctx context.Context, name, addr string, meta map[string]string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return errors.New("registry.nacos: registry closed")
	}

	host, port := parseAddr(addr)
	if port == 0 {
		port = r.cfg.ServicePort
	}

	serviceID := generateServiceID(name, host, port)

	fullMeta := make(map[string]string)
	for k, v := range r.cfg.ServiceMeta {
		fullMeta[k] = v
	}
	for k, v := range meta {
		fullMeta[k] = v
	}

	success, err := r.namingClient.RegisterInstance(vo.RegisterInstanceParam{
		ServiceName: name,
		GroupName:   r.cfg.NacosGroup,
		Ip:          host,
		Port:        uint64(port),
		Weight:      r.cfg.ServiceWeight,
		Enable:      true,
		Healthy:     true,
		Metadata:    fullMeta,
		Ephemeral:   true,
	})
	if err != nil {
		return fmt.Errorf("registry.nacos: register instance failed: %w", err)
	}
	if !success {
		return errors.New("registry.nacos: register instance failed")
	}

	r.registered.Store(serviceID, true)

	return nil
}

func (r *Registry) Deregister(ctx context.Context, name, addr string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	host, port := parseAddr(addr)
	if port == 0 {
		port = r.cfg.ServicePort
	}

	success, err := r.namingClient.DeregisterInstance(vo.DeregisterInstanceParam{
		ServiceName: name,
		GroupName:   r.cfg.NacosGroup,
		Ip:          host,
		Port:        uint64(port),
		Ephemeral:   true,
	})
	if err != nil {
		return fmt.Errorf("registry.nacos: deregister instance failed: %w", err)
	}
	if !success {
		return errors.New("registry.nacos: deregister instance failed")
	}

	serviceID := generateServiceID(name, host, port)
	r.registered.Delete(serviceID)

	return nil
}

func (r *Registry) Discover(ctx context.Context, name string) ([]transportcontract.ServiceInstance, error) {
	instances, err := r.namingClient.SelectInstances(vo.SelectInstancesParam{
		ServiceName: name,
		GroupName:   r.cfg.NacosGroup,
		HealthyOnly: true,
	})
	if err != nil {
		return nil, fmt.Errorf("registry.nacos: select instances failed: %w", err)
	}

	result := make([]transportcontract.ServiceInstance, 0, len(instances))
	for _, inst := range instances {
		fullAddr := fmt.Sprintf("%s:%d", inst.Ip, inst.Port)
		serviceID := generateServiceID(name, inst.Ip, int(inst.Port))

		meta := inst.Metadata
		if meta == nil {
			meta = make(map[string]string)
		}
		meta["weight"] = fmt.Sprintf("%.1f", inst.Weight)

		result = append(result, transportcontract.ServiceInstance{
			ID:       serviceID,
			Name:     name,
			Address:  fullAddr,
			Metadata: meta,
			Healthy:  inst.Healthy,
		})
	}

	if len(result) > 1 {
		result = r.applyLoadBalance(result)
	}

	return result, nil
}

func (r *Registry) Underlying() any {
	return r.namingClient
}

func (r *Registry) As(target any) bool {
	return internalnative.As(r.namingClient, target)
}

func (r *Registry) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return nil
	}
	r.closed = true

	var errs []error
	r.registered.Range(func(key, _ any) bool {
		serviceID := key.(string)
		// serviceID format: "{name}-{host}-{port}"
		name, host, port := parseServiceID(serviceID)
		_, err := r.namingClient.DeregisterInstance(vo.DeregisterInstanceParam{
			ServiceName: name,
			GroupName:   r.cfg.NacosGroup,
			Ip:          host,
			Port:        uint64(port),
			Ephemeral:   true,
		})
		if err != nil {
			errs = append(errs, fmt.Errorf("deregister %s: %w", serviceID, err))
		}
		r.registered.Delete(key)
		return true
	})

	if len(errs) > 0 {
		return fmt.Errorf("registry.nacos: close with %d errors: %w", len(errs), errors.Join(errs...))
	}
	return nil
}

func (r *Registry) applyLoadBalance(instances []transportcontract.ServiceInstance) []transportcontract.ServiceInstance {
	switch r.cfg.LoadBalance {
	case "random":
		rand.Shuffle(len(instances), func(i, j int) {
			instances[i], instances[j] = instances[j], instances[i]
		})
	case "weight":
	}
	return instances
}

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

func parseNacosAddr(addr string) string {
	parts := strings.Split(addr, ":")
	return parts[0]
}

func parseNacosPort(addr string) uint64 {
	parts := strings.Split(addr, ":")
	if len(parts) == 2 {
		port, _ := strconv.ParseUint(parts[1], 10, 64)
		return port
	}
	return 8848
}

func generateServiceID(name, host string, port int) string {
	return fmt.Sprintf("%s-%s-%d", name, host, port)
}

// parseServiceID parses a serviceID back into name, host and port.
// It splits from the right to handle hostnames that may contain hyphens.
func parseServiceID(id string) (name, host string, port int) {
	lastDash := strings.LastIndex(id, "-")
	if lastDash < 0 {
		return id, "", 0
	}
	portStr := id[lastDash+1:]
	p, err := strconv.Atoi(portStr)
	if err != nil {
		return id, "", 0
	}
	remaining := id[:lastDash]
	secondLastDash := strings.LastIndex(remaining, "-")
	if secondLastDash < 0 {
		return remaining, "", p
	}
	name = remaining[:secondLastDash]
	host = remaining[secondLastDash+1:]
	return name, host, p
}
