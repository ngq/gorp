package zookeeper

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/go-zookeeper/zk"
	internalnative "github.com/ngq/gorp/contrib/internal/native"
	"github.com/ngq/gorp/framework/contract"
)

var (
	ErrNoServers         = errors.New("zookeeper: no servers configured")
	ErrServiceNotFound   = errors.New("zookeeper: service not found")
	ErrRegistryClosed    = errors.New("zookeeper: registry closed")
	ErrAlreadyRegistered = errors.New("zookeeper: instance already registered")
)

// Provider 提供 Zookeeper 服务发现实现。
//
// 中文说明：
//   - 使用 Zookeeper 实现服务注册与发现；
//   - 支持临时节点（Ephemeral ZNode）实现健康检查；
//   - 支持服务元数据；
//   - 适用于已有 Zookeeper 集群的环境。
//   - 当前状态：部分可用
//   - 说明：已完成 P2 第一版最小注册/发现闭环，具备真实 Zookeeper 后端抽象与 fake backend 行为测试；
//     但当前仍未覆盖 watcher、重连与更完整 session 产品化语义。
type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string  { return "registry.zookeeper" }
func (p *Provider) IsDefer() bool { return true }
func (p *Provider) Provides() []string {
	return []string{contract.RPCRegistryKey}
}

func (p *Provider) Register(c contract.Container) error {
	c.Bind(contract.RPCRegistryKey, func(c contract.Container) (any, error) {
		cfg, err := getZookeeperConfig(c)
		if err != nil {
			return nil, err
		}
		return NewRegistry(cfg)
	}, true)

	return nil
}

func (p *Provider) Boot(c contract.Container) error { return nil }

type ZookeeperConfig struct {
	Servers            []string
	SessionTimeout     time.Duration
	WatchRetryInterval time.Duration
	BasePath           string
	ServiceName        string
	ServiceAddr        string
	ServicePort        int
	ServiceMeta        map[string]string
}

func getZookeeperConfig(c contract.Container) (*ZookeeperConfig, error) {
	cfgAny, err := c.Make(contract.ConfigKey)
	if err != nil {
		return nil, err
	}

	cfg, ok := cfgAny.(contract.Config)
	if !ok {
		return nil, errors.New("zookeeper: invalid config service")
	}

	zkCfg := &ZookeeperConfig{
		BasePath:           "/services",
		SessionTimeout:     30 * time.Second,
		WatchRetryInterval: 200 * time.Millisecond,
	}

	if v := cfg.Get("discovery.zookeeper.servers"); v != nil {
		if servers, ok := v.([]string); ok {
			zkCfg.Servers = servers
		}
	}
	if v := cfg.Get("discovery.zookeeper.base_path"); v != nil {
		zkCfg.BasePath = cfg.GetString("discovery.zookeeper.base_path")
	}
	if v := cfg.Get("discovery.zookeeper.session_timeout"); v != nil {
		if seconds := cfg.GetInt("discovery.zookeeper.session_timeout"); seconds > 0 {
			zkCfg.SessionTimeout = time.Duration(seconds) * time.Second
		}
	}
	if v := cfg.Get("discovery.zookeeper.watch_retry_interval_ms"); v != nil {
		if ms := cfg.GetInt("discovery.zookeeper.watch_retry_interval_ms"); ms > 0 {
			zkCfg.WatchRetryInterval = time.Duration(ms) * time.Millisecond
		}
	}

	return zkCfg, nil
}

type zkBackend interface {
	EnsurePath(path string) error
	CreateEphemeral(path string, data []byte) error
	Delete(path string) error
	Children(path string) ([]string, error)
	Get(path string) ([]byte, error)
	WatchChildren(ctx context.Context, path string, onUpdate func()) error
	Close() error
}

type nativeBackendProvider interface {
	Underlying() any
}

type Registry struct {
	config  *ZookeeperConfig
	backend zkBackend

	mu                  sync.RWMutex
	endpointCache       map[string][]contract.ServiceInstance
	watchSnapshots      map[string]string
	registeredInstances map[string]string
	closeMu             sync.Mutex
	closed              bool
	watchCancels        []context.CancelFunc
}

func NewRegistry(cfg *ZookeeperConfig) (*Registry, error) {
	backend, err := newZKBackend(cfg)
	if err != nil {
		return nil, err
	}
	return NewRegistryWithBackend(cfg, backend)
}

func NewRegistryWithBackend(cfg *ZookeeperConfig, backend zkBackend) (*Registry, error) {
	if len(cfg.Servers) == 0 {
		return nil, ErrNoServers
	}
	if backend == nil {
		return nil, errors.New("zookeeper: backend is required")
	}
	return &Registry{
		config:              cfg,
		backend:             backend,
		endpointCache:       make(map[string][]contract.ServiceInstance),
		watchSnapshots:      make(map[string]string),
		registeredInstances: make(map[string]string),
	}, nil
}

func (r *Registry) Register(ctx context.Context, name, addr string, meta map[string]string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return ErrRegistryClosed
	}

	key := instanceKey(name, addr)
	if _, exists := r.registeredInstances[key]; exists {
		return ErrAlreadyRegistered
	}

	servicePath := path.Join(r.config.BasePath, name)
	if err := r.backend.EnsurePath(servicePath); err != nil {
		return fmt.Errorf("zookeeper: ensure service path failed: %w", err)
	}

	record := serviceRecord{
		ID:       generateInstanceID(name, addr),
		Name:     name,
		Address:  addr,
		Metadata: mergeMeta(r.config.ServiceMeta, meta),
		Healthy:  true,
	}
	payload, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("zookeeper: encode instance data failed: %w", err)
	}

	instancePath := path.Join(servicePath, sanitizeNodeName(addr))
	if err := r.backend.CreateEphemeral(instancePath, payload); err != nil {
		return fmt.Errorf("zookeeper: create ephemeral node failed: %w", err)
	}

	r.registeredInstances[key] = instancePath
	delete(r.endpointCache, name)
	delete(r.watchSnapshots, name)
	return nil
}

func (r *Registry) Deregister(ctx context.Context, name, addr string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return ErrRegistryClosed
	}

	key := instanceKey(name, addr)
	instancePath, ok := r.registeredInstances[key]
	if !ok {
		instancePath = path.Join(r.config.BasePath, name, sanitizeNodeName(addr))
	}

	if err := r.backend.Delete(instancePath); err != nil {
		if errors.Is(err, zk.ErrNoNode) {
			return ErrServiceNotFound
		}
		return fmt.Errorf("zookeeper: delete instance failed: %w", err)
	}

	delete(r.registeredInstances, key)
	delete(r.endpointCache, name)
	delete(r.watchSnapshots, name)
	return nil
}

func (r *Registry) Discover(ctx context.Context, name string) ([]contract.ServiceInstance, error) {
	r.mu.RLock()
	if instances, ok := r.endpointCache[name]; ok && len(instances) > 0 {
		cached := append([]contract.ServiceInstance(nil), instances...)
		r.mu.RUnlock()
		return cached, nil
	}
	closed := r.closed
	r.mu.RUnlock()
	if closed {
		return nil, ErrRegistryClosed
	}

	servicePath := path.Join(r.config.BasePath, name)
	children, err := r.backend.Children(servicePath)
	if err != nil {
		if errors.Is(err, zk.ErrNoNode) {
			return nil, ErrServiceNotFound
		}
		return nil, fmt.Errorf("zookeeper: list instances failed: %w", err)
	}
	if len(children) == 0 {
		return nil, ErrServiceNotFound
	}
	sort.Strings(children)

	result := make([]contract.ServiceInstance, 0, len(children))
	for _, child := range children {
		data, getErr := r.backend.Get(path.Join(servicePath, child))
		if getErr != nil {
			return nil, fmt.Errorf("zookeeper: read instance failed: %w", getErr)
		}

		var record serviceRecord
		if err := json.Unmarshal(data, &record); err != nil {
			return nil, fmt.Errorf("zookeeper: decode instance failed: %w", err)
		}
		result = append(result, contract.ServiceInstance{
			ID:       record.ID,
			Name:     record.Name,
			Address:  record.Address,
			Metadata: record.Metadata,
			Healthy:  record.Healthy,
		})
	}
	sortServiceInstances(result)
	r.mu.Lock()
	r.endpointCache[name] = append([]contract.ServiceInstance(nil), result...)
	r.mu.Unlock()
	return result, nil
}

func (r *Registry) Watch(ctx context.Context, name string) (<-chan []contract.ServiceInstance, error) {
	r.closeMu.Lock()
	defer r.closeMu.Unlock()

	if r.closed {
		return nil, ErrRegistryClosed
	}

	watchCtx, cancel := context.WithCancel(ctx)
	r.watchCancels = append(r.watchCancels, cancel)

	ch := make(chan []contract.ServiceInstance, 10)
	servicePath := path.Join(r.config.BasePath, name)
	emit := func(instances []contract.ServiceInstance) bool {
		snapshot := snapshotKey(instances)

		r.mu.Lock()
		last := r.watchSnapshots[name]
		if last == snapshot {
			r.mu.Unlock()
			return true
		}
		r.watchSnapshots[name] = snapshot
		r.mu.Unlock()

		select {
		case ch <- append([]contract.ServiceInstance(nil), instances...):
			return true
		case <-watchCtx.Done():
			return false
		default:
			return true
		}
	}

	go func() {
		defer close(ch)
		for {
			err := r.backend.WatchChildren(watchCtx, servicePath, func() {
				r.mu.Lock()
				delete(r.endpointCache, name)
				r.mu.Unlock()

				instances, err := r.Discover(watchCtx, name)
				if err != nil {
					if errors.Is(err, ErrServiceNotFound) {
						r.mu.Lock()
						delete(r.endpointCache, name)
						r.mu.Unlock()
						emit(nil)
					}
					return
				}
				emit(instances)
			})
			if err == nil || watchCtx.Err() != nil {
				return
			}
			if !isRetryableWatchError(err) {
				return
			}
			select {
			case <-watchCtx.Done():
				return
			case <-time.After(r.config.WatchRetryInterval):
			}
		}
	}()

	go func() {
		instances, err := r.Discover(watchCtx, name)
		if err != nil {
			if errors.Is(err, ErrServiceNotFound) {
				emit(nil)
			}
			return
		}
		emit(instances)
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
	return r.backend.Close()
}

func (r *Registry) Underlying() any {
	if provider, ok := r.backend.(nativeBackendProvider); ok {
		if native := provider.Underlying(); native != nil {
			return native
		}
	}
	return r.backend
}

func (r *Registry) As(target any) bool {
	return internalnative.As(r.Underlying(), target)
}

type serviceRecord struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Address  string            `json:"address"`
	Metadata map[string]string `json:"metadata"`
	Healthy  bool              `json:"healthy"`
}

type realZKBackend struct {
	conn *zk.Conn
}

func newZKBackend(cfg *ZookeeperConfig) (zkBackend, error) {
	if len(cfg.Servers) == 0 {
		return nil, ErrNoServers
	}
	conn, _, err := zk.Connect(cfg.Servers, cfg.SessionTimeout)
	if err != nil {
		return nil, fmt.Errorf("zookeeper: connect failed: %w", err)
	}
	return &realZKBackend{conn: conn}, nil
}

func (b *realZKBackend) EnsurePath(target string) error {
	parts := strings.Split(strings.Trim(target, "/"), "/")
	current := ""
	for _, part := range parts {
		current += "/" + part
		exists, _, err := b.conn.Exists(current)
		if err != nil {
			return err
		}
		if exists {
			continue
		}
		_, err = b.conn.Create(current, nil, 0, zk.WorldACL(zk.PermAll))
		if err != nil && !errors.Is(err, zk.ErrNodeExists) {
			return err
		}
	}
	return nil
}

func (b *realZKBackend) CreateEphemeral(target string, data []byte) error {
	_, err := b.conn.Create(target, data, zk.FlagEphemeral, zk.WorldACL(zk.PermAll))
	return err
}

func (b *realZKBackend) Delete(target string) error {
	return b.conn.Delete(target, -1)
}

func (b *realZKBackend) Children(target string) ([]string, error) {
	children, _, err := b.conn.Children(target)
	return children, err
}

func (b *realZKBackend) Get(target string) ([]byte, error) {
	data, _, err := b.conn.Get(target)
	return data, err
}

func (b *realZKBackend) WatchChildren(ctx context.Context, target string, onUpdate func()) error {
	for {
		_, _, events, err := b.conn.ChildrenW(target)
		if err != nil {
			if errors.Is(err, zk.ErrNoNode) {
				onUpdate()
				select {
				case <-ctx.Done():
					return nil
				case <-time.After(200 * time.Millisecond):
					continue
				}
			}
			return err
		}

		select {
		case <-ctx.Done():
			return nil
		case event, ok := <-events:
			if !ok {
				return zk.ErrConnectionClosed
			}
			if event.Err != nil {
				return event.Err
			}
			if event.Type == zk.EventSession {
				switch event.State {
				case zk.StateExpired:
					return zk.ErrSessionExpired
				case zk.StateAuthFailed:
					return errors.New("zookeeper: watch session auth failed")
				case zk.StateDisconnected, zk.StateConnecting:
					return zk.ErrConnectionClosed
				}
			}
			onUpdate()
		}
	}
}

func (b *realZKBackend) Close() error {
	b.conn.Close()
	return nil
}

func (b *realZKBackend) Underlying() any {
	return b.conn
}

func sanitizeNodeName(addr string) string {
	return strings.ReplaceAll(addr, "/", "_")
}

func generateInstanceID(name, addr string) string {
	return name + "-" + addr
}

func instanceKey(name, addr string) string {
	return name + "|" + addr
}

func mergeMeta(base map[string]string, extra map[string]string) map[string]string {
	result := make(map[string]string, len(base)+len(extra))
	for k, v := range base {
		result[k] = v
	}
	for k, v := range extra {
		result[k] = v
	}
	return result
}

func sortServiceInstances(instances []contract.ServiceInstance) {
	sort.Slice(instances, func(i, j int) bool {
		if instances[i].ID != instances[j].ID {
			return instances[i].ID < instances[j].ID
		}
		return instances[i].Address < instances[j].Address
	})
}

func snapshotKey(instances []contract.ServiceInstance) string {
	if len(instances) == 0 {
		return "<empty>"
	}
	parts := make([]string, 0, len(instances))
	for _, instance := range instances {
		parts = append(parts, instance.ID+"|"+instance.Address)
	}
	sort.Strings(parts)
	return strings.Join(parts, ";")
}

func isRetryableWatchError(err error) bool {
	return errors.Is(err, zk.ErrConnectionClosed) || errors.Is(err, zk.ErrSessionExpired)
}
