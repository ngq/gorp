package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
	"github.com/ngq/gorp/framework/provider/proto"
	"github.com/spf13/cobra"
)

// protoCmd 为 proto 命令组，支持两种工作流。
//
// 工作流一：Proto -> pb.go + handler + service（主路径）
//
//	gorp proto all -f ./proto/user.proto
//
// 工作流二：Service -> Proto -> pb.go（迁移工具）
//
//	gorp proto from-service -s ./service.go -o ./proto/
//	gorp proto all -f ./proto/user.proto
var protoCmd = &cobra.Command{
	Use:     "proto",
	Short:   "Proto generator",
	GroupID: commandGroupCodegen,
	Long: `gorp proto is a Proto generator CLI.

Main workflow (Proto-first):
  gorp proto all -f ./proto/user.proto

Migration tool (Service-first, for existing Go code):
  gorp proto from-service -s ./service.go -o ./proto/`,
}

// outputDir is the output directory flag, shared by proto subcommands.
var outputDir string

func init() {
	rootCmd.AddCommand(protoCmd)
}

// createProtoGenerator creates a Proto generator.
func createProtoGenerator(includeHTTP bool) (integrationcontract.ProtoGenerator, error) {
	return proto.NewGenerator(&integrationcontract.ProtoGeneratorConfig{
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
