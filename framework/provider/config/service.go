// Package config provides configuration service implementation for gorp framework.
// Implements Config contract with viper-based layered loading.
// Supports YAML, environment variables, env(KEY) placeholders, and config source extension.
//
// 配置服务包，提供 gorp 框架的配置服务实现。
// 基于 viper 实现分层加载的 Config 契约。
// 支持 YAML、环境变量、env(KEY) 占位符和配置源扩展。
package config

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	"github.com/spf13/viper"
	"github.com/subosito/gotenv"
)

// Service is the default config service implementation based on viper.
// Supports YAML, layered overlay, environment variable override, and env(KEY) placeholder replacement.
// Supports config source extension (local files/remote config).
// Core logic: Hold viper instance, delegate Get/Unmarshal operations.
//
// Service 是基于 viper 的默认配置服务实现。
// 支持 YAML、分层覆盖、环境变量覆盖与 env(KEY) 占位符替换。
// 支持配置源扩展（本地文件/远程配置）。
// 核心逻辑：持有 viper 实例、委托 Get/Unmarshal 操作。
type Service struct {
	env    string
	v      *viper.Viper
	source datacontract.ConfigSource // 配置源（可选）
}

func NewService() *Service {
	return &Service{v: viper.New()}
}

// NewServiceWithSource 创建带配置源的 Service。
//
// 中文说明：
// - 支持远程配置源（Consul KV / etcd）；
// - 本地文件优先，远程配置覆盖。
func NewServiceWithSource(source datacontract.ConfigSource) *Service {
	return &Service{
		v:      viper.New(),
		source: source,
	}
}

// Env returns the current loaded environment name.
//
// Env 返回当前加载的环境名。
func (s *Service) Env() string { return s.env }

// Load loads configuration in a fixed order for deterministic results.
// Order: local files + config source + environment variables.
// Core logic: Normalize env, load base files, merge env overlay, apply env vars.
//
// Load 按固定顺序加载配置，确保确定性结果。
// 顺序：本地文件 + 配置源 + 环境变量。
// 核心逻辑：规范化环境名、加载基础文件、合并环境覆盖、应用环境变量。
func (s *Service) Load(env string) error {
	s.env = NormalizeEnv(env)

	root := projectRoot()
	v := viper.New()
	if err := LoadLocalConfigToViper(v, s.env, root); err != nil {
		return err
	}

	// 若配置源存在，则用配置源结果覆盖本地配置。
	if s.source != nil {
		remoteCfg, err := s.source.Load(context.Background())
		if err != nil {
			return fmt.Errorf("config: load from source failed: %w", err)
		}
		for k, val := range remoteCfg {
			v.Set(k, val)
		}
	}

	s.v = v
	return nil
}

// LoadLocalConfigToViper loads local config files to specified viper instance.
// Shared by Config provider and ConfigSource.local provider.
// Core logic: Discover base files, merge env overlay, apply env substitution.
//
// LoadLocalConfigToViper 将本地配置文件加载到指定 viper 实例。
// 由 Config provider 和 ConfigSource.local provider 共用。
// 核心逻辑：发现基础文件、合并环境覆盖、应用环境变量替换。
func LoadLocalConfigToViper(v *viper.Viper, env, root string) error {
	if v == nil {
		return errors.New("config: viper instance is nil")
	}
	env = NormalizeEnv(env)
	if strings.TrimSpace(root) == "" {
		root = projectRoot()
	}

	configDir := filepath.Join(root, "config")

	// Optional .env
	_ = gotenv.OverLoad(filepath.Join(root, ".env"))

	v.SetConfigType("yaml")

	// base files
	baseFiles, err := discoverBaseFiles(configDir)
	if err != nil {
		return err
	}
	if len(baseFiles) == 0 {
		return fmt.Errorf("no config files found under %s", configDir)
	}

	for i, p := range baseFiles {
		b, err := readFileWithEnvSubst(p)
		if err != nil {
			return err
		}
		if i == 0 {
			if err := v.ReadConfig(bytes.NewReader(b)); err != nil {
				return fmt.Errorf("read base config (%s): %w", p, err)
			}
			continue
		}
		if err := v.MergeConfig(bytes.NewReader(b)); err != nil {
			return fmt.Errorf("merge base config (%s): %w", p, err)
		}
	}

	// env overlay file (app.<env>.yaml)
	if env != "" {
		envFile := filepath.Join(configDir, fmt.Sprintf("app.%s.yaml", env))
		if _, err := os.Stat(envFile); err == nil {
			b, err := readFileWithEnvSubst(envFile)
			if err != nil {
				return err
			}
			if err := v.MergeConfig(bytes.NewReader(b)); err != nil {
				return fmt.Errorf("merge env config (%s): %w", envFile, err)
			}
		}

		// env directory overlay (config/<env>/*.yaml)
		envDir := filepath.Join(configDir, env)
		if entries, err := os.ReadDir(envDir); err == nil {
			files := make([]string, 0, len(entries))
			for _, e := range entries {
				if e.IsDir() {
					continue
				}
				name := e.Name()
				if !strings.HasSuffix(name, ".yaml") {
					continue
				}
				files = append(files, filepath.Join(envDir, name))
			}
			sort.Strings(files)
			for _, p := range files {
				b, err := readFileWithEnvSubst(p)
				if err != nil {
					return err
				}
				if err := v.MergeConfig(bytes.NewReader(b)); err != nil {
					return fmt.Errorf("merge env dir config (%s): %w", p, err)
				}
			}
		}
	}

	// Environment variables override config.
	//
	// Support underscore env keys, e.g. REDIS_ADDR -> redis.addr
	// (used by framework/testing with miniredis).
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	return nil
}

func discoverBaseFiles(configDir string) ([]string, error) {
	entries, err := os.ReadDir(configDir)
	if err != nil {
		return nil, err
	}

	reEnv := regexp.MustCompile(`^app\.[^.]+\.yaml$`)
	files := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".yaml") {
			continue
		}
		if reEnv.MatchString(name) {
			continue
		}
		files = append(files, filepath.Join(configDir, name))
	}

	sort.Strings(files)
	// ensure app.yaml first if present
	app := filepath.Join(configDir, "app.yaml")
	if idx := indexOf(files, app); idx > 0 {
		files[0], files[idx] = files[idx], files[0]
	}
	return files, nil
}

func indexOf(ss []string, v string) int {
	for i := range ss {
		if ss[i] == v {
			return i
		}
	}
	return -1
}

var reEnvPlaceholder = regexp.MustCompile(`env\(([^)]+)\)`) // env(KEY)

func readFileWithEnvSubst(path string) ([]byte, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	out := reEnvPlaceholder.ReplaceAllFunc(b, func(m []byte) []byte {
		key := strings.TrimSpace(string(m[4 : len(m)-1]))
		val, ok := os.LookupEnv(key)
		if !ok {
			// keep a stable, explicit marker so caller gets a meaningful error
			return []byte("__MISSING_ENV__(" + key + ")")
		}
		return []byte(val)
	})
	if bytes.Contains(out, []byte("__MISSING_ENV__(")) {
		return nil, fmt.Errorf("config env() placeholder has no env var set (file=%s)", path)
	}
	return out, nil
}

func projectRoot() string {
	if base := strings.TrimSpace(os.Getenv("APP_BASE_PATH")); base != "" {
		if filepath.IsAbs(base) {
			return filepath.Clean(base)
		}
		wd, _ := os.Getwd()
		return filepath.Clean(filepath.Join(wd, base))
	}
	wd, _ := os.Getwd()
	return wd
}

func (s *Service) Get(key string) any          { return s.v.Get(key) }
func (s *Service) GetString(key string) string { return s.v.GetString(key) }
func (s *Service) GetInt(key string) int       { return s.v.GetInt(key) }
func (s *Service) GetBool(key string) bool     { return s.v.GetBool(key) }
func (s *Service) GetFloat(key string) float64 { return s.v.GetFloat64(key) }

func (s *Service) Unmarshal(key string, out any) error {
	return s.v.UnmarshalKey(key, out)
}

// Watch watches configuration changes if config source supports it.
// For local file source, use a lightweight polling watcher to detect changes and reload config.
// Core logic: Delegate to config source when present; otherwise poll local files and emit key changes.
//
// Watch 监听配置变化（如果配置源支持）。
// 对于本地文件源，使用轻量轮询监听配置文件变化并自动重载。
// 核心逻辑：存在配置源时委托给配置源，否则轮询本地文件并发出 key 变更回调。
func (s *Service) Watch(ctx context.Context, key string) (datacontract.ConfigWatcher, error) {
	if s.source != nil {
		return s.source.Watch(ctx, key)
	}
	return newLocalConfigWatcher(ctx, s, key, 500*time.Millisecond), nil
}

// Reload forces reload of configuration from all sources.
// Core logic: Reload from remote source first, then reload local files.
//
// Reload 强制重新加载配置。
// 核心逻辑：先从远程源重新加载，然后重新加载本地文件。
func (s *Service) Reload(ctx context.Context) error {
	// 从远程配置源拉取（如果存在）
	if s.source != nil {
		remoteCfg, err := s.source.Load(ctx)
		if err != nil {
			return fmt.Errorf("config: load from source failed: %w", err)
		}
		// 合并远程配置到 viper
		for k, v := range remoteCfg {
			s.v.Set(k, v)
		}
	}

	// 重新加载本地文件
	return s.Load(s.env)
}

type localConfigWatcher struct {
	ctx       context.Context
	cancel    context.CancelFunc
	service   *Service
	interval  time.Duration
	mu        sync.RWMutex
	callbacks map[string][]func(value any)
	lastState map[string]time.Time
}

func newLocalConfigWatcher(parent context.Context, service *Service, key string, interval time.Duration) datacontract.ConfigWatcher {
	ctx, cancel := context.WithCancel(parent)
	w := &localConfigWatcher{
		ctx:       ctx,
		cancel:    cancel,
		service:   service,
		interval:  interval,
		callbacks: make(map[string][]func(value any)),
		lastState: collectLocalConfigState(service.env),
	}
	if strings.TrimSpace(key) != "" {
		w.callbacks[key] = nil
	}
	go w.loop()
	return w
}

func (w *localConfigWatcher) OnChange(key string, callback func(value any)) {
	if strings.TrimSpace(key) == "" || callback == nil {
		return
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	w.callbacks[key] = append(w.callbacks[key], callback)
}

func (w *localConfigWatcher) Stop() error {
	w.cancel()
	return nil
}

func (w *localConfigWatcher) loop() {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			return
		case <-ticker.C:
			w.checkAndReload()
		}
	}
}

func (w *localConfigWatcher) checkAndReload() {
	currentState := collectLocalConfigState(w.service.env)
	if reflect.DeepEqual(currentState, w.lastState) {
		return
	}
	w.lastState = currentState

	w.mu.RLock()
	keys := make([]string, 0, len(w.callbacks))
	previous := make(map[string]any, len(w.callbacks))
	for key := range w.callbacks {
		keys = append(keys, key)
		previous[key] = w.service.Get(key)
	}
	w.mu.RUnlock()

	if err := w.service.Reload(w.ctx); err != nil {
		return
	}

	w.mu.RLock()
	defer w.mu.RUnlock()
	for _, key := range keys {
		value := w.service.Get(key)
		if reflect.DeepEqual(previous[key], value) {
			continue
		}
		for _, callback := range w.callbacks[key] {
			callback(value)
		}
	}
}

func collectLocalConfigState(env string) map[string]time.Time {
	root := projectRoot()
	configDir := filepath.Join(root, "config")
	state := make(map[string]time.Time)

	baseFiles, err := discoverBaseFiles(configDir)
	if err == nil {
		for _, path := range baseFiles {
			if info, statErr := os.Stat(path); statErr == nil {
				state[path] = info.ModTime()
			}
		}
	}

	if strings.TrimSpace(env) != "" {
		envFile := filepath.Join(configDir, fmt.Sprintf("app.%s.yaml", env))
		if info, err := os.Stat(envFile); err == nil {
			state[envFile] = info.ModTime()
		}

		envDir := filepath.Join(configDir, env)
		if entries, err := os.ReadDir(envDir); err == nil {
			for _, entry := range entries {
				if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
					continue
				}
				path := filepath.Join(envDir, entry.Name())
				if info, statErr := os.Stat(path); statErr == nil {
					state[path] = info.ModTime()
				}
			}
		}
	}

	return state
}
