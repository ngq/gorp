package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
	"github.com/spf13/cobra"
)

var (
	protoServiceOutputDir      string
	protoServiceName           string
	protoServiceIncludeHTTP    bool
	protoServiceIncludeGRPC    bool
	protoServiceRegisterRoutes bool

	// gen-service 自己的 proto 输入 flag，不再借用 gen 的共享变量。
	protoServiceFile string
	protoServiceDir  string
)

// protoGenServiceCmd 从 Proto 生成 HTTP handler + gRPC service skeleton。
var protoGenServiceCmd = &cobra.Command{
	Use:   "gen-service",
	Short: "从 Proto 生成 HTTP handler + gRPC service 骨架",
	Long: `从 Proto 文件生成服务端代码骨架。

生成内容：
  --include-http     : HTTP handler skeleton（默认启用）
  --include-grpc     : gRPC service skeleton（默认启用）
  --register-routes  : 路由注册代码（默认启用）

使用:
  gorp proto gen-service -f api/proto/user/v1/user.proto
  gorp proto gen-service -f api/proto/user/v1/user.proto --include-http --register-routes
  gorp proto gen-service -f api/proto/user/v1/user.proto --include-grpc
  gorp proto gen-service -f api/proto/user/v1/user.proto -o internal/server/http`,
	RunE: runProtoGenService,
}

func init() {
	protoCmd.AddCommand(protoGenServiceCmd)

	protoGenServiceCmd.Flags().StringVarP(&protoServiceFile, "proto-file", "f", "", "Proto 文件路径")
	protoGenServiceCmd.Flags().StringVarP(&protoServiceDir, "proto-dir", "d", "api/proto", "Proto 文件目录（未指定 -f 时自动扫描）")
	protoGenServiceCmd.Flags().StringVarP(&protoServiceOutputDir, "output", "o", "internal/server/http", "输出目录")
	protoGenServiceCmd.Flags().StringVarP(&protoServiceName, "service", "s", "", "指定服务名（空=为所有服务生成）")
	protoGenServiceCmd.Flags().BoolVar(&protoServiceIncludeHTTP, "include-http", true, "生成 HTTP handler")
	protoGenServiceCmd.Flags().BoolVar(&protoServiceIncludeGRPC, "include-grpc", true, "生成 gRPC service")
	protoGenServiceCmd.Flags().BoolVar(&protoServiceRegisterRoutes, "register-routes", true, "生成路由注册")
}

func runProtoGenService(cmd *cobra.Command, args []string) error {
	// 确定 proto 文件路径：优先使用 -f 指定，否则从 -d 目录自动扫描。
	protoFile := protoServiceFile
	if protoFile == "" {
		found, err := findFirstProtoFile(protoServiceDir)
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

	gen, err := createProtoGenerator(protoServiceIncludeHTTP)
	if err != nil {
		return err
	}

	opts := integrationcontract.ServiceGenOptions{
		ProtoFile:      protoFile,
		OutputDir:      protoServiceOutputDir,
		PackageName:    filepath.Base(protoServiceOutputDir),
		Module:         module,
		ServiceName:    protoServiceName,
		IncludeHTTP:    protoServiceIncludeHTTP,
		IncludeGRPC:    protoServiceIncludeGRPC,
		RegisterRoutes: protoServiceRegisterRoutes,
	}

	if err := gen.GenService(context.Background(), opts); err != nil {
		return fmt.Errorf("generate service failed: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Service skeleton generated in %s\n", protoServiceOutputDir)
	return nil
}

// findFirstProtoFile 在指定目录下查找第一个 .proto 文件。
func findFirstProtoFile(dir string) (string, error) {
	var found string
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".proto") && found == "" {
			found = path
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return found, nil
}

// detectGoModule 从 go.mod 文件检测当前模块路径。
func detectGoModule() (string, error) {
	data, err := os.ReadFile("go.mod")
	if err != nil {
		return "", fmt.Errorf("go.mod not found: %w", err)
	}

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module ")), nil
		}
	}

	return "", fmt.Errorf("module declaration not found in go.mod")
}
