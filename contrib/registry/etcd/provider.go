package etcd

import (
	"errors"

	"github.com/ngq/gorp/contrib/internal/baseregistry"
	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	configprovider "github.com/ngq/gorp/framework/provider/config"
)

// Provider 提供 etcd 服务发现实现。
type Provider struct {
	baseregistry.BaseRegistryProvider
}

// NewProvider creates a new etcd registry provider.
func NewProvider() *Provider {
	p := &Provider{}
	p.NameStr = "registry.etcd"
	p.GetConfig = func(c runtimecontract.Container) (any, error) {
		return getDiscoveryConfig(c)
	}
	p.NewRegistry = func(cfg any) (transportcontract.ServiceRegistry, error) {
		return NewRegistry(cfg.(*DiscoveryConfig))
	}
	return p
}

type DiscoveryConfig struct {
	EtcdEndpoints []string
	EtcdUsername  string
	EtcdPassword  string

	ServicePath string
	LeaseTTL    int64

	ServiceName string
	ServiceAddr string
	ServicePort int
	ServiceMeta map[string]string

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
		ServicePath: "/services/",
		LeaseTTL:    10,
		LoadBalance: "random",
	}

	endpoints := configprovider.GetStringSliceAny(cfg,
		"discovery.etcd.endpoints",
		"discovery.etcd_endpoints",
	)
	if len(endpoints) == 0 {
		endpoints = []string{"localhost:2379"}
	}
	discCfg.EtcdEndpoints = endpoints

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

	if servicePath := configprovider.GetStringAny(cfg,
		"discovery.service.path",
		"discovery.service_path",
	); servicePath != "" {
		discCfg.ServicePath = servicePath
	}

	if ttl := configprovider.GetIntAny(cfg,
		"discovery.etcd.lease_ttl",
		"discovery.lease_ttl",
	); ttl > 0 {
		discCfg.LeaseTTL = int64(ttl)
	}

	sc := baseregistry.ReadServiceConfig(cfg)
	discCfg.ServiceName = sc.ServiceName
	discCfg.ServiceAddr = sc.ServiceAddr
	discCfg.ServicePort = sc.ServicePort
	discCfg.LoadBalance = sc.LoadBalance

	return discCfg, nil
}
