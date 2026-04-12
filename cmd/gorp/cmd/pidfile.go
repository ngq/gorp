package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ngq/gorp/framework/provider/app"
)

const pidFileName = "gorp.pid"

// appPIDPath 返回 HTTP app 默认 pidfile 路径；如果传入 override，则优先使用 override。
func appPIDPath(appSvc app.App, override string) string {
	if override != "" {
		return override
	}
	return filepath.Join(appSvc.RuntimePath(), pidFileName)
}

// defaultPIDPath 返回 HTTP app 默认 pidfile 路径。
func defaultPIDPath(appSvc app.App) string {
	return appPIDPath(appSvc, "")
}

// ensureRuntimeDirs 确保 runtime 目录存在。
func ensureRuntimeDirs(appSvc app.App) error {
	return os.MkdirAll(appSvc.RuntimePath(), 0o755)
}

func writePID(pidPath string, pid int) error {
	if err := os.MkdirAll(filepath.Dir(pidPath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(pidPath, []byte(strconv.Itoa(pid)+"\n"), 0o644)
}

func readPID(pidPath string) (int, error) {
	b, err := os.ReadFile(pidPath)
	if err != nil {
		return 0, err
	}
	s := strings.TrimSpace(string(b))
	pid, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid pid in %s: %w", pidPath, err)
	}
	return pid, nil
}

func getAppService(c any) (app.App, error) {
	container, ok := c.(interface{ Make(string) (any, error) })
	if !ok {
		return nil, fmt.Errorf("invalid container")
	}
	v, err := container.Make(app.AppKey)
	if err != nil {
		return nil, err
	}
	svc, ok := v.(app.App)
	if !ok {
		return nil, fmt.Errorf("invalid app service")
	}
	return svc, nil
}
