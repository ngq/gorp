// Package cmd from-service 子命令
package cmd

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/ngq/gorp/framework/contract"
	"github.com/spf13/cobra"
)

var (
	// from-service flags
	servicePath   string
	protoPackage  string
	goPackage     string
	serviceName   string
	includeHTTP   bool
	importPaths   []string
)

// fromServiceCmd from-service 子命令
var fromServiceCmd = &cobra.Command{
	Use:   "from-service",
	Short: "从 Go Service 接口生成 Proto 文件",
	Long: `从 Go Service 接口反向生成 Proto 文件。

解析 Go AST 提取：
- Service 接口方法定义
- 请求/响应类型结构体
- 字段类型映射（支持 Map、切片、嵌套结构体）
- 结构体注释和字段 remark tag

使用:
  gorp-proto from-service -s ./service.go -o ./proto/customer.proto
  gorp-proto from-service -s ./service.go -o ./proto/ --include-http --package customer

示例 Go Service:
  type CustomerServiceRPC interface {
      Register(ctx context.Context, req *RegisterRequest) (*Customer, error)
      GetCustomer(ctx context.Context, req *GetCustomerRequest) (*Customer, error)
  }

  type RegisterRequest struct {
      Username string ` + "`json:\"username\" remark:\"用户名\"`" + `
      Email    string ` + "`json:\"email\" remark:\"邮箱地址\"`" + `
  }`,
	RunE: runFromService,
}

func init() {
	rootCmd.AddCommand(fromServiceCmd)

	fromServiceCmd.Flags().StringVarP(&servicePath, "service-path", "s", "", "Go Service 接口文件路径（必填）")
	fromServiceCmd.Flags().StringVarP(&outputDir, "output", "o", "", "输出 Proto 文件路径（必填）")
	fromServiceCmd.Flags().StringVarP(&protoPackage, "package", "p", "", "Proto package 名称（默认从文件名推断）")
	fromServiceCmd.Flags().StringVar(&goPackage, "go-package", "", "Go package 路径（默认从输出路径推断）")
	fromServiceCmd.Flags().StringVar(&serviceName, "service-name", "", "Service 名称（默认从接口名推断）")
	fromServiceCmd.Flags().BoolVar(&includeHTTP, "include-http", false, "生成 google.api.http 注解")
	fromServiceCmd.Flags().StringSliceVar(&importPaths, "import-paths", nil, "额外的 import 路径（用于跨包类型解析）")

	fromServiceCmd.MarkFlagRequired("service-path")
	fromServiceCmd.MarkFlagRequired("output")
}

func runFromService(cmd *cobra.Command, args []string) error {
	// 确保输出目录存在
	if err := ensureDir(outputDir); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}

	// 推断默认值
	if protoPackage == "" {
		// 从输出文件名推断 package
		base := filepath.Base(outputDir)
		if ext := filepath.Ext(base); ext != "" {
			protoPackage = base[:len(base)-len(ext)]
		} else {
			protoPackage = "generated"
		}
	}

	if goPackage == "" {
		// 从输出路径推断 go_package
		// 用户需要根据实际项目调整
		goPackage = fmt.Sprintf("./%s;%s", filepath.Dir(outputDir), protoPackage)
		fmt.Printf("⚠️  go-package 未指定，使用推断值: %s\n", goPackage)
		fmt.Printf("   建议手动指定完整的 Go module 路径\n")
	}

	// 创建生成器
	gen, err := createGenerator(includeHTTP)
	if err != nil {
		return fmt.Errorf("创建生成器失败: %w", err)
	}

	// 构建选项
	opts := contract.ServiceToProtoOptions{
		ServicePath:  servicePath,
		OutputPath:   outputDir,
		Package:      protoPackage,
		GoPackage:    goPackage,
		ServiceName:  serviceName,
		IncludeHTTP:  includeHTTP,
		ImportPaths:  importPaths,
	}

	// 执行生成
	fmt.Printf("🔄 正在解析: %s\n", servicePath)
	if err := gen.GenFromService(context.Background(), opts); err != nil {
		printError("Service→Proto", err)
		return err
	}

	printSuccess("Service→Proto", outputDir)
	return nil
}