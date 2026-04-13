package cmd

import (
	"github.com/spf13/cobra"
)

// dddCmd 是 DDD 专属代码生成命令组的入口。
//
// 中文说明：
// - DDD 项目结构强调领域边界，不应硬套通用三层脚手架；
// - DDD 强调 domain / application / infrastructure / interfaces 四层边界；
// - 本命令组提供 DDD 风格的代码生成能力，当前主入口是：
//   1. context：在**已有 DDD starter 项目内部**生成完整 bounded context
// - 当前不建议把它当成“在任意空 Go 模块中独立起一个 DDD 项目”的命令使用；
// - 推荐先在已有符合 DDD 四层目录约定的宿主项目中运行，再用 `gorp ddd context` 追加上下文；
// - 当前公开 starter 模板中并没有 `--template ddd`，因此这里不再把 DDD 生成器表述成独立 starter 模板入口。
var dddCmd = &cobra.Command{
	Use:   "ddd",
	Short: "DDD-specific code generation tools",
	Long: `DDD-specific code generation tools for domain-driven design projects.

DDD generators respect the four-layer boundary:
  - domain: entities, value objects, repository ports
  - application: use cases, application services
  - infrastructure: repository implementations, external adapters
  - interfaces: HTTP handlers, request/response DTOs

Recommended workflow:
  1. Use 'gorp ddd context' to scaffold a complete bounded context
  2. Customize the generated domain entities and business rules
  3. Implement application use cases
  4. Add infrastructure implementations
  5. Expose through HTTP handlers

This approach keeps your DDD layers clean and maintainable.`,
}

func init() {
	rootCmd.AddCommand(dddCmd)
}