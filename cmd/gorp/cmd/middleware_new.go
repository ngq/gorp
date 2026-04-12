package cmd

import (
	"fmt"
	"path/filepath"
	"github.com/spf13/cobra"
)

// middlewareNewCmd 创建一个新的业务中间件骨架。
//
// 中文说明：
// - 目录为 `app/http/middleware/<name>/`。
// - 文件默认创建 `middleware.go`，并提供一个 `Middleware()` 返回 `gin.HandlerFunc`。
// - 它只生成最小可运行骨架，真正业务逻辑需要后续手写。
var middlewareNewCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new business middleware skeleton under app/http/middleware/",
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := repoRootFromCWD()
		if err != nil {
			return err
		}

		name, err := promptString(cmd.InOrStdin(), cmd.OutOrStdout(), "请输入中间件名称：", "", true)
		if err != nil {
			return err
		}
		if err := requireIdent(name, "middleware name"); err != nil {
			return err
		}

		targetDir := absJoin(root, "app", "http", "middleware", name)
		if dirExists(targetDir) {
			return fmt.Errorf("target folder already exists: %s", targetDir)
		}
		if err := ensureDir(targetDir); err != nil {
			return err
		}

		src := fmt.Sprintf(`package %s

import "github.com/gin-gonic/gin"

// Middleware 返回该中间件的 gin.HandlerFunc。
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}
`, name)

		if err := writeGoFile(filepath.Join(targetDir, "middleware.go"), src); err != nil {
			return err
		}

		fmt.Fprintf(cmd.OutOrStdout(), "创建中间件成功, 文件夹地址: %s\n", targetDir)
		return nil
	},
}

func init() {
	middlewareCmd.AddCommand(middlewareNewCmd)
}
