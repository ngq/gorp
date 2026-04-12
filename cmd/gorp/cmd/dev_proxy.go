package cmd

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/ngq/gorp/framework/contract"
	"github.com/ngq/gorp/framework/provider/app"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
)

type devConfig struct {
	Port string
	Backend struct {
		RefreshTimeSec int
		Port           string
		MonitorFolder  string
	}
	Frontend struct {
		Port string
	}
}

// initDevConfig 从配置中心装载开发态参数，并补默认值。
//
// 中文说明：
// - dev.port 是统一代理端口。
// - dev.backend.port / dev.frontend.port 分别是后端与前端真实监听端口。
// - monitor_folder 为空时默认监听 app 目录源码变更。
func initDevConfig(c contract.Container) (*devConfig, error) {
	cfg := &devConfig{
		Port: "8070",
	}
	cfg.Backend.RefreshTimeSec = 1
	cfg.Backend.Port = "8072"
	cfg.Backend.MonitorFolder = ""
	cfg.Frontend.Port = "8071"

	configAny, err := c.Make(contract.ConfigKey)
	if err != nil {
		return nil, err
	}
	config := configAny.(contract.Config)

	// 兼容教程写法：dev.port / dev.backend.* / dev.frontend.*
	if v := config.GetString("dev.port"); v != "" {
		cfg.Port = v
	}
	if v := config.GetInt("dev.backend.refresh_time"); v > 0 {
		cfg.Backend.RefreshTimeSec = v
	}
	if v := config.GetString("dev.backend.port"); v != "" {
		cfg.Backend.Port = v
	}
	if v := config.GetString("dev.backend.monitor_folder"); v != "" {
		cfg.Backend.MonitorFolder = v
	}
	if v := config.GetString("dev.frontend.port"); v != "" {
		cfg.Frontend.Port = v
	}

	if cfg.Backend.MonitorFolder == "" {
		appAny, err := c.Make(app.AppKey)
		if err != nil {
			return nil, err
		}
		appSvc := appAny.(app.App)
		cfg.Backend.MonitorFolder = filepath.Join(appSvc.BasePath(), "app")
	}

	return cfg, nil
}

type devProxy struct {
	cfg *devConfig

	mu sync.RWMutex
	// 当前 backend/http server 是否已经启动成功（简化：只做端口转发）
	backendURL  *url.URL
	frontendURL *url.URL
}

// newDevProxy 创建本地开发代理。
//
// 中文说明：
// - backendURL 指向 Go 后端开发服务。
// - frontendURL 指向 Vite/前端 dev server。
// - 代理策略由 handler() 决定：优先后端，404 再落到前端。
func newDevProxy(cfg *devConfig) (*devProxy, error) {
	be, err := url.Parse("http://127.0.0.1:" + cfg.Backend.Port)
	if err != nil {
		return nil, err
	}
	fe, err := url.Parse("http://127.0.0.1:" + cfg.Frontend.Port)
	if err != nil {
		return nil, err
	}
	return &devProxy{cfg: cfg, backendURL: be, frontendURL: fe}, nil
}

func (p *devProxy) handler() http.Handler {
	backend := httputil.NewSingleHostReverseProxy(p.backendURL)
	frontend := httputil.NewSingleHostReverseProxy(p.frontendURL)

	notFoundErr := errors.New("backend returned 404")

	backend.ModifyResponse = func(resp *http.Response) error {
		if resp.StatusCode == http.StatusNotFound {
			return notFoundErr
		}
		return nil
	}
	backend.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		if errors.Is(err, notFoundErr) {
			frontend.ServeHTTP(w, r)
			return
		}
		http.Error(w, err.Error(), http.StatusBadGateway)
	}

	// 中文说明：
	// - 开发态下优先把请求交给后端，覆盖 API、SSR、静态路由等场景。
	// - 如果后端明确返回 404，则认为该路径更可能属于前端路由，再转发给前端 dev server。
	// - 这样用户只需要访问一个统一端口。
	return backend
}

// startGoBackend starts backend via `go run ./cmd/gorp app start` with APP_ENV=development and custom address.
//
// 中文说明：
// - 我们不直接调用现有 `app start` 的内部函数，而是启动一个子进程，便于 kill+restart。
// - 通过环境变量 APP_ADDRESS 覆盖监听地址（需要在 gin provider 支持，后续若没有则改为配置文件）。
func startGoBackend(root string, port string) (*exec.Cmd, error) {
	cmd := exec.Command("go", "run", "./cmd/gorp", "app", "start")
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "APP_ENV=development", "APP_ADDRESS=127.0.0.1:"+port)
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return cmd, nil
}

func startFrontend(root string) (*exec.Cmd, error) {
	frontendDir := filepath.Join(root, "frontend")
	if !fileExists(filepath.Join(frontendDir, "package.json")) {
		return nil, fmt.Errorf("frontend package.json not found: %s", filepath.Join(frontendDir, "package.json"))
	}
	cmd := exec.Command("npm", "run", "dev", "--", "--port", "8071")
	cmd.Dir = frontendDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return cmd, nil
}

func killCmd(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}
	_ = cmd.Process.Kill()
	_, _ = cmd.Process.Wait()
}

func watchAndRestart(monitorDir string, refresh time.Duration, restart func() error) error {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer w.Close()

	// 中文说明：
	// - 这里采用“递归加目录 + debounce 重启”的简化策略。
	// - 文件频繁变更时不会每次都立刻重启，而是合并成 refresh 窗口后的单次重启。
	// - 只关心 .go 文件变动，避免前端构建产物或临时文件引起无意义重启。
	// add dirs recursively
	_ = filepath.WalkDir(monitorDir, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			// skip hidden and vendor
			base := filepath.Base(p)
			if strings.HasPrefix(base, ".") || base == "vendor" || base == "node_modules" {
				return filepath.SkipDir
			}
			_ = w.Add(p)
		}
		return nil
	})

	var (
		mu       sync.Mutex
		timer    *time.Timer
		restartE error
	)

	schedule := func() {
		mu.Lock()
		defer mu.Unlock()
		if timer != nil {
			timer.Stop()
		}
		timer = time.AfterFunc(refresh, func() {
			mu.Lock()
			mu.Unlock()
			restartE = restart()
		})
	}

	for {
		select {
		case ev := <-w.Events:
			if ev.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) != 0 {
				if strings.HasSuffix(ev.Name, ".go") {
					schedule()
				}
			}
		case err := <-w.Errors:
			return err
		default:
			if restartE != nil {
				return restartE
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}

var devBackendCmd = &cobra.Command{
	Use:   "backend",
	Short: "Dev backend: watch and restart on changes",
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := repoRootFromCWD()
		if err != nil {
			return err
		}
		_, c, err := bootstrap(WithAppEnv("development"))
		if err != nil {
			return err
		}
		cfg, err := initDevConfig(c)
		if err != nil {
			return err
		}

		// 中文说明：
		// - backend 开发态的核心是“先启动一个子进程，再监听源码变化重启它”。
		// - 这样可以最大程度复用正式启动命令，而不是额外维护一套仅 dev 使用的启动路径。
		var cur *exec.Cmd
		restart := func() error {
			fmt.Fprintln(cmd.OutOrStdout(), "restarting backend...")
			killCmd(cur)
			newCmd, err := startGoBackend(root, cfg.Backend.Port)
			if err != nil {
				return err
			}
			cur = newCmd
			return nil
		}

		if err := restart(); err != nil {
			return err
		}

		return watchAndRestart(cfg.Backend.MonitorFolder, time.Duration(cfg.Backend.RefreshTimeSec)*time.Second, restart)
	},
}

var devFrontendCmd = &cobra.Command{
	Use:   "frontend",
	Short: "Dev frontend: npm run dev",
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := repoRootFromCWD()
		if err != nil {
			return err
		}
		// 中文说明：
		// - frontend 开发态不做代理封装，直接复用前端项目自己的 dev server。
		// - 这样热更新、HMR、前端工具链能力都保持原生行为。
		c, err := startFrontend(root)
		if err != nil {
			return err
		}
		defer killCmd(c)
		return c.Wait()
	},
}

var devAllCmd = &cobra.Command{
	Use:   "all",
	Short: "Dev all: backend + frontend + proxy",
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := repoRootFromCWD()
		if err != nil {
			return err
		}
		_, c, err := bootstrap(WithAppEnv("development"))
		if err != nil {
			return err
		}
		cfg, err := initDevConfig(c)
		if err != nil {
			return err
		}

		// 中文说明：
		// - dev all 会同时拉起 frontend dev server、backend 进程和统一 proxy。
		// - 目前没有对任一子进程退出做复杂联动治理，保持最小实现。
		// - 用户统一访问 cfg.Port，即可透过 proxy 访问前后端。
		// start frontend
		fe, err := startFrontend(root)
		if err != nil {
			return err
		}
		defer killCmd(fe)

		// start backend
		be, err := startGoBackend(root, cfg.Backend.Port)
		if err != nil {
			return err
		}
		defer killCmd(be)

		proxy, err := newDevProxy(cfg)
		if err != nil {
			return err
		}

		srv := &http.Server{Addr: "127.0.0.1:" + cfg.Port, Handler: proxy.handler()}
		fmt.Fprintf(cmd.OutOrStdout(), "dev proxy listening: %s\n", srv.Addr)
		return srv.ListenAndServe()
	},
}

func init() {
	devCmd.AddCommand(devFrontendCmd)
	devCmd.AddCommand(devBackendCmd)
	devCmd.AddCommand(devAllCmd)

	_ = runtime.GOOS // keep runtime import used when extending
}
