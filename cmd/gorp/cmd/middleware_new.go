package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// middlewareNewCmd 创建一个新的业务中间件骨架。
//
// 中文说明：
// - 目录为 `app/http/middleware/<name>/`。
// - 文件默认创建 `middleware.go`，并提供一个 `Middleware()` 返回 `gorp.HTTPMiddleware`；
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

		modulePath := detectCurrentModulePath(root)
		src := fmt.Sprintf(`package %s

import (
	"net/http"

	gorp "%s"
)

// Middleware 返回该中间件的 framework HTTPMiddleware。
//
// 使用示例：
// - c.Get(key) 获取中间件传递的数据
// - c.Set(key, value) 存储数据供后续 handler 使用
// - c.Abort(status) 中止请求链
// - c.AbortWithJSON(status, body) 中止并返回 JSON 响应
// - c.Next() 继续执行下一个 handler
func Middleware() gorp.HTTPMiddleware {
	return func(next gorp.HTTPHandler) gorp.HTTPHandler {
		return func(c gorp.HTTPContext) {
			// 示例：认证检查
			// token := c.GetHeader("Authorization")
			// if token == "" {
			//     c.AbortWithJSON(http.StatusUnauthorized, map[string]any{"error": "unauthorized"})
			//     return
			// }
			// c.Set("user_id", "xxx")

			if next != nil {
				next(c)
			}
		}
	}
}
`, name, modulePath)

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

func detectCurrentModulePath(root string) string {
	content, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		return "github.com/ngq/gorp"
	}
	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module "))
		}
	}
	return "github.com/ngq/gorp"
}
