package cmd

import (
	"context"
	"fmt"

	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
	"github.com/spf13/cobra"
)

var (
	// openapi 子命令 flags
	openapiProtoFile   string
	openapiOutputFile  string
	openapiTitle       string
	openapiDescription string
	openapiVersion     string
	openapiHost        string
	openapiBasePath    string
	openapiServiceName string
)

// protoOpenapiCmd openapi 子命令：从 Proto 文件生成 OpenAPI 文档
var protoOpenapiCmd = &cobra.Command{
	Use:   "openapi",
	Short: "从 Proto 文件生成 OpenAPI 文档",
	Long: `从 Proto 文件生成 OpenAPI/Swagger 文档。

解析 proto 文件中的 google.api.http 注解，生成 OpenAPI 3.0 规范。

使用:
  gorp proto openapi -f api/proto/user/v1/user.proto -o docs/openapi.yaml
  gorp proto openapi -f api/proto/user/v1/user.proto -o docs/openapi.json

前置条件:
  Proto 文件需要使用 google.api.http 注解定义 HTTP 路由。`,
	RunE: runProtoOpenapi,
}

func init() {
	protoCmd.AddCommand(protoOpenapiCmd)

	protoOpenapiCmd.Flags().StringVarP(&openapiProtoFile, "proto-file", "f", "", "Proto 文件路径（必需）")
	protoOpenapiCmd.Flags().StringVarP(&openapiOutputFile, "output", "o", "docs/openapi.yaml", "输出文件路径（.yaml 或 .json）")
	protoOpenapiCmd.Flags().StringVar(&openapiTitle, "title", "", "API 标题")
	protoOpenapiCmd.Flags().StringVar(&openapiDescription, "description", "", "API 描述")
	protoOpenapiCmd.Flags().StringVar(&openapiVersion, "version", "1.0.0", "API 版本")
	protoOpenapiCmd.Flags().StringVar(&openapiHost, "host", "", "API 主机")
	protoOpenapiCmd.Flags().StringVar(&openapiBasePath, "base-path", "", "API 基础路径")
	protoOpenapiCmd.Flags().StringVar(&openapiServiceName, "service", "", "服务名称（可选，用于过滤）")

	protoOpenapiCmd.MarkFlagRequired("proto-file")
}

func runProtoOpenapi(cmd *cobra.Command, args []string) error {
	// 创建生成器
	gen, err := createProtoGenerator(true)
	if err != nil {
		return fmt.Errorf("创建生成器失败: %w", err)
	}

	// 构建选项
	opts := integrationcontract.OpenAPIGenOptions{
		ProtoFile:   openapiProtoFile,
		OutputFile:  openapiOutputFile,
		Title:       openapiTitle,
		Description: openapiDescription,
		Version:     openapiVersion,
		Host:        openapiHost,
		BasePath:    openapiBasePath,
		ServiceName: openapiServiceName,
	}

	// 执行生成
	fmt.Printf("🔄 正在生成 OpenAPI: %s → %s\n", openapiProtoFile, openapiOutputFile)

	if err := gen.GenOpenAPI(context.Background(), opts); err != nil {
		printProtoError("Proto→OpenAPI", err)
		return err
	}

	printProtoSuccess("Proto→OpenAPI", openapiOutputFile)
	return nil
}
