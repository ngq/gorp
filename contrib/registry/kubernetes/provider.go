package kubernetes

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	internalnative "github.com/ngq/gorp/contrib/internal/native"
	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const defaultRegistryPollInterval = 5 * time.Second

var (
	ErrNotInCluster         = errors.New("kubernetes: not in cluster, please provide master or kubeconfig")
	ErrServiceNotFound      = errors.New("kubernetes: service endpoints not found")
	ErrRegistryClosed       = errors.New("kubernetes: registry closed")
	ErrRegisterNotSupported = errors.New("kubernetes: register is not supported")
)

type Provider struct{}

func NewProvider() *Provider                               { return &Provider{} }
func (p *Provider) Name() string                           { return "registry.kubernetes" }
func (p *Provider) IsDefer() bool                          { return true }
func (p *Provider) Provides() []string                     { return []string{transportcontract.RPCRegistryKey} }
func (p *Provider) Boot(c runtimecontract.Container) error { return nil }

func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(transportcontract.RPCRegistryKey, func(c runtimecontract.Container) (any, error) {
		cfg, err := getKubernetesConfig(c)
		if err != nil {
			return nil, err
		}
		return NewRegistry(cfg)
	}, true)
	return nil
}

type KubernetesConfig struct {
	KubeConfig         string
	Master             string
	Namespace          string
	ServiceName        string
	ServiceAddr        string
	ServicePort        int
	ServiceMeta        map[string]string
	InCluster          bool
	BearerToken        string
	CAFile             string
	InsecureSkipVerify bool
	PollInterval       time.Duration
}

func getKubernetesConfig(c runtimecontract.Container) (*KubernetesConfig, error) {
	cfgAny, err := c.Make(datacontract.ConfigKey)
	if err != nil {
		return nil, err
	}
	cfg, ok := cfgAny.(datacontract.Config)
	if !ok {
		return nil, errors.New("kubernetes: invalid config service")
	}

	kubeCfg := &KubernetesConfig{
		Namespace:    "default",
		InCluster:    true,
		PollInterval: defaultRegistryPollInterval,
	}
	if v := cfg.Get("discovery.kubernetes.kubeconfig"); v != nil {
		kubeCfg.KubeConfig = cfg.GetString("discovery.kubernetes.kubeconfig")
	}
	if v := cfg.Get("discovery.kubernetes.namespace"); v != nil {
		kubeCfg.Namespace = cfg.GetString("discovery.kubernetes.namespace")
	}
	if v := cfg.Get("discovery.kubernetes.master"); v != nil {
		kubeCfg.Master = cfg.GetString("discovery.kubernetes.master")
	}
	if v := cfg.Get("discovery.kubernetes.in_cluster"); v != nil {
		kubeCfg.InCluster = cfg.GetBool("discovery.kubernetes.in_cluster")
	}
	if v := cfg.Get("discovery.kubernetes.bearer_token"); v != nil {
		kubeCfg.BearerToken = cfg.GetString("discovery.kubernetes.bearer_token")
	}
	if v := cfg.Get("discovery.kubernetes.ca_file"); v != nil {
		kubeCfg.CAFile = cfg.GetString("discovery.kubernetes.ca_file")
	}
	if v := cfg.Get("discovery.kubernetes.insecure_skip_verify"); v != nil {
		kubeCfg.InsecureSkipVerify = cfg.GetBool("discovery.kubernetes.insecure_skip_verify")
	}
	if v := cfg.Get("discovery.kubernetes.poll_interval_seconds"); v != nil {
		if seconds := cfg.GetInt("discovery.kubernetes.poll_interval_seconds"); seconds > 0 {
			kubeCfg.PollInterval = time.Duration(seconds) * time.Second
		}
	}
	return kubeCfg, nil
}

type discoveryClient interface {
	Discover(ctx context.Context, namespace, name string) ([]transportcontract.ServiceInstance, error)
	Watch(ctx context.Context, namespace, name string, onUpdate func([]transportcontract.ServiceInstance)) error
}

type nativeClientProvider interface {
	Underlying() any
}

type Registry struct {
	config *KubernetesConfig
	client discoveryClient

	mu            sync.RWMutex
	endpointCache map[string][]transportcontract.ServiceInstance
	closeMu       sync.Mutex
	closed        bool
	watchCancels  []context.CancelFunc
}

func NewRegistry(cfg *KubernetesConfig) (*Registry, error) {
	client, err := newDiscoveryClient(cfg)
	if err != nil {
		return nil, err
	}
	return NewRegistryWithClient(cfg, client)
}

func NewRegistryWithClient(cfg *KubernetesConfig, client discoveryClient) (*Registry, error) {
	if client == nil {
		return nil, errors.New("kubernetes: discovery client is required")
	}
	return &Registry{
		config:        cfg,
		client:        client,
		endpointCache: make(map[string][]transportcontract.ServiceInstance),
	}, nil
}

func (r *Registry) Register(ctx context.Context, name, addr string, meta map[string]string) error {
	return ErrRegisterNotSupported
}

func (r *Registry) Deregister(ctx context.Context, name, addr string) error {
	return ErrRegisterNotSupported
}

func (r *Registry) Discover(ctx context.Context, name string) ([]transportcontract.ServiceInstance, error) {
	r.mu.RLock()
	if instances, ok := r.endpointCache[name]; ok && len(instances) > 0 {
		cached := append([]transportcontract.ServiceInstance(nil), instances...)
		r.mu.RUnlock()
		return cached, nil
	}
	closed := r.closed
	r.mu.RUnlock()
	if closed {
		return nil, ErrRegistryClosed
	}

	instances, err := r.client.Discover(ctx, r.config.Namespace, name)
	if err != nil {
		return nil, err
	}
	r.mu.Lock()
	r.endpointCache[name] = append([]transportcontract.ServiceInstance(nil), instances...)
	r.mu.Unlock()
	return instances, nil
}

func (r *Registry) Watch(ctx context.Context, name string) (<-chan []transportcontract.ServiceInstance, error) {
	r.closeMu.Lock()
	defer r.closeMu.Unlock()
	if r.closed {
		return nil, ErrRegistryClosed
	}

	watchCtx, cancel := context.WithCancel(ctx)
	r.watchCancels = append(r.watchCancels, cancel)
	ch := make(chan []transportcontract.ServiceInstance, 10)

	go func() {
		defer close(ch)
		_ = r.client.Watch(watchCtx, r.config.Namespace, name, func(instances []transportcontract.ServiceInstance) {
			r.mu.Lock()
			r.endpointCache[name] = append([]transportcontract.ServiceInstance(nil), instances...)
			r.mu.Unlock()
			select {
			case ch <- append([]transportcontract.ServiceInstance(nil), instances...):
			case <-watchCtx.Done():
			default:
			}
		})
	}()

	go func() {
		instances, err := r.Discover(watchCtx, name)
		if err != nil {
			if errors.Is(err, ErrServiceNotFound) {
				select {
				case ch <- []transportcontract.ServiceInstance{}:
				case <-watchCtx.Done():
				default:
				}
			}
			return
		}
		select {
		case ch <- append([]transportcontract.ServiceInstance(nil), instances...):
		case <-watchCtx.Done():
		default:
		}
	}()

	return ch, nil
}

func (r *Registry) Close() error {
	r.closeMu.Lock()
	defer r.closeMu.Unlock()
	if r.closed {
		return nil
	}
	r.closed = true
	for _, cancel := range r.watchCancels {
		cancel()
	}
	r.watchCancels = nil
	return nil
}

func (r *Registry) Underlying() any {
	if provider, ok := r.client.(nativeClientProvider); ok {
		if native := provider.Underlying(); native != nil {
			return native
		}
	}
	return r.client
}

func (r *Registry) As(target any) bool {
	return internalnative.As(r.Underlying(), target)
}

type clientGoDiscoveryClient struct {
	client kubernetes.Interface
}

func newDiscoveryClient(cfg *KubernetesConfig) (discoveryClient, error) {
	restConfig, err := buildRegistryRESTConfig(cfg)
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("kubernetes: create clientset failed: %w", err)
	}
	return &clientGoDiscoveryClient{client: clientset}, nil
}

func (c *clientGoDiscoveryClient) Underlying() any {
	return c.client
}

func (c *clientGoDiscoveryClient) Discover(ctx context.Context, namespace, name string) ([]transportcontract.ServiceInstance, error) {
	endpoints, err := c.client.CoreV1().Endpoints(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, ErrServiceNotFound
		}
		return nil, fmt.Errorf("kubernetes: discover endpoints failed: %w", err)
	}
	instances := endpointsToInstances(name, endpoints)
	if len(instances) == 0 {
		return nil, ErrServiceNotFound
	}
	return instances, nil
}

func (c *clientGoDiscoveryClient) Watch(ctx context.Context, namespace, name string, onUpdate func([]transportcontract.ServiceInstance)) error {
	for {
		watcher, err := c.client.CoreV1().Endpoints(namespace).Watch(ctx, metav1.ListOptions{
			FieldSelector: fields.OneTermEqualSelector("metadata.name", name).String(),
		})
		if err != nil {
			return fmt.Errorf("kubernetes: watch endpoints failed: %w", err)
		}
		restart, err := consumeEndpointsWatch(ctx, name, watcher, onUpdate)
		if err != nil {
			return err
		}
		if !restart {
			return nil
		}
	}
}

func consumeEndpointsWatch(ctx context.Context, serviceName string, watcher watch.Interface, onUpdate func([]transportcontract.ServiceInstance)) (bool, error) {
	defer watcher.Stop()
	for {
		select {
		case <-ctx.Done():
			return false, nil
		case event, ok := <-watcher.ResultChan():
			if !ok {
				return true, nil
			}
			endpoints, ok := event.Object.(*corev1.Endpoints)
			if !ok || endpoints == nil {
				continue
			}
			if event.Type == watch.Deleted {
				onUpdate([]transportcontract.ServiceInstance{})
				continue
			}
			onUpdate(endpointsToInstances(serviceName, endpoints))
		}
	}
}

func buildRegistryRESTConfig(cfg *KubernetesConfig) (*rest.Config, error) {
	if cfg.KubeConfig != "" {
		return clientcmd.BuildConfigFromFlags(cfg.Master, cfg.KubeConfig)
	}
	if cfg.Master != "" {
		return &rest.Config{
			Host:            cfg.Master,
			BearerToken:     cfg.BearerToken,
			TLSClientConfig: rest.TLSClientConfig{CAFile: cfg.CAFile, Insecure: cfg.InsecureSkipVerify},
		}, nil
	}
	if cfg.InCluster {
		restConfig, err := rest.InClusterConfig()
		if err != nil {
			return nil, ErrNotInCluster
		}
		if cfg.BearerToken != "" {
			restConfig.BearerToken = cfg.BearerToken
		}
		if cfg.CAFile != "" {
			restConfig.TLSClientConfig.CAFile = cfg.CAFile
		}
		restConfig.TLSClientConfig.Insecure = cfg.InsecureSkipVerify
		return restConfig, nil
	}
	return nil, ErrNotInCluster
}

func endpointsToInstances(serviceName string, endpoints *corev1.Endpoints) []transportcontract.ServiceInstance {
	if endpoints == nil {
		return nil
	}
	result := make([]transportcontract.ServiceInstance, 0)
	for _, subset := range endpoints.Subsets {
		for _, address := range subset.Addresses {
			for _, port := range subset.Ports {
				fullAddr := fmt.Sprintf("%s:%d", address.IP, port.Port)
				meta := map[string]string{}
				if port.Name != "" {
					meta["port_name"] = port.Name
				}
				if address.Hostname != "" {
					meta["hostname"] = address.Hostname
				}
				if address.TargetRef != nil && address.TargetRef.Name != "" {
					meta["target_name"] = address.TargetRef.Name
				}
				result = append(result, transportcontract.ServiceInstance{
					ID:       generateInstanceID(serviceName, fullAddr),
					Name:     serviceName,
					Address:  fullAddr,
					Metadata: meta,
					Healthy:  true,
				})
			}
		}
	}
	return result
}

func generateInstanceID(name, addr string) string {
	return strings.TrimSpace(name) + "-" + strings.TrimSpace(addr)
}
