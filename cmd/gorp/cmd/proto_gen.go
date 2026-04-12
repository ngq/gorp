package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

var protoGenCmd = &cobra.Command{
	Use:   "gen",
	Short: "Generate *.pb.go from app/grpc/proto/*.proto using protoc",
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := repoRootFromCWD()
		if err != nil {
			return err
		}

		protoRoot := filepath.Join(root, "app", "grpc", "proto")
		// 中文说明：
		// - 这里把 app/grpc/proto 视为统一 proto 根目录。
		// - 会递归扫描所有 .proto 文件，再一次性交给 protoc 生成。
		// - source_relative 选项用于让生成文件尽量贴着源 proto 存放，目录结构更直观。
		// collect all .proto files
		var files []string
		err = filepath.WalkDir(protoRoot, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			if !strings.HasSuffix(path, ".proto") {
				return nil
			}
			rel, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			files = append(files, filepath.ToSlash(rel))
			return nil
		})
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("proto root not found: %s", protoRoot)
			}
			return err
		}
		sort.Strings(files)
		if len(files) == 0 {
			return fmt.Errorf("no .proto files under %s", protoRoot)
		}

		// build protoc args (run from repo root)
		args2 := []string{
			"-I", "app/grpc/proto",
			"--go_out=app/grpc/proto",
			"--go_opt=paths=source_relative",
			"--go-grpc_out=app/grpc/proto",
			"--go-grpc_opt=paths=source_relative",
		}
		args2 = append(args2, files...)

		c := exec.Command("protoc", args2...)
		c.Dir = root
		c.Stdout = cmd.OutOrStdout()
		c.Stderr = cmd.ErrOrStderr()

		if err := c.Run(); err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), "proto 生成失败：请确认已安装 protoc 与插件")
			fmt.Fprintln(cmd.ErrOrStderr(), "- protoc: https://github.com/protocolbuffers/protobuf")
			fmt.Fprintln(cmd.ErrOrStderr(), "- protoc-gen-go: go install google.golang.org/protobuf/cmd/protoc-gen-go@latest")
			fmt.Fprintln(cmd.ErrOrStderr(), "- protoc-gen-go-grpc: go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest")
			return err
		}

		fmt.Fprintf(cmd.OutOrStdout(), "generated %d proto files\n", len(files))
		return nil
	},
}

func init() {
	protoCmd.AddCommand(protoGenCmd)
}
