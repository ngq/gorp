// Application scenarios:
// - Provide the runtime root-path capability used by bootstrap and filesystem-aware services.
// - Resolve application base, storage, runtime, log, config, and temp directories from config and environment.
// - Centralize default path conventions behind one provider bound to contract.RootKey.
//
// 适用场景：
// - 提供 bootstrap 和文件系统相关服务所需的运行时根路径能力。
// - 根据配置和环境变量解析应用的 base、storage、runtime、log、config 和 temp 目录。
// - 通过绑定到 contract.RootKey 的单一 provider 统一默认路径约定。
package app

import (
	"os"
	"path/filepath"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

// AppKey is the binding key used for the root-path capability.
//
// AppKey 是根路径能力使用的绑定 key。
const AppKey = runtimecontract.RootKey

const (
	configBasePathKey    = "app.paths.base"
	configStoragePathKey = "app.paths.storage"
	configRuntimePathKey = "app.paths.runtime"
	configLogPathKey     = "app.paths.log"
	configConfigPathKey  = "app.paths.config"
	configTempPathKey    = "app.paths.temp"
	envBasePathKey       = "APP_BASE_PATH"
)

// App defines application path-related capabilities.
//
// App 定义应用路径相关能力。
//
// 中文说明：
// - 继承 contract.Root 接口，提供统一的路径管理。
// - 这是 “dedicated root contract 正式化” 的具体实现。
// - 当前它更接近 runtime convention provider：提供 framework 默认业务起步路径所需的宿主目录约定。
// - framework 现阶段先把它统一收进 `contract.RootKey` 语义，再继续观察是否需要进一步从 core provider 中下沉约定。
type App interface {
	runtimecontract.Root
}

// Provider provides the root-path capability.
//
// Provider 提供根路径能力。
type Provider struct{}

// NewProvider creates an app provider.
//
// NewProvider 创建 app provider。
func NewProvider() *Provider { return &Provider{} }

// Name returns the provider name.
//
// Name 返回 provider 名称。
func (p *Provider) Name() string { return "app" }

// IsDefer reports that the app provider is not lazily loaded.
//
// IsDefer 表示 app provider 不采用延迟加载。
func (p *Provider) IsDefer() bool { return false }

// Provides returns the capability keys exposed by the app provider.
//
// Provides 返回 app provider 暴露的能力 key。
func (p *Provider) Provides() []string { return []string{runtimecontract.RootKey} }

// Register binds the application path service into the container.
//
// Register 将应用路径服务绑定到容器。
func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(AppKey, func(c runtimecontract.Container) (any, error) {
		wd, err := os.Getwd()
		if err != nil {
			return nil, err
		}

		cfg := getConfig(c)
		basePath := ""
		storagePath := ""
		runtimePath := ""
		logPath := ""
		configPath := ""
		tempPath := ""
		if cfg != nil {
			basePath = cfg.GetString(configBasePathKey)
			storagePath = cfg.GetString(configStoragePathKey)
			runtimePath = cfg.GetString(configRuntimePathKey)
			logPath = cfg.GetString(configLogPathKey)
			configPath = cfg.GetString(configConfigPathKey)
			tempPath = cfg.GetString(configTempPathKey)
		}

		base := resolveBasePath(wd, basePath)
		storage := resolvePath(storagePath, filepath.Join(base, "storage"), base)
		runtime := resolvePath(runtimePath, filepath.Join(storage, "runtime"), storage)
		log := resolvePath(logPath, filepath.Join(storage, "log"), storage)
		cfgPath := resolvePath(configPath, filepath.Join(base, "config"), base)
		tmpPath := resolvePath(tempPath, filepath.Join(runtime, "tmp"), runtime)
		return &service{
			base:    base,
			storage: storage,
			runtime: runtime,
			log:     log,
			config:  cfgPath,
			temp:    tmpPath,
		}, nil
	}, true)
	return nil
}

// Boot does not require extra startup logic for the app provider.
//
// Boot 表示 app provider 不需要额外启动逻辑。
func (p *Provider) Boot(runtimecontract.Container) error { return nil }

type service struct {
	base    string
	storage string
	runtime string
	log     string
	config  string
	temp    string
}

func (s *service) BasePath() string    { return s.base }
func (s *service) StoragePath() string { return s.storage }
func (s *service) RuntimePath() string { return s.runtime }
func (s *service) LogPath() string     { return s.log }
func (s *service) ConfigPath() string  { return s.config }
func (s *service) TempPath() string    { return s.temp }

func getConfig(c runtimecontract.Container) datacontract.Config {
	if !c.IsBind(datacontract.ConfigKey) {
		return nil
	}
	v, err := c.Make(datacontract.ConfigKey)
	if err != nil {
		return nil
	}
	cfg, _ := v.(datacontract.Config)
	return cfg
}

func resolveBasePath(wd string, configured string) string {
	// Prefer explicit config, then environment, then working directory as the final fallback.
	// 优先使用显式配置，其次环境变量，最后回退到当前工作目录。
	if configured != "" {
		return resolvePath(configured, wd, wd)
	}
	if envBase := os.Getenv(envBasePathKey); envBase != "" {
		return resolvePath(envBase, wd, wd)
	}
	return filepath.Clean(wd)
}

func resolvePath(value string, fallback string, relativeBase string) string {
	if value == "" {
		return fallback
	}
	if filepath.IsAbs(value) {
		return filepath.Clean(value)
	}
	return filepath.Clean(filepath.Join(relativeBase, value))
}
