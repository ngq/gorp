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
	// proto all 自己的 flag，不和 gen / gen-service 共享变量。
	protoAllFile string
	protoAllDir  string
)

// protoAllCmd 一键从 Proto 生成全套代码。
var protoAllCmd = &cobra.Command{
	Use:   "all",
	Short: "一键生成 proto-first 全套代码",
	Long: `一键从 Proto 文件生成完整代码套件。

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
  gorp proto all -f api/proto/user/v1/user.proto
  gorp proto all -d api/proto  # 批量处理`,
	RunE: runProtoAll,
}

func init() {
	protoCmd.AddCommand(protoAllCmd)

	protoAllCmd.Flags().StringVarP(&protoAllFile, "proto-file", "f", "", "Proto 文件路径")
	protoAllCmd.Flags().StringVarP(&protoAllDir, "proto-dir", "d", "api/proto", "Proto 文件目录（未指定 -f 时自动扫描）")
}

func runProtoAll(cmd *cobra.Command, args []string) error {
	// 确定 proto 文件路径。
	protoFile := protoAllFile
	if protoFile == "" {
		found, err := findFirstProtoFile(protoAllDir)
		if err != nil {
			return fmt.Errorf("find proto file failed: %w", err)
		}
		protoFile = found
	}
	if protoFile == "" {
		return fmt.Errorf("proto file not found, use -f or -d to specify")
	}

	// 检测 Go module。
	module, err := detectGoModule()
	if err != nil {
		return fmt.Errorf("detect go module failed: %w", err)
	}

	// Step 1: proto gen（生成 pb.go）。
	fmt.Fprintf(cmd.OutOrStdout(), "Step 1/3: Generating pb.go files...\n")
	if err := runProtoAllGen(cmd, protoFile); err != nil {
		return fmt.Errorf("proto gen failed: %w", err)
	}

	// Step 2: proto gen-service（生成 handler/service/routes）。
	fmt.Fprintf(cmd.OutOrStdout(), "Step 2/3: Generating service skeleton...\n")
	if err := runProtoAllGenService(cmd, protoFile, module); err != nil {
		return fmt.Errorf("proto gen-service failed: %w", err)
	}

	// Step 3: proto gen-client（生成类型化 client wrapper）。
	fmt.Fprintf(cmd.OutOrStdout(), "Step 3/3: Generating RPC client wrapper...\n")
	if err := runProtoAllGenClient(cmd, protoFile); err != nil {
		// client wrapper 生成失败不阻塞主流程。
		fmt.Fprintf(cmd.OutOrStdout(), "Warning: proto gen-client skipped: %v\n", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "\nProto-first complete! Generated:\n")
	fmt.Fprintf(cmd.OutOrStdout(), "  - pb.go files (from protoc)\n")
	fmt.Fprintf(cmd.OutOrStdout(), "  - HTTP handler skeleton\n")
	fmt.Fprintf(cmd.OutOrStdout(), "  - gRPC service skeleton\n")
	fmt.Fprintf(cmd.OutOrStdout(), "  - Route registration\n")
	fmt.Fprintf(cmd.OutOrStdout(), "  - RPC client wrapper\n")
	fmt.Fprintf(cmd.OutOrStdout(), "\nNext: implement the TODO methods in handler/service files\n")
	return nil
}

// runProtoAllGen 在 all 流程中执行 protoc 生成 pb.go。
func runProtoAllGen(cmd *cobra.Command, protoFile string) error {
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

	if err := gen.GenFromProto(context.Background(), opts); err != nil {
		return err
	}
	return nil
}

// runProtoAllGenService 在 all 流程中生成 service skeleton。
func runProtoAllGenService(cmd *cobra.Command, protoFile, module string) error {
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

// runProtoAllGenClient 在 all 流程中生成 client wrapper。
func runProtoAllGenClient(cmd *cobra.Command, protoFile string) error {
	// 推导输出路径：internal/client/<proto_name>_client.go。
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
