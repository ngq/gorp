package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	// protoAllDir 指定 proto 文件目录。
	protoAllDir string
)

// protoAllCmd 批量从目录下所有 Proto 文件生成全套代码。
var protoAllCmd = &cobra.Command{
	Use:   "all",
	Short: "批量生成目录下所有 Proto 文件的代码",
	Long: `批量从指定目录下所有 Proto 文件生成完整代码套件。

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
  gorp proto all -d api/proto
  gorp proto all -d api/proto/user/v1

单个 proto 文件请使用: gorp proto gen -f api/proto/user/v1/user.proto`,
	RunE: runProtoAll,
}

func init() {
	protoCmd.AddCommand(protoAllCmd)

	protoAllCmd.Flags().StringVarP(&protoAllDir, "proto-dir", "d", "api/proto", "Proto 文件目录")
}

func runProtoAll(cmd *cobra.Command, args []string) error {
	// 检测 Go module。
	module, err := detectGoModule()
	if err != nil {
		return fmt.Errorf("detect go module failed: %w", err)
	}

	// 查找所有 proto 文件。
	var protoFiles []string
	err = filepath.WalkDir(protoAllDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // 忽略错误，继续处理
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".proto") {
			protoFiles = append(protoFiles, path)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("walk proto directory failed: %w", err)
	}

	if len(protoFiles) == 0 {
		return fmt.Errorf("no proto files found in %s", protoAllDir)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Found %d proto files in %s\n\n", len(protoFiles), protoAllDir)

	// 逐个处理 proto 文件。
	for i, protoFile := range protoFiles {
		fmt.Fprintf(cmd.OutOrStdout(), "[%d/%d] Processing: %s\n", i+1, len(protoFiles), protoFile)

		// Step 1: 生成 pb.go。
		if err := genPB(cmd, protoFile); err != nil {
			fmt.Fprintf(cmd.OutOrStdout(), "  Error generating pb.go: %v\n", err)
			continue
		}

		// Step 2: 生成 handler/service/routes。
		if err := genServiceSkeleton(cmd, protoFile, module); err != nil {
			fmt.Fprintf(cmd.OutOrStdout(), "  Error generating service: %v\n", err)
			continue
		}

		// Step 3: 生成 client wrapper。
		if err := genClientWrapper(cmd, protoFile); err != nil {
			fmt.Fprintf(cmd.OutOrStdout(), "  Warning: client wrapper skipped: %v\n", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "  Done!\n\n")
	}

	fmt.Fprintf(cmd.OutOrStdout(), "All proto files processed!\n")
	return nil
}
