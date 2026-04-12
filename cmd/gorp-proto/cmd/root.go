// Package cmd gorp-proto CLI 命令实现
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ngq/gorp/framework/contract"
	"github.com/ngq/gorp/framework/provider/proto"
	"github.com/spf13/cobra"
)

var (
	// 全局 flag
	outputDir string
)

// rootCmd 根命令
var rootCmd = &cobra.Command{
	Use:   "gorp-proto",
	Short: "Proto 生成器 - 支持三种工作流",
	Long: `gorp-proto 是一个 Proto 生成器 CLI 工具，支持三种工作流：

1. from-service: 从 Go Service 接口生成 Proto 文件
   gorp-proto from-service -s ./service.go -o ./proto/

2. from-proto: 从 Proto 文件生成 Go pb 代码（调用 protoc）
   gorp-proto from-proto -d ./proto -o ./pb/

3. from-route: 从 Gin 路由生成 Proto 文件（实验性）
   gorp-proto from-route -r ./routes.go -o ./proto/

安装:
  go install github.com/ngq/gorp/cmd/gorp-proto@latest`,
}

// Execute 执行 CLI 命令
func Execute() error {
	return rootCmd.Execute()
}

// createGenerator 创建 Proto 生成器
func createGenerator(includeHTTP bool) (contract.ProtoGenerator, error) {
	return proto.NewGenerator(&contract.ProtoGeneratorConfig{
		Enabled:               true,
		Strategy:              "noop",
		IncludeHTTPAnnotation: includeHTTP,
	})
}

// ensureDir 确保目录存在
func ensureDir(path string) error {
	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		return os.MkdirAll(dir, 0755)
	}
	return nil
}

// printSuccess 打印成功消息
func printSuccess(action, output string) {
	fmt.Printf("✅ %s 成功\n", action)
	fmt.Printf("   输出: %s\n", output)
}

// printError 打印错误消息
func printError(action string, err error) {
	fmt.Fprintf(os.Stderr, "❌ %s 失败: %v\n", action, err)
}