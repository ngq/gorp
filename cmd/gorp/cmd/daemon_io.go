package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ngq/gorp/framework/provider/app"
)

// openDaemonLogFiles 为 daemon 子进程打开 stdout/stderr 的日志文件。
//
// 为什么需要它：
// - daemon 模式下子进程脱离终端，stdout/stderr 如果不重定向，很容易变成“黑洞输出”，排障困难。
// - 这里统一把输出落到 app log path（由 app.App.LogPath() 决定），并按 name 区分不同 daemon。
//
// 注意：
// - 该函数只负责打开文件句柄并返回 cleanup；真正的重定向发生在 exec.Command 的 Stdout/Stderr 赋值。
func openDaemonLogFiles(c any, name string) (*os.File, *os.File, func(), error) {
	container, ok := c.(interface{ Make(string) (any, error) })
	if !ok {
		return nil, nil, nil, fmt.Errorf("invalid container")
	}
	v, err := container.Make(app.AppKey)
	if err != nil {
		return nil, nil, nil, err
	}
	appSvc := v.(app.App)

	logDir := appSvc.LogPath()
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return nil, nil, nil, err
	}

	stdoutPath := filepath.Join(logDir, fmt.Sprintf("%s.out.log", name))
	stderrPath := filepath.Join(logDir, fmt.Sprintf("%s.err.log", name))

	outF, err := os.OpenFile(stdoutPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, nil, nil, err
	}
	errF, err := os.OpenFile(stderrPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		_ = outF.Close()
		return nil, nil, nil, err
	}

	cleanup := func() {
		_ = outF.Close()
		_ = errF.Close()
	}
	return outF, errF, cleanup, nil
}
