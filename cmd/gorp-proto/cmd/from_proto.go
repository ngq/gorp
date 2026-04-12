// Package cmd from-proto 子命令
package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ngq/gorp/framework/contract"
	"github.com/spf13/cobra"
)

var (
	// from-proto flags
	protoDir     string
	protoFiles   []string
	importPathsP []string
	goOpt        string
	goGrpcOpt    string
	gatewayOpt   string
	includeHTTPP bool
	plugins      []string
)

// fromProtoCmd from-proto 子命令
var fromProtoCmd = &cobra.Command{
	Use:   "from-proto",
	Short: "从 Proto 文件生成 Go pb 代码",
	Long: `从 Proto 文件生成 Go pb 代码（调用 protoc）。

支持生成：
- Go message 代码 (--go_out)
- gRPC client/server 代码 (--go-grpc_out)
- HTTP 转码代码 (--grpc-gateway_out, 可选)

使用:
  gorp-proto from-proto -d ./proto -o ./pb
  gorp-proto from-proto -f ./proto/customer.proto -o ./pb --include-http
  gorp-proto from-proto -d ./proto --go-opt "paths=source_relative,module=github.com/myproject"

前置条件:
  1. 安装 protoc: https://grpc.io/docs/protoc-installation/
  2. 安装 Go 插件:
     go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
     go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
  3. HTTP 转码需要额外安装:
     go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest`,
	RunE: runFromProto,
}

func init() {
	rootCmd.AddCommand(fromProtoCmd)

	fromProtoCmd.Flags().StringVarP(&protoDir, "proto-dir", "d", "", "Proto 文件目录")
	fromProtoCmd.Flags().StringSliceVarP(&protoFiles, "proto-files", "f", nil, "指定的 Proto 文件（多个用逗号分隔）")
	fromProtoCmd.Flags().StringVarP(&outputDir, "output", "o", "", "输出目录（必填）")
	fromProtoCmd.Flags().StringSliceVar(&importPathsP, "import-paths", nil, "protoc 导入路径 (-I)")
	fromProtoCmd.Flags().StringVar(&goOpt, "go-opt", "paths=source_relative", "--go_opt 参数")
	fromProtoCmd.Flags().StringVar(&goGrpcOpt, "go-grpc-opt", "paths=source_relative", "--go-grpc_opt 参数")
	fromProtoCmd.Flags().StringVar(&gatewayOpt, "gateway-opt", "paths=source_relative", "--grpc-gateway_opt 参数")
	fromProtoCmd.Flags().BoolVar(&includeHTTPP, "include-http", false, "生成 grpc-gateway HTTP 转码")
	fromProtoCmd.Flags().StringSliceVar(&plugins, "plugins", nil, "额外的 protoc 插件")

	fromProtoCmd.MarkFlagRequired("output")
}

func runFromProto(cmd *cobra.Command, args []string) error {
	// 检查 protoc 是否安装
	if _, err := exec.LookPath("protoc"); err != nil {
		return fmt.Errorf("protoc 未安装，请先安装: https://grpc.io/docs/protoc-installation/")
	}

	// 检查 Go 插件
	if _, err := exec.LookPath("protoc-gen-go"); err != nil {
		return fmt.Errorf("protoc-gen-go 未安装，请运行: go install google.golang.org/protobuf/cmd/protoc-gen-go@latest")
	}
	if _, err := exec.LookPath("protoc-gen-go-grpc"); err != nil {
		return fmt.Errorf("protoc-gen-go-grpc 未安装，请运行: go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest")
	}

	// 检查 grpc-gateway 插件（如果启用 HTTP）
	if includeHTTPP {
		if _, err := exec.LookPath("protoc-gen-grpc-gateway"); err != nil {
			return fmt.Errorf("protoc-gen-grpc-gateway 未安装，请运行: go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest")
		}
	}

	// 验证输入
	if protoDir == "" && len(protoFiles) == 0 {
		return fmt.Errorf("必须指定 --proto-dir 或 --proto-files")
	}

	// 确保输出目录存在
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}

	// 创建生成器
	gen, err := createGenerator(includeHTTPP)
	if err != nil {
		return fmt.Errorf("创建生成器失败: %w", err)
	}

	// 构建选项
	opts := contract.ProtoGenOptions{
		ProtoDir:     protoDir,
		ProtoFiles:   protoFiles,
		OutputDir:    outputDir,
		ImportPaths:  importPathsP,
		GoOpt:        goOpt,
		GoGrpcOpt:    goGrpcOpt,
		GatewayOpt:   gatewayOpt,
		IncludeHTTP:  includeHTTPP,
		Plugins:      plugins,
	}

	// 执行生成
	srcDesc := protoDir
	if len(protoFiles) > 0 {
		srcDesc = fmt.Sprintf("%v", protoFiles)
	}
	fmt.Printf("🔄 正在生成: %s → %s\n", srcDesc, outputDir)

	if err := gen.GenFromProto(context.Background(), opts); err != nil {
		printError("Proto→pb.go", err)
		return err
	}

	// 列出生成的文件
	files, _ := filepath.Glob(filepath.Join(outputDir, "*.go"))
	fmt.Printf("✅ Proto→pb.go 成功\n")
	fmt.Printf("   输出目录: %s\n", outputDir)
	if len(files) > 0 {
		fmt.Printf("   生成文件:\n")
		for _, f := range files {
			fmt.Printf("     - %s\n", filepath.Base(f))
		}
	}

	return nil
}