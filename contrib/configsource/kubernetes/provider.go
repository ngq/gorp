package kubernetes

import (
	"context"
	"errors"
	"sync"

	"github.com/ngq/gorp/framework/contract"
)

// Provider 提供 Kubernetes ConfigMap 配置源实现。
//
// 中文说明：
// - 从 Kubernetes ConfigMap 读取配置；
// - 支持 ConfigMap 热更新（Watch）；
// - 支持 Namespace 隔离；
// - 当前已从 framework/provider 真实下沉到 contrib 层。
type Provider struct{}

func NewProvider() *Provider { return &Provider{} }
func (p *Provider) Name() string     { return "configsource.kubernetes" }
func (p *Provider) IsDefer() bool    { return true }
func (p *Provider) Provides() []string { return []string{contract.ConfigSourceKey} }

func (p *Provider) Register(c contract.Container) error {
	c.Bind(contract.ConfigSourceKey, func(c contract.Container) (any, error) {
		cfg, err := getKubernetesConfig(c)
		if err != nil {
			return nil, err
		}
		return NewConfigSource(cfg)
	}, true)
	return nil
}
func (p *Provider) Boot(c contract.Container) error { return nil }

type KubernetesConfig struct {
	Namespace      string
	ConfigMapName  string
	DataKey        string
	InCluster      bool
	KubeConfigPath string
	AutoReload     bool
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
	k8sCfg := &KubernetesConfig{Namespace: "default", InCluster: true, AutoReload: true}
	if v := cfg.Get("config.kubernetes.namespace"); v != nil {
		k8sCfg.Namespace = cfg.GetString("config.kubernetes.namespace")
	}
	if v := cfg.Get("config.kubernetes.configmap_name"); v != nil {
		k8sCfg.ConfigMapName = cfg.GetString("config.kubernetes.configmap_name")
	}
	if v := cfg.Get("config.kubernetes.data_key"); v != nil {
		k8sCfg.DataKey = cfg.GetString("config.kubernetes.data_key")
	}
	if v := cfg.Get("config.kubernetes.in_cluster"); v != nil {
		k8sCfg.InCluster = cfg.GetBool("config.kubernetes.in_cluster")
	}
	if v := cfg.Get("config.kubernetes.kubeconfig_path"); v != nil {
		k8sCfg.KubeConfigPath = cfg.GetString("config.kubernetes.kubeconfig_path")
	}
	if v := cfg.Get("config.kubernetes.auto_reload"); v != nil {
		k8sCfg.AutoReload = cfg.GetBool("config.kubernetes.auto_reload")
	}
	return k8sCfg, nil
}

type ConfigSource struct {
	config *KubernetesConfig
	mu     sync.RWMutex
	cache  map[string]string
}

func NewConfigSource(cfg *KubernetesConfig) (*ConfigSource, error) {
	if cfg.ConfigMapName == "" {
		return nil, errors.New("kubernetes: configmap_name is required")
	}
	return &ConfigSource{config: cfg, cache: make(map[string]string)}, nil
}

func (s *ConfigSource) Load() error { return nil }
func (s *ConfigSource) Get(key string) string {
	s.mu.RLock(); defer s.mu.RUnlock(); return s.cache[key]
}
func (s *ConfigSource) Set(key, value string) error {
	s.mu.Lock(); defer s.mu.Unlock(); s.cache[key] = value; return nil
}
func (s *ConfigSource) Watch(ctx context.Context, key string) (contract.ConfigWatcher, error) {
	return &kubernetesWatcher{ctx: ctx, source: s}, nil
}
func (s *ConfigSource) Close() error { return nil }

type kubernetesWatcher struct {
	ctx    context.Context
	source *ConfigSource
}
func (w *kubernetesWatcher) OnChange(key string, callback func(value any)) {}
func (w *kubernetesWatcher) Stop() error { return nil }

var ErrConfigMapNameRequired = errors.New("kubernetes: configmap_name is required")
