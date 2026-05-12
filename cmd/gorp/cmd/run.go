package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

// defaultAirToml 是 gorp 项目默认的 air 配置模板。
//
// 中文说明：
// - 适配 gorp 生成的项目结构：主入口为 cmd/app/main.go，构建产物输出到 tmp/main；
// - 监听 .go/.yaml/.yml/.toml 文件变更，自动重新编译和重启服务；
// - 排除 vendor/node_modules/.git/tmp 等无需监听的目录；
// - 排除 _test.go 文件，避免测试文件变更触发重启；
// - config/ 目录变更会触发重启（不排除），因为配置变更通常需要服务重新加载。
const defaultAirToml = `root = "."
tmp_dir = "tmp"

[build]
  bin = "./tmp/main"
  cmd = "go build -o ./tmp/main ./cmd/app"
  delay = 1000
  exclude_dir = ["vendor", "node_modules", ".git", "tmp"]
  exclude_regex = ["_test\\.go$"]
  include_ext = ["go", "yaml", "yml", "toml"]
  kill_delay = "0s"
  send_interrupt = false
  stop_on_error = true

[log]
  time = false

[misc]
  clean_on_exit = true
`

// runCmd 启动开发热重载（基于 air）。
//
// 中文说明：
// - 封装 air 工具，为 gorp 项目提供开箱即用的开发热重载能力；
// - 如果当前目录不存在 .air.toml，会自动生成一份适配 gorp 项目结构的默认配置；
// - 如果 air 未安装，会给出明确的安装指引。
var runCmd = &cobra.Command{
	Use:     "run",
	Short:   "启动开发热重载（基于 air）",
	GroupID: commandGroupCodegen,
	Long: `启动开发热重载。

监听代码变更，自动重新编译和重启服务。
需要先安装 air：go install github.com/air-verse/air@latest

首次运行时，如果当前目录不存在 .air.toml，会自动生成一份
适配 gorp 项目结构的默认配置：
  - 主入口：cmd/app/main.go
  - 构建命令：go build -o ./tmp/main ./cmd/app
  - 监听扩展：.go, .yaml, .yml, .toml
  - 排除目录：vendor/, node_modules/, .git/, tmp/`,
	RunE: runDev,
}

// runDev 执行热重载命令。
//
// 中文说明：
// - 1) 检查 air 是否已安装（通过 exec.LookPath）；
// - 2) 如果当前目录不存在 .air.toml，生成默认配置；
// - 3) 将 air 作为子进程执行，继承当前终端的 stdin/stdout/stderr。
func runDev(cmd *cobra.Command, args []string) error {
	// 检查 air 是否已安装
	airPath, err := exec.LookPath("air")
	if err != nil {
		return fmt.Errorf("未找到 air，请先安装：go install github.com/air-verse/air@latest")
	}

	// 如果当前目录不存在 .air.toml，生成默认配置
	airTomlPath := ".air.toml"
	if !fileExists(airTomlPath) {
		fmt.Fprintln(cmd.OutOrStdout(), "未找到 .air.toml，正在生成默认配置...")
		if err := os.WriteFile(airTomlPath, []byte(defaultAirToml), 0o644); err != nil {
			return fmt.Errorf("生成 .air.toml 失败：%w", err)
		}
		fmt.Fprintln(cmd.OutOrStdout(), "已生成 .air.toml（适配 gorp 项目结构）")
	}

	// 执行 air
	fmt.Fprintf(cmd.OutOrStdout(), "启动 air 热重载（%s）...\n", airPath)
	airCmd := exec.Command(airPath)
	airCmd.Stdin = os.Stdin
	airCmd.Stdout = os.Stdout
	airCmd.Stderr = os.Stderr

	if err := airCmd.Run(); err != nil {
		return fmt.Errorf("air 执行失败：%w", err)
	}
	return nil
}

func init() {
	rootCmd.AddCommand(runCmd)
}
