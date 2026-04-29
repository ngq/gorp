package config

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/ngq/gorp/framework/contract"
	"github.com/spf13/viper"
	"github.com/subosito/gotenv"
)

// Service 是基于 viper 的默认配置服务实现。
//
// 中文说明：
// - env 记录当前加载的环境名，例如 dev / test / prod；
// - v 持有真正的配置读取器，后续 Get/Unmarshal 都委托给它。
// - 该实现同时支持 yaml、多层覆盖、环境变量覆盖与 env(KEY) 占位符替换。
// - 支持配置源扩展（本地文件/远程配置）。
type Service struct {
	env    string
	v      *viper.Viper
	source contract.ConfigSource // 配置源（可选）
}

func NewService() *Service {
	return &Service{v: viper.New()}
}

// NewServiceWithSource 创建带配置源的 Service。
//
// 中文说明：
// - 支持远程配置源（Consul KV / etcd）；
// - 本地文件优先，远程配置覆盖。
func NewServiceWithSource(source contract.ConfigSource) *Service {
	return &Service{
		v:      viper.New(),
		source: source,
	}
}

func (s *Service) Env() string { return s.env }

// Load 按固定顺序加载配置，确保同一套输入始终得到确定性的最终结果。
//
// 加载顺序：
// 1. 本地文件层：合并 config/*.yaml + app.<env>.yaml + config/<env>/*.yaml
// 2. 配置源层：若存在 ConfigSource，则把 source.Load 的结果覆盖到当前配置
// 3. 环境变量层：通过 AutomaticEnv 覆盖最终值
//
// 额外能力：
// - YAML 内容中的 env(KEY) 会在读取文件时先替换成真实环境变量。
// - 如果引用了不存在的环境变量，会直接报错，避免带着半残配置继续运行。
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

// LoadLocalConfigToViper 统一按 framework 约定把本地配置加载到指定 viper 实例。
//
// 中文说明：
// - 这是 Config 与 ConfigSource.local 共用的本地加载主链；
// - env 会统一走 NormalizeEnv，兼容 dev/test/prod 与历史别名；
// - root 为空时自动按 APP_BASE_PATH/工作目录推导项目根。
func LoadLocalConfigToViper(v *viper.Viper, env, root string) error {
	if v == nil {
		return fmt.Errorf("config: viper instance is nil")
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
		envFiles := []string{filepath.Join(configDir, fmt.Sprintf("app.%s.yaml", env))}
		// 中文说明：
		// - framework 统一环境名后，仍兼容历史命名：
		//   dev <-> development
		//   test <-> testing
		//   prod <-> production
		// - 这样旧项目不需要立刻批量改名配置文件，也能平滑过渡。
		switch env {
		case EnvDev:
			envFiles = append(envFiles, filepath.Join(configDir, "app.development.yaml"))
		case EnvTest:
			envFiles = append(envFiles, filepath.Join(configDir, "app.testing.yaml"))
		case EnvProd:
			envFiles = append(envFiles, filepath.Join(configDir, "app.production.yaml"))
		}
		for _, envFile := range envFiles {
			if _, err := os.Stat(envFile); err == nil {
				b, err := readFileWithEnvSubst(envFile)
				if err != nil {
					return err
				}
				if err := v.MergeConfig(bytes.NewReader(b)); err != nil {
					return fmt.Errorf("merge env config (%s): %w", envFile, err)
				}
			}
		}

		// env directory overlay (config/<env>/*.yaml)
		envDirs := []string{filepath.Join(configDir, env)}
		switch env {
		case EnvDev:
			envDirs = append(envDirs, filepath.Join(configDir, "development"))
		case EnvTest:
			envDirs = append(envDirs, filepath.Join(configDir, "testing"))
		case EnvProd:
			envDirs = append(envDirs, filepath.Join(configDir, "production"))
		}
		for _, envDir := range envDirs {
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
	// 中文说明：
	// - 当前 config provider 采用“host root + framework config convention”模型：
	//   1. 优先取环境变量 `APP_BASE_PATH`
	//   2. 否则回退到当前工作目录
	// - 这说明它已经不是早期的裸 `Getwd()` 推断，但也还不是完全无宿主假设的极简中立 provider；
	// - framework 冻仓阶段先把这层语义写清并稳定下来，后续若继续抽仓再判断是否还需要进一步下沉约定。
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

// Watch 监听配置变化。
//
// 中文说明：
// - 如果配置源支持 Watch，返回 ConfigWatcher；
// - 否则返回 nil（本地文件不支持热更新）。
func (s *Service) Watch(ctx context.Context, key string) (contract.ConfigWatcher, error) {
	if s.source != nil {
		return s.source.Watch(ctx, key)
	}
	// 本地文件不支持热更新
	return nil, fmt.Errorf("config: watch not supported for local file source")
}

// Reload 强制重新加载配置。
//
// 中文说明：
// - 重新读取本地 config/*.yaml；
// - 如果配置源存在，先从远程拉取配置。
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
