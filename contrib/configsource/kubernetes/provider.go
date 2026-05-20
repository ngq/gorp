package kubernetes

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	defaultNamespace            = "default"
	defaultKubernetesConfigPoll = 5 * time.Second
)

var (
	ErrConfigMapNameRequired = errors.New("configsource.kubernetes: configmap_name is required")
	ErrSetNotSupported       = errors.New("configsource.kubernetes: set is not supported")
	ErrConfigMapNotFound     = errors.New("configsource.kubernetes: configmap not found")
	ErrNotInCluster          = errors.New("configsource.kubernetes: not in cluster, please provide api_server or kubeconfig_path")
	ErrConfigSourceClosed    = errors.New("configsource.kubernetes: config source closed")
)

// Provider 提供 Kubernetes ConfigMap 配置源实现。
type Provider struct {
	BaseConfigSourceProvider
}

func NewProvider() *Provider {
	p := &Provider{}
	p.NameStr = "configsource.kubernetes"
	p.GetConfig = func(c runtimecontract.Container) (any, error) {
		return getKubernetesConfig(c)
	}
	p.NewSource = func(cfg any) (datacontract.ConfigSource, error) {
		return NewConfigSource(cfg.(*KubernetesConfig))
	}
	return p
}

type KubernetesConfig struct {
	Namespace          string
	ConfigMapName      string
	DataKey            string
	InCluster          bool
	KubeConfigPath     string
	APIServer          string
	BearerToken        string
	CAFile             string
	InsecureSkipVerify bool
	AutoReload         bool
	PollInterval       time.Duration
}

func getKubernetesConfig(c runtimecontract.Container) (*KubernetesConfig, error) {
	cfg, err := ReadConfig(c)
	if err != nil {
		return nil, err
	}

	k8sCfg := &KubernetesConfig{
		Namespace:    defaultNamespace,
		InCluster:    true,
		AutoReload:   true,
		PollInterval: defaultKubernetesConfigPoll,
	}

	k8sCfg.Namespace = GetStringFallback(cfg, "kubernetes", "namespace")
	if k8sCfg.Namespace == "" {
		k8sCfg.Namespace = defaultNamespace
	}
	k8sCfg.ConfigMapName = GetStringFallback(cfg, "kubernetes", "configmap_name")
	k8sCfg.DataKey = GetStringFallback(cfg, "kubernetes", "data_key")
	if inCluster, ok := GetBoolFallback(cfg, "kubernetes", "in_cluster"); ok {
		k8sCfg.InCluster = inCluster
	}
	k8sCfg.KubeConfigPath = GetStringFallback(cfg, "kubernetes", "kubeconfig_path")
	k8sCfg.APIServer = GetStringFallback(cfg, "kubernetes", "api_server")
	k8sCfg.BearerToken = GetStringFallback(cfg, "kubernetes", "bearer_token")
	k8sCfg.CAFile = GetStringFallback(cfg, "kubernetes", "ca_file")
	if skipVerify, ok := GetBoolFallback(cfg, "kubernetes", "insecure_skip_verify"); ok {
		k8sCfg.InsecureSkipVerify = skipVerify
	}
	if autoReload, ok := GetBoolFallback(cfg, "kubernetes", "auto_reload"); ok {
		k8sCfg.AutoReload = autoReload
	}
	if d := GetDurationSecondsFallback(cfg, "kubernetes", "poll_interval_seconds"); d > 0 {
		k8sCfg.PollInterval = d
	}

	return k8sCfg, nil
}

type configMapClient interface {
	LoadConfigMap(ctx context.Context, namespace, name string) (map[string]string, error)
	WatchConfigMap(ctx context.Context, namespace, name string, onUpdate func(map[string]string)) error
}

type nativeClientProvider interface {
	Underlying() any
}

type ConfigSource struct {
	config   *KubernetesConfig
	client   configMapClient
	mu       sync.RWMutex
	cache    map[string]any
	watchers sync.Map
	closed   bool
	closeMu  sync.Mutex
}

func NewConfigSource(cfg *KubernetesConfig) (*ConfigSource, error) {
	client, err := newConfigMapClient(cfg)
	if err != nil {
		return nil, err
	}
	return NewConfigSourceWithClient(cfg, client)
}

func NewConfigSourceWithClient(cfg *KubernetesConfig, client configMapClient) (*ConfigSource, error) {
	if cfg.ConfigMapName == "" {
		return nil, ErrConfigMapNameRequired
	}
	if client == nil {
		return nil, errors.New("configsource.kubernetes: configmap client is required")
	}
	return &ConfigSource{
		config: cfg,
		client: client,
		cache:  make(map[string]any),
	}, nil
}

func (s *ConfigSource) Load(ctx context.Context) (map[string]any, error) {
	data, err := s.client.LoadConfigMap(ctx, s.config.Namespace, s.config.ConfigMapName)
	if err != nil {
		return nil, err
	}

	loaded, err := s.decodeConfigMapData(data)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	s.cache = cloneMap(loaded)
	s.mu.Unlock()
	return cloneMap(loaded), nil
}

func (s *ConfigSource) Get(ctx context.Context, key string) (any, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, ok := lookupNestedValue(s.cache, key)
	if !ok {
		return nil, fmt.Errorf("configsource.kubernetes: key %s not found", key)
	}
	return value, nil
}

func (s *ConfigSource) Set(ctx context.Context, key string, value any) error {
	return ErrSetNotSupported
}

func (s *ConfigSource) Underlying() any {
	if provider, ok := s.client.(nativeClientProvider); ok {
		if native := provider.Underlying(); native != nil {
			return native
		}
	}
	return s.client
}

func (s *ConfigSource) As(target any) bool {
	return As(s.Underlying(), target)
}

func (s *ConfigSource) Watch(ctx context.Context, key string) (datacontract.ConfigWatcher, error) {
	s.closeMu.Lock()
	defer s.closeMu.Unlock()

	if s.closed {
		return nil, ErrConfigSourceClosed
	}
	if err := s.ensureLoaded(ctx); err != nil {
		return nil, err
	}

	watchCtx, cancel := context.WithCancel(ctx)
	watcher := &kubernetesWatcher{
		cancel:    cancel,
		source:    s,
		callbacks: &sync.Map{},
	}
	s.watchers.Store(watcher, struct{}{})

	if s.config.AutoReload {
		go func() {
			_ = s.client.WatchConfigMap(watchCtx, s.config.Namespace, s.config.ConfigMapName, func(data map[string]string) {
				loaded, err := s.decodeConfigMapData(data)
				if err != nil {
					return
				}
				s.mu.Lock()
				s.cache = cloneMap(loaded)
				s.mu.Unlock()
				watcher.dispatch()
			})
		}()
	}

	return watcher, nil
}

func (s *ConfigSource) Close() error {
	s.closeMu.Lock()
	defer s.closeMu.Unlock()

	if s.closed {
		return nil
	}
	s.closed = true
	s.watchers.Range(func(key, value any) bool {
		if watcher, ok := key.(*kubernetesWatcher); ok {
			_ = watcher.Stop()
		}
		return true
	})
	return nil
}

func (s *ConfigSource) ensureLoaded(ctx context.Context) error {
	s.mu.RLock()
	loaded := len(s.cache) > 0
	s.mu.RUnlock()
	if loaded {
		return nil
	}
	_, err := s.Load(ctx)
	return err
}

func (s *ConfigSource) decodeConfigMapData(data map[string]string) (map[string]any, error) {
	if s.config.DataKey != "" {
		raw, ok := data[s.config.DataKey]
		if !ok {
			return nil, fmt.Errorf("configsource.kubernetes: data key %s not found", s.config.DataKey)
		}
		decoded, err := decodeStructuredString(raw)
		if err != nil {
			return nil, err
		}
		if asMap, ok := decoded.(map[string]any); ok {
			return asMap, nil
		}
		return map[string]any{s.config.DataKey: decoded}, nil
	}

	result := make(map[string]any, len(data))
	for key, value := range data {
		result[key] = value
	}
	return result, nil
}

type kubernetesWatcher struct {
	cancel    context.CancelFunc
	source    *ConfigSource
	callbacks *sync.Map
}

func (w *kubernetesWatcher) OnChange(key string, callback func(value any)) {
	w.callbacks.Store(key, callback)
	w.source.mu.RLock()
	current, exists := lookupNestedValue(w.source.cache, key)
	w.source.mu.RUnlock()
	if exists {
		callback(current)
	}
}

func (w *kubernetesWatcher) Stop() error {
	w.cancel()
	w.source.watchers.Delete(w)
	return nil
}

func (w *kubernetesWatcher) dispatch() {
	w.callbacks.Range(func(key, value any) bool {
		path, ok := key.(string)
		if !ok {
			return true
		}
		callback, ok := value.(func(value any))
		if !ok {
			return true
		}
		w.source.mu.RLock()
		current, exists := lookupNestedValue(w.source.cache, path)
		w.source.mu.RUnlock()
		if exists {
			callback(current)
		}
		return true
	})
}

type clientGoConfigMapClient struct {
	client kubernetes.Interface
}

func newConfigMapClient(cfg *KubernetesConfig) (configMapClient, error) {
	restConfig, err := buildConfigSourceRESTConfig(cfg)
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("configsource.kubernetes: create clientset failed: %w", err)
	}
	return &clientGoConfigMapClient{client: clientset}, nil
}

func (c *clientGoConfigMapClient) Underlying() any {
	return c.client
}

func (c *clientGoConfigMapClient) LoadConfigMap(ctx context.Context, namespace, name string) (map[string]string, error) {
	configMap, err := c.client.CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, ErrConfigMapNotFound
		}
		return nil, fmt.Errorf("configsource.kubernetes: load configmap failed: %w", err)
	}
	return cloneStringMap(configMap.Data), nil
}

func (c *clientGoConfigMapClient) WatchConfigMap(ctx context.Context, namespace, name string, onUpdate func(map[string]string)) error {
	for {
		watcher, err := c.client.CoreV1().ConfigMaps(namespace).Watch(ctx, metav1.ListOptions{
			FieldSelector: fields.OneTermEqualSelector("metadata.name", name).String(),
		})
		if err != nil {
			return fmt.Errorf("configsource.kubernetes: watch configmap failed: %w", err)
		}

		restart, err := consumeConfigMapWatch(ctx, watcher, onUpdate)
		if err != nil {
			return err
		}
		if !restart {
			return nil
		}
	}
}

func consumeConfigMapWatch(ctx context.Context, watcher watch.Interface, onUpdate func(map[string]string)) (bool, error) {
	defer watcher.Stop()
	for {
		select {
		case <-ctx.Done():
			return false, nil
		case event, ok := <-watcher.ResultChan():
			if !ok {
				return true, nil
			}
			configMap, ok := event.Object.(*corev1.ConfigMap)
			if !ok || configMap == nil || event.Type == watch.Deleted {
				continue
			}
			onUpdate(cloneStringMap(configMap.Data))
		}
	}
}

func buildConfigSourceRESTConfig(cfg *KubernetesConfig) (*rest.Config, error) {
	if cfg.KubeConfigPath != "" {
		return clientcmd.BuildConfigFromFlags(cfg.APIServer, cfg.KubeConfigPath)
	}
	if cfg.APIServer != "" {
		return &rest.Config{
			Host:            cfg.APIServer,
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

func decodeStructuredString(raw string) (any, error) {
	var value any
	if err := yaml.Unmarshal([]byte(raw), &value); err != nil {
		return nil, fmt.Errorf("configsource.kubernetes: decode config content failed: %w", err)
	}
	return normalizeYAMLValue(value), nil
}

func normalizeYAMLValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		result := make(map[string]any, len(typed))
		for key, nested := range typed {
			result[key] = normalizeYAMLValue(nested)
		}
		return result
	case map[any]any:
		result := make(map[string]any, len(typed))
		for key, nested := range typed {
			result[fmt.Sprint(key)] = normalizeYAMLValue(nested)
		}
		return result
	case []any:
		result := make([]any, 0, len(typed))
		for _, nested := range typed {
			result = append(result, normalizeYAMLValue(nested))
		}
		return result
	default:
		return typed
	}
}

func lookupNestedValue(data map[string]any, path string) (any, bool) {
	if path == "" {
		return cloneMap(data), true
	}
	current := any(data)
	for _, segment := range strings.Split(path, ".") {
		asMap, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		next, exists := asMap[segment]
		if !exists {
			return nil, false
		}
		current = next
	}
	return current, true
}

func cloneMap(source map[string]any) map[string]any {
	result := make(map[string]any, len(source))
	for key, value := range source {
		if nested, ok := value.(map[string]any); ok {
			result[key] = cloneMap(nested)
			continue
		}
		result[key] = value
	}
	return result
}

func cloneStringMap(source map[string]string) map[string]string {
	result := make(map[string]string, len(source))
	for key, value := range source {
		result[key] = value
	}
	return result
}
