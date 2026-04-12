package cmd

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"
)

// middlewareMigrateCmd 迁移 gin-contrib 中间件到本项目的业务中间件目录。
//
// 教程目标：
// - clone `https://github.com/gin-contrib/<repo>.git`
// - 删除 `.git/`、`go.mod`、`go.sum`
// - 替换源码中 `github.com/gin-gonic/gin` 为项目期望的 gin import
//
// 本仓库实际情况：
// - 我们直接依赖官方 gin（github.com/gin-gonic/gin），因此**替换等价于不替换**。
// - 这里仍保留替换流程，但默认 newImport 设为原始 gin import，用于“对齐教程步骤 + 便于未来改动”。
var middlewareMigrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate gin-contrib middleware into app/http/middleware/",
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := repoRootFromCWD()
		if err != nil {
			return err
		}

		repo, err := promptString(cmd.InOrStdin(), cmd.OutOrStdout(), "请输入中间件名称：", "", true)
		if err != nil {
			return err
		}
		if err := requireIdent(repo, "middleware repo"); err != nil {
			return err
		}

		mwRoot := filepath.Join(root, "app", "http", "middleware")
		if err := os.MkdirAll(mwRoot, 0o755); err != nil {
			return err
		}
		target := filepath.Join(mwRoot, repo)
		if dirExists(target) {
			return fmt.Errorf("target already exists: %s", target)
		}

		url := fmt.Sprintf("https://github.com/gin-contrib/%s.git", repo)
		fmt.Fprintf(cmd.OutOrStdout(), "cloning: %s\n", url)
		_, err = git.PlainClone(target, false, &git.CloneOptions{URL: url, Depth: 1})
		if err != nil {
			return err
		}

		// 删除项目级文件，让它变成“源码目录”而不是独立 Go module
		_ = os.RemoveAll(filepath.Join(target, ".git"))
		_ = os.Remove(filepath.Join(target, "go.mod"))
		_ = os.Remove(filepath.Join(target, "go.sum"))

		// import 替换：对齐教程步骤（在本仓库里替换前后是一样的）
		oldImport := "github.com/gin-gonic/gin"
		newImport := "github.com/gin-gonic/gin"
		err = filepath.WalkDir(target, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			if !strings.HasSuffix(p, ".go") {
				return nil
			}
			b, err := os.ReadFile(p)
			if err != nil {
				return err
			}
			if !bytes.Contains(b, []byte(oldImport)) {
				return nil
			}
			out := bytes.ReplaceAll(b, []byte(oldImport), []byte(newImport))
			return os.WriteFile(p, out, 0o644)
		})
		if err != nil {
			return err
		}

		fmt.Fprintf(cmd.OutOrStdout(), "migrated into: %s\n", target)
		return nil
	},
}

func init() {
	middlewareCmd.AddCommand(middlewareMigrateCmd)
}
