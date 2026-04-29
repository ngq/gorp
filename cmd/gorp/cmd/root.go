package cmd

import "github.com/spf13/cobra"

const (
	commandGroupStarter  = "starter"
	commandGroupCodegen  = "codegen"
	commandGroupAdvanced = "advanced"
	rootHelpTemplate     = `{{with (or .Long .Short)}}{{. | trimTrailingWhitespaces}}

{{end}}Usage:
{{if .Runnable}}  {{.UseLine}}
{{end}}{{if .HasAvailableSubCommands}}  {{.CommandPath}} [command]
{{end}}
{{if .HasAvailableSubCommands}}{{range $group := .Groups}}{{$group.Title}}
{{range $.Commands}}{{if and (eq .GroupID $group.ID) (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}

{{end}}{{if .HasAvailableLocalFlags}}Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}

{{end}}Additional commands (kept outside the default reading path):
{{range .Commands}}{{if and (eq .GroupID "") (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{else if .HasAvailableLocalFlags}}Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}
{{end}}`
)

// rootCmd 是 gorp CLI 的根命令。
//
// 中文说明：
// - `gorp` 的公开心智模型：framework + starter templates + developer toolchain；
// - 一级命令分为三层：
//   1. 默认起步主命令：`new`；
//   2. 高频开发命令：`proto / model`；
//   3. 按需进入的辅助工具链：`template / provider / middleware / command / doc / swagger / openapi / version`；
//
// - legacy runtime 命令（app / grpc / cron / build / dev / deploy）已退役；
// - 用户通过项目自己的 `cmd/*/main.go` 启动服务，不依赖 CLI runtime。
var rootCmd = &cobra.Command{
	Use:   "gorp",
	Short: "Framework, starter templates, and developer tooling for gorp",
	Long: `Framework, starter templates, and developer tooling for gorp.

Default starter path:
  - gorp new            : default single-service quick start
  - gorp new multi-wire : default multi-service quick start

High-frequency developer tools:
  - gorp proto
  - gorp model

Supplementary delivery path:
  - gorp new from-release : published release asset delivery path for the same public starter set

On-demand tools when needed:
  - Enter only after the default starter path and high-frequency tools no longer answer your current task.
  - gorp template / doc / provider / middleware / command / swagger / openapi / version

Service runtime entry:
  - Generated services should start from the project's own cmd/*/main.go entrypoints.
  - The CLI is toolchain-first and is no longer the default starter runtime path.`,
}

// Execute 执行整个 Cobra 命令树。
//
// 中文说明：
// - main 函数通常只需要调用这一层；
// - 具体命令匹配、flag 解析、help 输出、RunE 执行，都由 Cobra 在这里统一调度。
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.SetHelpTemplate(rootHelpTemplate)
	rootCmd.AddGroup(
		&cobra.Group{ID: commandGroupStarter, Title: "Default starter path"},
		&cobra.Group{ID: commandGroupCodegen, Title: "High-frequency developer tools"},
		&cobra.Group{ID: commandGroupAdvanced, Title: "On-demand tools after the default path"},
	)
}
