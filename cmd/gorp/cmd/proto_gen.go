package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
	"github.com/spf13/cobra"
)

var (
	// protoGenFile 指定单个 proto 文件。
	protoGenFile string
)

// protoGenCmd 从单个 Proto 文件生成全套代码。
var protoGenCmd = &cobra.Command{
	Use:   "gen",
	Short: "从单个 Proto 文件生成全套代码",
	Long: `从单个 Proto 文件生成完整代码套件。

生成内容：
  1. Go pb.go 代码（调用 protoc）
  2. gRPC service skeleton
  3. HTTP handler skeleton
  4. 路由注册代码
  5. 类型化 RPC client wrapper

前置条件：
  1. 安装 protoc: https://grpc.io/docs/protoc-installation/
  2. 安装 Go 插件:
     go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
     go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

使用:
  gorp proto gen -f api/proto/user/v1/user.proto

批量处理多个 proto 文件请使用: gorp proto all -d api/proto`,
	RunE: runProtoGen,
}

func init() {
	protoCmd.AddCommand(protoGenCmd)

	protoGenCmd.Flags().StringVarP(&protoGenFile, "proto-file", "f", "", "Proto 文件路径（必填）")
	protoGenCmd.MarkFlagRequired("proto-file")
}

func runProtoGen(cmd *cobra.Command, args []string) error {
	if protoGenFile == "" {
		return fmt.Errorf("proto file is required, use -f to specify")
	}

	// 检测 Go module。
	module, err := detectGoModule()
	if err != nil {
		return fmt.Errorf("detect go module failed: %w", err)
	}

	// Step 1: 生成 pb.go。
	fmt.Fprintf(cmd.OutOrStdout(), "Step 1/3: Generating pb.go files...\n")
	if err := genPB(cmd, protoGenFile); err != nil {
		return fmt.Errorf("proto gen failed: %w", err)
	}

	// Step 2: 生成 handler/service/routes。
	fmt.Fprintf(cmd.OutOrStdout(), "Step 2/3: Generating service skeleton...\n")
	if err := genServiceSkeleton(cmd, protoGenFile, module); err != nil {
		return fmt.Errorf("proto gen-service failed: %w", err)
	}

	// Step 3: 生成 client wrapper。
	fmt.Fprintf(cmd.OutOrStdout(), "Step 3/3: Generating RPC client wrapper...\n")
	if err := genClientWrapper(cmd, protoGenFile); err != nil {
		fmt.Fprintf(cmd.OutOrStdout(), "Warning: proto gen-client skipped: %v\n", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "\nProto-first complete! Generated:\n")
	fmt.Fprintf(cmd.OutOrStdout(), "  - pb.go files (from protoc)\n")
	fmt.Fprintf(cmd.OutOrStdout(), "  - HTTP handler skeleton\n")
	fmt.Fprintf(cmd.OutOrStdout(), "  - gRPC service skeleton\n")
	fmt.Fprintf(cmd.OutOrStdout(), "  - Route registration\n")
	fmt.Fprintf(cmd.OutOrStdout(), "  - RPC client wrapper\n")
	fmt.Fprintf(cmd.OutOrStdout(), "\nNext: implement business logic in service files\n")
	return nil
}

// genPB 执行 protoc 生成 pb.go。
func genPB(cmd *cobra.Command, protoFile string) error {
	if _, err := exec.LookPath("protoc"); err != nil {
		return fmt.Errorf("protoc 未安装，请先安装: https://grpc.io/docs/protoc-installation/")
	}
	if _, err := exec.LookPath("protoc-gen-go"); err != nil {
		return fmt.Errorf("protoc-gen-go 未安装，请运行: go install google.golang.org/protobuf/cmd/protoc-gen-go@latest")
	}
	if _, err := exec.LookPath("protoc-gen-go-grpc"); err != nil {
		return fmt.Errorf("protoc-gen-go-grpc 未安装，请运行: go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest")
	}

	out := filepath.Dir(protoFile)
	if err := os.MkdirAll(out, 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}

	gen, err := createProtoGenerator(false)
	if err != nil {
		return fmt.Errorf("创建生成器失败: %w", err)
	}

	opts := integrationcontract.ProtoGenOptions{
		ProtoDir:   filepath.Dir(protoFile),
		ProtoFiles: []string{protoFile},
		OutputDir:  out,
		GoOpt:      "paths=source_relative",
		GoGrpcOpt:  "paths=source_relative",
	}

	return gen.GenFromProto(context.Background(), opts)
}

// genServiceSkeleton 生成 service skeleton。
func genServiceSkeleton(cmd *cobra.Command, protoFile, module string) error {
	gen, err := createProtoGenerator(true)
	if err != nil {
		return err
	}

	opts := integrationcontract.ServiceGenOptions{
		ProtoFile:      protoFile,
		OutputDir:      "internal/server/http",
		PackageName:    "http",
		Module:         module,
		IncludeHTTP:    true,
		IncludeGRPC:    true,
		RegisterRoutes: true,
	}

	return gen.GenService(context.Background(), opts)
}

// genClientWrapper 生成 client wrapper。
func genClientWrapper(cmd *cobra.Command, protoFile string) error {
	base := strings.TrimSuffix(filepath.Base(protoFile), ".proto")
	clientDir := filepath.Join("internal", "client")
	if err := os.MkdirAll(clientDir, 0755); err != nil {
		return fmt.Errorf("创建 client 目录失败: %w", err)
	}
	outputFile := filepath.Join(clientDir, base+"_client.go")

	gen, err := createProtoGenerator(false)
	if err != nil {
		return err
	}

	opts := integrationcontract.ClientGenOptions{
		ProtoFile:   protoFile,
		OutputFile:  outputFile,
		PackageName: "client",
	}

	return gen.GenClient(context.Background(), opts)
}