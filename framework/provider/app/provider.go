package app

import (
	"os"
	"path/filepath"

	"github.com/ngq/gorp/framework/contract"
)

const AppKey = "framework.app"

const (
	configBasePathKey    = "app.paths.base"
	configStoragePathKey = "app.paths.storage"
	configRuntimePathKey = "app.paths.runtime"
	configLogPathKey     = "app.paths.log"
	configConfigPathKey  = "app.paths.config"
	configTempPathKey    = "app.paths.temp"
	envBasePathKey       = "APP_BASE_PATH"
)

// App 定义应用路径相关的能力。
//
// 中文说明：
// - 继承 contract.Root 接口，提供统一的路径管理；
// - 这是"dedicated root contract 正式化"的具体实现。
type App interface {
	contract.Root
}

type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string       { return "app" }
func (p *Provider) IsDefer() bool      { return false }
func (p *Provider) Provides() []string { return []string{AppKey} }

func (p *Provider) Register(c contract.Container) error {
	c.Bind(AppKey, func(c contract.Container) (any, error) {
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

func (p *Provider) Boot(contract.Container) error { return nil }

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

func getConfig(c contract.Container) contract.Config {
	if !c.IsBind(contract.ConfigKey) {
		return nil
	}
	v, err := c.Make(contract.ConfigKey)
	if err != nil {
		return nil
	}
	cfg, _ := v.(contract.Config)
	return cfg
}

func resolveBasePath(wd string, configured string) string {
	// 中文说明：
	// - framework 抽离阶段先补一层更明确的“宿主根目录”约定：
	//   1. 优先显式配置 `app.paths.base`
	//   2. 再看环境变量 `APP_BASE_PATH`
	//   3. 最后才回退到当前工作目录
	// - 这样 config/provider 与 runtime/provider 至少可以共享同一套 host root 覆盖入口。
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
