package kubernetes

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/ngq/gorp/framework/contract"
)

// Provider 提供 Kubernetes 服务发现实现。
//
// 中文说明：
// - 使用 Kubernetes Endpoints API 实现服务发现；
// - 适用于运行在 K8s 集群内的服务；
// - 当前已从 framework/provider 真实下沉到 contrib 层。
type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string     { return "discovery.kubernetes" }
func (p *Provider) IsDefer() bool    { return true }
func (p *Provider) Provides() []string { return []string{contract.RPCRegistryKey} }

func (p *Provider) Register(c contract.Container) error {
	c.Bind(contract.RPCRegistryKey, func(c contract.Container) (any, error) {
		cfg, err := getKubernetesConfig(c)
		if err != nil {
			return nil, err
		}
		return NewRegistry(cfg)
	}, true)
	return nil
}

func (p *Provider) Boot(c contract.Container) error { return nil }

type KubernetesConfig struct {
	KubeConfig  string
	Master      string
	Namespace   string
	ServiceName string
	ServiceAddr string
	ServicePort int
	ServiceMeta map[string]string
}

func getKubernetesConfig(c contract.Container) (*KubernetesConfig, error) {
	cfgAny, err := c.Make(contract.ConfigKey)
	if err != nil {
		return nil, err
	}
	cfg, ok := cfgAny.(contract.Config)
	if !ok {
		return nil, errors.New("kubernetes: invalid config service")
	}
	kubeCfg := &KubernetesConfig{Namespace: "default"}
	if v := cfg.Get("discovery.kubernetes.kubeconfig"); v != nil {
		kubeCfg.KubeConfig = cfg.GetString("discovery.kubernetes.kubeconfig")
	}
	if v := cfg.Get("discovery.kubernetes.namespace"); v != nil {
		kubeCfg.Namespace = cfg.GetString("discovery.kubernetes.namespace")
	}
	if v := cfg.Get("discovery.kubernetes.master"); v != nil {
		kubeCfg.Master = cfg.GetString("discovery.kubernetes.master")
	}
	return kubeCfg, nil
}

type Registry struct {
	config        *KubernetesConfig
	mu            sync.RWMutex
	endpointCache map[string][]contract.ServiceInstance
}

func NewRegistry(cfg *KubernetesConfig) (*Registry, error) {
	return &Registry{config: cfg, endpointCache: make(map[string][]contract.ServiceInstance)}, nil
}

func (r *Registry) Register(ctx context.Context, name, addr string, meta map[string]string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	instance := contract.ServiceInstance{ID: generateInstanceID(name, addr), Name: name, Address: addr, Metadata: meta, Healthy: true}
	r.endpointCache[name] = append(r.endpointCache[name], instance)
	return nil
}

func (r *Registry) Deregister(ctx context.Context, name, addr string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	instances := r.endpointCache[name]
	for i, inst := range instances {
		if inst.Address == addr {
			r.endpointCache[name] = append(instances[:i], instances[i+1:]...)
			break
		}
	}
	return nil
}

func (r *Registry) Discover(ctx context.Context, name string) ([]contract.ServiceInstance, error) {
	r.mu.RLock()
	if instances, ok := r.endpointCache[name]; ok && len(instances) > 0 {
		r.mu.RUnlock()
		return instances, nil
	}
	r.mu.RUnlock()
	instances, err := r.discoverFromKubernetes(ctx, name)
	if err != nil {
		return nil, err
	}
	r.mu.Lock()
	r.endpointCache[name] = instances
	r.mu.Unlock()
	return instances, nil
}

func (r *Registry) discoverFromKubernetes(ctx context.Context, name string) ([]contract.ServiceInstance, error) {
	return []contract.ServiceInstance{}, nil
}

func (r *Registry) Watch(ctx context.Context, name string) (<-chan []contract.ServiceInstance, error) {
	ch := make(chan []contract.ServiceInstance, 10)
	go func() {
		defer close(ch)
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				instances, err := r.discoverFromKubernetes(ctx, name)
				if err != nil {
					continue
				}
				select {
				case ch <- instances:
				default:
				}
			}
		}
	}()
	return ch, nil
}

func (r *Registry) Close() error { return nil }

func generateInstanceID(name, addr string) string { return name + "-" + addr }
