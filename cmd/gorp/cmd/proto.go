package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ngq/gorp/framework/contract"
	"github.com/ngq/gorp/framework/provider/proto"
	"github.com/spf13/cobra"
)

// protoCmd 为 proto 命令组，支持三种工作流。
//
// 工作流一：Service -> Proto -> pb.go（Go 开发者习惯）
//
//	gorp proto from-service -s ./service.go -o ./proto/
//	gorp proto gen -d ./proto -o ./pb/
//
// 工作流二：Route -> Proto -> pb.go（HTTP 开发者习惯）
//
//	gorp proto from-route -r ./routes.go -o ./proto/
//	gorp proto gen -d ./proto -o ./pb/
//
// 工作流三：Proto -> pb.go（传统 gRPC 习惯）
//
//	gorp proto gen  # 默认 api/proto
var protoCmd = &cobra.Command{
	Use:     "proto",
	Short:   "Proto generator - three workflows",
	GroupID: commandGroupCodegen,
	Long: `gorp proto is a Proto generator CLI with three workflows:

1. from-service: Generate Proto from Go Service interface
   gorp proto from-service -s ./service.go -o ./proto/

2. gen: Generate Go pb code from Proto files (calls protoc)
   gorp proto gen -d ./proto -o ./pb/

3. from-route: Generate Proto from Gin routes (experimental)
   gorp proto from-route -r ./routes.go -o ./proto/`,
}

// outputDir is the output directory flag, shared by proto subcommands.
var outputDir string

func init() {
	rootCmd.AddCommand(protoCmd)
}

// createProtoGenerator creates a Proto generator.
func createProtoGenerator(includeHTTP bool) (contract.ProtoGenerator, error) {
	return proto.NewGenerator(&contract.ProtoGeneratorConfig{
		Enabled:               true,
		Strategy:              "noop",
		IncludeHTTPAnnotation: includeHTTP,
	})
}

// ensureProtoDir ensures the directory exists.
func ensureProtoDir(path string) error {
	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		return os.MkdirAll(dir, 0755)
	}
	return nil
}

// printProtoSuccess prints success message.
func printProtoSuccess(action, output string) {
	fmt.Printf("success: %s\n", action)
	fmt.Printf("   output: %s\n", output)
}

// printProtoError prints error message.
func printProtoError(action string, err error) {
	fmt.Fprintf(os.Stderr, "error: %s failed: %v\n", action, err)
}
