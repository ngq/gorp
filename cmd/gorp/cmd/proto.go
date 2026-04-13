package cmd

import "github.com/spf13/cobra"

// protoCmd 为 proto 命令组（对齐教程“新增特性：gRPC proto 工作流”）。
//
// 当前仅提供 `gorp proto gen`：用于调用 protoc 生成 *.pb.go / *_grpc.pb.go。
var protoCmd = &cobra.Command{
	Use:   "proto",
	Short: "Proto and gRPC code generation tools",
}

func init() {
	rootCmd.AddCommand(protoCmd)
}
